// WebSocket 连接管理器 - 统一管理 WebSocket 连接和消息处理
import { WS_BASE_URL } from '../config/ports';

// ==================== 类型定义 ====================
export interface WebSocketMessage {
  type: string;
  task_id?: string;
  data?: any;
  timestamp?: number;
}

export interface TaskUpdate {
  task_id: string;
  status: string;
  progress: number;
  message?: string;
  result?: any;
  error?: string;
  images_found?: number;
  images_downloaded?: number;
  images_generated?: number;
  images_success?: number;
  time?: string;
}

export interface ProgressUpdate {
  task_id: string;
  progress: number;
  message: string;
  data?: any;
}

export interface StatusUpdate {
  task_id: string;
  status: string;
  message: string;
}

export interface WebUIStatusUpdate {
  status: string;
  port_open: boolean;
  api_responding: boolean;
  process_id: boolean;
  managed: boolean;
}

export interface GlobalLogMessage {
  level: string;
  message: string;
  timestamp: string;
  source?: string;
}

export interface TaskLogMessage {
  task_id: string;
  level: string;
  message: string;
  timestamp: string;
  source?: string;
}

// ==================== WebSocket 管理器类 ====================
class WebSocketManager {
  private connection: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private heartbeatInterval: NodeJS.Timeout | null = null;
  private heartbeatTimeout: NodeJS.Timeout | null = null;
  private isConnected = false;
  private eventHandlers = new Map<string, Function[]>();
  private url: string = '';
  private isConnecting = false;
  private connectionId: string = '';
  private connectTimer: NodeJS.Timeout | null = null;

