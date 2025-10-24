// 应用常量定义

export const API_ENDPOINTS = {
  CONFIG: '/api/config',
  TASKS: '/api/tasks',
  DATA: '/api/data',
  SYSTEM: '/api/system',
} as const;

export const TASK_TYPES = {
  CRAWL: 'crawl',
  GENERATE: 'generate',
  TRAIN: 'train',
  CLASSIFY: 'classify',
} as const;

export const TASK_STATUS = {
  PENDING: 'pending',
  RUNNING: 'running',
  COMPLETED: 'completed',
  FAILED: 'failed',
  CANCELLED: 'cancelled',
} as const;

export const PAGINATION = {
  DEFAULT_PAGE_SIZE: 20,
  MAX_PAGE_SIZE: 100,
} as const;

export const ROUTES = {
  DASHBOARD: '/',
  TASKS: '/tasks',
  DATA: '/data',
  CONFIG: '/config',
  SYSTEM: '/system',
} as const;

export const STORAGE_KEYS = {
  THEME: 'pixiv-tailor-theme',
  LANGUAGE: 'pixiv-tailor-language',
  USER_PREFERENCES: 'pixiv-tailor-preferences',
} as const;
