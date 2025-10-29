package grpc

import (
	"context"
	"fmt"
	"time"

	"pixiv-tailor/backend/internal/logger"
	"pixiv-tailor/backend/internal/repository"
	"pixiv-tailor/backend/internal/service"
	pb "pixiv-tailor/proto"
)

// PixivTailorServer gRPC 服务器实现
type PixivTailorServer struct {
	pb.UnimplementedPixivTailorServiceServer
	TaskService   service.TaskService
	ConfigService service.ConfigService
	DataService   service.DataService
	SystemService service.SystemService
}

// ============================================================================
// 转换函数
// ============================================================================

// convertTask 转换任务类型
func convertTask(task *repository.Task) *pb.Task {
	if task == nil {
		return nil
	}
	return &pb.Task{
		Id:           task.ID,
		Type:         task.Type,
		Status:       task.Status,
		Config:       task.Config,
		CreatedAt:    task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    task.UpdatedAt.Format(time.RFC3339),
		ErrorMessage: task.ErrorMessage,
	}
}

// convertTasks 转换任务列表
func convertTasks(tasks []*repository.Task) []*pb.Task {
	result := make([]*pb.Task, len(tasks))
	for i, task := range tasks {
		result[i] = convertTask(task)
	}
	return result
}

// convertCrawlResult 转换爬取结果
func convertCrawlResult(result *repository.CrawlResult) *pb.CrawlResult {
	if result == nil {
		return nil
	}
	return &pb.CrawlResult{
		Id:        result.ID,
		Url:       result.URL,
		Title:     result.Title,
		Author:    result.Author,
		Tags:      []string{}, // TODO: 解析JSON字符串
		ImageUrl:  result.ImageURL,
		CreatedAt: result.CreatedAt.Format(time.RFC3339),
	}
}

// convertCrawlResults 转换爬取结果列表
func convertCrawlResults(results []*repository.CrawlResult) []*pb.CrawlResult {
	result := make([]*pb.CrawlResult, len(results))
	for i, res := range results {
		result[i] = convertCrawlResult(res)
	}
	return result
}

// convertGeneratedImage 转换生成图像
func convertGeneratedImage(image *repository.GeneratedImage) *pb.GeneratedImage {
	if image == nil {
		return nil
	}
	return &pb.GeneratedImage{
		Id:        image.ID,
		Prompt:    image.Prompt,
		ImageUrl:  image.ImageURL,
		Model:     image.Model,
		CreatedAt: image.CreatedAt.Format(time.RFC3339),
	}
}

// convertGeneratedImages 转换生成图像列表
func convertGeneratedImages(images []*repository.GeneratedImage) []*pb.GeneratedImage {
	result := make([]*pb.GeneratedImage, len(images))
	for i, img := range images {
		result[i] = convertGeneratedImage(img)
	}
	return result
}

// convertTrainedModel 转换训练模型
func convertTrainedModel(model *repository.TrainedModel) *pb.TrainedModel {
	if model == nil {
		return nil
	}
	return &pb.TrainedModel{
		Id:        model.ID,
		Name:      model.Name,
		Type:      model.Type,
		Path:      model.Path,
		CreatedAt: model.CreatedAt.Format(time.RFC3339),
	}
}

// convertTrainedModels 转换训练模型列表
func convertTrainedModels(models []*repository.TrainedModel) []*pb.TrainedModel {
	result := make([]*pb.TrainedModel, len(models))
	for i, model := range models {
		result[i] = convertTrainedModel(model)
	}
	return result
}

// ============================================================================
// 配置管理
// ============================================================================

// GetConfig 获取配置
func (s *PixivTailorServer) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.GetConfigResponse, error) {
	logger.Infof("收到获取配置请求: %v", req)

	config, err := s.ConfigService.GetConfig(req.Module)
	if err != nil {
		return &pb.GetConfigResponse{
			Status: &pb.Status{
				Code:    1,
				Message: "获取配置失败",
				Details: err.Error(),
			},
		}, nil
	}

	return &pb.GetConfigResponse{
		Status: &pb.Status{
			Code:    0,
			Message: "成功",
		},
		Config: &pb.ConfigModule{
			Name:   req.Module,
			Config: config,
		},
	}, nil
}

// UpdateConfig 更新配置
func (s *PixivTailorServer) UpdateConfig(ctx context.Context, req *pb.UpdateConfigRequest) (*pb.UpdateConfigResponse, error) {
	logger.Infof("收到更新配置请求: %v", req)

	err := s.ConfigService.SetConfig(req.Module, req.Config)
	if err != nil {
		return &pb.UpdateConfigResponse{
			Status: &pb.Status{
				Code:    1,
				Message: "更新配置失败",
				Details: err.Error(),
			},
		}, nil
	}

	return &pb.UpdateConfigResponse{
		Status: &pb.Status{
			Code:    0,
			Message: "成功",
		},
	}, nil
}

