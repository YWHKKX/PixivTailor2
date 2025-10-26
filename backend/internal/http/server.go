package http

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"pixiv-tailor/backend/internal/repository"
	"pixiv-tailor/backend/internal/service"
	"pixiv-tailor/backend/pkg/paths"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// HTTP 服务器 - 基于 plan.md 设计
type HTTPServer struct {
	TaskService             service.TaskService
	ConfigService           service.ConfigService
	DataService             service.DataService
	SystemService           service.SystemService
	GenerationConfigService service.GenerationConfigService
	router                  *mux.Router
	upgrader                websocket.Upgrader
	clients                 map[*websocket.Conn]bool
	clientsMutex            sync.RWMutex
	broadcast               chan []byte
	// WebUI相关字段
	webUIProcess     *os.Process
	webUILogChannels []chan string
	webUIStatus      string
	webUILogMutex    sync.RWMutex
}

// 响应结构
type APIResponse struct {
	Status Status      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}

type Status struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// 生成请求结构
type GenerationRequest struct {
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negative_prompt"`
	Steps          int     `json:"steps"`
	CFGScale       float64 `json:"cfg_scale"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	Seed           int64   `json:"seed"`
	Model          string  `json:"model"`
	Sampler        string  `json:"sampler"`
	BatchSize      int     `json:"batch_size"`
	EnableHR       bool    `json:"enable_hr"`
}

// 生成响应结构
type GenerationResponse struct {
	TaskID      string     `json:"task_id"`
	Status      string     `json:"status"`
	Progress    int        `json:"progress"`
	Message     string     `json:"message"`
	Result      []string   `json:"result,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// 任务状态响应
type TaskStatusResponse struct {
	TaskID       string     `json:"task_id"`
	Status       string     `json:"status"`
	Progress     int        `json:"progress"`
	Message      string     `json:"message"`
	ErrorMessage string     `json:"error_message,omitempty"`
	Result       []string   `json:"result,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// WebSocket 消息结构
type WSMessage struct {
	Type    string      `json:"type"`
	TaskID  string      `json:"task_id,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// 创建新的 HTTP 服务器
func NewHTTPServer(
	taskService service.TaskService,
	configService service.ConfigService,
	dataService service.DataService,
	systemService service.SystemService,
	generationConfigService service.GenerationConfigService,
) *HTTPServer {
	server := &HTTPServer{
		TaskService:             taskService,
		ConfigService:           configService,
		DataService:             dataService,
		SystemService:           systemService,
		GenerationConfigService: generationConfigService,
		router:                  mux.NewRouter(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源
			},
		},
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
	}

	// 设置任务服务的日志回调，将任务日志广播到WebSocket客户端
	taskService.SetLogCallback(func(taskID, level, message string) {
		server.broadcastLogMessage(taskID, level, message)
	})

	// 设置任务服务的状态回调，将任务状态更新广播到WebSocket客户端
	taskService.SetStatusCallback(func(taskID, status string, progress int) {
		server.broadcastTaskUpdate(taskID, status, progress)
		// 同时发送全局日志
		server.broadcastGlobalLog("info", fmt.Sprintf("任务 %s 状态更新: %s (进度: %d%%)", taskID, status, progress))
	})

	server.setupRoutes()
	server.setupWebSocket()
	return server
}

// 设置路由
func (s *HTTPServer) setupRoutes() {
	// 全局CORS中间件
	s.router.Use(s.corsMiddleware)

	// 添加一个通用的OPTIONS处理器，必须在其他路由之前
	s.router.Methods("OPTIONS").HandlerFunc(s.handleAllOptions)

	// 为所有API路径添加OPTIONS支持
	s.router.PathPrefix("/api").Methods("OPTIONS").HandlerFunc(s.handleAllOptions)

	// API 路由
	api := s.router.PathPrefix("/api").Subrouter()

	// 生成图像
	api.HandleFunc("/generate", s.handleGenerate).Methods("POST", "OPTIONS")
	api.HandleFunc("/generate-with-config", s.handleGenerateWithConfig).Methods("POST", "OPTIONS")

	// WebUI管理
	api.HandleFunc("/webui/start", s.handleStartWebUI).Methods("POST", "OPTIONS")
	api.HandleFunc("/webui/start-external", s.handleStartWebUIExternal).Methods("POST", "OPTIONS")
	api.HandleFunc("/webui/stop", s.handleStopWebUI).Methods("POST", "OPTIONS")
	api.HandleFunc("/webui/status", s.handleWebUIStatus).Methods("GET", "OPTIONS")
	api.HandleFunc("/webui/logs", s.handleWebUILogs).Methods("GET", "OPTIONS")

	// 配置文件管理
	api.HandleFunc("/configs", s.handleListConfigs).Methods("GET", "OPTIONS")
	api.HandleFunc("/configs/create", s.handleCreateConfig).Methods("POST", "OPTIONS")
	api.HandleFunc("/configs/categories", s.handleGetCategories).Methods("GET", "POST", "OPTIONS")
	api.HandleFunc("/configs/default", s.handleGetDefaultConfig).Methods("GET", "OPTIONS")
	api.HandleFunc("/configs/default", s.handleSetDefaultConfig).Methods("POST", "OPTIONS")
	api.HandleFunc("/configs/import", s.handleImportConfig).Methods("POST", "OPTIONS")
	api.HandleFunc("/configs/export", s.handleExportConfigs).Methods("POST", "OPTIONS")
	api.HandleFunc("/configs/import-file", s.handleImportFromFile).Methods("POST", "OPTIONS")
	api.HandleFunc("/configs/export-file", s.handleExportToFile).Methods("POST", "OPTIONS")
	api.HandleFunc("/configs/load-from-files", s.handleLoadConfigsFromFiles).Methods("POST", "OPTIONS")
	api.HandleFunc("/configs/name/{name}", s.handleGetConfigByName).Methods("GET", "OPTIONS")
	api.HandleFunc("/configs/{id}/use", s.handleUseConfig).Methods("POST", "OPTIONS")
	api.HandleFunc("/configs/{id}", s.handleGetConfig).Methods("GET", "OPTIONS")
	api.HandleFunc("/configs/{id}", s.handleUpdateConfig).Methods("PUT", "OPTIONS")
	api.HandleFunc("/configs/{id}", s.handleDeleteConfig).Methods("DELETE", "OPTIONS")

	// 文件系统配置管理API
	api.HandleFunc("/configs/file/create", s.handleCreateConfigFile).Methods("POST", "OPTIONS")
	api.HandleFunc("/configs/file/{id}/update", s.handleUpdateConfigFile).Methods("PUT", "OPTIONS")
	api.HandleFunc("/configs/file/{id}/delete", s.handleDeleteConfigFile).Methods("DELETE", "OPTIONS")

	// 任务管理
	api.HandleFunc("/status", s.handleGetTaskStatus).Methods("POST", "OPTIONS")
	api.HandleFunc("/cancel", s.handleCancelTask).Methods("POST", "OPTIONS")
	api.HandleFunc("/delete", s.handleDeleteTask).Methods("POST", "OPTIONS")
	api.HandleFunc("/tasks", s.handleGetTasks).Methods("POST", "OPTIONS")

	// 任务图片服务
	api.HandleFunc("/tasks/{taskId}/images/{imageIndex}", s.handleGetTaskImage).Methods("GET")

	// 配置管理
	api.HandleFunc("/config/get", s.handleGetConfig).Methods("POST", "OPTIONS")
	api.HandleFunc("/config/update", s.handleUpdateConfig).Methods("POST", "OPTIONS")

	// 爬虫管理
	api.HandleFunc("/crawl/create", s.handleCreateCrawlTask).Methods("POST", "OPTIONS")
	api.HandleFunc("/crawl/results", s.handleGetCrawlResults).Methods("POST", "OPTIONS")
	api.HandleFunc("/generated/images", s.handleGetGeneratedImages).Methods("POST", "OPTIONS")

	// 任务管理
	api.HandleFunc("/task/start", s.handleStartTask).Methods("POST", "OPTIONS")
	api.HandleFunc("/task/stop", s.handleStopTask).Methods("POST", "OPTIONS")
	api.HandleFunc("/task/cleanup", s.handleCleanupTasks).Methods("POST", "OPTIONS")

	// 系统信息
	api.HandleFunc("/system/info", s.handleGetSystemInfo).Methods("POST", "OPTIONS")

	// 图片服务
	api.PathPrefix("/images/").HandlerFunc(s.handleServeImage).Methods("GET", "OPTIONS")

	// 文件树服务
	api.HandleFunc("/filetree", s.handleGetFileTree).Methods("POST", "OPTIONS")

	// 健康检查
	s.router.HandleFunc("/health", s.handleHealthCheck).Methods("GET", "OPTIONS")

	// WebSocket
	s.router.HandleFunc("/ws", s.handleWebSocket)
}

// CORS 中间件
func (s *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// 处理所有OPTIONS请求，不管路由如何
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// 处理OPTIONS请求
func (s *HTTPServer) handleOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(http.StatusOK)
}

// 处理所有OPTIONS请求的通用处理器
func (s *HTTPServer) handleAllOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(http.StatusOK)
}

