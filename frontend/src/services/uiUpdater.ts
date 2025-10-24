// UI 更新管理器 - 基于 plan.md 设计
import { Task, GenerationParams, TaskStatus } from './appState';
import { wsManager } from './websocket';
import { apiService } from './api';
import { appStateManager } from './appState';

class UIUpdater {
  private elements: {
    progressBar: HTMLElement | null;
    statusDisplay: HTMLElement | null;
    generateBtn: HTMLElement | null;
    stopBtn: HTMLElement | null;
    gallery: HTMLElement | null;
  } = {
    progressBar: null,
    statusDisplay: null,
    generateBtn: null,
    stopBtn: null,
    gallery: null,
  };

  constructor() {
    this.initElements();
    this.setupEventListeners();
  }

  private initElements(): void {
    // 获取DOM元素
    this.elements.progressBar = document.querySelector('#progress_bar');
    this.elements.statusDisplay = document.querySelector('#status_display');
    this.elements.generateBtn = document.querySelector('#generate_btn');
    this.elements.stopBtn = document.querySelector('#stop_btn');
    this.elements.gallery = document.querySelector('#result_gallery');
  }

  private setupEventListeners(): void {
    // 监听应用状态变化
    appStateManager.on('task_updated', (task: Task) => {
      this.updateTaskUI(task);
    });

    appStateManager.on('generation_started', () => {
      this.setGeneratingState(true);
    });

    appStateManager.on('generation_stopped', () => {
      this.setGeneratingState(false);
    });

    appStateManager.on('system_status_changed', (status: string) => {
      this.updateSystemStatus(status);
    });

    // 监听WebSocket事件
    wsManager.on('taskUpdate', (task: Task) => {
      appStateManager.updateTask(task);
    });

    wsManager.on('progressUpdate', (data: any) => {
      this.updateProgress(data);
    });

    wsManager.on('statusUpdate', (data: any) => {
      this.updateStatus(data);
    });

    wsManager.on('connected', () => {
      appStateManager.updateSystemStatus('connected');
    });

    wsManager.on('disconnected', () => {
      appStateManager.updateSystemStatus('disconnected');
    });

    wsManager.on('error', (error: any) => {
      appStateManager.updateSystemStatus('error');
      this.showError(error.message || 'Connection error');
    });

    // 绑定按钮事件
    this.bindButtonEvents();
  }

  private bindButtonEvents(): void {
    // 生成按钮
    if (this.elements.generateBtn) {
      this.elements.generateBtn.addEventListener('click', () => this.handleGenerate());
    }

    // 停止按钮
    if (this.elements.stopBtn) {
      this.elements.stopBtn.addEventListener('click', () => this.handleStop());
    }

    // 清空按钮
    const clearBtn = document.querySelector('#clear_btn');
    if (clearBtn) {
      clearBtn.addEventListener('click', () => this.handleClear());
    }

    // 下载按钮
    const downloadBtn = document.querySelector('#download_all_btn');
    if (downloadBtn) {
      downloadBtn.addEventListener('click', () => this.handleDownload());
    }

    // 保存配置按钮
    const saveConfigBtn = document.querySelector('#save_config_btn');
    if (saveConfigBtn) {
      saveConfigBtn.addEventListener('click', () => this.handleSaveConfig());
    }
  }

  private async handleGenerate(): Promise<void> {
    try {
      // 收集表单数据
      const params = this.collectFormData();
      
      // 验证参数
      if (!params.prompt.trim()) {
        this.showError('请输入提示词');
        return;
      }

      // 开始生成
      appStateManager.startGeneration(params);
      
      // 调用API
      const result = await apiService.generateImage(params);
      
      // 设置当前任务
      const task: Task = {
        id: result.task_id,
        name: '图像生成任务',
        type: 'generate',
        status: result.status as TaskStatus,
        config: JSON.stringify(params),
        progress: result.progress,
        created_at: result.created_at,
        updated_at: result.created_at,
        completed_at: result.completed_at,
        error_message: result.message,
      };
      
      appStateManager.setCurrentTask(task);
      appStateManager.addTask(task);

    } catch (error) {
      console.error('生成失败:', error);
      this.showError(error instanceof Error ? error.message : '生成失败');
      appStateManager.stopGeneration();
    }
  }

  private async handleStop(): Promise<void> {
    try {
      const currentTask = appStateManager.getState().currentTask;
      if (currentTask) {
        await apiService.cancelTask(currentTask.id.toString());
        appStateManager.stopGeneration();
      }
    } catch (error) {
      console.error('停止任务失败:', error);
      this.showError('停止任务失败');
    }
  }

  private handleClear(): void {
    // 清空画廊
    if (this.elements.gallery) {
      this.elements.gallery.innerHTML = `
        <div class="gallery-placeholder">
          <div class="placeholder-content">
            <div class="placeholder-icon">🎨</div>
            <div class="placeholder-text">生成结果将显示在这里</div>
          </div>
        </div>
      `;
    }

    // 重置状态显示
    if (this.elements.statusDisplay) {
      (this.elements.statusDisplay as HTMLInputElement).value = '就绪';
    }

    // 重置进度条
    if (this.elements.progressBar) {
      (this.elements.progressBar as any).percent = 0;
    }

    // 清除已完成的任务
    appStateManager.clearCompletedTasks();
  }

