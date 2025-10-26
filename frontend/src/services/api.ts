// API 服务层 - 统一管理所有后端 API 调用
import { GenerationParams, Task, CrawlRequest, PixivImage, GeneratedImage, SystemStatus, LoraConfig } from './appState';
import { API_BASE_URL } from '../config/ports';

// ==================== 类型定义 ====================
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

export interface WebUIStatus {
  status: string;
  port_open: boolean;
  api_responding: boolean;
  process_id: boolean;
  managed: boolean;
}

export interface ConfigItem {
  id: string;
  name: string;
  description?: string;
  category: string;
  prompt: string;
  negative_prompt: string;
  steps: number;
  cfg_scale: number;
  width: number;
  height: number;
  sampler: string;
  batch_size: number;
  enable_hr: boolean;
  hr_scale: number;
  hr_upscaler: string;
  hr_steps: number;
  hr_denoising_strength: number;
  loras?: LoraConfig[];
  vae?: string;
  restore_faces?: boolean;
  tiling?: boolean;
  clip_skip?: number;
  created_at: string;
  updated_at: string;
}

// ==================== API 服务类 ====================
class ApiService {
  private baseUrl: string;
  private timeout: number;

  constructor() {
    this.baseUrl = import.meta.env.VITE_API_URL || API_BASE_URL;
    this.timeout = 0; // 无超时限制，参考PixivTailor项目
  }

  // ==================== 通用请求方法 ====================
  private async request<T>(endpoint: string, data?: any, method: string = 'POST'): Promise<ApiResponse<T>> {
    const controller = new AbortController();
    let timeoutId: NodeJS.Timeout | null = null;
    
    // 只有当超时时间大于0时才设置超时
    if (this.timeout > 0) {
      timeoutId = setTimeout(() => controller.abort(), this.timeout);
    }

    try {
      // 对于GET请求，将参数添加到URL中
      let url = `${this.baseUrl}${endpoint}`;
      if (method === 'GET' && data) {
        const params = new URLSearchParams();
        Object.keys(data).forEach(key => {
          if (data[key] !== undefined && data[key] !== null) {
            params.append(key, data[key]);
          }
        });
          url += `?${params.toString()}`;
      }

      const options: RequestInit = {
        method,
        headers: {
          'Content-Type': 'application/json',
        },
        signal: controller.signal,
      };

      if (method !== 'GET' && data) {
        options.body = JSON.stringify(data);
      }

      const response = await fetch(url, options);
      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.message || `HTTP ${response.status}`);
      }