// 生成图像处理器 - 已删除，使用 handleGenerateWithConfig 替代
func (s *HTTPServer) handleGenerate(w http.ResponseWriter, r *http.Request) {
	s.sendErrorResponse(w, http.StatusNotImplemented, "此端点已废弃", "请使用 /api/generate-with-config 端点")
}

// 获取任务状态处理器
func (s *HTTPServer) handleGetTaskStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TaskID string `json:"task_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	task, err := s.TaskService.GetTask(req.TaskID)
	if err != nil {
		s.sendErrorResponse(w, http.StatusNotFound, "Task not found", err.Error())
		return
	}

	response := TaskStatusResponse{
		TaskID:       task.ID,
		Status:       task.Status,
		Progress:     task.Progress,
		Message:      task.Config,
		ErrorMessage: task.ErrorMessage,
		CreatedAt:    task.CreatedAt,
		CompletedAt:  nil, // 默认不设置完成时间
	}

	// 只有当任务完成、失败或取消时才设置完成时间
	if task.Status == "completed" || task.Status == "failed" || task.Status == "cancelled" {
		response.CompletedAt = &task.UpdatedAt
	}

	s.sendSuccessResponse(w, response)
}

// 取消任务处理器
func (s *HTTPServer) handleCancelTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TaskID string `json:"task_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// 获取任务信息
	task, err := s.TaskService.GetTask(req.TaskID)
	if err != nil {
		s.sendErrorResponse(w, http.StatusNotFound, "Task not found", err.Error())
		return
	}

	// 如果是AI生成任务且正在运行，先停止WebUI的工作
	if task.Type == "generate" && (task.Status == "running" || task.Status == "pending") {
		log.Printf("取消AI生成任务 %s，停止WebUI工作", req.TaskID)
		if err := s.stopWebUIGeneration(); err != nil {
			log.Printf("停止WebUI工作失败: %v", err)
		}
	}

	err = s.TaskService.CancelTask(req.TaskID)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to cancel task", err.Error())
		return
	}

	s.sendSuccessResponse(w, map[string]string{"message": "Task cancelled successfully"})
}

// 删除任务处理器
func (s *HTTPServer) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TaskID string `json:"task_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := s.TaskService.DeleteTask(req.TaskID)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to delete task", err.Error())
		return
	}

	s.sendSuccessResponse(w, map[string]string{"message": "Task deleted successfully"})
}

// 获取任务图片处理器
func (s *HTTPServer) handleGetTaskImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["taskId"]
	imageIndex := vars["imageIndex"]

	// 获取路径管理器
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Path manager not initialized", "")
		return
	}

	// 构建图片文件路径
	taskDir := pathManager.GetTaskImagesDir(taskID)

	// 尝试不同的文件扩展名
	extensions := []string{"png", "jpg", "jpeg", "webp"}
	var imagePath string
	var found bool

	for _, ext := range extensions {
		imagePath = filepath.Join(taskDir, fmt.Sprintf("generated_%s_%s.%s", taskID, imageIndex, ext))
		if _, err := os.Stat(imagePath); err == nil {
			found = true
			break
		}
	}

	if !found {
		s.sendErrorResponse(w, http.StatusNotFound, "Image not found", "")
		return
	}

	// 根据文件扩展名设置Content-Type
	ext := filepath.Ext(imagePath)
	var contentType string
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".webp":
		contentType = "image/webp"
	default:
		contentType = "image/png"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// 读取并返回图片文件
	http.ServeFile(w, r, imagePath)
}

// 获取任务列表处理器
func (s *HTTPServer) handleGetTasks(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pagination struct {
			Page     int `json:"page"`
			PageSize int `json:"page_size"`
			Total    int `json:"total"`
		} `json:"pagination"`
		Status string `json:"status"`
		Type   string `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	tasks, total, err := s.TaskService.ListTasks(int32(req.Pagination.Page), int32(req.Pagination.PageSize), req.Status, req.Type)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to get tasks", err.Error())
		return
	}

	// 确保tasks字段始终是数组而不是null
	if tasks == nil {
		tasks = []*repository.Task{}
	}

	response := map[string]interface{}{
		"tasks": tasks,
		"pagination": map[string]interface{}{
			"page":      req.Pagination.Page,
			"page_size": req.Pagination.PageSize,
			"total":     total,
		},
	}

	s.sendSuccessResponse(w, response)
}

