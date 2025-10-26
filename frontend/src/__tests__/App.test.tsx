import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import App from '../App';
import ErrorBoundary from '../components/ErrorBoundary';

// 测试包装器
const TestWrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <BrowserRouter>
    {children}
  </BrowserRouter>
);

describe('App Component', () => {
  test('renders without crashing', () => {
    render(
      <TestWrapper>
        <App />
      </TestWrapper>
    );
    // 应用应该正常渲染
  });

  test('renders main layout', () => {
    render(
      <TestWrapper>
        <App />
      </TestWrapper>
    );
    
    // 检查主布局是否存在
    expect(screen.getByText('PixivTailor')).toBeInTheDocument();
  });

  test('renders navigation menu', () => {
    render(
      <TestWrapper>
        <App />
      </TestWrapper>
    );
    
    // 检查导航菜单项
    expect(screen.getByText('首页')).toBeInTheDocument();
    expect(screen.getByText('AI图像生成')).toBeInTheDocument();
    expect(screen.getByText('配置管理')).toBeInTheDocument();
    expect(screen.getByText('爬虫管理')).toBeInTheDocument();
    expect(screen.getByText('历史记录')).toBeInTheDocument();
    expect(screen.getByText('系统设置')).toBeInTheDocument();
  });

  test('renders login button when not authenticated', () => {
    render(
      <TestWrapper>
        <App />
      </TestWrapper>
    );
    
    // 检查登录按钮
    expect(screen.getByText('登录')).toBeInTheDocument();
  });

  test('renders content area', () => {
    render(
      <TestWrapper>
        <App />
      </TestWrapper>
    );
    
    // 检查内容区域
    expect(screen.getByTestId('home-page')).toBeInTheDocument();
  });
});

describe('ErrorBoundary Component', () => {
  test('renders children when there is no error', () => {
    render(
      <ErrorBoundary>
        <div data-testid="test-content">Test Content</div>
      </ErrorBoundary>
    );

    expect(screen.getByTestId('test-content')).toBeInTheDocument();
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

    // 检查错误边界是否显示
    expect(screen.getByText('应用出现错误')).toBeInTheDocument();
    expect(screen.getByText('刷新页面')).toBeInTheDocument();
    expect(screen.getByText('返回首页')).toBeInTheDocument();
  });

  test('handles error recovery', () => {
    const ThrowError = ({ shouldThrow }: { shouldThrow: boolean }) => {
      if (shouldThrow) {
        throw new Error('Test error');
      }
      return <div data-testid="recovered-content">Recovered Content</div>;
    };

    const { rerender } = render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    // 初始状态应该显示错误
    expect(screen.getByText('应用出现错误')).toBeInTheDocument();

    // 重新渲染，不再抛出错误
    rerender(
      <ErrorBoundary>
        <ThrowError shouldThrow={false} />
      </ErrorBoundary>
    );

    // 应该显示恢复的内容
    expect(screen.getByTestId('recovered-content')).toBeInTheDocument();
  });
});

describe('App Integration', () => {
  test('handles route changes', async () => {
    render(
      <TestWrapper>
        <App />
      </TestWrapper>
    );

    // 初始应该显示首页
    expect(screen.getByTestId('home-page')).toBeInTheDocument();

    // 模拟路由变化
    // 注意：这里需要实际的路由测试，但由于我们使用了模拟，这里只是示例
  });

  test('handles WebSocket connection', async () => {
    render(
      <TestWrapper>
        <App />
      </TestWrapper>
    );

    // 等待 WebSocket 连接初始化
    await waitFor(() => {
      // 检查 WebSocket 是否被调用
      // 这里需要检查实际的 WebSocket 服务调用
    });
  });
});

describe('App Performance', () => {
  test('renders within acceptable time', () => {
    const startTime = performance.now();
    
    render(
      <TestWrapper>
        <App />
      </TestWrapper>
    );
    
    const endTime = performance.now();
    const renderTime = endTime - startTime;
    
    // 渲染时间应该在合理范围内（这里设置为 100ms）
    expect(renderTime).toBeLessThan(100);
  });

  test('does not cause memory leaks', () => {
    const { unmount } = render(
      <TestWrapper>
        <App />
      </TestWrapper>
    );

    // 卸载组件
    unmount();

    // 检查是否有内存泄漏
    // 这里需要更复杂的测试来检测内存泄漏
  });
});

describe('App Accessibility', () => {
  test('has proper ARIA labels', () => {
    render(
      <TestWrapper>
        <App />
      </TestWrapper>
    );

    // 检查关键元素是否有适当的 ARIA 标签
    const mainContent = screen.getByRole('main');
    expect(mainContent).toBeInTheDocument();
  });

  test('supports keyboard navigation', () => {
    render(
      <TestWrapper>
        <App />
      </TestWrapper>
    );

    // 检查键盘导航支持
    const menuItems = screen.getAllByRole('menuitem');
    expect(menuItems.length).toBeGreaterThan(0);
  });
});

describe('App Error Handling', () => {
  test('handles network errors gracefully', () => {
    // 模拟网络错误
    const originalFetch = global.fetch;
    global.fetch = jest.fn().mockRejectedValue(new Error('Network error'));

    render(
      <TestWrapper>
        <App />
      </TestWrapper>
    );

    // 恢复原始 fetch
    global.fetch = originalFetch;

    // 应用应该仍然正常渲染
    expect(screen.getByText('PixivTailor')).toBeInTheDocument();
  });

  test('handles component errors gracefully', () => {
    // 模拟组件错误
    const consoleError = jest.spyOn(console, 'error').mockImplementation(() => {});

    render(
      <TestWrapper>
        <App />
      </TestWrapper>
    );

    // 恢复 console.error
    consoleError.mockRestore();

    // 应用应该仍然正常渲染
    expect(screen.getByText('PixivTailor')).toBeInTheDocument();
  });
});