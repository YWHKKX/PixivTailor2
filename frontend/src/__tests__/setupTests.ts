import '@testing-library/jest-dom';

// 全局测试设置
beforeEach(() => {
  // 清理所有模拟
  jest.clearAllMocks();
  
  // 重置所有模拟
  jest.resetAllMocks();
  
  // 清理 DOM
  document.body.innerHTML = '';
  
  // 清理本地存储
  localStorage.clear();
  sessionStorage.clear();
});

// 全局测试后清理
afterEach(() => {
  // 清理定时器
  jest.clearAllTimers();
  
  // 清理模拟
  jest.restoreAllMocks();
});

// 模拟 IntersectionObserver
global.IntersectionObserver = class IntersectionObserver {
  constructor() {}
  disconnect() {}
  observe() {}
  unobserve() {}
};

// 模拟 ResizeObserver
global.ResizeObserver = class ResizeObserver {
  constructor() {}
  disconnect() {}
  observe() {}
  unobserve() {}
};

// 模拟 matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(), // deprecated
    removeListener: jest.fn(), // deprecated
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});

// 模拟 scrollTo
Object.defineProperty(window, 'scrollTo', {
  writable: true,
  value: jest.fn(),
});

// 模拟 getComputedStyle
Object.defineProperty(window, 'getComputedStyle', {
  writable: true,
  value: jest.fn().mockImplementation(() => ({
    getPropertyValue: jest.fn(),
  })),
});

// 模拟 URL.createObjectURL
Object.defineProperty(URL, 'createObjectURL', {
  writable: true,
  value: jest.fn().mockImplementation(() => 'mock-url'),
});

// 模拟 URL.revokeObjectURL
Object.defineProperty(URL, 'revokeObjectURL', {
  writable: true,
  value: jest.fn(),
});

// 模拟 fetch
global.fetch = jest.fn();

// 模拟 WebSocket
global.WebSocket = class WebSocket {
  constructor() {}
  close() {}
  send() {}
  addEventListener() {}
  removeEventListener() {}
  dispatchEvent() {}
};

// 模拟 EventSource
global.EventSource = class EventSource {
  constructor() {}
  close() {}
  addEventListener() {}
  removeEventListener() {}
  dispatchEvent() {}
};

// 模拟 console 方法
const originalConsole = { ...console };
beforeEach(() => {
  console.log = jest.fn();
  console.warn = jest.fn();
  console.error = jest.fn();
  console.info = jest.fn();
});

afterEach(() => {
  console.log = originalConsole.log;
  console.warn = originalConsole.warn;
  console.error = originalConsole.error;
  console.info = originalConsole.info;
});

// 模拟环境变量
process.env.NODE_ENV = 'test';
process.env.REACT_APP_API_URL = 'http://localhost:50052/api';
process.env.REACT_APP_WS_URL = 'ws://localhost:50052/ws';

// 模拟 Ant Design 的 message 组件
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
    error: jest.fn(),
    warning: jest.fn(),
    info: jest.fn(),
    loading: jest.fn(),
  },
  notification: {
    success: jest.fn(),
    error: jest.fn(),
    warning: jest.fn(),
    info: jest.fn(),
    open: jest.fn(),
  },
  Modal: {
    confirm: jest.fn(),
    info: jest.fn(),
    success: jest.fn(),
    error: jest.fn(),
    warning: jest.fn(),
  },
}));

// 模拟 React Router
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => jest.fn(),
  useLocation: () => ({
    pathname: '/',
    search: '',
    hash: '',
    state: null,
  }),
  useParams: () => ({}),
  useSearchParams: () => [new URLSearchParams(), jest.fn()],
}));

// 模拟 WebSocket 服务
jest.mock('../services/websocket', () => ({
  wsManager: {
    connect: jest.fn(),
    disconnect: jest.fn(),
    send: jest.fn(),
    on: jest.fn(),
    off: jest.fn(),
    isWebSocketConnected: jest.fn(() => true),
    getConnectionId: jest.fn(() => 'test-connection-id'),
    getReconnectAttempts: jest.fn(() => 0),
  },
}));