  // ==================== 连接管理 ====================
  connect(url: string): void {
    // 清除之前的连接定时器
    if (this.connectTimer) {
      clearTimeout(this.connectTimer);
      this.connectTimer = null;
    }

    // 如果已经在连接中，直接返回
    if (this.isConnecting) {
      console.log('WebSocket 正在连接中，跳过重复连接');
      return;
    }

    // 如果已经连接，先断开
    if (this.connection && this.connection.readyState === WebSocket.OPEN) {
      console.log('WebSocket 已连接，先断开旧连接');
      this.disconnect();
    }

    this.url = url;
    this.isConnecting = true;
    this.connectionId = `ws_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    console.log(`WebSocket 连接尝试 ${this.connectionId}`);

    try {
      this.connection = new WebSocket(url);
      this.setupEventHandlers();
    } catch (error) {
      console.error('WebSocket 连接失败:', error);
      this.isConnecting = false;
      this.scheduleReconnect();
    }
  }

  private setupEventHandlers(): void {
    if (!this.connection) return;

    this.connection.onopen = () => {
      console.log(`WebSocket连接已建立 ${this.connectionId}`);
      this.isConnected = true;
      this.isConnecting = false;
      this.reconnectAttempts = 0;
      this.startHeartbeat();
      this.emit('connected', { connectionId: this.connectionId });
    };

    this.connection.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        this.handleMessage(message);
      } catch (error) {
        console.error('WebSocket 消息解析失败:', error);
      }
    };

    this.connection.onclose = (event) => {
      console.log(`WebSocket连接已关闭 ${this.connectionId}: ${event.code}`);
      this.isConnected = false;
      this.isConnecting = false;
      this.stopHeartbeat();
      this.emit('disconnected', { code: event.code, reason: event.reason });

      // 如果不是正常关闭，启动重连
      if (event.code !== 1000) {
        console.log('WebSocket异常关闭，启动重连机制');
        this.scheduleReconnect();
      }
    };

    this.connection.onerror = (error) => {
      console.error('WebSocket错误:', error);
      this.isConnecting = false;
      this.emit('error', error);
    };
  }

  private handleMessage(message: WebSocketMessage): void {
    console.log('WebSocket 收到消息:', message);

    switch (message.type) {
      case 'welcome':
        console.log('WebSocket 欢迎消息:', message.message || message.data);
        this.emit('welcome', message);
        break;
      case 'task_update':
        this.emit('task_update', message.data as TaskUpdate);
        break;
      case 'progress_update':
        this.emit('progress_update', message.data as ProgressUpdate);
        break;
      case 'status_update':
        this.emit('status_update', message.data as StatusUpdate);
        break;
      case 'webui_status':
        this.emit('webui_status', message.data as WebUIStatusUpdate);
        break;
      case 'system_status':
        this.emit('system_status', message.data);
        break;
      case 'global_log':
        this.emit('global_log', message.data as GlobalLogMessage);
        break;
      case 'log_message':
        this.emit('log_message', message.data as TaskLogMessage);
        break;
      case 'error':
        console.error('WebSocket 服务器错误:', message.data);
        this.emit('error', message.data);
        break;
      default:
        console.log('WebSocket 未知消息类型:', message.type);
        this.emit('message', message);
    }
  }

  // ==================== 重连机制 ====================
  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('WebSocket 重连次数已达上限，停止重连');
      this.emit('max_reconnect_attempts_reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    
    console.log(`WebSocket重连尝试 ${this.reconnectAttempts}/${this.maxReconnectAttempts}，${delay}ms后重试`);
    
    this.connectTimer = setTimeout(() => {
      if (!this.isConnected && !this.isConnecting) {
        this.connect(this.url);
      }
    }, delay);
  }

  // ==================== 心跳机制 ====================
  private startHeartbeat(): void {
    this.stopHeartbeat();
    
    this.heartbeatInterval = setInterval(() => {
      if (this.connection && this.connection.readyState === WebSocket.OPEN) {
        this.send({ type: 'ping' });
        
        // 设置心跳超时
        this.heartbeatTimeout = setTimeout(() => {
          console.warn('WebSocket 心跳超时，断开连接');
          this.disconnect();
        }, 10000);
      }
    }, 30000);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
    if (this.heartbeatTimeout) {
      clearTimeout(this.heartbeatTimeout);
      this.heartbeatTimeout = null;
    }
  }

  // ==================== 消息发送 ====================
  send(message: WebSocketMessage): boolean {
    if (!this.connection || this.connection.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket 未连接，无法发送消息');
      return false;
    }

    try {
      this.connection.send(JSON.stringify(message));
      return true;
    } catch (error) {
      console.error('WebSocket 发送消息失败:', error);
      return false;
    }
  }

  // ==================== 事件管理 ====================
  on(event: string, handler: Function): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, []);
    }
    this.eventHandlers.get(event)!.push(handler);
  }

  off(event: string, handler: Function): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  private emit(event: string, data?: any): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(data);
        } catch (error) {
          console.error(`WebSocket 事件处理器错误 (${event}):`, error);
        }
      });
    }
  }

  // ==================== 连接控制 ====================
  disconnect(): void {
    this.stopHeartbeat();
    
    if (this.connectTimer) {
      clearTimeout(this.connectTimer);
      this.connectTimer = null;
    }

    if (this.connection) {
      this.connection.close(1000, '主动断开');
      this.connection = null;
    }

    this.isConnected = false;
    this.isConnecting = false;
    this.reconnectAttempts = 0;
  }

  reconnect(): void {
    if (this.url) {
      console.log('WebSocket 手动重连');
      this.disconnect();
      this.connect(this.url);
    } else {
      console.warn('WebSocket 重连失败：没有保存的连接URL');
    }
  }

  // ==================== 状态查询 ====================
  isWebSocketConnected(): boolean {
    return this.isConnected && this.connection?.readyState === WebSocket.OPEN;
  }

  getConnectionStatus(): { isConnected: boolean; isConnecting: boolean; readyState: number } {
    return {
      isConnected: this.isConnected,
      isConnecting: this.isConnecting,
      readyState: this.connection?.readyState || WebSocket.CLOSED
    };
  }

  getConnectionId(): string {
    return this.connectionId;
  }

  getReconnectAttempts(): number {
    return this.reconnectAttempts;
  }

  // ==================== 工具方法 ====================
  requestTaskUpdate(taskId: string): void {
    this.send({
      type: 'get_task_status',
      task_id: taskId,
      timestamp: Date.now()
    });
  }

  requestSystemStatus(): void {
    this.send({
      type: 'get_system_status',
      timestamp: Date.now()
    });
  }

  requestWebUIStatus(): void {
    this.send({
      type: 'get_webui_status',
      timestamp: Date.now()
    });
  }
}

// ==================== 导出单例 ====================
export const wsManager = new WebSocketManager();
export default wsManager;