// 获取爬取结果处理器
func (s *HTTPServer) handleGetCrawlResults(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pagination struct {
			Page     int `json:"page"`
			PageSize int `json:"page_size"`
			Total    int `json:"total"`
		} `json:"pagination"`
		Tags   []string `json:"tags"`
		Author string   `json:"author"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	results, total, err := s.DataService.GetCrawlResults(int32(req.Pagination.Page), int32(req.Pagination.PageSize), req.Tags, req.Author)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to get crawl results", err.Error())
		return
	}

	// 确保results字段始终是数组而不是null
	if results == nil {
		results = []*repository.CrawlResult{}
	}

	// 将CrawlResult转换为PixivImage格式
	pixivImages := make([]map[string]interface{}, len(results))
	for i, result := range results {
		// 解析tags字符串为数组
		var tags []string
		if result.Tags != "" {
			json.Unmarshal([]byte(result.Tags), &tags)
		}

		pixivImages[i] = map[string]interface{}{
			"id":            result.ID,
			"title":         result.Title,
			"author":        result.Author,
			"author_id":     0, // CrawlResult中没有author_id
			"tags":          tags,
			"url":           result.ImageURL, // 使用ImageURL作为主URL
			"thumbnail_url": result.ImageURL, // 使用ImageURL作为缩略图URL
			"width":         0,               // CrawlResult中没有尺寸信息
			"height":        0,
			"bookmarks":     0,     // CrawlResult中没有收藏数
			"views":         0,     // CrawlResult中没有浏览数
			"is_r18":        false, // CrawlResult中没有R18标记
			"created_at":    result.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updated_at":    result.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	response := map[string]interface{}{
		"results": pixivImages,
		"pagination": map[string]interface{}{
			"page":      req.Pagination.Page,
			"page_size": req.Pagination.PageSize,
			"total":     total,
		},
	}

	s.sendSuccessResponse(w, response)
}

// 获取生成图像处理器
func (s *HTTPServer) handleGetGeneratedImages(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pagination struct {
			Page     int `json:"page"`
			PageSize int `json:"page_size"`
			Total    int `json:"total"`
		} `json:"pagination"`
		Model string `json:"model"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	images, total, err := s.DataService.GetGeneratedImages(int32(req.Pagination.Page), int32(req.Pagination.PageSize), req.Model)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to get generated images", err.Error())
		return
	}

	// 确保images字段始终是数组而不是null
	if images == nil {
		images = []*repository.GeneratedImage{}
	}

	response := map[string]interface{}{
		"images": images,
		"pagination": map[string]interface{}{
			"page":      req.Pagination.Page,
			"page_size": req.Pagination.PageSize,
			"total":     total,
		},
	}

	s.sendSuccessResponse(w, response)
}

// 获取系统信息处理器
func (s *HTTPServer) handleGetSystemInfo(w http.ResponseWriter, r *http.Request) {
	sysInfo, err := s.SystemService.GetSystemInfo()
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to get system info", err.Error())
		return
	}

	metrics, err := s.SystemService.GetMetrics()
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to get metrics", err.Error())
		return
	}

	health, err := s.SystemService.GetHealthStatus()
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to get health status", err.Error())
		return
	}

	response := map[string]interface{}{
		"version":            sysInfo.Version,
		"status":             health.Status,
		"uptime":             metrics.Uptime.String(),
		"active_connections": metrics.Goroutines,
		"memory": map[string]interface{}{
			"used":      metrics.Memory.Alloc,
			"total":     metrics.Memory.Sys,
			"available": metrics.Memory.Sys - metrics.Memory.Alloc,
		},
	}

	s.sendSuccessResponse(w, response)
}

// 健康检查处理器
func (s *HTTPServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "healthy",
		"message": "Service is running",
		"time":    time.Now().Format(time.RFC3339),
	}

	s.sendSuccessResponse(w, response)
}

// 图片服务处理器
func (s *HTTPServer) handleServeImage(w http.ResponseWriter, r *http.Request) {
	// 从URL路径中提取图片路径
	imagePath := r.URL.Path[len("/api/images/"):]

	// 使用 PathManager 获取图片目录
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		http.Error(w, "Path manager not initialized", http.StatusInternalServerError)
		return
	}

	// 构建完整文件路径
	fullPath := filepath.Join(pathManager.GetImagesDir(), imagePath)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		log.Printf("Image not found: %s", fullPath)
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// 设置正确的Content-Type
	ext := filepath.Ext(fullPath)
	switch ext {
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// 设置缓存头
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// 提供文件
	http.ServeFile(w, r, fullPath)
}

// 文件树处理器
func (s *HTTPServer) handleGetFileTree(w http.ResponseWriter, r *http.Request) {
	// 使用 PathManager 获取图片目录
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Path manager not initialized", "")
		return
	}

	imagesDir := pathManager.GetImagesDir()

	// 构建文件树结构
	fileTree := buildFileTree(imagesDir)

	response := map[string]interface{}{
		"fileTree": fileTree,
	}

	s.sendSuccessResponse(w, response)
}

