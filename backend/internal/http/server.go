package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"pixiv-tailor/backend/internal/repository"
	"pixiv-tailor/backend/internal/service"
	"pixiv-tailor/backend/pkg/paths"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// HTTP 服务器 - 基于 plan.md 设计
type HTTPServer struct {
	TaskService   service.TaskService
	ConfigService service.ConfigService
	DataService   service.DataService
	SystemService service.SystemService
	router        *mux.Router
	upgrader      websocket.Upgrader
	clients       map[*websocket.Conn]bool
	broadcast     chan []byte
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
	TaskID      string     `json:"task_id"`
	Status      string     `json:"status"`
	Progress    int        `json:"progress"`
	Message     string     `json:"message"`
	Result      []string   `json:"result,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
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
) *HTTPServer {
	server := &HTTPServer{
		TaskService:   taskService,
		ConfigService: configService,
		DataService:   dataService,
		SystemService: systemService,
		router:        mux.NewRouter(),
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

	// 任务管理
	api.HandleFunc("/status", s.handleGetTaskStatus).Methods("POST", "OPTIONS")
	api.HandleFunc("/cancel", s.handleCancelTask).Methods("POST", "OPTIONS")
	api.HandleFunc("/tasks", s.handleGetTasks).Methods("POST", "OPTIONS")

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

// 生成图像处理器
func (s *HTTPServer) handleGenerate(w http.ResponseWriter, r *http.Request) {
	var req GenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// 验证请求参数
	if req.Prompt == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "Prompt is required", "")
		return
	}

	// 创建任务配置
	config := map[string]interface{}{
		"prompt":          req.Prompt,
		"negative_prompt": req.NegativePrompt,
		"steps":           req.Steps,
		"cfg_scale":       req.CFGScale,
		"width":           req.Width,
		"height":          req.Height,
		"seed":            req.Seed,
		"model":           req.Model,
		"sampler":         req.Sampler,
		"batch_size":      req.BatchSize,
		"enable_hr":       req.EnableHR,
	}

	configJSON, _ := json.Marshal(config)

	// 创建任务
	task, err := s.TaskService.CreateTask("generate", string(configJSON))
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to create task", err.Error())
		return
	}

	// 构造响应
	response := GenerationResponse{
		TaskID:    task.ID,
		Status:    task.Status,
		Progress:  task.Progress,
		Message:   "任务已创建，正在处理中...",
		CreatedAt: task.CreatedAt,
	}

	s.sendSuccessResponse(w, response)

	// 启动后台处理
	go s.processGenerationTask(task.ID)
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
		TaskID:      task.ID,
		Status:      task.Status,
		Progress:    task.Progress,
		Message:     task.Config,
		CreatedAt:   task.CreatedAt,
		CompletedAt: &task.UpdatedAt,
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

	err := s.TaskService.CancelTask(req.TaskID)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to cancel task", err.Error())
		return
	}

	s.sendSuccessResponse(w, map[string]string{"message": "Task cancelled successfully"})
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

// 获取配置处理器
func (s *HTTPServer) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Module string `json:"module"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	config, err := s.ConfigService.GetConfig(req.Module)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to get config", err.Error())
		return
	}

	s.sendSuccessResponse(w, config)
}

// 更新配置处理器
func (s *HTTPServer) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Module string      `json:"module"`
		Config interface{} `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	configJSON, _ := json.Marshal(req.Config)
	err := s.ConfigService.SetConfig(req.Module, string(configJSON))
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to update config", err.Error())
		return
	}

	s.sendSuccessResponse(w, map[string]string{"message": "Config updated successfully"})
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
	defer conn.Close()

	// 注册客户端
	s.clients[conn] = true
	log.Printf("WebSocket client connected. Total clients: %d", len(s.clients))

	// 发送欢迎消息
	welcomeMsg := WSMessage{
		Type:    "welcome",
		Message: "Connected to PixivTailor WebSocket server",
	}
	conn.WriteJSON(welcomeMsg)
	log.Printf("WebSocket welcome message sent to client\n")

	// 处理消息
	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// 处理心跳
		if msg.Type == "ping" {
			conn.WriteJSON(WSMessage{Type: "pong"})
			continue
		}

		// 处理其他消息
		s.handleWebSocketMessage(conn, msg)
	}

	// 注销客户端
	delete(s.clients, conn)
	log.Printf("WebSocket client disconnected. Total clients: %d", len(s.clients))
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
			for client := range s.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("WebSocket write error: %v", err)
					client.Close()
					delete(s.clients, client)
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
	log.Printf("广播日志消息: %s", string(msgBytes))
	log.Printf("当前WebSocket客户端数量: %d", len(s.clients))
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

	msg := WSMessage{
		Type:   "task_update",
		TaskID: taskID,
		Data: map[string]interface{}{
			"task_id":           taskID,
			"status":            status,
			"progress":          progress,
			"images_found":      task.ImagesFound,
			"images_downloaded": task.ImagesDownloaded,
			"time":              time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	msgBytes, _ := json.Marshal(msg)
	log.Printf("广播任务更新: %s", string(msgBytes))
	log.Printf("当前WebSocket客户端数量: %d", len(s.clients))
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

// 处理生成任务
func (s *HTTPServer) processGenerationTask(taskID string) {
	// 设置任务服务的日志回调，将日志转发到WebSocket客户端
	s.TaskService.SetLogCallback(func(taskID, level, message string) {
		s.broadcastLogMessage(taskID, level, message)
	})

	// 启动任务
	err := s.TaskService.StartTask(taskID)
	if err != nil {
		log.Printf("Failed to start generation task: %v", err)
		s.broadcastLogMessage(taskID, "error", fmt.Sprintf("启动生成任务失败: %v", err))
		return
	}

	log.Printf("Generation task started: %s", taskID)
	s.broadcastLogMessage(taskID, "info", "生成任务已启动")
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

	log.Printf("HTTP: 手动停止任务: %s", taskID)
	err := s.TaskService.StopTask(taskID)
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

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.CleanupType == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "Cleanup type is required", "")
		return
	}

	log.Printf("HTTP: 开始清理任务, 类型: %s", req.CleanupType)

	// 根据清理类型执行清理
	var cleanedCount int
	var err error

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

// 启动服务器
func (s *HTTPServer) Start(port string) error {
	// 使用统一的logger系统
	fmt.Printf("Starting HTTP server on port %s\n", port)
	fmt.Printf("HTTP server is ready to accept connections\n")

	// 启动服务器
	err := http.ListenAndServe(":"+port, s.router)
	if err != nil {
		fmt.Printf("HTTP server error: %v\n", err)
	}
	return err
}
