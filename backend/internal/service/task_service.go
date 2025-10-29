package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"pixiv-tailor/backend/internal/crawler"
	"pixiv-tailor/backend/internal/logger"
	"pixiv-tailor/backend/internal/repository"
	"pixiv-tailor/backend/internal/tagger"
	"pixiv-tailor/backend/pkg/models"
	"pixiv-tailor/backend/pkg/paths"

	"github.com/google/uuid"
)

// TaskService 任务服务接口
type TaskService interface {
	CreateTask(taskType, config string) (*repository.Task, error)
	GetTask(id string) (*repository.Task, error)
	UpdateTaskStatus(id, status string) error
	UpdateTaskProgress(id string, progress int) error
	UpdateTaskError(id, errorMsg string) error
	UpdateTaskCompletedAt(id string, completedAt time.Time) error
	UpdateTaskImagesFound(id string, count int) error
	UpdateTaskImagesDownloaded(id string, count int) error
	UpdateTaskResult(id string, result map[string]interface{}) error
	ListTasks(page, pageSize int32, status, taskType string) ([]*repository.Task, int, error)
	StartTask(id string) error
	StopTask(id string) error
	CancelTask(id string) error
	DeleteTask(id string) error
	CleanupTasks(cleanupType string) (int, error)
	SetLogCallback(callback func(taskID, level, message string))
	SetStatusCallback(callback func(taskID, status string, progress int))
	RegisterExecutor(taskType string, executor TaskExecutor)
}

// TaskExecutor 任务执行器接口（用于 handler 注册自己的执行逻辑）
type TaskExecutor interface {
	ExecuteGenerateTask(ctx context.Context, taskID string, config map[string]interface{})
}

// taskServiceImpl 任务服务实现
type taskServiceImpl struct {
	storage        repository.Storage
	logCallback    func(taskID, level, message string)
	statusCallback func(taskID, status string, progress int)
	// 任务上下文管理
	runningTasks map[string]context.CancelFunc
	taskMutex    sync.RWMutex
	// 任务队列管理 - 按类型管理，确保同类型任务串行执行
	waitingTasksByType map[string][]string // 按任务类型分组的等待队列
	queueMutex         sync.Mutex
	// 任务执行器映射 - 用于不同任务类型的自定义执行器
	executors map[string]TaskExecutor
}

// NewTaskService 创建任务服务实例
func NewTaskService(storage repository.Storage) TaskService {
	service := &taskServiceImpl{
		storage:            storage,
		runningTasks:       make(map[string]context.CancelFunc),
		waitingTasksByType: make(map[string][]string),
		executors:          make(map[string]TaskExecutor),
	}

	// 启动后台监控 goroutine，循环检查等待队列
	go service.monitorWaitingQueue()

	return service
}

// generateShortTaskID 生成8位hash任务ID
func generateShortTaskID() string {
	// 生成UUID
	uuidStr := uuid.New().String()

	// 计算MD5哈希
	hash := md5.Sum([]byte(uuidStr))

	// 取前8位作为任务ID
	return hex.EncodeToString(hash[:])[:8]
}

// CreateTask 创建任务
func (s *taskServiceImpl) CreateTask(taskType, config string) (*repository.Task, error) {
	logger.Infof("CreateTask 开始执行: taskType=%s, config=%s", taskType, config)

	// 验证配置JSON格式
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(config), &configMap); err != nil {
		return nil, fmt.Errorf("配置格式无效: %v", err)
	}

	// 验证任务类型
	validTypes := []string{"crawl", "generate", "train", "tag", "classify"}
	if !contains(validTypes, taskType) {
		return nil, fmt.Errorf("不支持的任务类型: %s", taskType)
	}

	task := &repository.Task{
		ID:        generateShortTaskID(),
		Type:      taskType,
		Status:    "pending",
		Config:    config,
		Progress:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.storage.CreateTask(task); err != nil {
		return nil, fmt.Errorf("创建任务失败: %v", err)
	}

	// 检查是否有同类型运行中的任务
	tasks, _, err := s.ListTasks(1, 100, "running", taskType)
	hasRunningTask := err == nil && len(tasks) > 0

	// 清理状态不一致的 running 任务（即数据库中有 running 状态但实际不在运行的任务）
	if hasRunningTask {
		s.queueMutex.Lock()
		actualRunningCount := len(s.runningTasks)
		s.queueMutex.Unlock()

		// 如果数据库中有 running 任务但实际没有任务在运行，说明状态不一致
		if actualRunningCount == 0 {
			logger.Warnf("检测到状态不一致：数据库中有 %d 个 running %s 任务，但实际没有任务在运行，强制清理", len(tasks), taskType)
			// 将所有 running 状态的任务标记为 failed
			for _, task := range tasks {
				logger.Warnf("强制清理不一致的 running 任务: %s", task.ID)
				s.UpdateTaskStatus(task.ID, "failed")
				s.UpdateTaskError(task.ID, "任务状态不一致，已自动清理")
			}
			hasRunningTask = false
		}
	}

	// 只有在没有同类型运行中的任务时才自动启动
	if !hasRunningTask {
		logger.Infof("开始自动启动任务: %s (类型: %s)", task.ID, taskType)
		if err := s.StartTask(task.ID); err != nil {
			// 如果启动失败，记录错误但不返回错误，因为任务已经创建成功
			logger.Warnf("任务启动失败 %s: %v", task.ID, err)
		} else {
			logger.Infof("任务启动成功: %s", task.ID)
		}
	} else {
		// 有同类型运行中的任务，设置为 pending 状态并加入等待队列
		if err := s.UpdateTaskStatus(task.ID, "pending"); err != nil {
			logger.Warnf("设置任务状态为pending失败 %s: %v", task.ID, err)
		} else {
			logger.Infof("任务 %s (类型: %s) 已设置为pending，加入等待队列", task.ID, taskType)

			// 将任务加入等待队列
			s.queueMutex.Lock()
			if s.waitingTasksByType[taskType] == nil {
				s.waitingTasksByType[taskType] = make([]string, 0)
			}
			s.waitingTasksByType[taskType] = append(s.waitingTasksByType[taskType], task.ID)
			logger.Infof("任务 %s 已加入 %s 类型等待队列（当前队列长度: %d）", task.ID, taskType, len(s.waitingTasksByType[taskType]))
			s.queueMutex.Unlock()
		}
	}

	return task, nil
}

