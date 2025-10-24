// 全局类型定义

export interface Task {
  id: string;
  type: string;
  status: string;
  progress: number;
  createdAt: string;
  updatedAt: string;
  params: any;
}

export interface SystemInfo {
  version: string;
  status: string;
  uptime: string;
  memory: {
    used: number;
    total: number;
    usage: number;
    available: number;
  };
  cpu: {
    usage: number;
    average: number;
    peak: number;
  };
  disk: {
    data: number;
    dataUsed: number;
    dataTotal: number;
    models: number;
    modelsUsed: number;
    modelsTotal: number;
    logs: number;
    logsUsed: number;
    logsTotal: number;
  };
  activeConnections: number;
  warnings: string[];
}

export interface ConfigModule {
  name: string;
  version: string;
  enabled: boolean;
  config: any;
}

export interface Pagination {
  page: number;
  page_size: number;
  total: number;
}

export interface ApiResponse<T = any> {
  status: {
    code: number;
    message: string;
    details?: string;
  };
  data?: T;
  pagination?: Pagination;
}
