// 应用状态管理器 - 统一管理应用状态和类型定义

// ==================== 生成参数类型 ====================
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
  batch_count: number;
  loop_count: number; // 循环数量 - 发包次数
  enable_hr: boolean;
  hr_scale: number;
  hr_upscaler: string;
  hr_steps: number;
  hr_denoising_strength: number;
  
  // LoRA配置
  loras?: LoraConfig[];
  
  // VAE配置
  vae?: string;
  
  // 其他参数
  restore_faces?: boolean;
  tiling?: boolean;
  clip_skip?: number;
  eta?: number;
  ensd?: number;
  
  // 输出设置
  save_images?: boolean;
  save_grid?: boolean;
  send_images?: boolean;
  do_not_save_grid?: boolean;
}

// ==================== 配置类型 ====================
export interface LoraConfig {
  name: string;
  lora_key: string;
  full_name: string;
  weight: number;
  path: string;
  description?: string;
  tags?: string[];
  extend_tags?: string[];
  use_mask?: boolean;
}

export interface VaeConfig {
  name: string;
  path: string;
}

export interface SamplerConfig {
  name: string;
  aliases: string[];
}

export interface ModelConfig {
  name: string;
  path: string;
  type: 'checkpoint' | 'lora' | 'vae' | 'embedding';
  size?: number;
  created_at?: string;
}

// ==================== 任务状态类型 ====================
export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

// 爬虫相关类型
export type CrawlType = 'tag' | 'user' | 'illust';
export type Order = 'date_d' | 'popular_d';
export type Mode = 'safe' | 'r18' | 'all';

export interface Task {
  id: string;
  type: 'generation' | 'crawl' | 'batch' | 'tag';
  status: TaskStatus;
  progress: number;
  message: string;
  result?: any;
  error?: string;
  error_message?: string;
  stage?: string; // 任务阶段信息
  created_at: string;
  started_at?: string;
  completed_at?: string;
  updated_at: string;
  
  // 图片统计字段
  images_generated?: number;
  images_success?: number;
  
  // 任务配置
  config?: string;
  
  // 任务特定数据
  params?: GenerationParams;
  config_id?: string;
  crawl_config?: CrawlRequest;
  
  // 统计信息
  total_items?: number;
  processed_items?: number;
  failed_items?: number;
  images_found?: number;
  images_downloaded?: number;
  name?: string;
}

// ==================== 爬虫相关类型 ====================
export interface CrawlRequest {
  task_id?: string;
  tags?: string[];
  date_range?: {
    start: string;
    end: string;
  };
  rating?: 'all' | 'safe' | 'questionable' | 'explicit';
  limit: number;
  max_images?: number; // 图片数量限制，0表示不限制
  quality?: 'original' | 'large' | 'medium' | 'small';
  save_path?: string;
  download_images?: boolean;
  download_metadata?: boolean;
  filter_duplicates?: boolean;
  custom_filters?: {
    min_score?: number;
    max_score?: number;
    min_bookmarks?: number;
    max_bookmarks?: number;
    min_views?: number;
    max_views?: number;
  };
  // Pixiv爬虫特定字段
  type?: 'tag' | 'user' | 'illust';
  query?: string;
  user_id?: number;
  illust_id?: number;
  order?: string;
  mode?: string;
  delay?: number;
  proxy_enabled?: boolean;
  proxy_url?: string;
  cookie?: string;
}

export interface CrawlConfig {
  id: string;
  name: string;
  description?: string;
  tags: string[];
  date_range: {
    start: string;
    end: string;
  };
  rating: 'all' | 'safe' | 'questionable' | 'explicit';
  limit: number;
  quality: 'original' | 'large' | 'medium' | 'small';
  save_path: string;
  download_images: boolean;
  download_metadata: boolean;
  filter_duplicates: boolean;
  custom_filters?: {
    min_score?: number;
    max_score?: number;
    min_bookmarks?: number;
    max_bookmarks?: number;
    min_views?: number;
    max_views?: number;
  };
  created_at: string;
  updated_at: string;
}

// ==================== 图片相关类型 ====================
export interface PixivImage {
  id: string;
  pixiv_id: string;
  title: string;
  author: string;
  author_id: string;
  tags: string[];
  rating: 'safe' | 'questionable' | 'explicit';
  score: number;
  bookmarks: number;
  views: number;
  width: number;
  height: number;
  file_size: number;
  file_type: string;
  file_path: string;
  thumbnail_path: string;
  metadata: {
  created_at: string;
  updated_at: string;
    source_url: string;
    pixiv_url: string;
    author_url: string;
  };
  task_id: string;
  downloaded_at: string;
}

export interface GeneratedImage {
  id: string;
  task_id: string;
  prompt: string;
  negative_prompt: string;
  params: GenerationParams;
  image_path: string;
  thumbnail_path: string;
  file_size: number;
  width: number;
  height: number;
  created_at: string;
  metadata: {
    model: string;
    sampler: string;
    steps: number;
    cfg_scale: number;
    seed: number;
    batch_size: number;
  };
}