// GetTask 获取任务
func (s *taskServiceImpl) GetTask(id string) (*repository.Task, error) {
	return s.storage.GetTask(id)
}

// UpdateTaskStatus 更新任务状态
func (s *taskServiceImpl) UpdateTaskStatus(id, status string) error {
	validStatuses := []string{"pending", "running", "completed", "failed", "cancelled"}
	if !contains(validStatuses, status) {
		return fmt.Errorf("不支持的任务状态: %s", status)
	}

	// 获取旧状态，用于日志
	task, err := s.GetTask(id)
	oldStatus := "unknown"
	if err == nil {
		oldStatus = task.Status
	}

	// 输出状态切换日志
	if oldStatus != status {
		logger.Infof("任务状态切换: %s [%s->%s]", id, oldStatus, status)
	}

	err = s.storage.UpdateTaskStatus(id, status)
	if err != nil {
		return err
	}

	// 发送WebSocket状态更新消息
	if s.statusCallback != nil {
		// 获取当前任务以获取进度信息
		task, err := s.GetTask(id)
		if err == nil {
			s.statusCallback(id, status, task.Progress)
		}
	}

	return nil
}

// UpdateTaskProgress 更新任务进度
func (s *taskServiceImpl) UpdateTaskProgress(id string, progress int) error {
	if progress < 0 || progress > 100 {
		return fmt.Errorf("进度值必须在0-100之间")
	}

	err := s.storage.UpdateTaskProgress(id, progress)
	if err != nil {
		return err
	}

	// 发送WebSocket进度更新消息
	// 使用异步方式确保获取到最新状态
	if s.statusCallback != nil {
		go func(taskID string, prog int) {
			// 等待一小段时间确保数据库更新完成
			time.Sleep(50 * time.Millisecond)
			// 获取最新的任务状态
			task, err := s.GetTask(taskID)
			if err == nil {
				s.statusCallback(taskID, task.Status, task.Progress)
			}
		}(id, progress)
	}

	return nil
}

// UpdateTaskProgressWithStage 更新任务进度（带阶段信息）
func (s *taskServiceImpl) UpdateTaskProgressWithStage(id string, progress int, stage string) error {
	if progress < 0 || progress > 100 {
		return fmt.Errorf("进度值必须在0-100之间")
	}

	err := s.storage.UpdateTaskProgress(id, progress)
	if err != nil {
		return err
	}

	// 发送阶段信息日志
	if s.logCallback != nil {
		s.logCallback(id, "info", fmt.Sprintf("阶段: %s - 进度: %d%%", stage, progress))
	}

	// 发送WebSocket进度更新消息
	// 使用异步方式确保获取到最新状态
	if s.statusCallback != nil {
		go func(taskID string, prog int) {
			// 等待一小段时间确保数据库更新完成
			time.Sleep(50 * time.Millisecond)
			// 获取最新的任务状态
			task, err := s.GetTask(taskID)
			if err == nil {
				s.statusCallback(taskID, task.Status, task.Progress)
			}
		}(id, progress)
	}

	return nil
}

// UpdateTaskError 更新任务错误信息
func (s *taskServiceImpl) UpdateTaskError(id, errorMsg string) error {
	return s.storage.UpdateTaskError(id, errorMsg)
}

// ListTasks 列出任务
func (s *taskServiceImpl) ListTasks(page, pageSize int32, status, taskType string) ([]*repository.Task, int, error) {
	if pageSize <= 0 {
		pageSize = 20 // 默认限制
	}
	if pageSize > 100 {
		pageSize = 100 // 最大限制
	}

	offset := int((page - 1) * pageSize)
	tasks, err := s.storage.ListTasks(status, taskType, int(pageSize), offset)
	if err != nil {
		return nil, 0, err
	}

	// 检查并修复僵尸任务（数据库中是running状态但实际没有运行的任务）
	s.checkAndFixZombieTasks(tasks)

	// 获取总数
	total, err := s.storage.CountTasks(status, taskType)
	if err != nil {
		return tasks, 0, err
	}

	return tasks, total, nil
}

// StartTask 启动任务
func (s *taskServiceImpl) StartTask(id string) error {
	logger.Infof("StartTask 开始执行: %s", id)

	task, err := s.GetTask(id)
	if err != nil {
		return fmt.Errorf("获取任务失败: %v", err)
	}

	// tag、crawl、generate 类型任务都由 task_service 统一处理

	// 允许重新启动失败或取消的任务
	if task.Status == "running" {
		return fmt.Errorf("任务正在运行中，无法重复启动")
	}

	// 如果任务已完成，需要重置状态
	if task.Status == "completed" {
		logger.Infof("任务 %s 已完成，重置状态为pending", id)
		if err := s.UpdateTaskStatus(id, "pending"); err != nil {
			return fmt.Errorf("重置任务状态失败: %v", err)
		}
	}

	// 如果任务失败或取消，清理错误信息并重新启动
	if task.Status == "failed" || task.Status == "cancelled" {
		logger.Infof("重新启动任务 %s (状态: %s)", id, task.Status)
		// 清理错误信息
		if err := s.UpdateTaskError(id, ""); err != nil {
			logger.Warnf("清理任务错误信息失败: %v", err)
		}
		// 重置进度
		if err := s.UpdateTaskProgress(id, 0); err != nil {
			logger.Warnf("重置任务进度失败: %v", err)
		}
		// 重置图片数量
		if err := s.UpdateTaskImagesFound(id, 0); err != nil {
			logger.Warnf("重置图片数量失败: %v", err)
		}
		if err := s.UpdateTaskImagesDownloaded(id, 0); err != nil {
			logger.Warnf("重置下载图片数量失败: %v", err)
		}
	}

	// 检查同类型任务是否正在运行
	hasRunningTaskOfSameType := s.hasRunningTaskOfType(task.Type, id)

	// 如果有同类型任务正在运行，将新任务加入对应类型的等待队列
	if hasRunningTaskOfSameType {
		logger.Infof("StartTask: 检测到同类型的运行中任务，任务 %s (类型: %s) 加入等待队列", id, task.Type)
		if err := s.UpdateTaskStatus(id, "pending"); err != nil {
			return fmt.Errorf("更新任务状态失败: %v", err)
		}
		s.queueMutex.Lock()
		if s.waitingTasksByType[task.Type] == nil {
			s.waitingTasksByType[task.Type] = make([]string, 0)
		}
		s.waitingTasksByType[task.Type] = append(s.waitingTasksByType[task.Type], id)
		logger.Infof("StartTask: 任务 %s 已加入 %s 类型等待队列（当前队列长度: %d）", id, task.Type, len(s.waitingTasksByType[task.Type]))
		s.queueMutex.Unlock()
		s.sendLog(id, "info", "任务等待中...")
		return nil
	}

	// 没有运行中的任务，立即启动
	return s.startTaskExecution(id)
}