// 构建文件树结构
func buildFileTree(rootPath string) map[string]interface{} {
	// 创建根节点
	root := map[string]interface{}{
		"key":      "images",
		"title":    "images",
		"isFolder": true,
		"children": []map[string]interface{}{},
	}

	// 读取图片目录
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return root
	}

	children := []map[string]interface{}{}

	for _, entry := range entries {
		if entry.IsDir() {
			// 处理任务文件夹（现在直接是 task_{taskID} 格式）
			taskPath := filepath.Join(rootPath, entry.Name())
			taskChildren := []map[string]interface{}{}

			// 读取任务文件夹中的图片文件
			taskEntries, err := os.ReadDir(taskPath)
			if err == nil {
				for _, taskEntry := range taskEntries {
					if !taskEntry.IsDir() {
						// 获取文件信息
						filePath := filepath.Join(taskPath, taskEntry.Name())
						fileInfo, err := os.Stat(filePath)
						if err == nil {
							// 构建文件节点
							fileNode := map[string]interface{}{
								"key":      entry.Name() + "_" + taskEntry.Name(),
								"title":    taskEntry.Name(),
								"isFolder": false,
								"filePath": filepath.Join(entry.Name(), taskEntry.Name()),
								"fileSize": fileInfo.Size(),
								"fileType": getFileMimeType(filepath.Ext(taskEntry.Name())),
							}
							taskChildren = append(taskChildren, fileNode)
						}
					}
				}
			}

			// 构建任务文件夹节点
			taskNode := map[string]interface{}{
				"key":      entry.Name(),
				"title":    entry.Name() + fmt.Sprintf(" (%d张)", len(taskChildren)),
				"isFolder": true,
				"children": taskChildren,
			}
			children = append(children, taskNode)
		}
	}

	root["children"] = children
	return root
}

// 获取文件MIME类型
func getFileMimeType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

// WebSocket 处理器
func (s *HTTPServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer func() {
		// 确保连接被正确关闭
		if err := conn.Close(); err != nil {
			log.Printf("WebSocket close error: %v", err)
		}
	}()

	// 注册客户端
	s.clientsMutex.Lock()
	s.clients[conn] = true
	s.clientsMutex.Unlock()

	// 发送欢迎消息
	welcomeMsg := WSMessage{
		Type:    "welcome",
		Message: "Connected to PixivTailor WebSocket server",
	}
	if err := conn.WriteJSON(welcomeMsg); err != nil {
		log.Printf("WebSocket welcome message send failed: %v", err)
		return
	}

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// 处理消息
	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket unexpected close error: %v", err)
			} else {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// 更新读取超时
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		// 处理心跳
		if msg.Type == "ping" {
			if err := conn.WriteJSON(WSMessage{Type: "pong"}); err != nil {
				log.Printf("WebSocket pong send failed: %v", err)
				break
			}
			continue
		}

		// 处理其他消息
		s.handleWebSocketMessage(conn, msg)
	}

	// 注销客户端
	s.clientsMutex.Lock()
	delete(s.clients, conn)
	s.clientsMutex.Unlock()
}

// 处理 WebSocket 消息
func (s *HTTPServer) handleWebSocketMessage(conn *websocket.Conn, msg WSMessage) {
	switch msg.Type {
	case "subscribe_task":
		// 订阅任务更新
		log.Printf("Client subscribed to task: %s", msg.TaskID)
	case "unsubscribe_task":
		// 取消订阅任务更新
		log.Printf("Client unsubscribed from task: %s", msg.TaskID)
	default:
		log.Printf("Unknown WebSocket message type: %s", msg.Type)
	}
}

// 设置 WebSocket 广播
func (s *HTTPServer) setupWebSocket() {
	go func() {
		for message := range s.broadcast {
			// 广播消息给所有客户端
			s.clientsMutex.RLock()
			clients := make([]*websocket.Conn, 0, len(s.clients))
			for client := range s.clients {
				clients = append(clients, client)
			}
			s.clientsMutex.RUnlock()

			// 向所有客户端发送消息
			for _, client := range clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("WebSocket write error: %v", err)
					client.Close()
					s.clientsMutex.Lock()
					delete(s.clients, client)
					s.clientsMutex.Unlock()
				}
			}
		}
	}()
}

// 广播日志消息
func (s *HTTPServer) broadcastLogMessage(taskID, level, message string) {
	msg := WSMessage{
		Type:   "log_message",
		TaskID: taskID,
		Data: map[string]interface{}{
			"task_id": taskID,
			"level":   level,
			"message": message,
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	msgBytes, _ := json.Marshal(msg)
	s.broadcast <- msgBytes
}

// 广播任务状态更新
func (s *HTTPServer) broadcastTaskUpdate(taskID, status string, progress int) {
	// 获取任务详细信息，包括图片数量
	task, err := s.TaskService.GetTask(taskID)
	if err != nil {
		log.Printf("获取任务详情失败: %v", err)
		// 如果获取失败，使用默认值
		task = &repository.Task{
			ID:               taskID,
			Status:           status,
			Progress:         progress,
			ImagesFound:      0,
			ImagesDownloaded: 0,
		}
	}

	// 解析任务结果（包含图片URL）
	var result map[string]interface{}
	if task.Result != "" {
		if err := json.Unmarshal([]byte(task.Result), &result); err != nil {
			log.Printf("解析任务结果失败: %v", err)
			result = make(map[string]interface{})
		}
	} else {
		result = make(map[string]interface{})
	}

	msg := WSMessage{
		Type:   "task_update",
		TaskID: taskID,
		Data: map[string]interface{}{
			"task_id":           taskID,
			"status":            status,
			"progress":          progress,
			"images_found":      task.ImagesFound,
			"images_downloaded": task.ImagesDownloaded,
			"images_generated":  task.ImagesFound,      // 前端期望的字段名
			"images_success":    task.ImagesDownloaded, // 前端期望的字段名
			"result":            result,                // 包含图片URL
			"time":              time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	msgBytes, _ := json.Marshal(msg)
	s.broadcast <- msgBytes
}

// 广播全局日志消息
func (s *HTTPServer) broadcastGlobalLog(level, message string) {
	msg := WSMessage{
		Type: "global_log",
		Data: map[string]interface{}{
			"level":   level,
			"message": message,
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	msgBytes, _ := json.Marshal(msg)
	s.broadcast <- msgBytes
}

// 发送成功响应
func (s *HTTPServer) sendSuccessResponse(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Status: Status{
			Code:    0,
			Message: "Success",
		},
		Data: data,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 发送错误响应
func (s *HTTPServer) sendErrorResponse(w http.ResponseWriter, statusCode int, message, details string) {
	response := APIResponse{
		Status: Status{
			Code:    1,
			Message: message,
			Details: details,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// 创建爬虫任务处理器
func (s *HTTPServer) handleCreateCrawlTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type         string `json:"type"`
		Query        string `json:"query"`
		UserID       *int   `json:"user_id,omitempty"`
		IllustID     *int   `json:"illust_id,omitempty"`
		Order        string `json:"order"`
		Mode         string `json:"mode"`
		Limit        int    `json:"limit"`
		Delay        int    `json:"delay"`
		ProxyEnabled *bool  `json:"proxy_enabled,omitempty"`
		ProxyURL     string `json:"proxy_url,omitempty"`
		Cookie       string `json:"cookie,omitempty"`
	}

	// 先读取原始请求体进行调试
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Failed to read request body", err.Error())
		return
	}

	log.Printf("收到爬取任务请求: %s", string(bodyBytes))

	// 解析JSON
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		log.Printf("JSON解析失败: %v", err)
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	log.Printf("解析后的请求: %+v", req)

	// 验证请求参数
	if req.Type == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "Type is required", "")
		return
	}

	// 根据类型验证必要参数
	if req.Type == "tag" && req.Query == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "Query is required for tag crawling", "")
		return
	}
	if req.Type == "user" && req.UserID == nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "UserID is required for user crawling", "")
		return
	}
	if req.Type == "illust" && req.IllustID == nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "IllustID is required for illust crawling", "")
		return
	}

	// 设置默认值
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Delay <= 0 {
		req.Delay = 1
	}

	// 创建任务配置
	config := map[string]interface{}{
		"type":  req.Type,
		"order": req.Order,
		"mode":  req.Mode,
		"limit": req.Limit,
		"delay": req.Delay,
	}

	// 添加Cookie配置
	if req.Cookie != "" {
		config["cookie"] = req.Cookie
	}

	// 根据类型设置相应的查询字段
	if req.Type == "tag" {
		config["query"] = req.Query
	} else if req.Type == "user" {
		config["user_id"] = *req.UserID
		config["query"] = fmt.Sprintf("%d", *req.UserID) // 为了兼容性，也设置query字段
	} else if req.Type == "illust" {
		config["illust_id"] = *req.IllustID
		config["query"] = fmt.Sprintf("%d", *req.IllustID) // 为了兼容性，也设置query字段
	}
	if req.ProxyEnabled != nil {
		config["proxy_enabled"] = *req.ProxyEnabled
	}
	if req.ProxyURL != "" {
		config["proxy_url"] = req.ProxyURL
	}

	configJSON, _ := json.Marshal(config)

	// 创建任务
	log.Printf("HTTP: 准备创建爬虫任务, 配置: %s", string(configJSON))
	task, err := s.TaskService.CreateTask("crawl", string(configJSON))
	if err != nil {
		log.Printf("HTTP: 创建任务失败: %v", err)
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to create crawl task", err.Error())
		return
	}

	// 发送全局日志
	s.broadcastGlobalLog("info", fmt.Sprintf("新任务已创建: %s (类型: %s)", task.ID, req.Type))

	log.Printf("HTTP: 任务创建成功, ID: %s, 状态: %s", task.ID, task.Status)
	s.sendSuccessResponse(w, task)
}

