package api

import (
	"encoding/json"
	"net/http"
)

// APIResponse 统一API响应结构
type APIResponse struct {
	Status Status      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}

// Status 响应状态
type Status struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ResponseHandler API响应处理器
type ResponseHandler struct{}

// NewResponseHandler 创建响应处理器
func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

// SendSuccessResponse 发送成功响应
func (h *ResponseHandler) SendSuccessResponse(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Status: Status{
			Code:    0,
			Message: "Success",
		},
		Data: data,
	}
	h.sendJSONResponse(w, http.StatusOK, response)
}

// SendErrorResponse 发送错误响应
func (h *ResponseHandler) SendErrorResponse(w http.ResponseWriter, statusCode int, message, details string) {
	response := APIResponse{
		Status: Status{
			Code:    1,
			Message: message,
			Details: details,
		},
	}
	h.sendJSONResponse(w, statusCode, response)
}

// SendBadRequestResponse 发送400错误响应
func (h *ResponseHandler) SendBadRequestResponse(w http.ResponseWriter, message, details string) {
	h.SendErrorResponse(w, http.StatusBadRequest, message, details)
}

// SendInternalServerErrorResponse 发送500错误响应
func (h *ResponseHandler) SendInternalServerErrorResponse(w http.ResponseWriter, message, details string) {
	h.SendErrorResponse(w, http.StatusInternalServerError, message, details)
}

// sendJSONResponse 发送JSON响应
func (h *ResponseHandler) sendJSONResponse(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// ParseJSONRequest 解析JSON请求
func (h *ResponseHandler) ParseJSONRequest(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// SetCORSHeaders 设置CORS头
func (h *ResponseHandler) SetCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// HandleOptionsRequest 处理OPTIONS请求
func (h *ResponseHandler) HandleOptionsRequest(w http.ResponseWriter, r *http.Request) {
	h.SetCORSHeaders(w)
	w.WriteHeader(http.StatusOK)
}
