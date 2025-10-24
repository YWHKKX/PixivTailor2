package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WSMessage WebSocket消息结构
type WSMessage struct {
	Type    string      `json:"type"`
	TaskID  string      `json:"task_id,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// WSManager WebSocket管理器
type WSManager struct {
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
	upgrader  websocket.Upgrader
	mutex     sync.RWMutex
}

// NewWSManager 创建WebSocket管理器
func NewWSManager() *WSManager {
	return &WSManager{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源
			},
		},
	}
}

// HandleWebSocket 处理WebSocket连接
func (ws *WSManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	// 注册客户端
	ws.mutex.Lock()
	ws.clients[conn] = true
	ws.mutex.Unlock()

	log.Printf("WebSocket客户端已连接，当前客户端数量: %d", len(ws.clients))

	// 启动广播循环
	go ws.broadcastLoop()

	// 处理客户端消息
	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket读取错误: %v", err)
			break
		}

		// 处理不同类型的消息
		ws.handleMessage(conn, &msg)
	}

	// 注销客户端
	ws.mutex.Lock()
	delete(ws.clients, conn)
	ws.mutex.Unlock()

	log.Printf("WebSocket客户端已断开，当前客户端数量: %d", len(ws.clients))
}

// broadcastLoop 广播循环
func (ws *WSManager) broadcastLoop() {
	for {
		select {
		case message := <-ws.broadcast:
			ws.mutex.RLock()
			for client := range ws.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("WebSocket发送错误: %v", err)
					client.Close()
					delete(ws.clients, client)
				}
			}
			ws.mutex.RUnlock()
		}
	}
}

// handleMessage 处理WebSocket消息
func (ws *WSManager) handleMessage(conn *websocket.Conn, msg *WSMessage) {
	switch msg.Type {
	case "ping":
		// 响应ping
		response := WSMessage{Type: "pong"}
		conn.WriteJSON(response)
	case "subscribe":
		// 订阅特定任务
		log.Printf("客户端订阅任务: %s", msg.TaskID)
	default:
		log.Printf("未知的WebSocket消息类型: %s", msg.Type)
	}
}

// BroadcastTaskUpdate 广播任务更新
func (ws *WSManager) BroadcastTaskUpdate(taskID, status string, progress int) {
	msg := WSMessage{
		Type:   "task_update",
		TaskID: taskID,
		Data: map[string]interface{}{
			"task_id":  taskID,
			"status":   status,
			"progress": progress,
			"time":     getCurrentTime(),
		},
	}
	ws.broadcastMessage(msg)
}

// BroadcastLogMessage 广播日志消息
func (ws *WSManager) BroadcastLogMessage(taskID, level, message string) {
	msg := WSMessage{
		Type:    "log",
		TaskID:  taskID,
		Message: message,
		Data: map[string]interface{}{
			"level": level,
			"time":  getCurrentTime(),
		},
	}
	ws.broadcastMessage(msg)
}

// BroadcastGlobalLog 广播全局日志
func (ws *WSManager) BroadcastGlobalLog(level, message string) {
	msg := WSMessage{
		Type:    "global_log",
		Message: message,
		Data: map[string]interface{}{
			"level": level,
			"time":  getCurrentTime(),
		},
	}
	ws.broadcastMessage(msg)
}

// broadcastMessage 广播消息
func (ws *WSManager) broadcastMessage(msg WSMessage) {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("序列化WebSocket消息失败: %v", err)
		return
	}

	ws.mutex.RLock()
	clientCount := len(ws.clients)
	ws.mutex.RUnlock()

	if clientCount > 0 {
		ws.broadcast <- msgBytes
	}
}

// GetClientCount 获取客户端数量
func (ws *WSManager) GetClientCount() int {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	return len(ws.clients)
}

// getCurrentTime 获取当前时间字符串
func getCurrentTime() string {
	return "2024-01-01 12:00:00" // 这里应该使用实际的时间格式化
}
