package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"pixiv-tailor/backend/internal/service"
)

// CrawlerHandler 爬虫任务处理器
type CrawlerHandler struct {
	taskService service.TaskService
}

// NewCrawlerHandler 创建爬虫处理器
func NewCrawlerHandler(taskService service.TaskService) *CrawlerHandler {
	return &CrawlerHandler{
		taskService: taskService,
	}
}

// CrawlRequest 爬虫请求
type CrawlRequest struct {
	Query        string                 `json:"query,omitempty"`
	UserID       int                    `json:"user_id,omitempty"`
	IllustID     int                    `json:"illust_id,omitempty"`
	Order        string                 `json:"order,omitempty"`
	Mode         string                 `json:"mode,omitempty"`
	Limit        int                    `json:"limit,omitempty"`
	Delay        int                    `json:"delay,omitempty"`
	ProxyEnabled bool                   `json:"proxy_enabled,omitempty"`
	ProxyURL     string                 `json:"proxy_url,omitempty"`
	Cookie       string                 `json:"cookie,omitempty"`
	CrawlType    string                 `json:"crawl_type"` // tag, user, illust
	Override     map[string]interface{} `json:"override,omitempty"`
}

// HandleCrawl 处理爬虫任务
func (h *CrawlerHandler) HandleCrawl(w http.ResponseWriter, r *http.Request) {
	var req CrawlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// 构建任务配置
	taskConfig, err := h.buildCrawlTaskConfig(&req)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid crawl request", err.Error())
		return
	}

	// 创建任务记录（task_service 会自动处理执行）
	configJSON, _ := json.Marshal(taskConfig)
	task, err := h.taskService.CreateTask("crawl", string(configJSON))
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Failed to create task", err.Error())
		return
	}

	h.sendSuccessResponse(w, map[string]interface{}{
		"task_id": task.ID,
		"message": "爬虫任务已创建",
		"status":  task.Status,
	})
}

// buildCrawlTaskConfig 构建爬虫任务配置
func (h *CrawlerHandler) buildCrawlTaskConfig(req *CrawlRequest) (map[string]interface{}, error) {
	config := make(map[string]interface{})

	// 根据爬取类型构建配置
	switch req.CrawlType {
	case "tag":
		config["type"] = "tag"
		config["query"] = req.Query
		config["order"] = req.Order
		config["mode"] = req.Mode
		config["limit"] = req.Limit
	case "user":
		config["type"] = "user"
		config["user_id"] = req.UserID
		config["limit"] = req.Limit
	case "illust":
		config["type"] = "illust"
		config["illust_id"] = req.IllustID
	default:
		return nil, fmt.Errorf("不支持的爬取类型: %s", req.CrawlType)
	}

	// 公共配置
	config["delay"] = req.Delay
	config["proxy_enabled"] = req.ProxyEnabled
	if req.ProxyURL != "" {
		config["proxy_url"] = req.ProxyURL
	}
	if req.Cookie != "" {
		config["cookie"] = req.Cookie
	}

	// 应用覆盖参数
	for k, v := range req.Override {
		config[k] = v
	}

	return config, nil
}

// sendSuccessResponse 发送成功响应
func (h *CrawlerHandler) sendSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": map[string]interface{}{
			"code":    0,
			"message": "success",
		},
		"data": data,
	})
}

// sendErrorResponse 发送错误响应
func (h *CrawlerHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": map[string]interface{}{
			"code":    statusCode,
			"message": message,
			"details": details,
		},
	})
}