// startTaskExecution 执行任务
func (s *taskServiceImpl) startTaskExecution(id string) error {
	logger.Infof("startTaskExecution: 开始执行任务 %s", id)

	// 更新状态为running
	if err := s.UpdateTaskStatus(id, "running"); err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	// 创建任务上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 保存取消函数
	s.taskMutex.Lock()
	s.runningTasks[id] = cancel
	s.taskMutex.Unlock()

	// 获取任务
	task, err := s.GetTask(id)
	if err != nil {
		return fmt.Errorf("获取任务失败: %v", err)
	}

	// 启动实际的任务执行逻辑
	go s.executeTaskWithContext(ctx, task)

	// 启动任务完成后继续处理等待队列
	go s.processNextTask()

	return nil
}

// hasRunningTaskOfType 检查是否有指定类型的任务正在运行
func (s *taskServiceImpl) hasRunningTaskOfType(taskType, excludeTaskID string) bool {
	// 先检查 task service 管理的运行中任务
	s.taskMutex.RLock()
	for runningTaskID := range s.runningTasks {
		if runningTaskID == excludeTaskID {
			continue
		}
		// 获取运行中的任务信息
		runningTask, err := s.GetTask(runningTaskID)
		if err == nil && runningTask.Type == taskType {
			s.taskMutex.RUnlock()
			return true
		}
	}
	s.taskMutex.RUnlock()

	// 然后检查数据库中的运行中任务
	tasks, _, err := s.ListTasks(1, 100, "running", taskType)
	if err != nil {
		logger.Warnf("获取运行中任务列表失败: %v", err)
		return false
	}

	// 排除当前任务本身
	for _, task := range tasks {
		if task.ID != excludeTaskID {
			return true
		}
	}

	return false
}

// processNextTask 处理下一个任务（按任务类型）
func (s *taskServiceImpl) processNextTask() {
	// 多等待几秒确保当前任务完全结束
	time.Sleep(3 * time.Second)

	s.queueMutex.Lock()
	defer s.queueMutex.Unlock()

	// 遍历所有任务类型的等待队列，启动可执行的任务
	for taskType, waitingTasks := range s.waitingTasksByType {
		// 检查该类型是否有正在运行的任务
		hasRunning := s.hasRunningTaskOfTypeInLock(taskType)

		if !hasRunning && len(waitingTasks) > 0 {
			// 该类型没有运行中的任务，且有等待中的任务，启动第一个
			nextTaskID := waitingTasks[0]
			s.waitingTasksByType[taskType] = waitingTasks[1:]

			logger.Infof("processNextTask: 启动 %s 类型任务 %s (该类型还有 %d 个等待任务)", taskType, nextTaskID, len(s.waitingTasksByType[taskType]))

			// 解锁后启动任务
			s.queueMutex.Unlock()
			if err := s.startTaskExecution(nextTaskID); err != nil {
				logger.Errorf("启动等待任务失败: %v", err)
			}
			s.queueMutex.Lock()
			return
		}
	}

	logger.Infof("processNextTask: 没有可执行的任务 (当前等待队列: %+v)", s.waitingTasksByType)
}

// monitorWaitingQueue 后台监控 goroutine，定期检查等待队列
func (s *taskServiceImpl) monitorWaitingQueue() {
	// 每 5 秒检查一次等待队列
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.checkAndProcessWaitingQueue()
	}
}

// checkAndProcessWaitingQueue 检查并处理等待队列
func (s *taskServiceImpl) checkAndProcessWaitingQueue() {
	s.queueMutex.Lock()
	defer s.queueMutex.Unlock()

	// 如果没有等待任务，直接返回
	hasWaitingTask := false
	for _, tasks := range s.waitingTasksByType {
		if len(tasks) > 0 {
			hasWaitingTask = true
			break
		}
	}

	if !hasWaitingTask {
		return
	}

	// logger.Infof("checkAndProcessWaitingQueue: 发现等待任务，开始处理 (当前等待队列: %+v)", s.waitingTasksByType)

	// 遍历所有任务类型的等待队列，启动可执行的任务
	for taskType, waitingTasks := range s.waitingTasksByType {
		// 检查该类型是否有正在运行的任务
		hasRunning := s.hasRunningTaskOfTypeInLock(taskType)

		if !hasRunning && len(waitingTasks) > 0 {
			// 该类型没有运行中的任务，且有等待中的任务，启动第一个
			nextTaskID := waitingTasks[0]
			s.waitingTasksByType[taskType] = waitingTasks[1:]

			logger.Infof("checkAndProcessWaitingQueue: 启动 %s 类型任务 %s (该类型还有 %d 个等待任务)", taskType, nextTaskID, len(s.waitingTasksByType[taskType]))

			// 解锁后启动任务
			s.queueMutex.Unlock()
			if err := s.startTaskExecution(nextTaskID); err != nil {
				logger.Errorf("启动等待任务失败: %v", err)
			}
			s.queueMutex.Lock()
			return // 一次只启动一个任务
		}
	}

	// logger.Infof("checkAndProcessWaitingQueue: 所有类型的任务都在运行中，等待下次检查")
}