// ExportConfig 导出配置
func (s *PixivTailorServer) ExportConfig(ctx context.Context, req *pb.ExportConfigRequest) (*pb.ExportConfigResponse, error) {
	logger.Infof("收到导出配置请求: %v", req)

	// 简化实现，返回空配置

	return &pb.ExportConfigResponse{
		Status: &pb.Status{
			Code:    0,
			Message: "成功",
		},
	}, nil
}

// ImportConfig 导入配置
func (s *PixivTailorServer) ImportConfig(ctx context.Context, req *pb.ImportConfigRequest) (*pb.ImportConfigResponse, error) {
	logger.Infof("收到导入配置请求: %v", req)

	// 简化实现，直接返回成功
	return &pb.ImportConfigResponse{
		Status: &pb.Status{
			Code:    0,
			Message: "成功",
		},
	}, nil
}

// ============================================================================
// 任务管理
// ============================================================================

// CreateTask 创建任务
func (s *PixivTailorServer) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	logger.Infof("收到创建任务请求: %v", req)

	task, err := s.TaskService.CreateTask(req.Type, req.Config)
	if err != nil {
		return &pb.CreateTaskResponse{
			Status: &pb.Status{
				Code:    1,
				Message: "创建任务失败",
				Details: err.Error(),
			},
		}, nil
	}

	return &pb.CreateTaskResponse{
		Status: &pb.Status{
			Code:    0,
			Message: "成功",
		},
		Task: convertTask(task),
	}, nil
}

// GetTaskStatus 获取任务状态
func (s *PixivTailorServer) GetTaskStatus(ctx context.Context, req *pb.GetTaskStatusRequest) (*pb.GetTaskStatusResponse, error) {
	logger.Infof("收到获取任务状态请求: %v", req)

	task, err := s.TaskService.GetTask(req.TaskId)
	if err != nil {
		return &pb.GetTaskStatusResponse{
			Status: &pb.Status{
				Code:    1,
				Message: "获取任务状态失败",
				Details: err.Error(),
			},
		}, nil
	}

	return &pb.GetTaskStatusResponse{
		Status: &pb.Status{
			Code:    0,
			Message: "成功",
		},
		Task: convertTask(task),
	}, nil
}

// ListTasks 列出任务
func (s *PixivTailorServer) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	logger.Infof("收到列出任务请求: %v", req)

	tasks, total, err := s.TaskService.ListTasks(req.Pagination.Page, req.Pagination.PageSize, req.Status, req.Type)
	if err != nil {
		return &pb.ListTasksResponse{
			Status: &pb.Status{
				Code:    1,
				Message: "列出任务失败",
				Details: err.Error(),
			},
		}, nil
	}

	return &pb.ListTasksResponse{
		Status: &pb.Status{
			Code:    0,
			Message: "成功",
		},
		Tasks: convertTasks(tasks),
		Pagination: &pb.Pagination{
			Page:     req.Pagination.Page,
			PageSize: req.Pagination.PageSize,
			Total:    int32(total),
		},
	}, nil
}

// CancelTask 取消任务
func (s *PixivTailorServer) CancelTask(ctx context.Context, req *pb.CancelTaskRequest) (*pb.CancelTaskResponse, error) {
	logger.Infof("收到取消任务请求: %v", req)

	err := s.TaskService.CancelTask(req.TaskId)
	if err != nil {
		return &pb.CancelTaskResponse{
			Status: &pb.Status{
				Code:    1,
				Message: "取消任务失败",
				Details: err.Error(),
			},
		}, nil
	}

	return &pb.CancelTaskResponse{
		Status: &pb.Status{
			Code:    0,
			Message: "成功",
		},
	}, nil
}

// ============================================================================
// 进度跟踪
// ============================================================================

// GetTaskProgress 获取任务进度（流式响应）
func (s *PixivTailorServer) GetTaskProgress(req *pb.GetTaskProgressRequest, stream pb.PixivTailorService_GetTaskProgressServer) error {
	logger.Infof("收到获取任务进度请求: %v", req)

	// 获取任务信息
	task, err := s.TaskService.GetTask(req.TaskId)
	if err != nil {
		// 发送错误状态
		progress := &pb.TaskProgressUpdate{
			TaskId:   req.TaskId,
			Progress: 0,
			Status:   "error",
			Message:  "获取任务失败",
			Details:  err.Error(),
		}
		return stream.Send(progress)
	}

	// 发送初始进度
	progress := &pb.TaskProgressUpdate{
		TaskId:   req.TaskId,
		Progress: int32(task.Progress),
		Status:   task.Status,
		Message:  "获取任务进度",
		Details:  task.Config,
	}

	if err := stream.Send(progress); err != nil {
		return err
	}

	// 如果任务正在运行，持续发送进度更新
	if task.Status == "running" {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for i := 0; i < 5; i++ { // 最多发送5次更新
			select {
			case <-ticker.C:
				// 重新获取任务状态
				updatedTask, err := s.TaskService.GetTask(req.TaskId)
				if err != nil {
					continue
				}

				progress := &pb.TaskProgressUpdate{
					TaskId:   req.TaskId,
					Progress: int32(updatedTask.Progress),
					Status:   updatedTask.Status,
					Message:  "进度更新",
					Details:  fmt.Sprintf("当前进度: %d%%", updatedTask.Progress),
				}

				if err := stream.Send(progress); err != nil {
					return err
				}

				// 如果任务完成或失败，停止更新
				if updatedTask.Status == "completed" || updatedTask.Status == "failed" || updatedTask.Status == "cancelled" {
					return nil
				}
			case <-stream.Context().Done():
				return stream.Context().Err()
			}
		}
	}

	return nil
}