// handleStartTask 手动启动任务
func (s *HTTPServer) handleStartTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TaskID string `json:"task_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.TaskID == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "Task ID is required", "")
		return
	}

	taskID := req.TaskID

	log.Printf("HTTP: 手动启动任务: %s", taskID)
	err := s.TaskService.StartTask(taskID)
	if err != nil {
		log.Printf("HTTP: 启动任务失败: %v", err)
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to start task", err.Error())
		return
	}

	log.Printf("HTTP: 任务启动成功: %s", taskID)
	s.sendSuccessResponse(w, map[string]interface{}{
		"message": "Task started successfully",
		"task_id": taskID,
	})
}

// handleStopTask 停止任务处理器
func (s *HTTPServer) handleStopTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TaskID string `json:"task_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.TaskID == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "Task ID is required", "")
		return
	}

	taskID := req.TaskID

	// 获取任务信息
	task, err := s.TaskService.GetTask(taskID)
	if err != nil {
		s.sendErrorResponse(w, http.StatusNotFound, "Task not found", err.Error())
		return
	}

	log.Printf("HTTP: 手动停止任务: %s (类型: %s)", taskID, task.Type)

	// 如果是AI生成任务，先停止WebUI的当前生成
	if task.Type == "generate" {
		aiHandler := NewAIHandler(nil, s.TaskService, s.GenerationConfigService)
		if err := aiHandler.stopWebUIGeneration(); err != nil {
			log.Printf("HTTP: 停止WebUI生成失败: %v", err)
			// 即使WebUI停止失败，也继续停止任务
		} else {
			log.Printf("HTTP: WebUI生成已停止")
		}
	}

	// 停止任务
	err = s.TaskService.StopTask(taskID)
	if err != nil {
		log.Printf("HTTP: 停止任务失败: %v", err)
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to stop task", err.Error())
		return
	}

	log.Printf("HTTP: 任务停止成功: %s", taskID)
	s.sendSuccessResponse(w, map[string]interface{}{
		"message": "Task stopped successfully",
		"task_id": taskID,
	})
}

// handleCleanupTasks 清理任务处理器
func (s *HTTPServer) handleCleanupTasks(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CleanupType string `json:"cleanup_type"` // "completed", "failed", "all"
	}

	// 添加调试日志
	log.Printf("收到清理任务请求，Content-Type: %s", r.Header.Get("Content-Type"))

	// 读取原始请求体
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("读取请求体失败: %v", err)
		s.sendErrorResponse(w, http.StatusBadRequest, "Failed to read request body", err.Error())
		return
	}
	log.Printf("原始请求体: %s", string(bodyBytes))

	// 解析JSON
	if err = json.Unmarshal(bodyBytes, &req); err != nil {
		log.Printf("解析请求体失败: %v", err)
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	log.Printf("解析后的清理类型: '%s'", req.CleanupType)

	if req.CleanupType == "" {
		log.Printf("清理类型为空")
		s.sendErrorResponse(w, http.StatusBadRequest, "Cleanup type is required", "")
		return
	}

	log.Printf("HTTP: 开始清理任务, 类型: %s", req.CleanupType)

	// 根据清理类型执行清理
	var cleanedCount int

	switch req.CleanupType {
	case "completed":
		cleanedCount, err = s.TaskService.CleanupTasks("completed")
	case "failed":
		cleanedCount, err = s.TaskService.CleanupTasks("failed")
	case "all":
		cleanedCount, err = s.TaskService.CleanupTasks("all")
	default:
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid cleanup type", "Must be 'completed', 'failed', or 'all'")
		return
	}

	if err != nil {
		log.Printf("HTTP: 清理任务失败: %v", err)
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to cleanup tasks", err.Error())
		return
	}

	log.Printf("HTTP: 任务清理成功, 清理了 %d 个任务", cleanedCount)
	s.sendSuccessResponse(w, map[string]interface{}{
		"message":       "Tasks cleaned up successfully",
		"cleaned_count": cleanedCount,
		"cleanup_type":  req.CleanupType,
	})
}