// hasRunningTaskOfTypeInLock 在持有 queueMutex 的情况下检查指定类型是否有运行中任务
// 注意：调用此方法时必须持有 queueMutex
func (s *taskServiceImpl) hasRunningTaskOfTypeInLock(taskType string) bool {
	// 检查 task service 管理的运行中任务
	s.taskMutex.RLock()
	for runningTaskID := range s.runningTasks {
		// 获取运行中的任务信息
		runningTask, err := s.GetTask(runningTaskID)
		if err == nil && runningTask.Type == taskType {
			s.taskMutex.RUnlock()
			return true
		}
	}
	s.taskMutex.RUnlock()

	// 检查数据库中的运行中任务
	tasks, _, err := s.ListTasks(1, 100, "running", taskType)
	if err != nil {
		return false
	}

	return len(tasks) > 0
}

// StopTask 停止任务
func (s *taskServiceImpl) StopTask(id string) error {
	task, err := s.GetTask(id)
	if err != nil {
		return fmt.Errorf("获取任务失败: %v", err)
	}

	if task.Status != "running" {
		return fmt.Errorf("任务状态不是running，无法停止")
	}

	// 取消任务上下文
	s.taskMutex.Lock()
	if cancel, exists := s.runningTasks[id]; exists {
		cancel() // 取消任务执行
		delete(s.runningTasks, id)
	}
	s.taskMutex.Unlock()

	// 更新状态为cancelled
	if err := s.UpdateTaskStatus(id, "cancelled"); err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	s.sendLog(id, "info", "任务已被用户取消")
	return nil
}

// CancelTask 取消任务
func (s *taskServiceImpl) CancelTask(id string) error {
	task, err := s.GetTask(id)
	if err != nil {
		return fmt.Errorf("获取任务失败: %v", err)
	}

	if task.Status == "completed" {
		return fmt.Errorf("任务已完成，无法取消")
	}

	// 更新状态为cancelled
	if err := s.UpdateTaskStatus(id, "cancelled"); err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	// TODO: 这里应该停止实际的任务执行逻辑

	return nil
}

// DeleteTask 删除任务
func (s *taskServiceImpl) DeleteTask(id string) error {
	// 检查任务是否存在
	task, err := s.GetTask(id)
	if err != nil {
		return fmt.Errorf("获取任务失败: %v", err)
	}

	// 如果任务正在运行，先停止它
	if task.Status == "running" {
		if err := s.StopTask(id); err != nil {
			logger.Infof("停止运行中的任务失败: %v", err)
			// 即使停止失败，也继续删除
		}
	}

	// 删除任务相关的图片文件（在删除数据库记录之前获取任务信息）
	pathManager := paths.GetPathManager()
	if pathManager != nil {
		var taskDir string
		if task != nil {
			// 使用新格式：[时间]_任务类型_哈希值
			taskDir = pathManager.GetTaskImagesDir(id, task.Type, task.CreatedAt)
		} else {
			// 回退到旧格式：task_{id}
			taskDir = filepath.Join(pathManager.GetImagesDir(), fmt.Sprintf("task_%s", id))
		}

		// 尝试删除新格式
		if err := os.RemoveAll(taskDir); err != nil {
			logger.Infof("删除任务图片目录失败: %v", err)
		}

		// 也尝试删除旧格式（向后兼容）
		oldTaskDir := filepath.Join(pathManager.GetImagesDir(), fmt.Sprintf("task_%s", id))
		if oldTaskDir != taskDir {
			if err := os.RemoveAll(oldTaskDir); err != nil {
				logger.Infof("删除旧格式任务图片目录失败: %v", err)
			}
		}
	}

	// 从数据库中删除任务
	if err := s.storage.DeleteTask(id); err != nil {
		return fmt.Errorf("删除任务失败: %v", err)
	}

	logger.Infof("任务 %s 已删除", id)
	return nil
}

// executeTask 执行任务
func (s *taskServiceImpl) executeTaskWithContext(ctx context.Context, task *repository.Task) {
	// 检查上下文是否已被取消
	select {
	case <-ctx.Done():
		s.UpdateTaskStatus(task.ID, "cancelled")
		s.sendLog(task.ID, "info", "任务已被取消")
		return
	default:
	}

	// 执行任务
	s.executeTask(ctx, task)
}

func (s *taskServiceImpl) executeTask(ctx context.Context, task *repository.Task) {
	logger.Infof("executeTask 开始执行: %s", task.ID)

	// 解析任务配置
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(task.Config), &config); err != nil {
		s.UpdateTaskError(task.ID, fmt.Sprintf("配置解析失败: %v", err))
		s.UpdateTaskStatus(task.ID, "failed")
		return
	}

	// 根据任务类型执行不同的逻辑
	switch task.Type {
	case "tag":
		// 标签任务使用专门的执行方法
		s.executeTagTask(ctx, task, config)
	case "crawl":
		// 爬虫任务使用专门的执行方法
		s.executeCrawlTask(ctx, task, config)
	case "generate":
		// AI生成任务使用专门的执行方法
		s.executeGenerateTask(ctx, task, config)
	case "train", "classify":
		// 这些类型的任务将在这里处理
		s.UpdateTaskError(task.ID, fmt.Sprintf("任务类型 %s 暂未实现", task.Type))
		s.UpdateTaskStatus(task.ID, "failed")
	default:
		logger.Warnf("executeTask 收到未处理的任务类型 %s (任务ID: %s)，这可能不应该发生", task.Type, task.ID)
	}
}

