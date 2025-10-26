import { aiService } from '../../services/aiService';
import { GenerationParams } from '../../services/appState';

// 模拟 API 服务
jest.mock('../../services/api', () => ({
  apiService: {
    generateImages: jest.fn(),
    generateWithConfig: jest.fn(),
    getTask: jest.fn(),
    cancelTask: jest.fn(),
  },
}));

import { apiService } from '../../services/api';

describe('AI Service', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('generateImages', () => {
    it('should generate images successfully', async () => {
      const mockResponse = {
        taskId: 'test-task-id',
        status: 'completed',
        progress: 100,
        message: 'Generation completed',
        images: ['image1.png', 'image2.png'],
        result: { images: ['image1.png', 'image2.png'] },
        createdAt: '2023-01-01T00:00:00Z',
        completedAt: '2023-01-01T00:01:00Z'
      };

      (apiService.generateImages as jest.Mock).mockResolvedValueOnce(mockResponse);

      const params: GenerationParams = {
        prompt: 'test prompt',
        negative_prompt: 'test negative prompt',
        steps: 20,
        cfg_scale: 7.0,
        width: 512,
        height: 512,
        seed: -1,
        model: 'test-model',
        sampler: 'DPM++ 2M Karras',
        batch_size: 1,
        enable_hr: false,
        hr_scale: 2.0,
        hr_upscaler: 'Latent',
        hr_steps: 0,
        hr_denoising_strength: 0.7
      };

      const result = await aiService.generateImages(params);

      expect(result).toEqual(mockResponse);
      expect(apiService.generateImages).toHaveBeenCalledWith(params);
    });

    it('should handle generation errors', async () => {
      const error = new Error('Generation failed');
      (apiService.generateImages as jest.Mock).mockRejectedValueOnce(error);

      const params: GenerationParams = {
        prompt: 'test prompt',
        negative_prompt: 'test negative prompt',
        steps: 20,
        cfg_scale: 7.0,
        width: 512,
        height: 512,
        seed: -1,
        model: 'test-model',
        sampler: 'DPM++ 2M Karras',
        batch_size: 1,
        enable_hr: false,
        hr_scale: 2.0,
        hr_upscaler: 'Latent',
        hr_steps: 0,
        hr_denoising_strength: 0.7
      };

      await expect(aiService.generateImages(params)).rejects.toThrow('使用配置生成AI图像失败: Generation failed');
    });
  });

  describe('generateWithConfig', () => {
    it('should generate with config successfully', async () => {
      const mockResponse = {
        taskId: 'test-task-id',
        status: 'completed',
        progress: 100,
        message: 'Generation completed',
        images: ['image1.png', 'image2.png'],
        result: { images: ['image1.png', 'image2.png'] },
        createdAt: '2023-01-01T00:00:00Z',
        completedAt: '2023-01-01T00:01:00Z'
      };

      (apiService.generateWithConfig as jest.Mock).mockResolvedValueOnce(mockResponse);

      const configId = 'test-config-id';
      const override = { prompt: 'override prompt' };

      const result = await aiService.generateWithConfig(configId, override);

      expect(result).toEqual(mockResponse);
      expect(apiService.generateWithConfig).toHaveBeenCalledWith(configId, override);
    });

    it('should handle config generation errors', async () => {
      const error = new Error('Config generation failed');
      (apiService.generateWithConfig as jest.Mock).mockRejectedValueOnce(error);

      const configId = 'test-config-id';
      const override = { prompt: 'override prompt' };

      await expect(aiService.generateWithConfig(configId, override)).rejects.toThrow('使用配置生成AI图像失败: Config generation failed');
    });
  });

  describe('convertToGenerationParams', () => {
    it('should convert options to generation params', () => {
      const options = {
        prompt: 'test prompt',
        negativePrompt: 'test negative prompt',
        steps: 30,
        cfgScale: 8.0,
        width: 768,
        height: 768,
        seed: 12345,
        model: 'test-model',
        sampler: 'Euler a',
        batchSize: 2,
        enableHR: true
      };

      const result = aiService.convertToGenerationParams(options);

      expect(result).toEqual({
        prompt: 'test prompt',
        negative_prompt: 'test negative prompt',
        steps: 30,
        cfg_scale: 8.0,
        width: 768,
        height: 768,
        seed: 12345,
        model: 'test-model',
        sampler: 'Euler a',
        batch_size: 2,
        enable_hr: true,
        hr_scale: 2.0,
        hr_upscaler: 'Latent',
        hr_steps: 0,
        hr_denoising_strength: 0.7
      });
    });

    it('should use default values for missing options', () => {
      const options = {
        prompt: 'test prompt'
      };

      const result = aiService.convertToGenerationParams(options);

      expect(result).toEqual({
        prompt: 'test prompt',
        negative_prompt: '',
        steps: 20,
        cfg_scale: 7.0,
        width: 512,
        height: 512,
        seed: -1,
        model: '',
        sampler: 'DPM++ 2M Karras',
        batch_size: 1,
        enable_hr: false,
        hr_scale: 2.0,
        hr_upscaler: 'Latent',
        hr_steps: 0,
        hr_denoising_strength: 0.7
      });
    });
  });

  describe('getTaskStatus', () => {
    it('should get task status successfully', async () => {
      const mockTask = {
        id: 'test-task-id',
        status: 'running',
        progress: 50,
        message: 'Task in progress',
        result: null,
        created_at: '2023-01-01T00:00:00Z',
        updated_at: '2023-01-01T00:00:30Z'
      };

      (apiService.getTask as jest.Mock).mockResolvedValueOnce(mockTask);

      const result = await aiService.getTaskStatus('test-task-id');

      expect(result).toEqual({
        taskId: 'test-task-id',
        status: 'running',
        progress: 50,
        message: 'Task in progress',
        images: [],
        result: null,
        createdAt: '2023-01-01T00:00:00Z',
        completedAt: undefined
      });
      expect(apiService.getTask).toHaveBeenCalledWith('test-task-id');
    });

    it('should handle task status errors', async () => {
      const error = new Error('Task not found');
      (apiService.getTask as jest.Mock).mockRejectedValueOnce(error);

      await expect(aiService.getTaskStatus('test-task-id')).rejects.toThrow('获取任务状态失败: Task not found');
    });
  });

  describe('cancelTask', () => {
    it('should cancel task successfully', async () => {
      (apiService.cancelTask as jest.Mock).mockResolvedValueOnce(undefined);

      await aiService.cancelTask('test-task-id');

      expect(apiService.cancelTask).toHaveBeenCalledWith('test-task-id');
    });

    it('should handle cancel task errors', async () => {
      const error = new Error('Cancel failed');
      (apiService.cancelTask as jest.Mock).mockRejectedValueOnce(error);

      await expect(aiService.cancelTask('test-task-id')).rejects.toThrow('取消任务失败: Cancel failed');
    });
  });

  describe('validateGenerationParams', () => {
    it('should validate valid parameters', () => {
      const params: GenerationParams = {
        prompt: 'test prompt',
        negative_prompt: 'test negative prompt',
        steps: 20,
        cfg_scale: 7.0,
        width: 512,
        height: 512,
        seed: -1,
        model: 'test-model',
        sampler: 'DPM++ 2M Karras',
        batch_size: 1,
        enable_hr: false,
        hr_scale: 2.0,
        hr_upscaler: 'Latent',
        hr_steps: 0,
        hr_denoising_strength: 0.7
      };

      const result = aiService.validateGenerationParams(params);

      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should validate invalid parameters', () => {
      const params: GenerationParams = {
        prompt: '', // 空提示词
        negative_prompt: 'test negative prompt',
        steps: 200, // 超出范围
        cfg_scale: 50, // 超出范围
        width: 50, // 超出范围
        height: 50, // 超出范围
        seed: -1,
        model: 'test-model',
        sampler: 'DPM++ 2M Karras',
        batch_size: 10, // 超出范围
        enable_hr: true,
        hr_scale: 5.0, // 超出范围
        hr_upscaler: 'Latent',
        hr_steps: 0,
        hr_denoising_strength: 0.7
      };

      const result = aiService.validateGenerationParams(params);

      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors).toContain('提示词不能为空');
      expect(result.errors).toContain('步数必须在1-150之间');
      expect(result.errors).toContain('CFG Scale必须在1-30之间');
      expect(result.errors).toContain('宽度必须在64-2048之间');
      expect(result.errors).toContain('高度必须在64-2048之间');
      expect(result.errors).toContain('批次大小必须在1-8之间');
      expect(result.errors).toContain('高分辨率修复倍数必须在1-4之间');
    });
  });

  describe('generateRandomSeed', () => {
    it('should generate random seed', () => {
      const seed1 = aiService.generateRandomSeed();
      const seed2 = aiService.generateRandomSeed();

      expect(typeof seed1).toBe('number');
      expect(seed1).toBeGreaterThanOrEqual(0);
      expect(seed1).toBeLessThan(1000000);
      expect(seed1).not.toBe(seed2); // 两次生成的种子应该不同
    });
  });

  describe('formatGenerationTime', () => {
    it('should format generation time correctly', () => {
      const createdAt = '2023-01-01T00:00:00Z';
      const completedAt = '2023-01-01T00:01:30Z';

      const result = aiService.formatGenerationTime(createdAt, completedAt);

      expect(result).toBe('1分钟');
    });

    it('should handle ongoing generation', () => {
      const createdAt = '2023-01-01T00:00:00Z';

      const result = aiService.formatGenerationTime(createdAt);

      expect(result).toBe('进行中...');
    });
  });

  describe('getPresetConfigs', () => {
    it('should return preset configs', () => {
      const presets = aiService.getPresetConfigs();

      expect(Array.isArray(presets)).toBe(true);
      expect(presets.length).toBeGreaterThan(0);
      expect(presets[0]).toHaveProperty('name');
      expect(presets[0]).toHaveProperty('params');
    });
  });

  describe('applyPresetConfig', () => {
    it('should apply preset config', () => {
      const params: GenerationParams = {
        prompt: 'test prompt',
        negative_prompt: 'test negative prompt',
        steps: 20,
        cfg_scale: 7.0,
        width: 512,
        height: 512,
        seed: -1,
        model: 'test-model',
        sampler: 'DPM++ 2M Karras',
        batch_size: 1,
        enable_hr: false,
        hr_scale: 2.0,
        hr_upscaler: 'Latent',
        hr_steps: 0,
        hr_denoising_strength: 0.7
      };

      const result = aiService.applyPresetConfig(params, '快速生成');

      expect(result).toEqual({
        ...params,
        steps: 10,
        cfg_scale: 7.0,
        sampler: 'Euler a'
      });
    });

    it('should return original params for unknown preset', () => {
      const params: GenerationParams = {
        prompt: 'test prompt',
        negative_prompt: 'test negative prompt',
        steps: 20,
        cfg_scale: 7.0,
        width: 512,
        height: 512,
        seed: -1,
        model: 'test-model',
        sampler: 'DPM++ 2M Karras',
        batch_size: 1,
        enable_hr: false,
        hr_scale: 2.0,
        hr_upscaler: 'Latent',
        hr_steps: 0,
        hr_denoising_strength: 0.7
      };

      const result = aiService.applyPresetConfig(params, 'Unknown Preset');

      expect(result).toEqual(params);
    });
  });
});