// handleLoadConfigsFromFiles 从文件系统加载配置文件
func (s *HTTPServer) handleLoadConfigsFromFiles(w http.ResponseWriter, r *http.Request) {
	// 使用 PathManager 获取配置目录
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Path manager not initialized", "")
		return
	}

	configsDir := filepath.Join(pathManager.GetDataDir(), "configs")

	// 读取配置目录中的所有JSON文件
	files, err := filepath.Glob(filepath.Join(configsDir, "*.json"))
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to read config directory", err.Error())
		return
	}

	var loadedConfigs []map[string]interface{}
	var errors []string

	for _, file := range files {
		// 读取文件内容
		content, err := ioutil.ReadFile(file)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to read %s: %v", file, err))
			continue
		}

		// 解析JSON
		var config map[string]interface{}
		if err := json.Unmarshal(content, &config); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to parse %s: %v", file, err))
			continue
		}

		// 直接返回文件配置，不存储到数据库
		loadedConfigs = append(loadedConfigs, map[string]interface{}{
			"id":   getStringFromMap(config, "id"),
			"name": getStringFromMap(config, "name"),
			"file": filepath.Base(file),
		})
	}

	response := map[string]interface{}{
		"loaded_configs": loadedConfigs,
		"total_loaded":   len(loadedConfigs),
		"errors":         errors,
	}

	s.sendSuccessResponse(w, response)
}

// handleStartWebUI 启动WebUI
func (s *HTTPServer) handleStartWebUI(w http.ResponseWriter, r *http.Request) {
	// 检查WebUI是否已经在运行
	if s.isWebUIRunning() {
		s.sendErrorResponse(w, http.StatusConflict, "WebUI is already running", "")
		return
	}

	// 启动WebUI进程
	go s.startWebUIProcess()

	s.sendSuccessResponse(w, map[string]interface{}{
		"message": "WebUI启动命令已执行",
		"status":  "starting",
	})
}

// handleStartWebUIExternal 启动外部WebUI（运行批处理脚本）
func (s *HTTPServer) handleStartWebUIExternal(w http.ResponseWriter, r *http.Request) {
	// 检查WebUI是否已经在运行
	status := s.getWebUIStatus()
	if status["status"] == "running" || status["status"] == "external" {
		s.sendErrorResponse(w, http.StatusConflict, "WebUI is already running", "")
		return
	}

	// 运行批处理脚本
	cmd := exec.Command("cmd", "/c", "start", "cmd", "/k", "start-webui-api.bat")
	cmd.Dir = ".." // 在项目根目录运行（从backend目录向上一级）

	if err := cmd.Start(); err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to start WebUI", err.Error())
		return
	}

	s.sendSuccessResponse(w, map[string]interface{}{
		"message": "WebUI批处理脚本已启动，请查看弹出的命令行窗口",
		"status":  "starting",
	})
}

// handleStopWebUI 停止WebUI
func (s *HTTPServer) handleStopWebUI(w http.ResponseWriter, r *http.Request) {
	if !s.isWebUIRunning() {
		// WebUI没有运行，返回成功状态而不是错误
		s.sendSuccessResponse(w, map[string]interface{}{
			"message": "WebUI未运行",
			"status":  "stopped",
		})
		return
	}

	// 先调用WebUI的interrupt API停止所有正在进行的生成任务
	s.broadcastWebUILog("正在停止WebUI的生成任务...")
	if err := s.stopWebUIGeneration(); err != nil {
		s.broadcastWebUILog(fmt.Sprintf("停止WebUI生成任务失败: %v", err))
		// 即使停止生成失败，也继续停止WebUI进程
	} else {
		s.broadcastWebUILog("WebUI生成任务已停止")
	}

	// 等待一段时间让WebUI完成当前操作
	time.Sleep(2 * time.Second)

	// 停止WebUI进程
	s.stopWebUIProcess()

	s.sendSuccessResponse(w, map[string]interface{}{
		"message": "WebUI停止命令已执行",
		"status":  "stopped",
	})
}

// handleWebUIStatus 获取WebUI状态
func (s *HTTPServer) handleWebUIStatus(w http.ResponseWriter, r *http.Request) {
	status := s.getWebUIStatus()
	s.sendSuccessResponse(w, status)
}

// handleWebUILogs 获取WebUI日志
func (s *HTTPServer) handleWebUILogs(w http.ResponseWriter, r *http.Request) {
	// 处理OPTIONS请求
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")
		w.WriteHeader(http.StatusOK)
		return
	}

	// 设置SSE头
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// 创建日志流
	logChan := make(chan string, 100)
	s.webUILogMutex.Lock()
	s.webUILogChannels = append(s.webUILogChannels, logChan)
	s.webUILogMutex.Unlock()

	// 发送初始状态
	fmt.Fprintf(w, "data: %s\n\n", "WebUI日志流已连接")
	if f, ok := w.(http.Flusher); ok && f != nil {
		// 使用defer recover来捕获任何panic
		defer func() {
			if r := recover(); r != nil {
				log.Printf("WebUI日志初始刷新时发生panic: %v", r)
			}
		}()
		f.Flush()
	}

	// 发送当前WebUI状态
	status := s.getWebUIStatus()

	// 简化的日志流 - 只发送状态更新
	go func() {
		defer func() {
			// 清理通道
			s.webUILogMutex.Lock()
			for i, ch := range s.webUILogChannels {
				if ch == logChan {
					s.webUILogChannels = append(s.webUILogChannels[:i], s.webUILogChannels[i+1:]...)
					break
				}
			}
			s.webUILogMutex.Unlock()
			close(logChan)
		}()

		// 监听客户端断开连接
		ctx := r.Context()
		done := ctx.Done()

		// 发送状态更新
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		// 用于安全写入的辅助函数
		safeWrite := func(data string) bool {
			// 检查连接是否仍然有效
			if ctx.Err() != nil {
				return false
			}

			// 检查 ResponseWriter 是否为 nil
			if w == nil {
				return false
			}

			// 尝试写入数据
			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				log.Printf("WebUI日志写入失败: %v", err)
				return false
			}

			// 尝试刷新 - 添加更严格的nil检查
			if f, ok := w.(http.Flusher); ok && f != nil {
				// 使用defer recover来捕获任何panic
				defer func() {
					if r := recover(); r != nil {
						log.Printf("WebUI日志刷新时发生panic: %v", r)
					}
				}()
				f.Flush()
			}
			return true
		}

		// 发送初始状态
		statusMsg := fmt.Sprintf("WebUI状态: %s (端口开放: %v, API响应: %v)",
			status["status"], status["port_open"], status["api_responding"])
		if !safeWrite(statusMsg) {
			return
		}

		for {
			select {
			case log, ok := <-logChan:
				if !ok {
					return
				}
				if !safeWrite(log) {
					return
				}
			case <-ticker.C:
				// 发送状态更新
				currentStatus := s.getWebUIStatus()
				statusMsg := fmt.Sprintf("WebUI状态: %s (端口开放: %v, API响应: %v)",
					currentStatus["status"], currentStatus["port_open"], currentStatus["api_responding"])
				if !safeWrite(statusMsg) {
					return
				}
			case <-done:
				return
			}
		}
	}()
}