// CleanupTasks 清理任务
func (s *taskServiceImpl) CleanupTasks(cleanupType string) (int, error) {
	logger.Infof("开始清理任务: cleanupType=%s", cleanupType)

	// 验证清理类型
	validTypes := []string{"completed", "failed", "all"}
	if !contains(validTypes, cleanupType) {
		return 0, fmt.Errorf("不支持的清理类型: %s", cleanupType)
	}

	var cleanedCount int
	var err error

	switch cleanupType {
	case "completed":
		// 清理已完成的任务
		cleanedCount, err = s.storage.CleanupTasksByStatus("completed")
	case "failed":
		// 清理失败的任务
		cleanedCount, err = s.storage.CleanupTasksByStatus("failed")
	case "all":
		// 清理所有任务
		cleanedCount, err = s.storage.CleanupAllTasks()
	}

	if err != nil {
		logger.Errorf("清理任务失败: %v", err)
		return 0, fmt.Errorf("清理任务失败: %v", err)
	}

	logger.Infof("清理任务成功, 清理了 %d 个任务", cleanedCount)
	return cleanedCount, nil
}

// SetLogCallback 设置日志回调函数
func (s *taskServiceImpl) SetLogCallback(callback func(taskID, level, message string)) {
	s.logCallback = callback
}

func (s *taskServiceImpl) SetStatusCallback(callback func(taskID, status string, progress int)) {
	s.statusCallback = callback
}

// sendLog 发送日志消息
func (s *taskServiceImpl) sendLog(taskID, level, message string) {
	if s.logCallback != nil {
		s.logCallback(taskID, level, message)
	}
}

// executeGenerateTask 已删除 - AI生成任务已迁移到 ai_handler.go

// UpdateTaskCompletedAt 更新任务完成时间
func (s *taskServiceImpl) UpdateTaskCompletedAt(id string, completedAt time.Time) error {
	// 这里需要添加一个方法来更新完成时间
	// 暂时返回nil，后续可以完善
	return nil
}

// UpdateTaskImagesFound 更新获取到的图片数量
func (s *taskServiceImpl) UpdateTaskImagesFound(id string, count int) error {
	err := s.storage.UpdateTaskImagesFound(id, count)
	if err != nil {
		return err
	}

	// 发送WebSocket更新消息
	if s.statusCallback != nil {
		// 获取当前任务以获取状态和进度信息
		task, err := s.GetTask(id)
		if err == nil {
			s.statusCallback(id, task.Status, task.Progress)
		}
	}

	return nil
}

// UpdateTaskImagesDownloaded 更新下载的图片数量
func (s *taskServiceImpl) UpdateTaskImagesDownloaded(id string, count int) error {
	err := s.storage.UpdateTaskImagesDownloaded(id, count)
	if err != nil {
		return err
	}

	// 发送WebSocket更新消息（异步获取最新状态）
	if s.statusCallback != nil {
		go func(taskID string, downloadedCount int) {
			// 等待一小段时间确保数据库更新完成
			time.Sleep(50 * time.Millisecond)
			// 获取最新的任务状态
			task, err := s.GetTask(taskID)
			if err == nil {
				// 发送状态更新，这样会触发完整的任务更新（包括下载计数）
				s.statusCallback(taskID, task.Status, task.Progress)
			}
		}(id, count)
	}

	return nil
}

// UpdateTaskResult 更新任务结果
func (s *taskServiceImpl) UpdateTaskResult(id string, result map[string]interface{}) error {
	// 将结果转换为JSON字符串
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("序列化任务结果失败: %v", err)
	}

	// 更新数据库
	err = s.storage.UpdateTaskResult(id, string(resultJSON))
	if err != nil {
		return err
	}

	// 发送WebSocket更新消息
	if s.statusCallback != nil {
		// 获取当前任务以获取状态和进度信息
		task, err := s.GetTask(id)
		if err == nil {
			s.statusCallback(id, task.Status, task.Progress)
		}
	}

	return nil
}

