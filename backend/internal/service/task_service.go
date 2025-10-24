package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"pixiv-tailor/backend/internal/ai"
	"pixiv-tailor/backend/internal/crawler"
	"pixiv-tailor/backend/internal/logger"
	"pixiv-tailor/backend/internal/repository"
	"pixiv-tailor/backend/pkg/models"

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
	ListTasks(page, pageSize int32, status, taskType string) ([]*repository.Task, int, error)
	StartTask(id string) error
	StopTask(id string) error
	CancelTask(id string) error
	CleanupTasks(cleanupType string) (int, error)
	SetLogCallback(callback func(taskID, level, message string))
	SetStatusCallback(callback func(taskID, status string, progress int))
}

// taskServiceImpl 任务服务实现
type taskServiceImpl struct {
	storage        repository.Storage
	logCallback    func(taskID, level, message string)
	statusCallback func(taskID, status string, progress int)
	// 任务上下文管理
	runningTasks map[string]context.CancelFunc
	taskMutex    sync.RWMutex
}

// NewTaskService 创建任务服务实例
func NewTaskService(storage repository.Storage) TaskService {
	return &taskServiceImpl{
		storage:      storage,
		runningTasks: make(map[string]context.CancelFunc),
	}
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
		ID:        uuid.New().String(),
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

	// 自动启动任务
	logger.Infof("开始自动启动任务: %s", task.ID)
	if err := s.StartTask(task.ID); err != nil {
		// 如果启动失败，记录错误但不返回错误，因为任务已经创建成功
		logger.Warnf("任务启动失败 %s: %v", task.ID, err)
	} else {
		logger.Infof("任务启动成功: %s", task.ID)
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

	err := s.storage.UpdateTaskStatus(id, status)
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
	if s.statusCallback != nil {
		// 获取当前任务以获取状态信息
		task, err := s.GetTask(id)
		if err == nil {
			s.statusCallback(id, task.Status, progress)
		}
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

	// 启动实际的任务执行逻辑
	go s.executeTaskWithContext(ctx, task)

	return nil
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
	case "crawl":
		s.executeCrawlTask(ctx, task, config)
	case "generate":
		s.executeGenerateTask(ctx, task, config)
	default:
		s.UpdateTaskError(task.ID, fmt.Sprintf("不支持的任务类型: %s", task.Type))
		s.UpdateTaskStatus(task.ID, "failed")
	}
}

// executeCrawlTask 执行爬虫任务（重构版）
func (s *taskServiceImpl) executeCrawlTask(ctx context.Context, task *repository.Task, config map[string]interface{}) {
	logger.Infof("开始执行爬虫任务: %s", task.ID)
	s.sendLog(task.ID, "info", "开始执行爬虫任务")

	// 检查上下文是否已被取消
	select {
	case <-ctx.Done():
		s.UpdateTaskStatus(task.ID, "cancelled")
		s.sendLog(task.ID, "info", "任务已被取消")
		return
	default:
	}

	// 更新进度到10%
	s.UpdateTaskProgress(task.ID, 10)
	s.sendLog(task.ID, "info", "任务初始化完成 (10%)")

	// 创建任务特定的爬虫实例（使用任务特定的缓存和下载目录）
	s.sendLog(task.ID, "info", "正在创建任务特定的爬虫实例...")
	crawler, err := crawler.NewCrawlerForTask(task.ID, config)
	if err != nil {
		s.sendLog(task.ID, "error", fmt.Sprintf("创建爬虫失败: %v", err))
		s.UpdateTaskError(task.ID, fmt.Sprintf("创建爬虫失败: %v", err))
		s.UpdateTaskStatus(task.ID, "failed")
		return
	}
	s.sendLog(task.ID, "info", "任务特定的爬虫实例创建成功")

	// 配置代理设置
	if proxyEnabled, exists := config["proxy_enabled"]; exists && proxyEnabled.(bool) {
		if proxyURL, exists := config["proxy_url"]; exists && proxyURL != "" {
			s.sendLog(task.ID, "info", fmt.Sprintf("启用代理: %s", proxyURL))
			crawler.SetProxy(true, proxyURL.(string))
		}
	} else {
		s.sendLog(task.ID, "info", "未启用代理，使用直连")
		crawler.SetProxy(false, "")
	}

	// 配置Cookie设置
	if cookie, exists := config["cookie"]; exists && cookie != "" {
		if cookie.(string) == "default" {
			s.sendLog(task.ID, "info", "使用配置文件中的默认Cookie")
			// 不需要调用SetCookie，因为NewCrawlerForTask已经加载了配置文件中的Cookie
		} else {
			s.sendLog(task.ID, "info", "设置自定义Pixiv Cookie")
			crawler.SetCookie(cookie.(string))
		}
	} else {
		s.sendLog(task.ID, "warn", "未设置Cookie，可能无法访问Pixiv")
	}

	// 设置爬虫的任务ID和日志回调
	crawler.SetTaskID(task.ID)
	crawler.SetLogCallback(func(level, message string) {
		s.sendLog(task.ID, level, message)
	})

	// 更新进度到20%
	s.UpdateTaskProgress(task.ID, 20)

	// 检查上下文是否已被取消
	select {
	case <-ctx.Done():
		s.UpdateTaskStatus(task.ID, "cancelled")
		s.sendLog(task.ID, "info", "任务已被取消")
		return
	default:
	}

	// 根据配置执行爬取 - 使用重构后的API
	crawlType := config["type"].(string)
	var results []*models.PixivImage
	var crawlErr error

	s.sendLog(task.ID, "info", fmt.Sprintf("开始执行爬取任务，类型: %s", crawlType))

	switch crawlType {
	case "tag":
		query := config["query"].(string)
		order := config["order"].(string)
		mode := config["mode"].(string)
		limit := int(config["limit"].(float64))

		s.sendLog(task.ID, "info", fmt.Sprintf("开始按标签爬取: %s (排序: %s, 模式: %s, 限制: %d)", query, order, mode, limit))
		results, crawlErr = crawler.CrawlByTag(query, order, mode, limit)

	case "user":
		userIDValue, exists := config["user_id"]
		if !exists || userIDValue == nil {
			crawlErr = fmt.Errorf("用户ID不能为空")
			break
		}
		userID := int(userIDValue.(float64))
		limit := int(config["limit"].(float64))

		s.sendLog(task.ID, "info", fmt.Sprintf("开始按用户爬取: 用户ID %d (限制: %d)", userID, limit))
		results, crawlErr = crawler.CrawlByUser(userID, limit)

	case "illust":
		illustIDValue, exists := config["illust_id"]
		if !exists || illustIDValue == nil {
			crawlErr = fmt.Errorf("插画ID不能为空")
			break
		}
		illustID := int(illustIDValue.(float64))

		s.sendLog(task.ID, "info", fmt.Sprintf("开始按插画ID爬取: 插画ID %d", illustID))
		result, err := crawler.CrawlByIllust(illustID)
		crawlErr = err
		if result != nil {
			results = []*models.PixivImage{result}
		}

	default:
		crawlErr = fmt.Errorf("不支持的类型: %s", crawlType)
	}

	// 更新进度到80%
	s.UpdateTaskProgress(task.ID, 80)
	s.sendLog(task.ID, "info", "爬取任务执行完成，正在处理结果... (80%)")

	if crawlErr != nil {
		s.sendLog(task.ID, "error", fmt.Sprintf("爬取失败: %v", crawlErr))
		s.UpdateTaskError(task.ID, fmt.Sprintf("爬取失败: %v", crawlErr))
		s.UpdateTaskStatus(task.ID, "failed")
		return
	}

	s.sendLog(task.ID, "info", fmt.Sprintf("爬取成功，获得 %d 张图片", len(results)))

	// 更新获取到的图片数量
	s.UpdateTaskImagesFound(task.ID, len(results))

	// 保存爬取结果到数据库
	s.sendLog(task.ID, "info", "正在保存爬取结果到数据库...")

	// 先删除该任务的旧结果，避免重复
	if err := s.storage.DeleteCrawlResultsByTaskID(task.ID); err != nil {
		s.sendLog(task.ID, "warn", fmt.Sprintf("删除旧结果失败: %v", err))
	}

	for i, image := range results {
		crawlResult := &repository.CrawlResult{
			ID:        fmt.Sprintf("%s_%d_%d", task.ID, time.Now().Unix(), i+1), // 添加时间戳确保唯一性
			URL:       image.URL,
			Title:     image.Title,
			Author:    image.Author,
			Tags:      strings.Join(image.Tags, ","),
			ImageURL:  image.URL, // 使用URL作为图片URL
			CreatedAt: time.Now(),
		}

		if err := s.storage.AddCrawlResult(crawlResult); err != nil {
			s.sendLog(task.ID, "error", fmt.Sprintf("保存爬取结果失败: %v", err))
		}
	}
	s.sendLog(task.ID, "info", fmt.Sprintf("已保存 %d 条爬取结果到数据库", len(results)))

	// 开始下载图片 - 使用简化的API
	s.sendLog(task.ID, "info", "开始下载图片...")
	downloadedCount := 0

	// 为每个图片创建下载任务
	imageCount := make(map[int64]int) // 记录每个插画ID的图片数量
	for i, image := range results {
		// 检查上下文是否已被取消
		select {
		case <-ctx.Done():
			s.UpdateTaskStatus(task.ID, "cancelled")
			s.sendLog(task.ID, "info", "任务已被取消")
			return
		default:
		}

		// 为每个插画ID的图片编号
		imageCount[image.ID]++
		imageIndex := imageCount[image.ID]

		// 生成文件名 - 使用插画ID，格式为 artworks_{id}_p{page}.{ext}
		var filename string
		if image.ID > 0 {
			// 从URL中提取文件扩展名
			ext := filepath.Ext(image.URL)
			if ext == "" {
				ext = ".jpg" // 默认扩展名
			}
			filename = fmt.Sprintf("artworks_%d_p%02d%s", image.ID, imageIndex, ext)
		} else {
			// 如果没有插画ID，使用序号
			ext := filepath.Ext(image.URL)
			if ext == "" {
				ext = ".jpg"
			}
			filename = fmt.Sprintf("image_%d%s", i+1, ext)
		}

		// 使用简化的下载API，自动为任务生成目录
		s.sendLog(task.ID, "info", fmt.Sprintf("正在下载图片 %d/%d: %s", i+1, len(results), filename))

		// 创建进度回调
		progressCallback := func(url, filename string, downloaded, total int64, percent float64) {
			s.sendLog(task.ID, "info", fmt.Sprintf("下载进度: %s %.1f%% (%d/%d bytes)", filename, percent, downloaded, total))
		}

		// 使用统一的DownloadImage方法，自动为任务生成目录
		if err := crawler.DownloadImage(image.URL, filename, task.ID, progressCallback); err != nil {
			s.sendLog(task.ID, "error", fmt.Sprintf("下载图片失败: %v", err))
			continue
		}

		downloadedCount++
		s.sendLog(task.ID, "info", fmt.Sprintf("图片下载成功: %s", filename))

		// 更新下载进度
		progress := 80 + int(float64(i+1)/float64(len(results))*20) // 80-100%
		s.UpdateTaskProgress(task.ID, progress)
	}

	// 更新下载的图片数量
	s.UpdateTaskImagesDownloaded(task.ID, downloadedCount)

	// 更新进度到100%
	s.UpdateTaskProgress(task.ID, 100)
	s.sendLog(task.ID, "info", "任务处理完成 (100%)")

	// 任务完成
	s.UpdateTaskStatus(task.ID, "completed")
	s.sendLog(task.ID, "info", fmt.Sprintf("任务完成！共爬取 %d 张图片，成功下载 %d 张", len(results), downloadedCount))
	logger.Infof("爬虫任务完成: %s, 共爬取 %d 张图片，成功下载 %d 张", task.ID, len(results), downloadedCount)
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

// executeGenerateTask 执行生成任务
func (s *taskServiceImpl) executeGenerateTask(ctx context.Context, task *repository.Task, config map[string]interface{}) {
	logger.Infof("开始执行生成任务: %s", task.ID)
	s.sendLog(task.ID, "info", "开始执行图片生成任务")

	// 检查上下文是否已被取消
	select {
	case <-ctx.Done():
		s.UpdateTaskStatus(task.ID, "cancelled")
		s.sendLog(task.ID, "info", "任务已被取消")
		return
	default:
	}

	// 更新进度到10%
	s.UpdateTaskProgress(task.ID, 10)
	s.sendLog(task.ID, "info", "任务初始化完成 (10%)")

	// 创建AI生成器实例
	s.sendLog(task.ID, "info", "正在创建AI生成器实例...")
	generator, err := ai.NewGenerator()
	if err != nil {
		s.sendLog(task.ID, "error", fmt.Sprintf("创建AI生成器失败: %v", err))
		s.UpdateTaskError(task.ID, fmt.Sprintf("创建AI生成器失败: %v", err))
		s.UpdateTaskStatus(task.ID, "failed")
		return
	}
	s.sendLog(task.ID, "info", "AI生成器实例创建成功")

	// 更新进度到20%
	s.UpdateTaskProgress(task.ID, 20)

	// 构造生成请求
	s.sendLog(task.ID, "info", "正在构造生成请求...")
	request := &models.GenerateRequest{
		Prompt:         config["prompt"].(string),
		NegativePrompt: config["negative_prompt"].(string),
		Steps:          int(config["steps"].(float64)),
		CFGScale:       config["cfg_scale"].(float64),
		Width:          int(config["width"].(float64)),
		Height:         int(config["height"].(float64)),
		Seed:           int(config["seed"].(float64)),
		Model:          config["model"].(string),
		Sampler:        config["sampler"].(string),
		BatchSize:      int(config["batch_size"].(float64)),
		EnableHR:       config["enable_hr"].(bool),
		Loras:          []models.LoraConfig{}, // TODO: 从配置中解析LoRA配置
	}

	s.sendLog(task.ID, "info", fmt.Sprintf("开始生成图片: %s", request.Prompt))

	// 更新进度到30%
	s.UpdateTaskProgress(task.ID, 30)

	// 调用AI生成
	s.sendLog(task.ID, "info", "正在调用AI生成服务...")
	images, err := generator.GenerateImages(request)
	if err != nil {
		s.sendLog(task.ID, "error", fmt.Sprintf("AI生成失败: %v", err))
		s.UpdateTaskError(task.ID, fmt.Sprintf("AI生成失败: %v", err))
		s.UpdateTaskStatus(task.ID, "failed")
		return
	}

	s.sendLog(task.ID, "info", fmt.Sprintf("AI生成成功，生成了 %d 张图片", len(images)))

	// 更新进度到80%
	s.UpdateTaskProgress(task.ID, 80)

	// 保存生成结果
	s.sendLog(task.ID, "info", "正在保存生成结果...")
	resultPaths := make([]string, len(images))
	for i, image := range images {
		// 生成保存路径
		outputDir := filepath.Join("backend", "data", "generated")
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			s.sendLog(task.ID, "error", fmt.Sprintf("创建输出目录失败: %v", err))
			continue
		}

		filename := fmt.Sprintf("generated_%s_%d.jpg", task.ID, i+1)
		filepath := filepath.Join(outputDir, filename)

		// 保存图片到文件系统
		if err := generator.SaveImage(image, filepath); err != nil {
			s.sendLog(task.ID, "error", fmt.Sprintf("保存图片失败: %v", err))
			continue
		}

		// 保存图片信息到数据库
		repoImage := &repository.GeneratedImage{
			ID:        fmt.Sprintf("%d", image.ID),
			Prompt:    image.Prompt,
			ImageURL:  filepath, // 存储文件路径而不是base64数据
			Model:     image.Model,
			CreatedAt: image.CreatedAt,
		}

		if err := s.storage.AddGeneratedImage(repoImage); err != nil {
			s.sendLog(task.ID, "error", fmt.Sprintf("保存图片信息到数据库失败: %v", err))
		}

		resultPaths[i] = filepath
		s.sendLog(task.ID, "info", fmt.Sprintf("图片已保存: %s", filepath))
	}

	// 更新进度到100%
	s.UpdateTaskProgress(task.ID, 100)

	// 更新任务状态为完成
	s.UpdateTaskStatus(task.ID, "completed")
	s.UpdateTaskCompletedAt(task.ID, time.Now())

	s.sendLog(task.ID, "info", fmt.Sprintf("生成任务执行完成，共生成 %d 张图片", len(images)))
	logger.Infof("生成任务执行完成: %s", task.ID)
}

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

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
