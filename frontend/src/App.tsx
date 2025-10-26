import React, { useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';

import ErrorBoundary from '@/components/ErrorBoundary';
import MainLayout from '@/components/layout/MainLayout';
import { AuthProvider } from '@/contexts/AuthContext';

// 导入页面组件
import HomePage from '@/pages/HomePage';
import AIGeneratorPage from '@/pages/AIGeneratorPage';
import CrawlerPage from '@/pages/CrawlerPage';
import HistoryPage from '@/pages/HistoryPage';
// 已移除 SettingsPage
import ConfigManagerPage from '@/pages/ConfigManagerPage';

// 导入服务
import { wsManager } from '@/services/websocket';

// 导入样式
import './styles/gradio.css';
import './styles/theme.css';

// ==================== 主应用组件 ====================
const App: React.FC = () => {
    useEffect(() => {
        // 在开发模式下延迟初始化 WebSocket 连接，避免热重载频繁重连
        const isDevelopment = import.meta.env.DEV;
        const delay = isDevelopment ? 2000 : 0; // 开发模式延迟2秒

        console.log('App 组件挂载，初始化 WebSocket 连接', isDevelopment ? '(开发模式，延迟连接)' : '');

        const timer = setTimeout(() => {
            // 初始化 WebSocket 连接
            const wsUrl = import.meta.env['VITE_WS_URL'] || 'ws://localhost:50052/ws';
            wsManager.connect(wsUrl);

            // 设置连接事件监听
            wsManager.on('connected', () => {
                console.log('WebSocket 连接已建立');
            });

            wsManager.on('disconnected', (data: any) => {
                console.log('WebSocket 连接已断开:', data);
            });

            wsManager.on('error', (error: any) => {
                console.error('WebSocket 连接错误:', error);
            });
        }, delay);

        // 清理函数
        return () => {
            clearTimeout(timer);
            console.log('App 组件卸载，清理 WebSocket 连接');
        };
    }, []);

    return (
        <ErrorBoundary>
            <ConfigProvider locale={zhCN}>
                <AuthProvider>
                    <Router>
                        <MainLayout>
                            <Routes>
                                <Route path="/" element={<HomePage />} />
                                <Route path="/ai-generator" element={<AIGeneratorPage />} />
                                <Route path="/crawler" element={<CrawlerPage />} />
                                <Route path="/history" element={<HistoryPage />} />
                                {/* 已移除系统设置页面 */}
                                <Route path="/config-manager" element={<ConfigManagerPage />} />
                            </Routes>
                        </MainLayout>
                    </Router>
                </AuthProvider>
            </ConfigProvider>
        </ErrorBoundary>
    );
};

export default App;