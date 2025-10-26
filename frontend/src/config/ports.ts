/**
 * 服务器端口配置
 * 与后端服务器端口保持一致
 */

// ==================== 服务器端口配置 ====================
export const SERVER_PORTS = {
  // HTTP API端口
  HTTP: 50052,
  // gRPC端口
  GRPC: 50051,
  // WebUI端口
  WEBUI: 7860,
} as const;

// ==================== API 配置 ====================
export const API_BASE_URL = `http://localhost:${SERVER_PORTS.HTTP}/api`;

// ==================== WebSocket 配置 ====================
export const WS_BASE_URL = `ws://localhost:${SERVER_PORTS.HTTP}/ws`;

// ==================== 图片服务配置 ====================
export const IMAGE_BASE_URL = `http://localhost:${SERVER_PORTS.HTTP}/api/images`;

// ==================== WebUI 配置 ====================
export const WEBUI_BASE_URL = `http://localhost:${SERVER_PORTS.WEBUI}`;
export const WEBUI_API_URL = `${WEBUI_BASE_URL}/sdapi/v1`;

// ==================== 开发环境配置 ====================
export const DEV_CONFIG = {
  // 开发模式下的API URL
  API_URL: import.meta.env.VITE_API_URL || API_BASE_URL,
  // 开发模式下的WebSocket URL
  WS_URL: import.meta.env.VITE_WS_URL || WS_BASE_URL,
  // 开发模式下的WebUI URL
  WEBUI_URL: import.meta.env.VITE_WEBUI_URL || WEBUI_BASE_URL,
} as const;

// ==================== 导出类型 ====================
export type ServerPort = typeof SERVER_PORTS[keyof typeof SERVER_PORTS];

// ==================== 工具函数 ====================
export const getApiUrl = (endpoint: string): string => {
  return `${DEV_CONFIG.API_URL}${endpoint}`;
};

export const getWebSocketUrl = (): string => {
  return DEV_CONFIG.WS_URL;
};

export const getWebUIUrl = (endpoint: string = ''): string => {
  return `${DEV_CONFIG.WEBUI_URL}${endpoint}`;
};

export const getImageUrl = (imagePath: string): string => {
  return `${IMAGE_BASE_URL}/${imagePath}`;
};