import { Component, ErrorInfo, ReactNode } from 'react';
import { Result, Button, Card } from 'antd';
import { ReloadOutlined, HomeOutlined } from '@ant-design/icons';

interface Props {
    children: ReactNode;
    fallback?: ReactNode;
}

interface State {
    hasError: boolean;
    error?: Error;
    errorInfo?: ErrorInfo;
}

class ErrorBoundary extends Component<Props, State> {
    constructor(props: Props) {
        super(props);
        this.state = { hasError: false };
    }

    static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error };
    }

    componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        console.error('ErrorBoundary caught an error:', error, errorInfo);
        this.setState({
            error,
            errorInfo,
        });
    }

    handleReload = () => {
        window.location.reload();
    };

    handleGoHome = () => {
        window.location.href = '/';
    };

    render() {
        if (this.state.hasError) {
            if (this.props.fallback) {
                return this.props.fallback;
            }

            return (
                <div style={{
                    display: 'flex',
                    justifyContent: 'center',
                    alignItems: 'center',
                    minHeight: '100vh',
                    padding: '20px'
                }}>
                    <Card style={{ maxWidth: 600, width: '100%' }}>
                        <Result
                            status="error"
                            title="应用出现错误"
                            subTitle="很抱歉，应用遇到了一个错误。请尝试刷新页面或返回首页。"
                            extra={[
                                <Button
                                    type="primary"
                                    key="reload"
                                    icon={<ReloadOutlined />}
                                    onClick={this.handleReload}
                                >
                                    刷新页面
                                </Button>,
                                <Button
                                    key="home"
                                    icon={<HomeOutlined />}
                                    onClick={this.handleGoHome}
                                >
                                    返回首页
                                </Button>,
                            ]}
                        />

                        {import.meta.env.DEV && this.state.error && (
                            <div style={{
                                marginTop: 20,
                                padding: 16,
                                background: '#f5f5f5',
                                borderRadius: 6
                            }}>
                                <h4>错误详情 (开发模式):</h4>
                                <pre style={{
                                    whiteSpace: 'pre-wrap',
                                    wordBreak: 'break-word',
                                    fontSize: '12px',
                                    color: '#ff4d4f'
                                }}>
                                    {this.state.error.toString()}
                                </pre>
                                {this.state.errorInfo && (
                                    <pre style={{
                                        whiteSpace: 'pre-wrap',
                                        wordBreak: 'break-word',
                                        fontSize: '12px',
                                        color: '#666',
                                        marginTop: 10
                                    }}>
                                        {this.state.errorInfo.componentStack}
                                    </pre>
                                )}
                            </div>
                        )}
                    </Card>
                </div>
            );
        }

        return this.props.children;
    }
}

export default ErrorBoundary;