// ==================== 系统状态类型 ====================
export interface SystemStatus {
  status: 'running' | 'stopped' | 'error';
  uptime: number;
  memory_usage: number;
  disk_usage: number;
  cpu_usage: number;
  active_tasks: number;
  total_tasks: number;
  webui_status: {
    status: string;
    port_open: boolean;
    api_responding: boolean;
    process_id: boolean;
    managed: boolean;
  };
  last_updated: string;
}

export interface SystemInfo {
  version: string;
  build_date: string;
  go_version: string;
  platform: string;
  architecture: string;
  uptime: number;
  memory_usage: {
    total: number;
    used: number;
    free: number;
    percentage: number;
  };
  disk_usage: {
    total: number;
    used: number;
    free: number;
    percentage: number;
  };
  cpu_usage: number;
  active_tasks: number;
  total_tasks: number;
}

// ==================== 配置管理类型 ====================
export interface AppConfig {
  // 应用设置
  theme: 'light' | 'dark' | 'auto';
  language: 'zh-CN' | 'en-US';
  auto_save: boolean;
  auto_save_interval: number;
  
  // 生成设置
  default_steps: number;
  default_cfg_scale: number;
  default_width: number;
  default_height: number;
  default_sampler: string;
  default_batch_size: number;
  
  // 爬虫设置
  default_rating: 'all' | 'safe' | 'questionable' | 'explicit';
  default_quality: 'original' | 'large' | 'medium' | 'small';
  default_limit: number;
  max_concurrent_downloads: number;
  
  // 存储设置
  default_save_path: string;
  auto_create_folders: boolean;
  organize_by_date: boolean;
  organize_by_author: boolean;
  
  // 通知设置
  enable_notifications: boolean;
  notification_sound: boolean;
  notification_desktop: boolean;
  
  // 高级设置
  debug_mode: boolean;
  log_level: 'debug' | 'info' | 'warn' | 'error';
  max_log_lines: number;
  auto_cleanup_logs: boolean;
}

// ==================== 事件类型 ====================
export interface AppEvent {
  type: string;
  data?: any;
  timestamp: number;
}

export interface TaskEvent extends AppEvent {
  type: 'task_created' | 'task_started' | 'task_completed' | 'task_failed' | 'task_cancelled';
  data: {
    task_id: string;
    task_type: string;
    status: TaskStatus;
    progress: number;
    message: string;
  };
}

export interface SystemEvent extends AppEvent {
  type: 'system_started' | 'system_stopped' | 'system_error' | 'webui_status_changed';
  data: {
    status: string;
    message: string;
    details?: any;
  };
}

// ==================== 工具函数 ====================
export const getTaskStatusColor = (status: TaskStatus): string => {
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
};

export const getTaskStatusText = (status: TaskStatus): string => {
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
};

export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

export const formatDuration = (seconds: number): string => {
  if (seconds < 60) {
    return `${Math.round(seconds)}秒`;
  } else if (seconds < 3600) {
    return `${Math.round(seconds / 60)}分钟`;
  } else {
    return `${Math.round(seconds / 3600)}小时`;
  }
};

export const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
};

// ==================== 默认配置 ====================
export const DEFAULT_GENERATION_PARAMS: GenerationParams = {
        prompt: '',
        negative_prompt: '',
        steps: 20,
        cfg_scale: 7.0,
        width: 512,
        height: 512,
        seed: -1,
  model: '',
  sampler: 'DPM++ 2M Karras',
        batch_size: 1,
        batch_count: 1,
        loop_count: 1, // 默认循环1次
        enable_hr: false,
        hr_scale: 2.0,
        hr_upscaler: 'Latent',
        hr_steps: 0,
        hr_denoising_strength: 0.7,
        restore_faces: false,
        tiling: false,
        save_images: true,
  save_grid: false,
        send_images: true,
  do_not_save_grid: false
};

export const DEFAULT_APP_CONFIG: AppConfig = {
  theme: 'auto',
  language: 'zh-CN',
  auto_save: true,
  auto_save_interval: 30,
  default_steps: 20,
  default_cfg_scale: 7.0,
  default_width: 512,
  default_height: 512,
  default_sampler: 'DPM++ 2M Karras',
  default_batch_size: 1,
  default_rating: 'all',
  default_quality: 'original',
  default_limit: 100,
  max_concurrent_downloads: 5,
  default_save_path: './downloads',
  auto_create_folders: true,
  organize_by_date: true,
  organize_by_author: false,
  enable_notifications: true,
  notification_sound: true,
  notification_desktop: true,
  debug_mode: false,
  log_level: 'info',
  max_log_lines: 1000,
  auto_cleanup_logs: true
};