// startWebUIProcess 启动WebUI进程
func (s *HTTPServer) startWebUIProcess() {
	s.webUIStatus = "starting"
	s.broadcastWebUILog("开始启动WebUI进程...")

	// 获取WebUI路径
	pathManager := paths.GetPathManager()
	webuiDir := pathManager.GetWebUIDir()
	webuiBat := pathManager.GetWebUIBat()

	s.broadcastWebUILog(fmt.Sprintf("WebUI目录: %s", webuiDir))
	s.broadcastWebUILog(fmt.Sprintf("WebUI批处理文件: %s", webuiBat))

	// 检查WebUI目录和文件是否存在
	if _, err := os.Stat(webuiDir); os.IsNotExist(err) {
		s.broadcastWebUILog(fmt.Sprintf("❌ 错误: WebUI目录不存在: %s", webuiDir))
		s.webUIStatus = "error"
		return
	}

	if _, err := os.Stat(webuiBat); os.IsNotExist(err) {
		s.broadcastWebUILog(fmt.Sprintf("❌ 错误: WebUI批处理文件不存在: %s", webuiBat))
		s.webUIStatus = "error"
		return
	}

	// 根据操作系统选择命令
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows: 直接运行webui.bat
		cmd = exec.Command("cmd", "/c", webuiBat, "--api", "--listen", "--port", "7860", "--skip-python-version-check")
		cmd.Dir = webuiDir
		s.broadcastWebUILog("使用Windows直接运行webui.bat")
	} else {
		// Linux/Mac: 运行webui.sh
		webuiSh := filepath.Join(webuiDir, "webui.sh")
		cmd = exec.Command("bash", webuiSh, "--api", "--listen", "--port", "7860", "--skip-python-version-check")
		cmd.Dir = webuiDir
		s.broadcastWebUILog("使用Linux/Mac直接运行webui.sh")
	}

	// 设置环境变量
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	// 设置输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		s.broadcastWebUILog(fmt.Sprintf("❌ 创建stdout管道失败: %v", err))
		s.webUIStatus = "error"
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		s.broadcastWebUILog(fmt.Sprintf("❌ 创建stderr管道失败: %v", err))
		s.webUIStatus = "error"
		return
	}

	s.broadcastWebUILog("✅ 输出管道创建成功")

	// 启动进程
	s.broadcastWebUILog("正在启动WebUI进程...")
	if err := cmd.Start(); err != nil {
		s.broadcastWebUILog(fmt.Sprintf("❌ 启动WebUI失败: %v", err))
		s.webUIStatus = "error"
		return
	}

	s.webUIProcess = cmd.Process
	s.webUIStatus = "running"
	s.broadcastWebUILog("✅ WebUI进程启动成功")
	s.broadcastWebUILog(fmt.Sprintf("📋 进程ID: %d", cmd.Process.Pid))
	s.broadcastWebUILog("⏳ 等待WebUI输出...")

	// 启动心跳日志
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if s.webUIStatus == "running" {
					s.broadcastWebUILog("WebUI heartbeat - process is running")
				} else {
					return
				}
			}
		}
	}()

	// 读取输出
	go func() {
		defer stdout.Close()
		s.broadcastWebUILog("📖 开始读取STDOUT...")
		scanner := bufio.NewScanner(stdout)
		lineCount := 0
		for scanner.Scan() {
			line := scanner.Text()
			lineCount++
			if line != "" {
				s.broadcastWebUILog(fmt.Sprintf("📤 [%d] %s", lineCount, line))
			}
		}
		if err := scanner.Err(); err != nil {
			s.broadcastWebUILog(fmt.Sprintf("❌ STDOUT扫描错误: %v", err))
		}
		s.broadcastWebUILog(fmt.Sprintf("📤 STDOUT流已关闭 (共读取 %d 行)", lineCount))
	}()

	// 如果WebUI已经运行，发送一些测试日志
	go func() {
		time.Sleep(2 * time.Second)
		if s.webUIStatus == "running" {
			s.broadcastWebUILog("🔍 正在检查WebUI输出...")
			s.broadcastWebUILog("💡 如果WebUI没有输出，可能是缓冲问题")
		}
	}()

	go func() {
		defer stderr.Close()
		s.broadcastWebUILog("📖 开始读取STDERR...")
		scanner := bufio.NewScanner(stderr)
		lineCount := 0
		for scanner.Scan() {
			line := scanner.Text()
			lineCount++
			if line != "" {
				s.broadcastWebUILog(fmt.Sprintf("⚠️ [%d] %s", lineCount, line))
			}
		}
		if err := scanner.Err(); err != nil {
			s.broadcastWebUILog(fmt.Sprintf("❌ STDERR扫描错误: %v", err))
		}
		s.broadcastWebUILog(fmt.Sprintf("⚠️ STDERR流已关闭 (共读取 %d 行)", lineCount))
	}()

	// 等待进程结束
	go func() {
		s.broadcastWebUILog("⏳ 等待WebUI进程结束...")
		err := cmd.Wait()
		s.webUIStatus = "stopped"
		if err != nil {
			s.broadcastWebUILog(fmt.Sprintf("❌ WebUI进程异常退出: %v", err))
		} else {
			s.broadcastWebUILog("✅ WebUI进程正常退出")
		}
		s.webUIProcess = nil
	}()
}

