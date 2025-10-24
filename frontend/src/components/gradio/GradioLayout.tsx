import React from 'react';
import { Layout, Typography, Space } from 'antd';
import { uiComponentFactory } from './UIComponentFactory';

const { Header, Content, Sider } = Layout;
const { Title } = Typography;

interface GradioLayoutProps {
    children?: React.ReactNode;
}

const GradioLayout: React.FC<GradioLayoutProps> = ({ children }) => {
    return (
        <Layout className="gradio-layout">
            <Header className="gradio-header">
                <div className="header-content">
                    <Title level={2} className="app-title">
                        🎨 AI图像生成器
                    </Title>
                    <div className="header-subtitle">
                        基于 Gradio 风格的 AI 图像生成界面
                    </div>
                </div>
            </Header>

            <Layout className="gradio-main">
                <Sider width={400} className="gradio-sider">
                    <div className="sider-content">
                        <Space direction="vertical" size="large" style={{ width: '100%' }}>
                            {uiComponentFactory.createInputPanel()}
                            {uiComponentFactory.createSettingsPanel()}
                            {uiComponentFactory.createControlPanel()}
                        </Space>
                    </div>
                </Sider>

                <Content className="gradio-content">
                    <div className="content-wrapper">
                        {uiComponentFactory.createOutputPanel()}
                        {children}
                    </div>
                </Content>
            </Layout>

            <Layout.Footer className="gradio-footer">
                <div className="footer-content">
                    <Space direction="vertical" size="small">
                        <div className="footer-title">使用说明</div>
                        <div className="footer-steps">
                            <div>1. 在左侧输入提示词</div>
                            <div>2. 调整生成参数</div>
                            <div>3. 点击生成按钮</div>
                            <div>4. 等待生成完成</div>
                        </div>
                    </Space>
                </div>
            </Layout.Footer>
        </Layout>
    );
};

export default GradioLayout;