// GetTaskLogs 获取任务日志（流式响应）
func (s *PixivTailorServer) GetTaskLogs(req *pb.GetTaskLogsRequest, stream pb.PixivTailorService_GetTaskLogsServer) error {
	logger.Infof("收到获取任务日志请求: %v", req)

	// 获取任务信息验证任务是否存在
	task, err := s.TaskService.GetTask(req.TaskId)
	if err != nil {
		logEntry := &pb.LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     "error",
			Message:   "获取任务失败",
			Details:   err.Error(),
		}
		return stream.Send(logEntry)
	}

	// 发送任务基本信息日志
	logs := []struct {
		level   string
		message string
		details string
	}{
		{"info", "任务信息", fmt.Sprintf("任务ID: %s, 类型: %s, 状态: %s", task.ID, task.Type, task.Status)},
		{"info", "任务配置", task.Config},
		{"info", "创建时间", task.CreatedAt.Format(time.RFC3339)},
		{"info", "更新时间", task.UpdatedAt.Format(time.RFC3339)},
	}

	// 如果有错误信息，添加错误日志
	if task.ErrorMessage != "" {
		logs = append(logs, struct {
			level   string
			message string
			details string
		}{"error", "任务错误", task.ErrorMessage})
	}

	// 根据任务状态添加状态日志
	switch task.Status {
	case "pending":
		logs = append(logs, struct {
			level   string
			message string
			details string
		}{"info", "任务状态", "任务等待执行"})
	case "running":
		logs = append(logs, struct {
			level   string
			message string
			details string
		}{"info", "任务状态", fmt.Sprintf("任务正在执行，进度: %d%%", task.Progress)})
	case "completed":
		logs = append(logs, struct {
			level   string
			message string
			details string
		}{"info", "任务状态", "任务已完成"})
	case "failed":
		logs = append(logs, struct {
			level   string
			message string
			details string
		}{"error", "任务状态", "任务执行失败"})
	case "cancelled":
		logs = append(logs, struct {
			level   string
			message string
			details string
		}{"warning", "任务状态", "任务已取消"})
	}

	// 流式发送日志
	limit := int(req.Limit)
	if limit <= 0 || limit > len(logs) {
		limit = len(logs)
	}

	for i := 0; i < limit; i++ {
		logEntry := &pb.LogEntry{
			Timestamp: time.Now().Add(-time.Duration(len(logs)-i) * time.Second).Format(time.RFC3339),
			Level:     logs[i].level,
			Message:   logs[i].message,
			Details:   logs[i].details,
		}

		if err := stream.Send(logEntry); err != nil {
			return err
		}

		// 模拟延迟
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// ============================================================================
// 数据查询
// ============================================================================

// GetCrawlResults 获取爬取结果
func (s *PixivTailorServer) GetCrawlResults(ctx context.Context, req *pb.GetCrawlResultsRequest) (*pb.GetCrawlResultsResponse, error) {
	logger.Infof("收到获取爬取结果请求: %v", req)

	results, total, err := s.DataService.GetCrawlResults(req.Pagination.Page, req.Pagination.PageSize, req.Tags, req.Author)
	if err != nil {
		return &pb.GetCrawlResultsResponse{
			Status: &pb.Status{
				Code:    1,
				Message: "获取爬取结果失败",
				Details: err.Error(),
			},
		}, nil
	}

	return &pb.GetCrawlResultsResponse{
		Status: &pb.Status{
			Code:    0,
			Message: "成功",
		},
		Results: convertCrawlResults(results),
		Pagination: &pb.Pagination{
			Page:     req.Pagination.Page,
			PageSize: req.Pagination.PageSize,
			Total:    int32(total),
		},
	}, nil
}

// GetGeneratedImages 获取生成图像
func (s *PixivTailorServer) GetGeneratedImages(ctx context.Context, req *pb.GetGeneratedImagesRequest) (*pb.GetGeneratedImagesResponse, error) {
	logger.Infof("收到获取生成图像请求: %v", req)

	images, total, err := s.DataService.GetGeneratedImages(req.Pagination.Page, req.Pagination.PageSize, req.Model)
	if err != nil {
		return &pb.GetGeneratedImagesResponse{
			Status: &pb.Status{
				Code:    1,
				Message: "获取生成图像失败",
				Details: err.Error(),
			},
		}, nil
	}

	return &pb.GetGeneratedImagesResponse{
		Status: &pb.Status{
			Code:    0,
			Message: "成功",
		},
		Images: convertGeneratedImages(images),
		Pagination: &pb.Pagination{
			Page:     req.Pagination.Page,
			PageSize: req.Pagination.PageSize,
			Total:    int32(total),
		},
	}, nil
}

// GetTrainedModels 获取训练模型
func (s *PixivTailorServer) GetTrainedModels(ctx context.Context, req *pb.GetTrainedModelsRequest) (*pb.GetTrainedModelsResponse, error) {
	logger.Infof("收到获取训练模型请求: %v", req)

	models, total, err := s.DataService.GetTrainedModels(req.Pagination.Page, req.Pagination.PageSize, req.Type)
	if err != nil {
		return &pb.GetTrainedModelsResponse{
			Status: &pb.Status{
				Code:    1,
				Message: "获取训练模型失败",
				Details: err.Error(),
			},
		}, nil
	}

	return &pb.GetTrainedModelsResponse{
		Status: &pb.Status{
			Code:    0,
			Message: "成功",
		},
		Models: convertTrainedModels(models),
		Pagination: &pb.Pagination{
			Page:     req.Pagination.Page,
			PageSize: req.Pagination.PageSize,
			Total:    int32(total),
		},
	}, nil
}

// ============================================================================
// 系统监控
// ============================================================================

// GetSystemInfo 获取系统信息
func (s *PixivTailorServer) GetSystemInfo(ctx context.Context, req *pb.GetSystemInfoRequest) (*pb.GetSystemInfoResponse, error) {
	logger.Infof("收到获取系统信息请求: %v", req)

	// 获取系统信息
	sysInfo, err := s.SystemService.GetSystemInfo()
	if err != nil {
		return &pb.GetSystemInfoResponse{
			Status: &pb.Status{
				Code:    1,
				Message: "获取系统信息失败",
				Details: err.Error(),
			},
		}, nil
	}

	// 获取系统指标
	metrics, err := s.SystemService.GetMetrics()
	if err != nil {
		return &pb.GetSystemInfoResponse{
			Status: &pb.Status{
				Code:    1,
				Message: "获取系统指标失败",
				Details: err.Error(),
			},
		}, nil
	}

	// 获取健康状态
	health, err := s.SystemService.GetHealthStatus()
	if err != nil {
		return &pb.GetSystemInfoResponse{
			Status: &pb.Status{
				Code:    1,
				Message: "获取健康状态失败",
				Details: err.Error(),
			},
		}, nil
	}

	// 计算内存使用率
	memoryUsage := int32(0)
	if metrics.Memory.Sys > 0 {
		memoryUsage = int32((metrics.Memory.Alloc * 100) / metrics.Memory.Sys)
	}

	// 构造响应
	systemInfo := &pb.SystemInfo{
		Version:           sysInfo.Version,
		Status:            health.Status,
		Uptime:            metrics.Uptime.String(),
		ActiveConnections: int32(metrics.Goroutines),
		Memory: &pb.MemoryInfo{
			Used:      int64(metrics.Memory.Alloc),
			Total:     int64(metrics.Memory.Sys),
			Usage:     memoryUsage,
			Available: int64(metrics.Memory.Sys - metrics.Memory.Alloc),
		},
		Cpu: &pb.CPUInfo{
			Usage:   0, // TODO: 实现CPU使用率监控
			Average: 0,
			Peak:    0,
		},
		Disk: &pb.DiskInfo{
			Data:        0, // TODO: 实现磁盘使用监控
			DataUsed:    0,
			DataTotal:   0,
			Models:      0,
			ModelsUsed:  0,
			ModelsTotal: 0,
			Logs:        0,
			LogsUsed:    0,
			LogsTotal:   0,
		},
		Warnings: []string{}, // TODO: 实现警告信息收集
	}

	// 添加警告信息
	if memoryUsage > 80 {
		systemInfo.Warnings = append(systemInfo.Warnings, "内存使用率超过80%")
	}
	if metrics.Goroutines > 1000 {
		systemInfo.Warnings = append(systemInfo.Warnings, "Goroutine数量过多")
	}

	return &pb.GetSystemInfoResponse{
		Status: &pb.Status{
			Code:    0,
			Message: "成功",
		},
		SystemInfo: systemInfo,
	}, nil
}