// stopWebUIGeneration 停止WebUI的当前生成任务
func (s *HTTPServer) stopWebUIGeneration() error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 调用WebUI的interrupt API
	resp, err := client.Post("http://127.0.0.1:7860/sdapi/v1/interrupt", "application/json", nil)
	if err != nil {
		return fmt.Errorf("调用WebUI停止API失败: %v", err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WebUI停止API返回错误状态码: %d", resp.StatusCode)
	}

	return nil
}

// stopWebUIProcess 停止WebUI进程
func (s *HTTPServer) stopWebUIProcess() {
	if s.webUIProcess != nil {
		s.webUIStatus = "stopping"
		s.broadcastWebUILog("Stopping WebUI process...")

		// 先尝试优雅关闭
		if err := s.webUIProcess.Signal(os.Interrupt); err != nil {
			s.broadcastWebUILog(fmt.Sprintf("Failed to send interrupt signal: %v", err))
		}

		// 等待进程优雅关闭
		done := make(chan error, 1)
		go func() {
			_, err := s.webUIProcess.Wait()
			done <- err
		}()

		select {
		case <-time.After(5 * time.Second):
			// 超时，强制终止
			s.broadcastWebUILog("WebUI process did not stop gracefully, forcing termination...")
			// 只终止特定的WebUI进程，不终止所有Python进程
			s.webUIProcess.Kill()
		case err := <-done:
			if err != nil {
				s.broadcastWebUILog(fmt.Sprintf("WebUI process exited with error: %v", err))
			} else {
				s.broadcastWebUILog("WebUI process stopped gracefully")
			}
		}

		s.webUIProcess = nil
		s.webUIStatus = "stopped"
		s.broadcastWebUILog("WebUI process stopped")
	}
}

// isWebUIRunning 检查WebUI是否在运行
func (s *HTTPServer) isWebUIRunning() bool {
	// 检查内部管理的WebUI进程
	if s.webUIProcess != nil && s.webUIStatus == "running" {
		return true
	}

	// 检查外部WebUI（通过端口和API响应）
	portOpen := s.checkPort(7860)
	apiResponding := s.checkWebUIAPI()

	return portOpen && apiResponding
}

// getWebUIStatus 获取WebUI状态
func (s *HTTPServer) getWebUIStatus() map[string]interface{} {
	// 检查端口是否被占用
	portOpen := s.checkPort(7860)

	// 检查 WebUI API 是否响应
	apiResponding := s.checkWebUIAPI()

	// 确定实际状态
	var actualStatus string
	if s.webUIProcess != nil && portOpen && apiResponding {
		actualStatus = "running"
	} else if s.webUIProcess != nil && !portOpen {
		actualStatus = "starting"
	} else if portOpen && apiResponding {
		actualStatus = "external" // 外部启动的 WebUI
	} else {
		actualStatus = "stopped"
	}

	return map[string]interface{}{
		"status":         actualStatus,
		"port_open":      portOpen,
		"api_responding": apiResponding,
		"process_id":     s.webUIProcess != nil,
		"managed":        s.webUIProcess != nil, // 是否由后端管理
	}
}

// checkPort 检查端口是否被占用
func (s *HTTPServer) checkPort(port int) bool {
	// 简单的端口检查实现
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// checkWebUIAPI 检查 WebUI API 是否响应
func (s *HTTPServer) checkWebUIAPI() bool {
	// 尝试访问 WebUI API 的状态端点
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get("http://127.0.0.1:7860/sdapi/v1/options")
	if err != nil {
		return false
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	return resp.StatusCode == http.StatusOK
}

// broadcastWebUILog 广播WebUI日志
func (s *HTTPServer) broadcastWebUILog(message string) {
	timestamp := time.Now().Format("15:04:05")
	logMessage := fmt.Sprintf("[%s] %s", timestamp, message)

	// 发送到所有日志通道
	s.webUILogMutex.RLock()
	channels := make([]chan string, len(s.webUILogChannels))
	copy(channels, s.webUILogChannels)
	s.webUILogMutex.RUnlock()

	for _, ch := range channels {
		select {
		case ch <- logMessage:
			// 成功发送
		default:
			// 通道已满，移除该通道
			s.webUILogMutex.Lock()
			for j, originalCh := range s.webUILogChannels {
				if originalCh == ch {
					s.webUILogChannels = append(s.webUILogChannels[:j], s.webUILogChannels[j+1:]...)
					break
				}
			}
			s.webUILogMutex.Unlock()
			close(ch)
		}
	}
}

// 启动服务器
func (s *HTTPServer) Start(port string) error {
	// 使用统一的logger系统
	fmt.Printf("Starting HTTP server on port %s\n", port)
	fmt.Printf("HTTP server is ready to accept connections\n")

	// 应用CORS中间件
	handler := s.corsMiddleware(s.router)

	// 启动服务器
	err := http.ListenAndServe(":"+port, handler)
	if err != nil {
		fmt.Printf("HTTP server error: %v\n", err)
	}
	return err
}

// StopWebUI 停止WebUI进程
func (s *HTTPServer) StopWebUI() {
	if s.webUIProcess != nil {
		s.webUIStatus = "stopping"
		s.broadcastWebUILog("主程序关闭，正在停止WebUI进程...")

		// 先尝试优雅关闭
		if err := s.webUIProcess.Signal(os.Interrupt); err != nil {
			s.broadcastWebUILog(fmt.Sprintf("Failed to send interrupt signal: %v", err))
		}

		// 等待进程优雅关闭
		done := make(chan error, 1)
		go func() {
			_, err := s.webUIProcess.Wait()
			done <- err
		}()

		select {
		case <-time.After(3 * time.Second):
			// 超时，强制终止
			s.broadcastWebUILog("WebUI process did not stop gracefully, forcing termination...")
			if runtime.GOOS == "windows" {
				// Windows: 使用taskkill强制终止
				exec.Command("taskkill", "/f", "/im", "python.exe").Run()
			}
			s.webUIProcess.Kill()
		case err := <-done:
			if err != nil {
				s.broadcastWebUILog(fmt.Sprintf("WebUI process exited with error: %v", err))
			} else {
				s.broadcastWebUILog("WebUI process stopped gracefully")
			}
		}

		s.webUIProcess = nil
		s.webUIStatus = "stopped"
		s.broadcastWebUILog("WebUI进程已停止")
	}
}
