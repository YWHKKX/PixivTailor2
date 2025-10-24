// API 服务层 - 基于 plan.md 设计
import { GenerationParams, Task, CrawlRequest, PixivImage, GeneratedImage, SystemStatus } from './appState';

export interface ApiResponse<T = any> {
  status: {
    code: number;
    message: string;
    details?: string;
  };
  data?: T;
}

export interface GenerationRequest extends GenerationParams {
  // 继承 GenerationParams 的所有属性
}

export interface GenerationResponse {
  task_id: string;
  status: string;
  progress: number;
  message: string;
  result?: any;
  created_at: string;
  completed_at?: string;
}

class ApiService {
  private baseUrl: string;
  private timeout: number;

  constructor() {
    this.baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:50052/api';
    this.timeout = 30000; // 30秒超时
  }

  private async request<T>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        body: data ? JSON.stringify(data) : undefined,
        signal: controller.signal,
        // 明确禁用代理
        mode: 'cors',
        credentials: 'omit',
      });

      clearTimeout(timeoutId);

      // 检查HTTP状态
      if (!response.ok) {
        const errorText = await response.text().catch(() => 'Unknown error');
        throw new Error(`HTTP error: ${response.statusText} - ${errorText}`);
      }

      // 解析响应
      const result = await response.json();
      return result;

    } catch (error) {
      clearTimeout(timeoutId);

      if (error instanceof Error) {
        if (error.name === 'AbortError') {
          throw new Error(`Request timeout: ${this.timeout}ms`);
        }
        throw error;
      }

      throw new Error(`Unknown error: ${String(error)}`);
    }
  }

  // 生成图像
  async generateImage(params: GenerationRequest): Promise<GenerationResponse> {
    const response = await this.request<GenerationResponse>('/generate', params);
    
    if (response.status.code !== 0) {
      throw new Error(response.status.message || 'Generation failed');
    }

    return response.data!;
  }

  // 获取任务状态
  async getTaskStatus(taskId: string): Promise<Task> {
    const response = await this.request<Task>('/status', { task_id: taskId });
    
    if (response.status.code !== 0) {
      throw new Error(response.status.message || 'Failed to get task status');
    }

    return response.data!;
  }

  // 取消任务
  async cancelTask(taskId: string): Promise<void> {
    const response = await this.request('/cancel', { task_id: taskId });
    
    if (response.status.code !== 0) {
      throw new Error(response.status.message || 'Failed to cancel task');
    }
  }

  // 获取任务列表
  async getTasks(page: number = 1, pageSize: number = 20, status?: string, type?: string): Promise<{
    tasks: Task[];
    total: number;
    page: number;
    pageSize: number;
  }> {
    const response = await this.request<{
      tasks: Task[];
      pagination: {
        total: number;
        page: number;
        page_size: number;
      };
    }>('/tasks', {
      pagination: { page, page_size: pageSize, total: 0 },
      status: status || '',
      type: type || '',
    });

    if (response.status.code !== 0) {
      throw new Error(response.status.message || 'Failed to get tasks');
    }

    return {
      tasks: response.data!.tasks,
      total: response.data!.pagination.total,
      page: response.data!.pagination.page,
      pageSize: response.data!.pagination.page_size,
    };
  }

  // 获取配置
  async getConfig(module: string): Promise<any> {
    const response = await this.request('/config/get', { module });
    
    if (response.status.code !== 0) {
      throw new Error(response.status.message || 'Failed to get config');
    }

    return response.data;
  }

  // 更新配置
  async updateConfig(module: string, config: any): Promise<void> {
    const response = await this.request('/config/update', { module, config });
    
    if (response.status.code !== 0) {
      throw new Error(response.status.message || 'Failed to update config');
    }
  }


  // 健康检查
  async healthCheck(): Promise<boolean> {
    try {
      const response = await this.request('/health');
      return response.status.code === 0;
    } catch (error) {
      console.error('Health check failed:', error);
      return false;
    }
  }

  // 创建爬虫任务
  async createCrawlTask(request: CrawlRequest): Promise<Task> {
    const response = await this.request<Task>('/crawl/create', request);
    
    if (response.status.code !== 0) {
      throw new Error(response.status.message || 'Failed to create crawl task');
    }

    return response.data!;
  }

  // 获取爬取结果
  async getCrawlResults(
    page: number = 1,
    pageSize: number = 20,
    tags?: string[],
    author?: string
  ): Promise<{
    results: PixivImage[];
    total: number;
    page: number;
    pageSize: number;
  }> {
    const response = await this.request<{
      results: PixivImage[];
      pagination: {
        total: number;
        page: number;
        page_size: number;
      };
    }>('/crawl/results', {
      pagination: { page, page_size: pageSize, total: 0 },
      tags: tags || [],
      author: author || '',
    });

    if (response.status.code !== 0) {
      throw new Error(response.status.message || 'Failed to get crawl results');
    }

    return {
      results: response.data!.results,
      total: response.data!.pagination.total,
      page: response.data!.pagination.page,
      pageSize: response.data!.pagination.page_size,
    };
  }

  // 获取生成图像
  async getGeneratedImages(
    page: number = 1,
    pageSize: number = 20,
    model?: string
  ): Promise<{
    images: GeneratedImage[];
    total: number;
    page: number;
    pageSize: number;
  }> {
    const response = await this.request<{
      images: GeneratedImage[];
      pagination: {
        total: number;
        page: number;
        page_size: number;
      };
    }>('/generated/images', {
      pagination: { page, page_size: pageSize, total: 0 },
      model: model || '',
    });

    if (response.status.code !== 0) {
      throw new Error(response.status.message || 'Failed to get generated images');
    }

    return {
      images: response.data!.images,
      total: response.data!.pagination.total,
      page: response.data!.pagination.page,
      pageSize: response.data!.pagination.page_size,
    };
  }

  // 获取系统信息
  async getSystemInfo(): Promise<SystemStatus> {
    const response = await this.request<SystemStatus>('/system/info');
    
    if (response.status.code !== 0) {
      throw new Error(response.status.message || 'Failed to get system info');
    }

    return response.data!;
  }

  // 清理任务
  async cleanupTasks(cleanupType: string): Promise<{ cleaned_count: number; cleanup_type: string }> {
    const response = await this.request<{ cleaned_count: number; cleanup_type: string }>('/task/cleanup', {
      cleanup_type: cleanupType
    });
    
    if (response.status.code !== 0) {
      throw new Error(response.status.message || 'Failed to cleanup tasks');
    }

    return response.data!;
  }

  // 启动任务
  async startTask(taskId: string): Promise<void> {
    const response = await this.request<void>('/task/start', {
      task_id: taskId
    });
    
    if (response.status.code !== 0) {
      throw new Error(response.status.message || 'Failed to start task');
    }
  }

  // 停止任务
  async stopTask(taskId: string): Promise<void> {
    const response = await this.request<void>('/task/stop', {
      task_id: taskId
    });
    
    if (response.status.code !== 0) {
      throw new Error(response.status.message || 'Failed to stop task');
    }
  }
}

// 创建单例实例
export const apiService = new ApiService();

export default apiService;
