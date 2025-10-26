import React from 'react';
import ReactDOM from 'react-dom/client';
import App from '@/App';

// 导入样式
import '@/styles/App.css';

// ==================== 应用初始化 ====================
const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);

// 渲染应用
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);

// ==================== 开发环境配置 ====================
if (import.meta.env.DEV) {
  // 开发环境下的全局配置
  console.log('🚀 PixivTailor 开发模式启动');
  
  // 启用 React DevTools
  if (window.__REACT_DEVTOOLS_GLOBAL_HOOK__) {
    window.__REACT_DEVTOOLS_GLOBAL_HOOK__.inject = () => {};
  }
  
  // 全局错误处理
  window.addEventListener('error', (event) => {
    console.error('全局错误:', event.error);
  });
  
  window.addEventListener('unhandledrejection', (event) => {
    console.error('未处理的 Promise 拒绝:', event.reason);
  });
}

// ==================== 生产环境配置 ====================
if (import.meta.env.PROD) {
  // 生产环境下的全局配置
  console.log('🚀 PixivTailor 生产模式启动');
  
  // 禁用 console.log 在生产环境
  if (process.env.NODE_ENV === 'production') {
    console.log = () => {};
    console.warn = () => {};
    console.info = () => {};
  }
}