// executeTagTask 执行标签任务
func (s *taskServiceImpl) executeTagTask(ctx context.Context, task *repository.Task, config map[string]interface{}) {
	logger.Infof("开始执行标签任务: %s", task.ID)
	s.sendLog(task.ID, "info", "开始执行标签任务")

	// 检查上下文是否已被取消
	select {
	case <-ctx.Done():
		s.UpdateTaskStatus(task.ID, "cancelled")
		s.sendLog(task.ID, "info", "任务已被取消")
		return
	default:
	}

	// 更新进度到10% - 初始化阶段
	s.UpdateTaskProgressWithStage(task.ID, 10, "初始化")
	s.sendLog(task.ID, "info", "标签任务初始化完成 (10%)")

	// 创建 tagger 实例
	tagger, err := tagger.NewWD14Tagger()
	if err != nil {
		s.sendLog(task.ID, "error", fmt.Sprintf("创建标签器失败: %v", err))
		s.UpdateTaskError(task.ID, fmt.Sprintf("创建标签器失败: %v", err))
		s.UpdateTaskStatus(task.ID, "failed")
		// 从运行任务列表中移除
		s.taskMutex.Lock()
		delete(s.runningTasks, task.ID)
		s.taskMutex.Unlock()
		return
	}

	// 解析配置参数（支持单个路径或路径数组）
	inputDir, ok := config["input_dir"]
	if !ok || inputDir == nil {
		s.sendLog(task.ID, "error", "输入目录不能为空")
		s.UpdateTaskError(task.ID, "输入目录不能为空")
		s.UpdateTaskStatus(task.ID, "failed")
		// 从运行任务列表中移除
		s.taskMutex.Lock()
		delete(s.runningTasks, task.ID)
		s.taskMutex.Unlock()
		return
	}

	// 转换为字符串数组以获取输入目录列表
	var inputDirs []string
	switch v := inputDir.(type) {
	case string:
		if v != "" {
			inputDirs = []string{v}
		}
	case []interface{}:
		for _, dir := range v {
			if dirStr, ok := dir.(string); ok && dirStr != "" {
				inputDirs = append(inputDirs, dirStr)
			}
		}
	case []string:
		for _, dir := range v {
			if dir != "" {
				inputDirs = append(inputDirs, dir)
			}
		}
	}

	if len(inputDirs) == 0 {
		s.sendLog(task.ID, "error", "输入目录不能为空")
		s.UpdateTaskError(task.ID, "输入目录不能为空")
		s.UpdateTaskStatus(task.ID, "failed")
		// 从运行任务列表中移除
		s.taskMutex.Lock()
		delete(s.runningTasks, task.ID)
		s.taskMutex.Unlock()
		return
	}

	outputDir, ok := config["output_dir"].(string)
	if !ok || outputDir == "" {
		outputDir = "tags/" // 默认输出目录
	}

	// 设置默认值
	tagRequest := &models.TagRequest{
		InputDir:   inputDirs, // 使用数组
		OutputDir:  outputDir,
		Analyzer:   "wd14tagger",
		TagOrder:   "score",
		SaveType:   "txt",
		Limit:      100,
		SkipTags:   []string{},
		ExtendTags: []string{},
	}

	// 读取配置参数
	if limit, ok := config["limit"].(float64); ok {
		tagRequest.Limit = int(limit)
	}

	if saveType, ok := config["save_type"].(string); ok {
		tagRequest.SaveType = saveType
	}

	if skipTags, ok := config["skip_tags"].([]interface{}); ok {
		for _, tag := range skipTags {
			if tagStr, ok := tag.(string); ok {
				tagRequest.SkipTags = append(tagRequest.SkipTags, tagStr)
			}
		}
	}

	if extendTags, ok := config["extend_tags"].([]interface{}); ok {
		for _, tag := range extendTags {
			if tagStr, ok := tag.(string); ok {
				tagRequest.ExtendTags = append(tagRequest.ExtendTags, tagStr)
			}
		}
	}

	s.sendLog(task.ID, "info", fmt.Sprintf("配置完成: 输入目录数量=%d, 输出目录=%s, 限制=%d", len(inputDirs), outputDir, tagRequest.Limit))

	// 更新进度到20% - 配置完成
	s.UpdateTaskProgressWithStage(task.ID, 20, "配置完成")

	// 检查上下文是否已被取消
	select {
	case <-ctx.Done():
		s.UpdateTaskStatus(task.ID, "cancelled")
		s.sendLog(task.ID, "info", "任务已被取消")
		// 从运行任务列表中移除
		s.taskMutex.Lock()
		delete(s.runningTasks, task.ID)
		s.taskMutex.Unlock()
		return
	default:
	}

	// 执行标签生成
	s.UpdateTaskProgressWithStage(task.ID, 30, "开始生成标签")
	s.sendLog(task.ID, "info", "开始调用 WD14 Tagger 生成标签...")

	// 创建进度回调函数，使用内部变量缓存上次进度，减少不必要的更新
	lastProgress := -1
	progressCallback := func(current int, total int) {
		// 计算进度：30%-95% (生成标签占65%)
		progress := 30 + int(float64(current)/float64(total)*65)

		// 只在进度变化超过5%或完成时更新，减少WebSocket消息频率
		if progress != lastProgress && (progress-lastProgress >= 5 || progress == 100 || current == total) {
			s.UpdateTaskProgressWithStage(task.ID, progress, fmt.Sprintf("生成标签 (%d/%d)", current, total))
			lastProgress = progress
		}
	}

	// 创建日志回调函数
	logCallback := func(level string, message string) {
		s.sendLog(task.ID, level, message)
	}

	if err := tagger.GenerateTagsWithCallback(tagRequest, progressCallback, logCallback); err != nil {
		s.sendLog(task.ID, "error", fmt.Sprintf("生成标签失败: %v", err))
		s.UpdateTaskError(task.ID, fmt.Sprintf("生成标签失败: %v", err))
		s.UpdateTaskStatus(task.ID, "failed")
		// 从运行任务列表中移除
		s.taskMutex.Lock()
		delete(s.runningTasks, task.ID)
		s.taskMutex.Unlock()
		return
	}

	// 更新进度到100% - 任务完成
	s.UpdateTaskProgressWithStage(task.ID, 100, "任务完成")
	s.sendLog(task.ID, "info", "标签生成完成 (100%)")

	// 任务完成
	logger.Infof("executeTagTask: 任务 %s 准备更新为 completed", task.ID)
	s.UpdateTaskStatus(task.ID, "completed")
	s.sendLog(task.ID, "info", "标签任务完成！")
	logger.Infof("标签任务完成: %s", task.ID)

	// 从运行任务列表中移除
	s.taskMutex.Lock()
	delete(s.runningTasks, task.ID)
	s.taskMutex.Unlock()
	logger.Infof("任务 %s 已完成，从运行列表移除", task.ID)

	// 处理下一个等待的任务
	logger.Infof("executeTagTask: 任务 %s 完成后，准备处理下一个任务", task.ID)
	// 延迟一点再处理下一个任务，确保状态完全更新
	go func() {
		time.Sleep(1 * time.Second)
		s.processNextTask()
	}()
}

