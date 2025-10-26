import { apiService } from '../../services/api';
import { GenerationParams } from '../../services/appState';

// 模拟 fetch
global.fetch = jest.fn();

describe('API Service', () => {
  beforeEach(() => {
    (global.fetch as jest.Mock).mockClear();
  });

  describe('generateImages', () => {
    it('should generate images successfully', async () => {
      const mockResponse = {
        task_id: 'test-task-id',
        status: 'completed',
        progress: 100,
        message: 'Generation completed',
        result: {
          images: ['image1.png', 'image2.png']
        },
        created_at: '2023-01-01T00:00:00Z',
        completed_at: '2023-01-01T00:01:00Z'
      };

      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: mockResponse })
      });

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

      const result = await apiService.generateImages(params);

      expect(result).toEqual(mockResponse);
      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:50052/api/generate',
        expect.objectContaining({
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(params)
        })
      );
    });

    it('should handle generation errors', async () => {
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        json: async () => ({ message: 'Generation failed' })
      });

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

      await expect(apiService.generateImages(params)).rejects.toThrow('Generation failed');
    });
  });

  describe('getWebUIStatus', () => {
    it('should get WebUI status successfully', async () => {
      const mockStatus = {
        status: 'running',
        port_open: true,
        api_responding: true,
        process_id: true,
        managed: true
      };

      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: mockStatus })
      });

      const result = await apiService.getWebUIStatus();

      expect(result).toEqual(mockStatus);
      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:50052/api/webui/status',
        expect.objectContaining({
          method: 'GET'
        })
      );
    });
  });

  describe('getConfigs', () => {
    it('should get configs successfully', async () => {
      const mockConfigs = [
        {
          id: 'config1',
          name: 'Test Config 1',
          category: 'test',
          prompt: 'test prompt',
          negative_prompt: 'test negative prompt',
          steps: 20,
          cfg_scale: 7.0,
          width: 512,
          height: 512,
          sampler: 'DPM++ 2M Karras',
          batch_size: 1,
          enable_hr: false,
          hr_scale: 2.0,
          hr_upscaler: 'Latent',
          hr_steps: 0,
          hr_denoising_strength: 0.7,
          created_at: '2023-01-01T00:00:00Z',
          updated_at: '2023-01-01T00:00:00Z'
        }
      ];

      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: mockConfigs })
      });

      const result = await apiService.getConfigs();

      expect(result).toEqual(mockConfigs);
      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:50052/api/configs',
        expect.objectContaining({
          method: 'GET'
        })
      );
    });
  });

  describe('createWebUILogStream', () => {
    it('should create EventSource for WebUI logs', () => {
      const stream = apiService.createWebUILogStream();

      expect(stream).toBeInstanceOf(EventSource);
      expect(stream.url).toBe('http://localhost:50052/api/webui/logs');
    });
  });

  describe('getTaskStatusColor', () => {
    it('should return correct colors for different task statuses', () => {
      expect(apiService.getTaskStatusColor('pending')).toBe('#faad14');
      expect(apiService.getTaskStatusColor('running')).toBe('#1890ff');
      expect(apiService.getTaskStatusColor('completed')).toBe('#52c41a');
      expect(apiService.getTaskStatusColor('failed')).toBe('#ff4d4f');
      expect(apiService.getTaskStatusColor('cancelled')).toBe('#d9d9d9');
      expect(apiService.getTaskStatusColor('unknown')).toBe('#d9d9d9');
    });
  });

  describe('getTaskStatusText', () => {
    it('should return correct text for different task statuses', () => {
      expect(apiService.getTaskStatusText('pending')).toBe('等待中');
      expect(apiService.getTaskStatusText('running')).toBe('运行中');
      expect(apiService.getTaskStatusText('completed')).toBe('已完成');
      expect(apiService.getTaskStatusText('failed')).toBe('失败');
      expect(apiService.getTaskStatusText('cancelled')).toBe('已取消');
      expect(apiService.getTaskStatusText('unknown')).toBe('未知');
    });
  });
});