// 模拟 API 服务
jest.mock('../services/api', () => ({
  apiService: {
    generateImages: jest.fn(),
    generateWithConfig: jest.fn(),
    startWebUI: jest.fn(),
    stopWebUI: jest.fn(),
    getWebUIStatus: jest.fn(),
    createWebUILogStream: jest.fn(),
    getConfigs: jest.fn(),
    getConfig: jest.fn(),
    createConfig: jest.fn(),
    updateConfig: jest.fn(),
    deleteConfig: jest.fn(),
    getConfigCategories: jest.fn(),
    getCrawlTasks: jest.fn(),
    getCrawlTask: jest.fn(),
    createCrawlTask: jest.fn(),
    stopCrawlTask: jest.fn(),
    deleteCrawlTask: jest.fn(),
    getImages: jest.fn(),
    downloadImage: jest.fn(),
    deleteImage: jest.fn(),
    getSystemStatus: jest.fn(),
    getSystemInfo: jest.fn(),
    getTasks: jest.fn(),
    getTask: jest.fn(),
    cancelTask: jest.fn(),
    deleteTask: jest.fn(),
    getImageUrl: jest.fn(),
    getTaskStatusColor: jest.fn(),
    getTaskStatusText: jest.fn(),
  },
}));

// 模拟 AI 服务
jest.mock('../services/aiService', () => ({
  aiService: {
    generateImages: jest.fn(),
    generateWithConfig: jest.fn(),
    convertToGenerationParams: jest.fn(),
    getTaskStatus: jest.fn(),
    cancelTask: jest.fn(),
    onGenerationComplete: jest.fn(),
    offGenerationComplete: jest.fn(),
    validateGenerationParams: jest.fn(),
    generateRandomSeed: jest.fn(),
    formatGenerationTime: jest.fn(),
    getPresetConfigs: jest.fn(),
    applyPresetConfig: jest.fn(),
  },
}));

// 模拟认证上下文
jest.mock('../contexts/AuthContext', () => ({
  useAuth: () => ({
    user: null,
    isAuthenticated: false,
    login: jest.fn(),
    logout: jest.fn(),
    loading: false,
    error: null,
  }),
  AuthProvider: ({ children }: { children: React.ReactNode }) => children,
}));

// 模拟错误边界
jest.mock('../components/ErrorBoundary', () => ({
  __esModule: true,
  default: ({ children }: { children: React.ReactNode }) => children,
}));

// 模拟主布局
jest.mock('../components/layout/MainLayout', () => ({
  __esModule: true,
  default: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

// 模拟登录模态框
jest.mock('../components/auth/LoginModal', () => ({
  __esModule: true,
  default: ({ visible, onCancel, onLogin }: any) => 
    visible ? <div data-testid="login-modal">Login Modal</div> : null,
}));

// 模拟页面组件
jest.mock('../pages/HomePage', () => ({
  __esModule: true,
  default: () => <div data-testid="home-page">Home Page</div>,
}));

jest.mock('../pages/AIGeneratorPage', () => ({
  __esModule: true,
  default: () => <div data-testid="ai-generator-page">AI Generator Page</div>,
}));

jest.mock('../pages/CrawlerPage', () => ({
  __esModule: true,
  default: () => <div data-testid="crawler-page">Crawler Page</div>,
}));

jest.mock('../pages/HistoryPage', () => ({
  __esModule: true,
  default: () => <div data-testid="history-page">History Page</div>,
}));

jest.mock('../pages/SettingsPage', () => ({
  __esModule: true,
  default: () => <div data-testid="settings-page">Settings Page</div>,
}));

jest.mock('../pages/ConfigManagerPage', () => ({
  __esModule: true,
  default: () => <div data-testid="config-manager-page">Config Manager Page</div>,
}));

// 模拟样式文件
jest.mock('../styles/App.css', () => ({}));
jest.mock('../styles/gradio.css', () => ({}));
jest.mock('../styles/theme.css', () => ({}));

// 模拟 Vite 环境变量
Object.defineProperty(import.meta, 'env', {
  value: {
    DEV: false,
    PROD: true,
    VITE_API_URL: 'http://localhost:50052/api',
    VITE_WS_URL: 'ws://localhost:50052/ws',
    VITE_WEBUI_URL: 'http://localhost:7860',
  },
  writable: true,
});

// 模拟 window 对象
Object.defineProperty(window, 'location', {
  value: {
    href: 'http://localhost:3000',
    origin: 'http://localhost:3000',
    protocol: 'http:',
    host: 'localhost:3000',
    hostname: 'localhost',
    port: '3000',
    pathname: '/',
    search: '',
    hash: '',
    assign: jest.fn(),
    replace: jest.fn(),
    reload: jest.fn(),
  },
  writable: true,
});

// 模拟 document 对象
Object.defineProperty(document, 'title', {
  value: 'PixivTailor',
  writable: true,
});

// 模拟 navigator 对象
Object.defineProperty(navigator, 'userAgent', {
  value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36',
  writable: true,
});

// 模拟 localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
  length: 0,
  key: jest.fn(),
};

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
  writable: true,
});

