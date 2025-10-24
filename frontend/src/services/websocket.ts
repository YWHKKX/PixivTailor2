// WebSocket 连接管理器 - 基于 plan.md 设计
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
  message: string;
  result?: any;
  error?: string;
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

  // 连接 WebSocket
  connect(url: string): void {
    // 清除之前的连接定时器
    if (this.connectTimer) {
      clearTimeout(this.connectTimer);
      this.connectTimer = null;
    }

    // 如果已有连接且状态正常，跳过重复连接
    if (this.connection && this.isConnected) {
      console.log('WebSocket 已连接，跳过重复连接');
      return;
    }

    // 如果正在连接中，跳过重复连接
    if (this.isConnecting) {
      console.log('WebSocket 正在连接中，跳过重复连接');
      return;
    }

    // 如果已有连接，先关闭
    if (this.connection) {
      this.connection.close();
      this.connection = null;
    }

    this.url = url;
    
    // 添加防抖延迟，避免频繁连接
    this.connectTimer = setTimeout(() => {
      this.isConnecting = true;
      
      try {
        // 生成连接ID用于去重
        this.connectionId = `ws_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
        console.log(`WebSocket 连接尝试 ${this.connectionId}`);
        
        this.connection = new WebSocket(url);
        this.setupEventHandlers();
      } catch (error) {
        console.error('WebSocket连接失败:', error);
        this.isConnecting = false;
      }
    }, 500); // 500ms防抖延迟
  }

  // 设置事件处理器
  private setupEventHandlers(): void {
    if (!this.connection) return;

    this.connection.onopen = (event) => {
      console.log(`WebSocket连接已建立 ${this.connectionId}`);
      this.isConnected = true;
      this.isConnecting = false;
      this.reconnectAttempts = 0;
      this.startHeartbeat();
      this.emit('connected', event);
    };

    this.connection.onmessage = (event) => {
      try {
        const data: WebSocketMessage = JSON.parse(event.data);
        this.handleMessage(data);
      } catch (error) {
        console.error('消息解析失败:', error);
      }
    };

    this.connection.onclose = (event) => {
      console.log(`WebSocket连接已关闭 ${this.connectionId}:`, event.code, event.reason);
      this.isConnected = false;
      this.isConnecting = false;
      this.stopHeartbeat();
      this.emit('disconnected', event);
      
      // 只有在非正常关闭时才重连
      if (event.code !== 1000) { // 1000 是正常关闭
        console.log('WebSocket异常关闭，启动重连机制');
        this.handleReconnect();
      } else {
        console.log('WebSocket正常关闭，不进行重连');
      }
    };

    this.connection.onerror = (error) => {
      console.error('WebSocket错误:', error);
      this.emit('error', error);
    };
  }

  // 处理消息
  private handleMessage(data: WebSocketMessage): void {
    const { type, data: payload } = data;
    const message = (data as any).message;

    switch (type) {
      case 'welcome':
        console.log('WebSocket 欢迎消息:', message);
        break;
      case 'task_update':
        this.emit('taskUpdate', payload);
        break;
      case 'progress_update':
        this.emit('progressUpdate', payload);
        break;
      case 'status_update':
        this.emit('statusUpdate', payload);
        break;
      case 'error':
        this.emit('error', payload);
        break;
      case 'pong':
        // 心跳响应
        if (this.heartbeatTimeout) {
          clearTimeout(this.heartbeatTimeout);
          this.heartbeatTimeout = null;
        }
        break;
      case 'log_message':
        this.emit('logMessage', data);
        break;
      case 'global_log':
        this.emit('logMessage', data);
        break;
      default:
        console.log('未知消息类型:', type);
    }
  }

  // 发送消息
  send(data: WebSocketMessage): boolean {
    if (this.connection && this.isConnected) {
      try {
        this.connection.send(JSON.stringify(data));
        return true;
      } catch (error) {
        console.error('发送消息失败:', error);
        return false;
      }
    }
    return false;
  }

  // 心跳机制
  private startHeartbeat(): void {
    this.heartbeatInterval = setInterval(() => {
      if (this.isConnected) {
        this.send({ type: 'ping' });
        
        // 设置心跳超时
        this.heartbeatTimeout = setTimeout(() => {
          console.warn('心跳超时，重新连接...');
          if (this.connection) {
            this.connection.close();
          }
        }, 5000);
      }
    }, 30000); // 30秒发送一次心跳
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

  // 重连机制
  private handleReconnect(): void {
    // 如果正在连接中，不重复重连
    if (this.isConnecting) {
      console.log('WebSocket正在连接中，跳过重连');
      return;
    }

    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = Math.min(this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1), 10000); // 最大延迟10秒
      
      console.log(`WebSocket重连尝试 ${this.reconnectAttempts}/${this.maxReconnectAttempts}，${delay}ms后重试`);
      
      setTimeout(() => {
        if (!this.isConnected && !this.isConnecting) {
          this.connect(this.url);
        }
      }, delay);
    } else {
      console.error('WebSocket重连失败，已达到最大重试次数');
      this.emit('reconnectFailed');
      
      // 重置重连计数，允许后续重连
      setTimeout(() => {
        this.reconnectAttempts = 0;
        console.log('WebSocket重连计数已重置，可以重新尝试连接');
      }, 30000); // 30秒后重置
    }
  }

  // 事件系统
  on(event: string, handler: Function): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, []);
    }
    this.eventHandlers.get(event)!.push(handler);
  }

  off(event: string, handler: Function): void {
    if (this.eventHandlers.has(event)) {
      const handlers = this.eventHandlers.get(event)!;
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  private emit(event: string, data?: any): void {
    if (this.eventHandlers.has(event)) {
      this.eventHandlers.get(event)!.forEach(handler => handler(data));
    }
  }

  // 关闭连接
  close(): void {
    this.stopHeartbeat();
    this.isConnecting = false;
    this.isConnected = false;
    
    // 清理连接定时器
    if (this.connectTimer) {
      clearTimeout(this.connectTimer);
      this.connectTimer = null;
    }
    
    if (this.connection) {
      this.connection.close();
      this.connection = null;
    }
  }

  // 手动重连
  reconnect(): void {
    console.log('手动触发WebSocket重连');
    this.reconnectAttempts = 0;
    if (this.url) {
      this.connect(this.url);
    }
  }

  // 获取连接状态
  getConnectionStatus(): { isConnected: boolean; isConnecting: boolean; reconnectAttempts: number } {
    return {
      isConnected: this.isConnected,
      isConnecting: this.isConnecting,
      reconnectAttempts: this.reconnectAttempts
    };
  }
}

// 全局 WebSocket 管理器实例
export const wsManager = new WebSocketManager();

// 初始化连接
export const initWebSocket = (): void => {
  const wsUrl = 'ws://localhost:50052/ws';
  const isDevelopment = import.meta.env.DEV;
  
  if (isDevelopment) {
    console.log('开发模式：WebSocket连接已初始化');
  }
  
  wsManager.connect(wsUrl);
};

// 页面卸载时清理连接
if (typeof window !== 'undefined') {
  window.addEventListener('beforeunload', () => {
    console.log('页面即将卸载，清理 WebSocket 连接');
    wsManager.close();
  });
}

// 导出供其他模块使用
export default wsManager;
