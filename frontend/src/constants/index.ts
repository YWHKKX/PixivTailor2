// 应用常量定义

// ==================== API 端点 ====================
export const API_ENDPOINTS = {
  // 配置管理
  CONFIG: '/api/config',
  CONFIGS: '/api/configs',
  CONFIG_CATEGORIES: '/api/configs/categories',
  
  // 任务管理
  TASKS: '/api/tasks',
  TASK_CANCEL: '/api/tasks/cancel',
  TASK_DELETE: '/api/tasks/delete',
  
  // 数据管理
  DATA: '/api/data',
  IMAGES: '/api/images',
  IMAGE_DOWNLOAD: '/api/images/download',
  IMAGE_DELETE: '/api/images/delete',
  
  // 系统管理
  SYSTEM: '/api/system',
  SYSTEM_STATUS: '/api/system/status',
  SYSTEM_INFO: '/api/system/info',
  
  // WebUI 管理
  WEBUI: '/api/webui',
  WEBUI_START: '/api/webui/start',
  WEBUI_STOP: '/api/webui/stop',
  WEBUI_STATUS: '/api/webui/status',
  WEBUI_LOGS: '/api/webui/logs',
  
  // 爬虫管理
  CRAWL: '/api/crawl',
  CRAWL_CREATE: '/api/crawl/create',
  CRAWL_TASKS: '/api/crawl/tasks',
  CRAWL_TASK_STOP: '/api/crawl/tasks/stop',
  CRAWL_TASK_DELETE: '/api/crawl/tasks/delete',
  
  // AI 生成
  GENERATE: '/api/generate',
  GENERATE_WITH_CONFIG: '/api/generate-with-config',
} as const;

// ==================== 任务类型 ====================
export const TASK_TYPES = {
  CRAWL: 'crawl',
  GENERATE: 'generate',
  TRAIN: 'train',
  CLASSIFY: 'classify',
  BATCH: 'batch',
} as const;

// ==================== 任务状态 ====================
export const TASK_STATUS = {
  PENDING: 'pending',
  RUNNING: 'running',
  COMPLETED: 'completed',
  FAILED: 'failed',
  CANCELLED: 'cancelled',
} as const;

// ==================== WebUI 状态 ====================
export const WEBUI_STATUS = {
  STOPPED: 'stopped',
  STARTING: 'starting',
  RUNNING: 'running',
  EXTERNAL: 'external',
  ERROR: 'error',
} as const;

// ==================== 分页配置 ====================
export const PAGINATION = {
  DEFAULT_PAGE_SIZE: 20,
  MAX_PAGE_SIZE: 100,
  MIN_PAGE_SIZE: 5,
} as const;

// ==================== 路由配置 ====================
export const ROUTES = {
  HOME: '/',
  AI_GENERATOR: '/ai-generator',
  CRAWLER: '/crawler',
  HISTORY: '/history',
  CONFIG_MANAGER: '/config-manager',
  SETTINGS: '/settings',
} as const;

// ==================== 存储键名 ====================
export const STORAGE_KEYS = {
  THEME: 'pixiv-tailor-theme',
  LANGUAGE: 'pixiv-tailor-language',
  USER_PREFERENCES: 'pixiv-tailor-preferences',
  AUTH_TOKEN: 'pixiv-tailor-auth-token',
  USER_INFO: 'pixiv-tailor-user-info',
  GENERATION_PARAMS: 'pixiv-tailor-generation-params',
  CRAWL_CONFIG: 'pixiv-tailor-crawl-config',
} as const;

// ==================== 文件类型 ====================
export const FILE_TYPES = {
  IMAGE: ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp'],
  VIDEO: ['mp4', 'avi', 'mov', 'wmv', 'flv', 'webm'],
  AUDIO: ['mp3', 'wav', 'flac', 'aac', 'ogg'],
  DOCUMENT: ['pdf', 'doc', 'docx', 'txt', 'md'],
  ARCHIVE: ['zip', 'rar', '7z', 'tar', 'gz'],
} as const;

// ==================== 图片质量 ====================
export const IMAGE_QUALITY = {
  ORIGINAL: 'original',
  LARGE: 'large',
  MEDIUM: 'medium',
  SMALL: 'small',
} as const;

// ==================== 内容评级 ====================
export const CONTENT_RATING = {
  ALL: 'all',
  SAFE: 'safe',
  QUESTIONABLE: 'questionable',
  EXPLICIT: 'explicit',
} as const;

// ==================== 采样器 ====================
export const SAMPLERS = [
  'DPM++ 2M Karras',
  'DPM++ SDE Karras',
  'Euler a',
  'Euler',
  'LMS',
  'Heun',
  'DPM2',
  'DPM2 a',
  'DPM++ 2S a',
  'DPM++ 2M',
  'DPM++ SDE',
  'DPM fast',
  'DPM adaptive',
  'LMS Karras',
  'DPM2 Karras',
  'DPM2 a Karras',
  'DPM++ 2S a Karras',
] as const;

// ==================== 高分辨率修复器 ====================
export const HR_UPSCALERS = [
  'Latent',
  'ESRGAN_4x',
  'LDSR',
  'ScuNET',
  'SwinIR',
  'GFPGAN',
  'RealESRGAN',
  'CodeFormers',
] as const;

// ==================== 主题配置 ====================
export const THEMES = {
  LIGHT: 'light',
  DARK: 'dark',
  AUTO: 'auto',
} as const;

// ==================== 语言配置 ====================
export const LANGUAGES = {
  ZH_CN: 'zh-CN',
  EN_US: 'en-US',
} as const;

// ==================== 通知类型 ====================
export const NOTIFICATION_TYPES = {
  SUCCESS: 'success',
  ERROR: 'error',
  WARNING: 'warning',
  INFO: 'info',
} as const;

// ==================== 日志级别 ====================
export const LOG_LEVELS = {
  DEBUG: 'debug',
  INFO: 'info',
  WARN: 'warn',
  ERROR: 'error',
} as const;

// ==================== 工具函数 ====================
export const getTaskStatusColor = (status: string): string => {
  switch (status) {
    case TASK_STATUS.PENDING:
      return '#faad14';
    case TASK_STATUS.RUNNING:
      return '#1890ff';
    case TASK_STATUS.COMPLETED:
      return '#52c41a';
    case TASK_STATUS.FAILED:
      return '#ff4d4f';
    case TASK_STATUS.CANCELLED:
      return '#d9d9d9';
    default:
      return '#d9d9d9';
  }
};

export const getWebUIStatusColor = (status: string): string => {
  switch (status) {
    case WEBUI_STATUS.RUNNING:
    case WEBUI_STATUS.EXTERNAL:
      return '#52c41a';
    case WEBUI_STATUS.STARTING:
      return '#faad14';
    case WEBUI_STATUS.STOPPED:
      return '#ff4d4f';
    case WEBUI_STATUS.ERROR:
      return '#ff4d4f';
    default:
      return '#d9d9d9';
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