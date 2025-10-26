// AI服务 - 统一管理AI图像生成相关功能
import { apiService } from './api';
import { GenerationParams, Task } from './appState';

// ==================== 类型定义 ====================
export interface AIGenerationOptions {
  prompt: string;
  negativePrompt?: string;
  steps?: number;
  cfgScale?: number;
  width?: number;
  height?: number;
  seed?: number;
  model?: string;
  sampler?: string;
  batchSize?: number;
  enableHR?: boolean;
  loras?: Array<{
    name: string;
    weight: number;
  }>;
}

export interface AIGenerationResult {
  taskId: string;
  status: string;
  progress: number;
  message: string;
  images?: string[];
  result?: any;
  createdAt: string;
  completedAt?: string;
}

export interface AIConfigGenerationOptions {
  configId: string;
  override?: Record<string, any>;
}

// ==================== AI 服务类 ====================
class AIService {
  private static instance: AIService;
  private generationCallbacks: Map<string, (result: AIGenerationResult) => void> = new Map();

  private constructor() {}

  static getInstance(): AIService {
    if (!AIService.instance) {
      AIService.instance = new AIService();
    }
    return AIService.instance;
  }

  // ==================== 图像生成 ====================
  async generateImages(params: GenerationParams): Promise<AIGenerationResult> {
    try {
      const response = await apiService.generateImages(params);
      
      return {
        taskId: response.task_id,
        status: response.status,
        progress: response.progress,
        message: response.message,
        images: response.result?.images || [],
        result: response.result,
        createdAt: response.created_at,
        completedAt: response.completed_at
      };
    } catch (error) {
      console.error('AI图像生成失败:', error);
      throw new Error(`使用配置生成AI图像失败: ${error instanceof Error ? error.message : '未知错误'}`);
    }
  }

  async generateWithConfig(configId: string, override: Partial<GenerationParams>): Promise<AIGenerationResult> {
    try {
      const response = await apiService.generateWithConfig(configId, override);
      
      return {
        taskId: response.task_id,
        status: response.status,
        progress: response.progress,
        message: response.message,
        images: response.result?.images || [],
        result: response.result,
        createdAt: response.created_at,
        completedAt: response.completed_at
      };
    } catch (error) {
      console.error('使用配置生成AI图像失败:', error);
      throw new Error(`使用配置生成AI图像失败: ${error instanceof Error ? error.message : '未知错误'}`);
    }
  }

  // ==================== 参数转换 ====================
  convertToGenerationParams(options: AIGenerationOptions): GenerationParams {
    return {
      prompt: options.prompt,
      negative_prompt: options.negativePrompt || '',
      steps: options.steps || 20,
      cfg_scale: options.cfgScale || 7.0,
      width: options.width || 512,
      height: options.height || 512,
      seed: options.seed || -1,
      model: options.model || '',
      sampler: options.sampler || 'DPM++ 2M Karras',
      batch_size: options.batchSize || 1,
      enable_hr: options.enableHR || false,
      hr_scale: 2.0,
      hr_upscaler: 'Latent',
      hr_steps: 0,
      hr_denoising_strength: 0.7
    };
  }

  // ==================== 任务管理 ====================
  async getTaskStatus(taskId: string): Promise<AIGenerationResult> {
    try {
      const task = await apiService.getTask(taskId);
      
      return {
        taskId: task.id,
        status: task.status,
        progress: task.progress,
        message: task.message,
        images: task.result?.images || [],
        result: task.result,
        createdAt: task.created_at,
        completedAt: task.completed_at
      };
    } catch (error) {
      console.error('获取任务状态失败:', error);
      throw new Error(`获取任务状态失败: ${error instanceof Error ? error.message : '未知错误'}`);
    }
  }

  async cancelTask(taskId: string): Promise<void> {
    try {
      await apiService.cancelTask(taskId);
    } catch (error) {
      console.error('取消任务失败:', error);
      throw new Error(`取消任务失败: ${error instanceof Error ? error.message : '未知错误'}`);
    }
  }

