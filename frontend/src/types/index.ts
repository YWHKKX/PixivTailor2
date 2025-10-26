// 全局类型定义

// ==================== 基础类型 ====================
export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
export type TaskType = 'generation' | 'crawl' | 'batch' | 'train' | 'classify';
export type WebUIStatus = 'stopped' | 'starting' | 'running' | 'external' | 'error';
export type ContentRating = 'all' | 'safe' | 'questionable' | 'explicit';
export type ImageQuality = 'original' | 'large' | 'medium' | 'small';
export type Theme = 'light' | 'dark' | 'auto';
export type Language = 'zh-CN' | 'en-US';

// ==================== 任务相关 ====================
export interface Task {
  id: string;
  type: TaskType;
  status: TaskStatus;
  progress: number;
  message: string;
  result?: any;
  error?: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  updated_at: string;
  
  // 任务特定数据
  params?: any;
  config_id?: string;
  crawl_config?: CrawlRequest;
  
  // 统计信息
  total_items?: number;
  processed_items?: number;
  failed_items?: number;
}

// ==================== 系统信息 ====================
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
  warnings: string[];
}

export interface SystemStatus {
  status: 'running' | 'stopped' | 'error';
  uptime: number;
  memory_usage: number;
  disk_usage: number;
  cpu_usage: number;
  active_tasks: number;
  total_tasks: number;
  webui_status: {
    status: WebUIStatus;
    port_open: boolean;
    api_responding: boolean;
    process_id: boolean;
    managed: boolean;
  };
  last_updated: string;
}

// ==================== 配置相关 ====================
export interface ConfigModule {
  name: string;
  version: string;
  enabled: boolean;
  config: any;
}

export interface AppConfig {
  // 应用设置
  theme: Theme;
  language: Language;
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
  default_rating: ContentRating;
  default_quality: ImageQuality;
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

// ==================== 分页相关 ====================
export interface Pagination {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}

// ==================== API 响应 ====================
export interface ApiResponse<T = any> {
  status: {
    code: number;
    message: string;
    details?: string;
  };
  data?: T;
  pagination?: Pagination;
}

// ==================== 生成参数 ====================
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

export interface LoraConfig {
  name: string;
  weight: number;
  path?: string;
}

// ==================== 爬虫相关 ====================
export interface CrawlRequest {
  task_id?: string;
  tags: string[];
  date_range: {
    start: string;
    end: string;
  };
  rating: ContentRating;
  limit: number;
  quality: ImageQuality;
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
  rating: ContentRating;
  limit: number;
  quality: ImageQuality;
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

// ==================== 图片相关 ====================
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

// ==================== 用户相关 ====================
export interface User {
  id: string;
  username: string;
  email: string;
  avatar?: string;
  role: 'admin' | 'user';
  created_at: string;
  updated_at: string;
  last_login?: string;
}

export interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  token: string | null;
  loading: boolean;
  error: string | null;
}

// ==================== 事件相关 ====================
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

// ==================== WebSocket 相关 ====================
export interface WebSocketMessage {
  type: string;
  task_id?: string;
  data?: any;
  timestamp?: number;
}

export interface TaskUpdate {
  task_id: string;
  status: TaskStatus;
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
  status: TaskStatus;
  message: string;
}

export interface WebUIStatusUpdate {
  status: WebUIStatus;
  port_open: boolean;
  api_responding: boolean;
  process_id: boolean;
  managed: boolean;
}

// ==================== 组件 Props ====================
export interface BaseComponentProps {
  className?: string;
  style?: React.CSSProperties;
  children?: React.ReactNode;
}

export interface ModalProps extends BaseComponentProps {
  visible: boolean;
  onCancel: () => void;
  onOk?: () => void;
  title?: string;
  footer?: React.ReactNode;
  width?: number | string;
  height?: number | string;
  closable?: boolean;
  maskClosable?: boolean;
  destroyOnClose?: boolean;
}

export interface FormProps extends BaseComponentProps {
  initialValues?: any;
  onFinish?: (values: any) => void;
  onFinishFailed?: (errorInfo: any) => void;
  loading?: boolean;
  disabled?: boolean;
}

// ==================== 工具类型 ====================
export type Optional<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;
export type Required<T, K extends keyof T> = T & { [P in K]-?: T[P] };
export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};
export type NonNullable<T> = T extends null | undefined ? never : T;
export type ValueOf<T> = T[keyof T];
export type KeysOf<T> = keyof T;
export type ValuesOf<T> = T[KeysOf<T>];