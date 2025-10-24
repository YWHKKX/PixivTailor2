import React from 'react';
import { render, screen } from '@testing-library/react';
import App from '../App';
import ErrorBoundary from '../components/ErrorBoundary';

describe('App Component', () => {
    test('renders AI image generator title', () => {
        render(<App />);
        const titleElement = screen.getByText(/AI图像生成器/i);
        expect(titleElement).toBeInTheDocument();
    });

    test('renders without crashing', () => {
        render(<App />);
        // 应用应该正常渲染
    });
});

describe('ErrorBoundary Component', () => {
    test('renders children when there is no error', () => {
        render(
            <ErrorBoundary>
                <div>Test Content</div>
            </ErrorBoundary>
        );

        expect(screen.getByText('Test Content')).toBeInTheDocument();
    });

    test('renders error fallback when there is an error', () => {
        const ThrowError = () => {
            throw new Error('Test error');
        };

        render(
            <ErrorBoundary>
                <ThrowError />
            </ErrorBoundary>
        );

        expect(screen.getByText('应用出现错误')).toBeInTheDocument();
        expect(screen.getByText('刷新页面')).toBeInTheDocument();
        expect(screen.getByText('返回首页')).toBeInTheDocument();
    });
});