  // ==================== 回调管理 ====================
  onGenerationComplete(taskId: string, callback: (result: AIGenerationResult) => void): void {
    this.generationCallbacks.set(taskId, callback);
  }

  offGenerationComplete(taskId: string): void {
    this.generationCallbacks.delete(taskId);
  }

  private notifyGenerationComplete(result: AIGenerationResult): void {
    const callback = this.generationCallbacks.get(result.taskId);
    if (callback) {
      callback(result);
    }
  }

  // ==================== 工具方法 ====================
  validateGenerationParams(params: GenerationParams): { valid: boolean; errors: string[] } {
    const errors: string[] = [];

    if (!params.prompt || params.prompt.trim().length === 0) {
      errors.push('提示词不能为空');
    }

    if (params.prompt && params.prompt.length > 1000) {
      errors.push('提示词长度不能超过1000个字符');
    }

    if (params.negative_prompt && params.negative_prompt.length > 500) {
      errors.push('负面提示词长度不能超过500个字符');
    }

    if (params.steps < 1 || params.steps > 150) {
      errors.push('步数必须在1-150之间');
    }

    if (params.cfg_scale < 1 || params.cfg_scale > 30) {
      errors.push('CFG Scale必须在1-30之间');
    }

    if (params.width < 64 || params.width > 2048) {
      errors.push('宽度必须在64-2048之间');
    }

    if (params.height < 64 || params.height > 2048) {
      errors.push('高度必须在64-2048之间');
    }

    if (params.batch_size < 1 || params.batch_size > 8) {
      errors.push('批次大小必须在1-8之间');
    }

    if (params.enable_hr && (params.hr_scale < 1 || params.hr_scale > 4)) {
      errors.push('高分辨率修复倍数必须在1-4之间');
    }

    return {
      valid: errors.length === 0,
      errors
    };
  }

  generateRandomSeed(): number {
    return Math.floor(Math.random() * 1000000);
  }

  formatGenerationTime(createdAt: string, completedAt?: string): string {
    if (!completedAt) {
      return '进行中...';
    }

    const start = new Date(createdAt).getTime();
    const end = new Date(completedAt).getTime();
    const duration = Math.round((end - start) / 1000);

    if (duration < 60) {
      return `${duration}秒`;
    } else if (duration < 3600) {
      return `${Math.round(duration / 60)}分钟`;
    } else {
      return `${Math.round(duration / 3600)}小时`;
    }
  }

  // ==================== 预设配置 ====================
  getPresetConfigs(): Array<{ name: string; params: Partial<GenerationParams> }> {
    return [
      {
        name: '快速生成',
        params: {
          steps: 10,
          cfg_scale: 7.0,
          sampler: 'Euler a'
        }
      },
      {
        name: '高质量',
        params: {
          steps: 50,
          cfg_scale: 7.0,
          sampler: 'DPM++ 2M Karras'
        }
      },
      {
        name: '动漫风格',
        params: {
          steps: 30,
          cfg_scale: 7.0,
          sampler: 'DPM++ 2M Karras',
          enable_hr: true,
          hr_scale: 2.0
        }
      },
      {
        name: '写实风格',
        params: {
          steps: 40,
          cfg_scale: 7.0,
          sampler: 'DPM++ SDE Karras',
          enable_hr: true,
          hr_scale: 1.5
        }
      }
    ];
  }

  applyPresetConfig(params: GenerationParams, presetName: string): GenerationParams {
    const presets = this.getPresetConfigs();
    const preset = presets.find(p => p.name === presetName);
    
    if (!preset) {
      return params;
    }

    return {
      ...params,
      ...preset.params
    };
  }
}

// ==================== 导出单例 ====================
export const aiService = AIService.getInstance();
export default aiService;