// 模拟 sessionStorage
const sessionStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
  length: 0,
  key: jest.fn(),
};

Object.defineProperty(window, 'sessionStorage', {
  value: sessionStorageMock,
  writable: true,
});

// 模拟 IntersectionObserver
global.IntersectionObserver = class IntersectionObserver {
  constructor() {}
  disconnect() {}
  observe() {}
  unobserve() {}
};

// 模拟 ResizeObserver
global.ResizeObserver = class ResizeObserver {
  constructor() {}
  disconnect() {}
  observe() {}
  unobserve() {}
};

// 模拟 matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(), // deprecated
    removeListener: jest.fn(), // deprecated
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});

// 模拟 scrollTo
Object.defineProperty(window, 'scrollTo', {
  writable: true,
  value: jest.fn(),
});

// 模拟 getComputedStyle
Object.defineProperty(window, 'getComputedStyle', {
  writable: true,
  value: jest.fn().mockImplementation(() => ({
    getPropertyValue: jest.fn(),
  })),
});

// 模拟 URL.createObjectURL
Object.defineProperty(URL, 'createObjectURL', {
  writable: true,
  value: jest.fn().mockImplementation(() => 'mock-url'),
});

// 模拟 URL.revokeObjectURL
Object.defineProperty(URL, 'revokeObjectURL', {
  writable: true,
  value: jest.fn(),
});

// 模拟 fetch
global.fetch = jest.fn();

// 模拟 WebSocket
global.WebSocket = class WebSocket {
  constructor() {}
  close() {}
  send() {}
  addEventListener() {}
  removeEventListener() {}
  dispatchEvent() {}
};

// 模拟 EventSource
global.EventSource = class EventSource {
  constructor() {}
  close() {}
  addEventListener() {}
  removeEventListener() {}
  dispatchEvent() {}
};

// 模拟 console 方法
const originalConsole = { ...console };
beforeEach(() => {
  console.log = jest.fn();
  console.warn = jest.fn();
  console.error = jest.fn();
  console.info = jest.fn();
});

afterEach(() => {
  console.log = originalConsole.log;
  console.warn = originalConsole.warn;
  console.error = originalConsole.error;
  console.info = originalConsole.info;
});

// 模拟环境变量
process.env.NODE_ENV = 'test';
process.env.REACT_APP_API_URL = 'http://localhost:50052/api';
process.env.REACT_APP_WS_URL = 'ws://localhost:50052/ws';

// 模拟 Ant Design 的 message 组件
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
    error: jest.fn(),
    warning: jest.fn(),
    info: jest.fn(),
    loading: jest.fn(),
  },
  notification: {
    success: jest.fn(),
    error: jest.fn(),
    warning: jest.fn(),
    info: jest.fn(),
    open: jest.fn(),
  },
  Modal: {
    confirm: jest.fn(),
    info: jest.fn(),
    success: jest.fn(),
    error: jest.fn(),
    warning: jest.fn(),
  },
}));

// 模拟 React Router
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => jest.fn(),
  useLocation: () => ({
    pathname: '/',
    search: '',
    hash: '',
    state: null,
  }),
  useParams: () => ({}),
  useSearchParams: () => [new URLSearchParams(), jest.fn()],
}));

// 模拟 WebSocket 服务
jest.mock('../services/websocket', () => ({
  wsManager: {
    connect: jest.fn(),
    disconnect: jest.fn(),
    send: jest.fn(),
    on: jest.fn(),
    off: jest.fn(),
    isWebSocketConnected: jest.fn(() => true),
    getConnectionId: jest.fn(() => 'test-connection-id'),
    getReconnectAttempts: jest.fn(() => 0),
  },
}));

// 模拟 API 服务
jest.mock('../services/api', () => ({
  apiService: {
    generateImages: jest.fn(),
    generateWithConfig: jest.fn(),
    startWebUI: jest.fn(),
    stopWebUI: jest.fn(),
    getWebUIStatus: jest.fn(),
    createWebUILogStream: jest.fn(),
    getConfigs: jest.fn(),
    getConfig: jest.fn(),
    createConfig: jest.fn(),
    updateConfig: jest.fn(),
    deleteConfig: jest.fn(),
    getConfigCategories: jest.fn(),
    getCrawlTasks: jest.fn(),
    getCrawlTask: jest.fn(),
    createCrawlTask: jest.fn(),
    stopCrawlTask: jest.fn(),
    deleteCrawlTask: jest.fn(),
    getImages: jest.fn(),
    downloadImage: jest.fn(),
    deleteImage: jest.fn(),
    getSystemStatus: jest.fn(),
    getSystemInfo: jest.fn(),
    getTasks: jest.fn(),
    getTask: jest.fn(),
    cancelTask: jest.fn(),
    deleteTask: jest.fn(),
    getImageUrl: jest.fn(),
    getTaskStatusColor: jest.fn(),
    getTaskStatusText: jest.fn(),
  },
}));