      return result;
    } catch (error) {
      if (error instanceof Error) {
        if (error.name === 'AbortError') {
          throw new Error('请求超时');
        }
        throw error;
      }
      throw new Error('网络请求失败');
    } finally {
      if (timeoutId) {
        clearTimeout(timeoutId);
      }
    }
  }

  // ==================== 图像生成 API ====================
  async generateImages(params: GenerationParams): Promise<GenerationResponse> {
      const response = await this.request<GenerationResponse>('/generate', params);
    return response.data!;
      }
      
  async generateWithConfig(configId: string, override: Partial<GenerationParams>): Promise<GenerationResponse> {
    const response = await this.request<GenerationResponse>('/generate-with-config', {
      config_id: configId,
      override
    });
      return response.data!;
  }

  // ==================== WebUI 管理 API ====================
  async startWebUI(): Promise<{ message: string; status: string }> {
    const response = await this.request<{ message: string; status: string }>('/webui/start');
    return response.data!;
  }

  async stopWebUI(): Promise<{ message: string; status: string }> {
    const response = await this.request<{ message: string; status: string }>('/webui/stop');
    return response.data!;
  }

  async getWebUIStatus(): Promise<WebUIStatus> {
    const response = await this.request<WebUIStatus>('/webui/status', undefined, 'GET');
    return response.data!;
  }

  createWebUILogStream(): EventSource {
    return new EventSource(`${this.baseUrl}/webui/logs`);
  }

  // ==================== 配置管理 API ====================
  async getConfigs(): Promise<ConfigItem[]> {
    try {
      const response = await this.request<{ categories: string[]; configs: ConfigItem[] }>('/configs', undefined, 'GET');
      const data = response.data || { configs: [] };
      return Array.isArray(data.configs) ? data.configs : [];
    } catch (error) {
      console.error('获取配置列表失败:', error);
      return [];
    }
  }

  async getConfig(id: string): Promise<ConfigItem | null> {
    try {
      const response = await this.request<ConfigItem>(`/configs/${id}`, undefined, 'GET');
      return response.data || null;
    } catch (error) {
      console.error('获取配置失败:', error);
      return null;
    }
  }

  async createConfig(config: Omit<ConfigItem, 'id' | 'created_at' | 'updated_at'>): Promise<ConfigItem> {
    const response = await this.request<ConfigItem>('/configs/create', config);
    return response.data!;
  }

  async updateConfig(id: string, config: Partial<ConfigItem>): Promise<ConfigItem> {
    const response = await this.request<ConfigItem>(`/configs/${id}`, config, 'PUT');
    return response.data!;
  }

  async deleteConfig(id: string): Promise<void> {
    await this.request(`/configs/${id}`, undefined, 'DELETE');
  }

  async getConfigCategories(): Promise<string[]> {
    try {
      const response = await this.request<{ categories: string[] }>('/configs/categories', undefined, 'GET');
      const data = response.data || { categories: [] };
      return Array.isArray(data.categories) ? data.categories : [];
    } catch (error) {
      console.error('获取配置分类失败:', error);
      return [];
    }
  }

  // 别名方法，用于 ConfigManagerPage
  async getCategories(): Promise<string[]> {
    return this.getConfigCategories();
  }

  // 配置管理相关方法
  async listConfigs(params?: {
    page?: number;
    pageSize?: number;
    category?: string;
    search?: string;
  }): Promise<{ configs: any[]; total: number }> {
    try {
      // 构建查询参数
      const queryParams = new URLSearchParams();
      if (params?.page) queryParams.append('page', params.page.toString());
      if (params?.pageSize) queryParams.append('pageSize', params.pageSize.toString());
      if (params?.category) queryParams.append('category', params.category);
      if (params?.search) queryParams.append('search', params.search);
      
      const url = `/configs${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
      const response = await this.request<{ categories: string[]; configs: any[] }>(url, undefined, 'GET');
      
      // 后端返回 { data: { categories: [], configs: [] } }
      const data = response.data || { configs: [], categories: [] };
      return { 
        configs: Array.isArray(data.configs) ? data.configs : [], 
        total: Array.isArray(data.configs) ? data.configs.length : 0 
      };
    } catch (error) {
      console.error('获取配置列表失败:', error);
      return { configs: [], total: 0 };
    }
  }

  async deleteConfigFile(id: string): Promise<void> {
    try {
      await this.request(`/configs/${id}`, undefined, 'DELETE');
    } catch (error) {
      console.error('删除配置文件失败:', error);
      throw error;
    }
  }

  async setDefaultConfig(id: string): Promise<void> {
    try {
      await this.request(`/configs/${id}/default`, undefined, 'POST');
    } catch (error) {
      console.error('设置默认配置失败:', error);
      throw error;
    }
  }

  async updateConfigFile(id: string, config: any): Promise<void> {
    try {
      await this.request(`/configs/file/${id}/update`, config, 'PUT');
    } catch (error) {
      console.error('更新配置文件失败:', error);
      throw error;
    }
  }

  async createConfigFile(config: any): Promise<void> {
    try {
      await this.request('/configs', config, 'POST');
    } catch (error) {
      console.error('创建配置文件失败:', error);
      throw error;
    }
  }

  async importConfig(configData: any): Promise<void> {
    try {
      await this.request('/configs/import', configData, 'POST');
    } catch (error) {
      console.error('导入配置失败:', error);
      throw error;
    }
  }

  async exportConfigs(configIds: string[]): Promise<Blob> {
    try {
      const response = await fetch(`${this.baseUrl}/configs/export`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ config_ids: configIds }),
      });
      
      if (!response.ok) {
        throw new Error('导出配置失败');
      }
      
      return response.blob();
    } catch (error) {
      console.error('导出配置失败:', error);
      throw error;
    }
  }

  // ==================== 爬虫管理 API ====================
  async createCrawlTask(request: CrawlRequest): Promise<{ task_id: string; message: string }> {
    const response = await this.request<{ task_id: string; message: string }>('/crawl/create', request);
    return response.data!;
  }

  async getCrawlTasks(): Promise<Task[]> {
    try {
      // 使用通用的任务列表接口
      return await this.getTasks();
    } catch (error) {
      console.error('获取爬虫任务失败:', error);
      return [];
    }
  }

  async getCrawlTask(taskId: string): Promise<Task> {
    try {
      // 使用通用的任务详情接口
      return await this.getTask(taskId);
    } catch (error) {
      console.error('获取爬虫任务详情失败:', error);
      throw error;
    }
  }

  async stopCrawlTask(taskId: string): Promise<{ message: string }> {
    try {
      // 使用通用的取消任务接口
      return await this.cancelTask(taskId);
    } catch (error) {
      console.error('停止爬虫任务失败:', error);
      throw error;
    }
  }

  async deleteCrawlTask(taskId: string): Promise<{ message: string }> {
    try {
      // 使用通用的删除任务接口
      return await this.deleteTask(taskId);
    } catch (error) {
      console.error('删除爬虫任务失败:', error);
      throw error;
    }
  }

  // ==================== 图片管理 API ====================
  async getImages(params?: {
    task_id?: string;
    page?: number;
    limit?: number;
    category?: string;
  }): Promise<{ images: PixivImage[]; total: number; page: number; limit: number }> {
    try {
      const response = await this.request<{ images: PixivImage[]; total: number; page: number; limit: number }>(
        '/generated/images',
        {
          pagination: {
            page: params?.page || 1,
            page_size: params?.limit || 20
          },
          model: params?.category || ''
        },
        'POST'
      );
      return response.data || { images: [], total: 0, page: 1, limit: 20 };
    } catch (error) {
      console.error('获取图片列表失败:', error);
      return { images: [], total: 0, page: 1, limit: 20 };
    }
  }

  async downloadImage(imageId: string): Promise<Blob> {
    try {
      // 后端没有下载端点，直接通过图片URL下载
      const response = await fetch(`${this.baseUrl}/images/${imageId}`);
      if (!response.ok) {
        throw new Error('下载图片失败');
      }
      return response.blob();
    } catch (error) {
      console.error('下载图片失败:', error);
      throw error;
    }
  }

  async deleteImage(imageId: string): Promise<{ message: string }> {
    try {
      // 后端没有删除图片的端点，这里先返回成功消息
      console.warn('删除图片功能暂未实现');
      return { message: '删除图片功能暂未实现' };
    } catch (error) {
      console.error('删除图片失败:', error);
      throw error;
    }
  }

  // ==================== 系统状态 API ====================
  async getSystemStatus(): Promise<SystemStatus> {
    try {
      const response = await this.request<SystemStatus>('/system/info', undefined, 'POST');
      return response.data!;
    } catch (error) {
      console.error('获取系统状态失败:', error);
      throw error;
    }
  }

  async getSystemInfo(): Promise<{
    version: string;
    uptime: number;
    memory_usage: number;
    disk_usage: number;
    cpu_usage: number;
  }> {
    try {
      const response = await this.request<{
        version: string;
        uptime: number;
        memory_usage: number;
        disk_usage: number;
        cpu_usage: number;
      }>('/system/info', undefined, 'POST');
      return response.data!;
    } catch (error) {
      console.error('获取系统信息失败:', error);
      throw error;
    }
  }

  // ==================== 任务管理 API ====================
  async getTasks(): Promise<Task[]> {
    try {
      const response = await this.request<{ tasks: Task[] }>('/tasks', {
        pagination: { page: 1, page_size: 100 },
        status: '',
        type: ''
      }, 'POST');
      return Array.isArray(response.data?.tasks) ? response.data.tasks : [];
    } catch (error) {
      console.error('获取任务列表失败:', error);
      return [];
    }
  }

  async getTask(taskId: string): Promise<Task> {
    try {
      const response = await this.request<Task>('/status', { task_id: taskId }, 'POST');
      return response.data!;
    } catch (error) {
      console.error('获取任务详情失败:', error);
      throw error;
    }
  }

  async cleanupTasks(cleanupType: string): Promise<{ cleaned_count: number }> {
    try {
      const response = await this.request<{ cleaned_count: number; cleanup_type: string; message: string }>('/task/cleanup', {
        cleanup_type: cleanupType
      });
      // 后端返回的是 { status: {...}, data: {...} } 格式，需要提取 data 字段
      if (response.data) {
        return { cleaned_count: response.data.cleaned_count };
      }
      return { cleaned_count: (response as any).cleaned_count || 0 };
    } catch (error) {
      console.error('清理任务失败:', error);
      throw error;
    }
  }

  async startTask(taskId: string): Promise<void> {
    try {
      await this.request<void>('/task/start', { task_id: taskId }, 'POST');
    } catch (error) {
      console.error('启动任务失败:', error);
      throw error;
    }
  }

  async stopTask(taskId: string): Promise<void> {
    try {
      await this.request<void>('/task/stop', { task_id: taskId }, 'POST');
    } catch (error) {
      console.error('停止任务失败:', error);
      throw error;
    }
  }

  async cancelTask(taskId: string): Promise<{ message: string }> {
    try {
      const response = await this.request<{ message: string }>('/cancel', { task_id: taskId }, 'POST');
    return response.data!;
    } catch (error) {
      console.error('取消任务失败:', error);
      throw error;
    }
  }

  async deleteTask(taskId: string): Promise<{ message: string }> {
    try {
      const response = await this.request<{ message: string }>('/delete', {
        task_id: taskId
      });
      // 后端返回的是 { status: {...}, data: {...} } 格式，需要提取 data 字段
      if (response.data) {
        return { message: response.data.message };
      }
      return { message: (response as any).message || '删除成功' };
    } catch (error) {
      console.error('删除任务失败:', error);
      throw error;
    }
  }

  // ==================== 工具方法 ====================
  getImageUrl(imagePath: string): string {
    return `${this.baseUrl}/images/${imagePath}`;
  }

  getTaskStatusColor(status: string): string {
    switch (status) {
      case 'pending':
        return '#faad14';
      case 'running':
        return '#1890ff';
      case 'completed':
        return '#52c41a';
      case 'failed':
        return '#ff4d4f';
      case 'cancelled':
        return '#d9d9d9';
      default:
        return '#d9d9d9';
    }
  }

  getTaskStatusText(status: string): string {
    switch (status) {
      case 'pending':
        return '等待中';
      case 'running':
        return '运行中';
      case 'completed':
        return '已完成';
      case 'failed':
        return '失败';
      case 'cancelled':
        return '已取消';
      default:
        return '未知';
    }
  }
}

// ==================== 导出单例 ====================
export const apiService = new ApiService();
export default apiService;