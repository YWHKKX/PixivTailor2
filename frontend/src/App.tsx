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
import SettingsPage from '@/pages/SettingsPage';

// 导入服务
import { initWebSocket } from '@/services/websocket';

// 导入样式
import './styles/gradio.css';
import './styles/theme.css';

const App: React.FC = () => {
    useEffect(() => {
        // 在开发模式下延迟初始化 WebSocket 连接，避免热重载频繁重连
        const isDevelopment = import.meta.env.DEV;
        const delay = isDevelopment ? 2000 : 0; // 开发模式延迟2秒

        console.log('App 组件挂载，初始化 WebSocket 连接', isDevelopment ? '(开发模式，延迟连接)' : '');

        const timer = setTimeout(() => {
            initWebSocket();
        }, delay);

        // 初始化 UI 更新器
        console.log('UI Updater initialized');

        // 清理函数
        return () => {
            // 清理定时器
            clearTimeout(timer);

            // 清理 WebSocket 连接
            console.log('App 组件卸载，清理 WebSocket 连接');
            // 注意：这里不关闭连接，因为可能还有其他组件在使用
            // 实际的连接清理会在页面卸载时进行
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
                                <Route path="/settings" element={<SettingsPage />} />
                            </Routes>
                        </MainLayout>
                    </Router>
                </AuthProvider>
            </ConfigProvider>
        </ErrorBoundary>
    );
};

export default App;