// executeCrawlTask 执行爬虫任务
func (s *taskServiceImpl) executeCrawlTask(ctx context.Context, task *repository.Task, config map[string]interface{}) {
	logger.Infof("开始执行爬虫任务: %s", task.ID)
	s.sendLog(task.ID, "info", "开始执行爬虫任务")

	// 检查上下文是否已被取消
	select {
	case <-ctx.Done():
		s.UpdateTaskStatus(task.ID, "cancelled")
		s.sendLog(task.ID, "info", "任务已被取消")
		// 从运行任务列表中移除
		s.taskMutex.Lock()
		delete(s.runningTasks, task.ID)
		s.taskMutex.Unlock()
		return
	default:
	}

	// 更新进度到5%
	s.UpdateTaskProgress(task.ID, 5)
	s.sendLog(task.ID, "info", "爬虫任务初始化完成 (5%)")

	// 创建任务特定的爬虫实例
	crawlerInstance, err := crawler.NewCrawlerForTask(task.ID, config)
	if err != nil {
		s.sendLog(task.ID, "error", fmt.Sprintf("创建爬虫失败: %v", err))
		s.UpdateTaskError(task.ID, fmt.Sprintf("创建爬虫失败: %v", err))
		s.UpdateTaskStatus(task.ID, "failed")
		// 从运行任务列表中移除
		s.taskMutex.Lock()
		delete(s.runningTasks, task.ID)
		s.taskMutex.Unlock()
		return
	}

	// 设置任务信息（用于生成任务文件夹名：[时间]_任务类型_哈希值）
	crawlerInstance.SetTaskInfo(task.Type, task.CreatedAt)

	// 配置代理设置
	if proxyEnabled, exists := config["proxy_enabled"].(bool); exists && proxyEnabled {
		if proxyURL, exists := config["proxy_url"].(string); exists && proxyURL != "" {
			crawlerInstance.SetProxy(true, proxyURL)
		}
	} else {
		crawlerInstance.SetProxy(false, "")
	}

	// 配置Cookie设置
	if cookie, exists := config["cookie"].(string); exists && cookie != "" {
		if cookie == "default" {
			// 使用默认Cookie
		} else {
			crawlerInstance.SetCookie(cookie)
		}
	}

	// 更新进度到20%
	s.UpdateTaskProgress(task.ID, 20)
	s.sendLog(task.ID, "info", "爬虫配置完成 (20%)")

	// 检查上下文是否已被取消
	select {
	case <-ctx.Done():
		s.UpdateTaskStatus(task.ID, "cancelled")
		s.sendLog(task.ID, "info", "任务已被取消")
		s.taskMutex.Lock()
		delete(s.runningTasks, task.ID)
		s.taskMutex.Unlock()
		return
	default:
	}

	// 根据配置执行爬取
	crawlType := config["type"].(string)
	var results []*models.PixivImage
	var crawlErr error

	switch crawlType {
	case "tag":
		query := config["query"].(string)
		order := config["order"].(string)
		mode := config["mode"].(string)
		limit := int(config["limit"].(float64))
		results, crawlErr = crawlerInstance.CrawlByTag(query, order, mode, limit)

	case "user":
		userID := int(config["user_id"].(float64))
		limit := int(config["limit"].(float64))
		results, crawlErr = crawlerInstance.CrawlByUser(userID, limit)

	case "illust":
		illustID := int(config["illust_id"].(float64))
		illustResults, err := crawlerInstance.CrawlByIllust(illustID)
		crawlErr = err
		if illustResults != nil {
			results = illustResults
		}

	default:
		s.sendLog(task.ID, "error", fmt.Sprintf("不支持的类型: %s", crawlType))
		s.UpdateTaskError(task.ID, fmt.Sprintf("不支持的类型: %s", crawlType))
		s.UpdateTaskStatus(task.ID, "failed")
		s.taskMutex.Lock()
		delete(s.runningTasks, task.ID)
		s.taskMutex.Unlock()
		return
	}

	if crawlErr != nil {
		s.sendLog(task.ID, "error", fmt.Sprintf("爬取失败: %v", crawlErr))
		s.UpdateTaskError(task.ID, fmt.Sprintf("爬取失败: %v", crawlErr))
		s.UpdateTaskStatus(task.ID, "failed")
		s.taskMutex.Lock()
		delete(s.runningTasks, task.ID)
		s.taskMutex.Unlock()
		return
	}

	// 保存原始获取到的图片数量
	originalImagesCount := len(results)

	// 应用图片数量限制
	maxImages := 0
	if maxImagesVal, exists := config["max_images"]; exists {
		maxImages = int(maxImagesVal.(float64))
	}

	// 如果有设置限制且获取到的图片数量超过限制，只保留前maxImages张
	if maxImages > 0 && len(results) > maxImages {
		results = results[:maxImages]
		s.sendLog(task.ID, "info", fmt.Sprintf("应用图片数量限制：已获取 %d 张，限制下载 %d 张", originalImagesCount, len(results)))
	}

	// 记录原始获取数量（不包括限制）
	s.UpdateTaskImagesFound(task.ID, originalImagesCount)
	s.sendLog(task.ID, "info", fmt.Sprintf("爬取成功，获得 %d 张图片", originalImagesCount))

	// 检查上下文是否已被取消
	select {
	case <-ctx.Done():
		s.UpdateTaskStatus(task.ID, "cancelled")
		s.sendLog(task.ID, "info", "任务已被取消")
		s.taskMutex.Lock()
		delete(s.runningTasks, task.ID)
		s.taskMutex.Unlock()
		return
	default:
	}

	// 开始下载图片
	if len(results) > 0 {
		s.sendLog(task.ID, "info", fmt.Sprintf("开始下载 %d 张图片", len(results)))
		downloadedCount := 0
		for i, image := range results {
			// 检查是否被取消
			select {
			case <-ctx.Done():
				s.UpdateTaskStatus(task.ID, "cancelled")
				s.sendLog(task.ID, "info", "任务已被取消")
				s.taskMutex.Lock()
				delete(s.runningTasks, task.ID)
				s.taskMutex.Unlock()
				return
			default:
			}

			// 更新进度: 50% + 45% * (i / total)
			progress := 50 + int(45*float64(i)/float64(len(results)))
			s.UpdateTaskProgress(task.ID, progress)

			// 每10张图片或每张图片发送下载进度日志
			if (i+1)%10 == 0 || i == 0 || i == len(results)-1 {
				s.sendLog(task.ID, "info", fmt.Sprintf("正在下载图片 %d/%d", i+1, len(results)))
			}

			// 生成文件名
			// 从URL中提取文件扩展名，如果无法提取则使用.jpg
			fileExt := ".jpg"
			if ext := filepath.Ext(image.URL); ext != "" {
				fileExt = ext
			}
			// 使用图片ID和索引生成唯一文件名
			filename := fmt.Sprintf("artworks_%d_p%02d%s", image.ID, i+1, fileExt)

			// 下载图片
			if err := crawlerInstance.DownloadImage(image.URL, filename, task.ID, func(url, filename string, downloaded, total int64, percent float64) {
				// 进度回调（可选）
			}); err != nil {
				logger.Warnf("下载图片失败 %s: %v", image.URL, err)
				s.sendLog(task.ID, "warning", fmt.Sprintf("下载失败: %s", image.Title))
			} else {
				downloadedCount++
				// 实时更新下载计数
				s.UpdateTaskImagesDownloaded(task.ID, downloadedCount)
				// 每10张或最后一张图片发送成功日志
				if downloadedCount%10 == 0 || downloadedCount == len(results) {
					s.sendLog(task.ID, "info", fmt.Sprintf("已下载 %d/%d 张图片", downloadedCount, len(results)))
				}
			}
		}
		s.sendLog(task.ID, "info", fmt.Sprintf("图片下载完成，成功下载 %d/%d 张", downloadedCount, len(results)))
	}

	// 更新进度到95%
	s.UpdateTaskProgress(task.ID, 95)

	// 检查上下文是否已被取消
	select {
	case <-ctx.Done():
		s.UpdateTaskStatus(task.ID, "cancelled")
		s.sendLog(task.ID, "info", "任务已被取消")
		s.taskMutex.Lock()
		delete(s.runningTasks, task.ID)
		s.taskMutex.Unlock()
		return
	default:
	}

	// 任务完成
	s.UpdateTaskProgress(task.ID, 100)
	s.sendLog(task.ID, "info", "爬虫任务完成！")
	logger.Infof("爬虫任务完成: %s", task.ID)

	// 保存任务结果信息（包括预期下载数量）
	result := map[string]interface{}{
		"expected_images": len(results), // 预期下载的图片数量
		"has_limit":       maxImages > 0,
	}
	s.UpdateTaskResult(task.ID, result)

	// 更新状态为completed
	s.UpdateTaskStatus(task.ID, "completed")

	// 从运行任务列表中移除
	s.taskMutex.Lock()
	delete(s.runningTasks, task.ID)
	s.taskMutex.Unlock()
	logger.Infof("任务 %s 已完成，从运行列表移除", task.ID)

	// 处理下一个等待的任务
	logger.Infof("executeCrawlTask: 任务 %s 完成后，准备处理下一个任务", task.ID)
	// 延迟一点再处理下一个任务，确保状态完全更新
	go func() {
		time.Sleep(1 * time.Second)
		s.processNextTask()
	}()
}

