// 应用状态管理器 - 基于 plan.md 设计
export interface GenerationParams {
  prompt: string;
  negative_prompt?: string;
  steps: number;
  cfg_scale: number;
  width: number;
  height: number;
  seed: number;
  model: string;
  sampler: string;
  batch_size: number;
  enable_hr: boolean;
}

// 任务状态枚举 - 与后端同步
export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

// 任务类型枚举 - 与后端同步
export type TaskType = 'crawl' | 'generate' | 'train' | 'tag' | 'classify';

// 爬取类型枚举 - 与后端同步
export type CrawlType = 'tag' | 'user' | 'illust';

// 排序方式枚举 - 与后端同步
export type Order = 'date_d' | 'popular_d';

// 模式枚举 - 与后端同步
export type Mode = 'safe' | 'r18' | 'all';

// 任务接口 - 与后端模型完全同步
export interface Task {
  id: string;
  name: string;
  type: TaskType;
  status: TaskStatus;
  config: string;
  progress: number;
  created_at: string;
  updated_at: string;
  started_at?: string;
  completed_at?: string;
  error_message?: string;
  images_found?: number;    // 获取到的图片数量
  images_downloaded?: number; // 下载的图片数量
}

// 爬取请求接口 - 与后端同步
export interface CrawlRequest {
  type: CrawlType;
  query: string;
  user_id?: number;
  illust_id?: number;
  order: Order;
  mode: Mode;
  limit: number;
  delay: number;
  proxy_enabled?: boolean;
  proxy_url?: string;
  cookie?: string;
}

// Pixiv图像接口 - 与后端同步
export interface PixivImage {
  id: number;
  title: string;
  author: string;
  author_id: number;
  tags: string[];
  url: string;
  thumbnail_url: string;
  width: number;
  height: number;
  bookmarks: number;
  views: number;
  is_r18: boolean;
  created_at: string;
  updated_at: string;
}

// 生成图像接口 - 与后端同步
export interface GeneratedImage {
  id: number;
  prompt: string;
  negative_prompt: string;
  model: string;
  loras: string[];
  image_url: string;
  width: number;
  height: number;
  seed: number;
  cfg_scale: number;
  steps: number;
  sampler: string;
  created_at: string;
  updated_at: string;
}

// 系统状态接口 - 与后端同步
export interface SystemStatus {
  version: string;
  status: string;
  uptime: number;
  modules: Record<string, string>;
  metrics: SystemMetrics;
}

// 系统指标接口 - 与后端同步
export interface SystemMetrics {
  cpu_usage: number;
  memory_usage: number;
  disk_usage: number;
  active_tasks: number;
  completed_tasks: number;
  failed_tasks: number;
}

export interface AppState {
  currentTask: Task | null;
  isGenerating: boolean;
  generationParams: GenerationParams;
  tasks: Task[];
  systemStatus: 'connected' | 'disconnected' | 'error';
}

class AppStateManager {
  private state: AppState = {
    currentTask: null,
    isGenerating: false,
    generationParams: {
      prompt: '',
      negative_prompt: '',
      steps: 20,
      cfg_scale: 7.0,
      width: 512,
      height: 512,
      seed: -1,
      model: 'Model A',
      sampler: 'Euler',
      batch_size: 1,
      enable_hr: false,
    },
    tasks: [],
    systemStatus: 'disconnected',
  };

  private callbacks = new Map<string, Function[]>();

  // 事件系统
  on(event: string, callback: Function): void {
    if (!this.callbacks.has(event)) {
      this.callbacks.set(event, []);
    }
    this.callbacks.get(event)!.push(callback);
  }

  private emit(event: string, data?: any): void {
    if (this.callbacks.has(event)) {
      this.callbacks.get(event)!.forEach(callback => callback(data));
    }
  }

  // 获取当前状态
  getState(): AppState {
    return { ...this.state };
  }

  // 更新生成参数
  updateGenerationParams(params: Partial<GenerationParams>): void {
    this.state.generationParams = { ...this.state.generationParams, ...params };
    this.emit('params_updated', this.state.generationParams);
  }

  // 开始生成任务
  startGeneration(params: GenerationParams): void {
    this.state.isGenerating = true;
    this.state.generationParams = params;
    this.emit('generation_started', params);
  }

  // 停止生成
  stopGeneration(): void {
    this.state.isGenerating = false;
    this.state.currentTask = null;
    this.emit('generation_stopped');
  }

  // 更新任务状态
  updateTask(task: Task): void {
    const existingIndex = this.state.tasks.findIndex(t => t.id === task.id);
    
    if (existingIndex >= 0) {
      this.state.tasks[existingIndex] = task;
    } else {
      this.state.tasks.unshift(task);
    }

    if (this.state.currentTask && this.state.currentTask.id === task.id) {
      this.state.currentTask = task;
    }

    this.emit('task_updated', task);
  }

  // 设置当前任务
  setCurrentTask(task: Task | null): void {
    this.state.currentTask = task;
    this.emit('current_task_changed', task);
  }

  // 更新系统状态
  updateSystemStatus(status: 'connected' | 'disconnected' | 'error'): void {
    this.state.systemStatus = status;
    this.emit('system_status_changed', status);
  }

  // 添加新任务
  addTask(task: Task): void {
    this.state.tasks.unshift(task);
    this.emit('task_added', task);
  }

  // 获取任务列表
  getTasks(): Task[] {
    return [...this.state.tasks];
  }

  // 根据ID获取任务
  getTaskById(taskId: string): Task | null {
    return this.state.tasks.find(task => task.id === taskId) || null;
  }

  // 清除已完成的任务
  clearCompletedTasks(): void {
    this.state.tasks = this.state.tasks.filter(task => 
      task.status !== 'completed' && task.status !== 'failed'
    );
    this.emit('tasks_cleared');
  }

  // 重置状态
  reset(): void {
    this.state = {
      currentTask: null,
      isGenerating: false,
      generationParams: {
        prompt: '',
        negative_prompt: '',
        steps: 20,
        cfg_scale: 7.0,
        width: 512,
        height: 512,
        seed: -1,
        model: 'Model A',
        sampler: 'Euler',
        batch_size: 1,
        enable_hr: false,
      },
      tasks: [],
      systemStatus: 'disconnected',
    };
    this.emit('state_reset');
  }
}

// 创建单例实例
export const appStateManager = new AppStateManager();

export default appStateManager;