  private handleDownload(): void {
    const currentTask = appStateManager.getState().currentTask;
    if (currentTask) {
      // 实现下载逻辑
      console.log('下载任务结果:', currentTask);
      this.showMessage('下载功能开发中...');
    } else {
      this.showError('没有可下载的内容');
    }
  }

  private handleSaveConfig(): void {
    const params = this.collectFormData();
    // 保存到本地存储
    localStorage.setItem('gradio_config', JSON.stringify(params));
    this.showMessage('配置已保存');
  }

  private collectFormData(): GenerationParams {
    const prompt = (document.querySelector('#prompt_input') as HTMLTextAreaElement)?.value || '';
    const negativePrompt = (document.querySelector('#negative_prompt_input') as HTMLTextAreaElement)?.value || '';
    const steps = parseInt((document.querySelector('#steps_slider') as any)?.value || '20');
    const cfgScale = parseFloat((document.querySelector('#cfg_scale_slider') as any)?.value || '7.0');
    const width = parseInt((document.querySelector('#width_slider') as any)?.value || '512');
    const height = parseInt((document.querySelector('#height_slider') as any)?.value || '512');
    const seed = parseInt((document.querySelector('#seed_input') as HTMLInputElement)?.value || '-1');
    const model = (document.querySelector('#model_dropdown') as any)?.value || 'Model A';
    const sampler = (document.querySelector('#sampler_dropdown') as any)?.value || 'Euler';
    const batchSize = parseInt((document.querySelector('#batch_size_slider') as any)?.value || '1');
    const enableHr = (document.querySelector('#enable_hr_checkbox') as any)?.checked || false;

    return {
      prompt,
      negative_prompt: negativePrompt,
      steps,
      cfg_scale: cfgScale,
      width,
      height,
      seed,
      model,
      sampler,
      batch_size: batchSize,
      enable_hr: enableHr,
    };
  }

  private updateTaskUI(task: Task): void {
    // 更新进度条
    this.updateProgress({ progress: task.progress, message: task.error_message || '' });
    
    // 更新状态显示
    this.updateStatus({ status: task.status, message: task.error_message || '' });
    
    // 更新按钮状态
    this.setGeneratingState(task.status === 'running');
    
    // 更新结果展示
    if (task.status === 'completed') {
      this.updateGallery([]);
    }
  }

  private updateProgress(data: { progress: number; message: string }): void {
    if (this.elements.progressBar) {
      // 更新进度条
      const progressElement = this.elements.progressBar as any;
      if (progressElement.setPercent) {
        progressElement.setPercent(data.progress);
      }
    }
    
    // 更新消息
    if (this.elements.statusDisplay) {
      (this.elements.statusDisplay as HTMLInputElement).value = data.message;
    }
  }

  private updateStatus(data: { status: string; message: string }): void {
    if (this.elements.statusDisplay) {
      const statusElement = this.elements.statusDisplay as HTMLInputElement;
      statusElement.value = data.message;
      
      // 更新状态样式
      statusElement.className = `status-display ${data.status}`;
    }
  }

  private setGeneratingState(isGenerating: boolean): void {
    if (this.elements.generateBtn) {
      const btn = this.elements.generateBtn as HTMLButtonElement;
      btn.disabled = isGenerating;
      btn.textContent = isGenerating ? '生成中...' : '生成';
    }
    
    if (this.elements.stopBtn) {
      const btn = this.elements.stopBtn as HTMLButtonElement;
      btn.disabled = !isGenerating;
    }
  }

  private updateGallery(images: any[]): void {
    if (this.elements.gallery && images && images.length > 0) {
      // 更新画廊显示
      this.elements.gallery.innerHTML = images.map((image, index) => `
        <div class="gallery-item" data-index="${index}">
          <img src="${image}" alt="Generated Image ${index + 1}" />
          <div class="gallery-overlay">
            <button class="download-btn" onclick="downloadImage('${image}', ${index})">⬇️</button>
            <button class="fullscreen-btn" onclick="openFullscreen('${image}')">🔍</button>
          </div>
        </div>
      `).join('');
    }
  }

  private updateSystemStatus(status: string): void {
    // 更新系统状态指示器
    const statusIndicator = document.querySelector('.system-status');
    if (statusIndicator) {
      statusIndicator.className = `system-status ${status}`;
      statusIndicator.textContent = status === 'connected' ? '🟢 已连接' : 
                                   status === 'disconnected' ? '🔴 未连接' : '🟡 错误';
    }
  }

  private showError(message: string): void {
    console.error('UI错误:', message);
    // 可以添加更复杂的错误显示逻辑
    if (this.elements.statusDisplay) {
      (this.elements.statusDisplay as HTMLInputElement).value = `错误: ${message}`;
    }
  }

  private showMessage(message: string): void {
    console.log('UI消息:', message);
    // 可以添加消息显示逻辑
  }
}

// 创建单例实例
export const uiUpdater = new UIUpdater();

export default uiUpdater;