// RegisterExecutor 注册任务执行器
func (s *taskServiceImpl) RegisterExecutor(taskType string, executor TaskExecutor) {
	if s.executors == nil {
		s.executors = make(map[string]TaskExecutor)
	}
	s.executors[taskType] = executor
	logger.Infof("注册了 %s 任务的执行器", taskType)
}

// executeGenerateTask 执行AI生成任务
func (s *taskServiceImpl) executeGenerateTask(ctx context.Context, task *repository.Task, config map[string]interface{}) {
	logger.Infof("executeGenerateTask: 开始执行AI生成任务 %s", task.ID)

	// 检查上下文是否已被取消
	select {
	case <-ctx.Done():
		s.UpdateTaskStatus(task.ID, "cancelled")
		s.sendLog(task.ID, "info", "任务已被取消")
		s.taskMutex.Lock()
		delete(s.runningTasks, task.ID)
		s.taskMutex.Unlock()
		return
	default:
	}

	// 查找是否有注册的执行器
	if executor, exists := s.executors["generate"]; exists {
		// 使用注册的执行器执行
		logger.Infof("executeGenerateTask: 使用注册的执行器执行任务 %s", task.ID)
		executor.ExecuteGenerateTask(ctx, task.ID, config)

		// 任务完成后，从运行任务列表中移除
		s.taskMutex.Lock()
		delete(s.runningTasks, task.ID)
		s.taskMutex.Unlock()

		// 处理下一个等待的任务
		logger.Infof("executeGenerateTask: 任务 %s 已完成，准备处理下一个任务", task.ID)
		// 延迟一点再处理下一个任务，确保状态完全更新
		go func() {
			time.Sleep(1 * time.Second)
			s.processNextTask()
		}()
	} else {
		// 如果没有注册的执行器，标记为失败
		logger.Warnf("executeGenerateTask: 未找到 generate 任务的执行器")
		s.UpdateTaskError(task.ID, "未找到AI生成任务的执行器")
		s.UpdateTaskStatus(task.ID, "failed")

		s.taskMutex.Lock()
		delete(s.runningTasks, task.ID)
		s.taskMutex.Unlock()
	}
}

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// checkAndFixZombieTasks 检查并修复僵尸任务（数据库中是running状态但实际没有运行的任务）
func (s *taskServiceImpl) checkAndFixZombieTasks(tasks []*repository.Task) {
	if len(tasks) == 0 {
		return
	}

	// 获取当前实际运行的任务
	s.taskMutex.RLock()
	actualRunningTasks := make(map[string]bool)
	for taskID := range s.runningTasks {
		actualRunningTasks[taskID] = true
	}
	s.taskMutex.RUnlock()

	// 检查每个任务
	for _, task := range tasks {
		// 只检查running状态的任务
		if task.Status != "running" {
			continue
		}

		// 检查是否在实际运行列表中
		if !actualRunningTasks[task.ID] {
			// 这是一个僵尸任务，自动修复
			logger.Warnf("检测到僵尸任务: %s (类型: %s)，状态不一致，自动修复为failed", task.ID, task.Type)

			// 更新状态为failed
			if err := s.UpdateTaskStatus(task.ID, "failed"); err != nil {
				logger.Errorf("修复僵尸任务失败 %s: %v", task.ID, err)
				continue
			}

			// 更新错误信息
			if err := s.UpdateTaskError(task.ID, "任务状态不一致，已自动修复"); err != nil {
				logger.Errorf("更新僵尸任务错误信息失败 %s: %v", task.ID, err)
			}

			// 同步更新任务状态
			task.Status = "failed"
			task.ErrorMessage = "任务状态不一致，已自动修复"
		}
	}
}