// 模拟 AI 服务
jest.mock('../services/aiService', () => ({
  aiService: {
    generateImages: jest.fn(),
    generateWithConfig: jest.fn(),
    convertToGenerationParams: jest.fn(),
    getTaskStatus: jest.fn(),
    cancelTask: jest.fn(),
    onGenerationComplete: jest.fn(),
    offGenerationComplete: jest.fn(),
    validateGenerationParams: jest.fn(),
    generateRandomSeed: jest.fn(),
    formatGenerationTime: jest.fn(),
    getPresetConfigs: jest.fn(),
    applyPresetConfig: jest.fn(),
  },
}));

// 模拟认证上下文
jest.mock('../contexts/AuthContext', () => ({
  useAuth: () => ({
    user: null,
    isAuthenticated: false,
    login: jest.fn(),
    logout: jest.fn(),
    loading: false,
    error: null,
  }),
  AuthProvider: ({ children }: { children: React.ReactNode }) => children,
}));

// 模拟错误边界
jest.mock('../components/ErrorBoundary', () => ({
  __esModule: true,
  default: ({ children }: { children: React.ReactNode }) => children,
}));

// 模拟主布局
jest.mock('../components/layout/MainLayout', () => ({
  __esModule: true,
  default: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

// 模拟登录模态框
jest.mock('../components/auth/LoginModal', () => ({
  __esModule: true,
  default: ({ visible, onCancel, onLogin }: any) => 
    visible ? <div data-testid="login-modal">Login Modal</div> : null,
}));

// 模拟页面组件
jest.mock('../pages/HomePage', () => ({
  __esModule: true,
  default: () => <div data-testid="home-page">Home Page</div>,
}));

jest.mock('../pages/AIGeneratorPage', () => ({
  __esModule: true,
  default: () => <div data-testid="ai-generator-page">AI Generator Page</div>,
}));

jest.mock('../pages/CrawlerPage', () => ({
  __esModule: true,
  default: () => <div data-testid="crawler-page">Crawler Page</div>,
}));

jest.mock('../pages/HistoryPage', () => ({
  __esModule: true,
  default: () => <div data-testid="history-page">History Page</div>,
}));

jest.mock('../pages/SettingsPage', () => ({
  __esModule: true,
  default: () => <div data-testid="settings-page">Settings Page</div>,
}));

jest.mock('../pages/ConfigManagerPage', () => ({
  __esModule: true,
  default: () => <div data-testid="config-manager-page">Config Manager Page</div>,
}));

// 模拟样式文件
jest.mock('../styles/App.css', () => ({}));
jest.mock('../styles/gradio.css', () => ({}));
jest.mock('../styles/theme.css', () => ({}));

// 模拟 Vite 环境变量
Object.defineProperty(import.meta, 'env', {
  value: {
    DEV: false,
    PROD: true,
    VITE_API_URL: 'http://localhost:50052/api',
    VITE_WS_URL: 'ws://localhost:50052/ws',
    VITE_WEBUI_URL: 'http://localhost:7860',
  },
  writable: true,
});

// 模拟 window 对象
Object.defineProperty(window, 'location', {
  value: {
    href: 'http://localhost:3000',
    origin: 'http://localhost:3000',
    protocol: 'http:',
    host: 'localhost:3000',
    hostname: 'localhost',
    port: '3000',
    pathname: '/',
    search: '',
    hash: '',
    assign: jest.fn(),
    replace: jest.fn(),
    reload: jest.fn(),
  },
  writable: true,
});

// 模拟 document 对象
Object.defineProperty(document, 'title', {
  value: 'PixivTailor',
  writable: true,
});

// 模拟 navigator 对象
Object.defineProperty(navigator, 'userAgent', {
  value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36',
  writable: true,
});

// 模拟 localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
  length: 0,
  key: jest.fn(),
};

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
  writable: true,
});

// 模拟 sessionStorage
const sessionStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
  length: 0,
  key: jest.fn(),
};

Object.defineProperty(window, 'sessionStorage', {
  value: sessionStorageMock,
  writable: true,
});