import React, { useState } from 'react';
import {
    Row,
    Col,
    Card,
    Form,
    Input,
    Switch,
    Button,
    Select,
    Slider,
    InputNumber,
    Space,
    Typography,
    Divider,
    message,
    Tabs
} from 'antd';
import {
    UserOutlined,
    SettingOutlined,
    SaveOutlined,
    ReloadOutlined
} from '@ant-design/icons';

const { Title } = Typography;
const { Option } = Select;
const { TabPane } = Tabs;

const SettingsPage: React.FC = () => {
    const [form] = Form.useForm();
    const [loading, setLoading] = useState(false);

    const handleSave = async () => {
        try {
            setLoading(true);
            const values = await form.validateFields();
            console.log('保存设置:', values);
            // 这里可以调用API保存设置
            await new Promise(resolve => setTimeout(resolve, 1000)); // 模拟API调用
            message.success('设置已保存');
        } catch (error) {
            message.error('保存失败');
        } finally {
            setLoading(false);
        }
    };

    const handleReset = () => {
        form.resetFields();
        message.info('设置已重置');
    };

    return (
        <div>
            <Title level={2} style={{ marginBottom: 24 }}>
                <SettingOutlined style={{ marginRight: 8 }} />
                系统设置
            </Title>

            <Tabs defaultActiveKey="general" type="card">
                {/* 常规设置 */}
                <TabPane tab="常规设置" key="general">
                    <Card>
                        <Form
                            form={form}
                            layout="vertical"
                            initialValues={{
                                username: 'user',
                                email: 'user@example.com',
                                language: 'zh-CN',
                                theme: 'light',
                                autoSave: true,
                                notifications: true
                            }}
                        >
                            <Row gutter={[24, 24]}>
                                <Col xs={24} md={12}>
                                    <Form.Item label="用户名" name="username">
                                        <Input prefix={<UserOutlined />} />
                                    </Form.Item>
                                </Col>
                                <Col xs={24} md={12}>
                                    <Form.Item label="邮箱" name="email">
                                        <Input type="email" />
                                    </Form.Item>
                                </Col>
                                <Col xs={24} md={12}>
                                    <Form.Item label="语言" name="language">
                                        <Select>
                                            <Option value="zh-CN">简体中文</Option>
                                            <Option value="en-US">English</Option>
                                            <Option value="ja-JP">日本語</Option>
                                        </Select>
                                    </Form.Item>
                                </Col>
                                <Col xs={24} md={12}>
                                    <Form.Item label="主题" name="theme">
                                        <Select>
                                            <Option value="light">浅色主题</Option>
                                            <Option value="dark">深色主题</Option>
                                            <Option value="auto">跟随系统</Option>
                                        </Select>
                                    </Form.Item>
                                </Col>
                                <Col span={24}>
                                    <Form.Item label="自动保存" name="autoSave" valuePropName="checked">
                                        <Switch />
                                    </Form.Item>
                                </Col>
                                <Col span={24}>
                                    <Form.Item label="桌面通知" name="notifications" valuePropName="checked">
                                        <Switch />
                                    </Form.Item>
                                </Col>
                            </Row>
                        </Form>
                    </Card>
                </TabPane>

                {/* AI生成设置 */}
                <TabPane tab="AI生成" key="ai">
                    <Card>
                        <Form
                            form={form}
                            layout="vertical"
                            initialValues={{
                                defaultModel: 'stable-diffusion-v1.5',
                                defaultSteps: 20,
                                defaultCFG: 7.0,
                                defaultWidth: 512,
                                defaultHeight: 512,
                                maxConcurrent: 2,
                                autoDownload: true,
                                quality: 'high'
                            }}
                        >
                            <Row gutter={[24, 24]}>
                                <Col xs={24} md={12}>
                                    <Form.Item label="默认模型" name="defaultModel">
                                        <Select>
                                            <Option value="stable-diffusion-v1.5">Stable Diffusion v1.5</Option>
                                            <Option value="stable-diffusion-v2.1">Stable Diffusion v2.1</Option>
                                            <Option value="dreamshaper">DreamShaper</Option>
                                            <Option value="realistic-vision">Realistic Vision</Option>
                                        </Select>
                                    </Form.Item>
                                </Col>
                                <Col xs={24} md={12}>
                                    <Form.Item label="默认采样步数" name="defaultSteps">
                                        <InputNumber min={1} max={100} style={{ width: '100%' }} />
                                    </Form.Item>
                                </Col>
                                <Col xs={24} md={12}>
                                    <Form.Item label="默认CFG Scale" name="defaultCFG">
                                        <Slider min={1} max={30} step={0.5} />
                                    </Form.Item>
                                </Col>
                                <Col xs={24} md={12}>
                                    <Form.Item label="默认尺寸" name="defaultSize">
                                        <Space>
                                            <InputNumber placeholder="宽度" min={64} max={2048} step={64} />
                                            <span>×</span>
                                            <InputNumber placeholder="高度" min={64} max={2048} step={64} />
                                        </Space>
                                    </Form.Item>
                                </Col>
                                <Col xs={24} md={12}>
                                    <Form.Item label="最大并发数" name="maxConcurrent">
                                        <InputNumber min={1} max={8} style={{ width: '100%' }} />
                                    </Form.Item>
                                </Col>
                                <Col xs={24} md={12}>
                                    <Form.Item label="生成质量" name="quality">
                                        <Select>
                                            <Option value="low">低质量 (快速)</Option>
                                            <Option value="medium">中等质量</Option>
                                            <Option value="high">高质量 (推荐)</Option>
                                            <Option value="ultra">超高质量 (慢速)</Option>
                                        </Select>
                                    </Form.Item>
                                </Col>
                                <Col span={24}>
                                    <Form.Item label="自动下载结果" name="autoDownload" valuePropName="checked">
                                        <Switch />
                                    </Form.Item>
                                </Col>
                            </Row>
                        </Form>
                    </Card>
                </TabPane>

                {/* 爬虫设置 */}
                <TabPane tab="爬虫设置" key="crawler">
                    <Card>
                        <Form
                            form={form}
                            layout="vertical"
                            initialValues={{
                                maxConcurrent: 3,
                                delay: 2,
                                retryCount: 3,
                                timeout: 30,
                                userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
                                autoStart: false,
                                savePath: './data/images'
                            }}
                        >
                            <Row gutter={[24, 24]}>
                                <Col xs={24} md={12}>
                                    <Form.Item label="最大并发数" name="maxConcurrent">
                                        <InputNumber min={1} max={10} style={{ width: '100%' }} />
                                    </Form.Item>
                                </Col>
                                <Col xs={24} md={12}>
                                    <Form.Item label="请求延迟 (秒)" name="delay">
                                        <InputNumber min={0.5} max={10} step={0.5} style={{ width: '100%' }} />
                                    </Form.Item>
                                </Col>
                                <Col xs={24} md={12}>
                                    <Form.Item label="重试次数" name="retryCount">
                                        <InputNumber min={0} max={10} style={{ width: '100%' }} />
                                    </Form.Item>
                                </Col>
                                <Col xs={24} md={12}>
                                    <Form.Item label="超时时间 (秒)" name="timeout">
                                        <InputNumber min={10} max={120} style={{ width: '100%' }} />
                                    </Form.Item>
                                </Col>
                                <Col span={24}>
                                    <Form.Item label="User Agent" name="userAgent">
                                        <Input.TextArea rows={2} />
                                    </Form.Item>
                                </Col>
                                <Col span={24}>
                                    <Form.Item label="保存路径" name="savePath">
                                        <Input />
                                    </Form.Item>
                                </Col>
                                <Col span={24}>
                                    <Form.Item label="自动开始爬取" name="autoStart" valuePropName="checked">
                                        <Switch />
                                    </Form.Item>
                                </Col>
                            </Row>
                        </Form>
                    </Card>
                </TabPane>

                {/* 安全设置 */}
                <TabPane tab="安全设置" key="security">
                    <Card>
                        <Form
                            form={form}
                            layout="vertical"
                            initialValues={{
                                enableAuth: false,
                                sessionTimeout: 30,
                                enableLogging: true,
                                logLevel: 'info'
                            }}
                        >
                            <Row gutter={[24, 24]}>
                                <Col span={24}>
                                    <Form.Item label="启用身份验证" name="enableAuth" valuePropName="checked">
                                        <Switch />
                                    </Form.Item>
                                </Col>
                                <Col xs={24} md={12}>
                                    <Form.Item label="会话超时 (分钟)" name="sessionTimeout">
                                        <InputNumber min={5} max={480} style={{ width: '100%' }} />
                                    </Form.Item>
                                </Col>
                                <Col xs={24} md={12}>
                                    <Form.Item label="日志级别" name="logLevel">
                                        <Select>
                                            <Option value="debug">Debug</Option>
                                            <Option value="info">Info</Option>
                                            <Option value="warn">Warning</Option>
                                            <Option value="error">Error</Option>
                                        </Select>
                                    </Form.Item>
                                </Col>
                                <Col span={24}>
                                    <Form.Item label="启用操作日志" name="enableLogging" valuePropName="checked">
                                        <Switch />
                                    </Form.Item>
                                </Col>
                                <Col span={24}>
                                    <Divider />
                                    <Space>
                                        <Button type="primary" danger>
                                            清除所有日志
                                        </Button>
                                        <Button>
                                            导出日志
                                        </Button>
                                    </Space>
                                </Col>
                            </Row>
                        </Form>
                    </Card>
                </TabPane>

                {/* 通知设置 */}
                <TabPane tab="通知设置" key="notifications">
                    <Card>
                        <Form
                            form={form}
                            layout="vertical"
                            initialValues={{
                                emailNotifications: true,
                                desktopNotifications: true,
                                taskComplete: true,
                                taskError: true,
                                systemUpdate: false
                            }}
                        >
                            <Row gutter={[24, 24]}>
                                <Col span={24}>
                                    <Form.Item label="邮件通知" name="emailNotifications" valuePropName="checked">
                                        <Switch />
                                    </Form.Item>
                                </Col>
                                <Col span={24}>
                                    <Form.Item label="桌面通知" name="desktopNotifications" valuePropName="checked">
                                        <Switch />
                                    </Form.Item>
                                </Col>
                                <Col span={24}>
                                    <Divider orientation="left">通知类型</Divider>
                                </Col>
                                <Col span={24}>
                                    <Form.Item label="任务完成通知" name="taskComplete" valuePropName="checked">
                                        <Switch />
                                    </Form.Item>
                                </Col>
                                <Col span={24}>
                                    <Form.Item label="任务错误通知" name="taskError" valuePropName="checked">
                                        <Switch />
                                    </Form.Item>
                                </Col>
                                <Col span={24}>
                                    <Form.Item label="系统更新通知" name="systemUpdate" valuePropName="checked">
                                        <Switch />
                                    </Form.Item>
                                </Col>
                            </Row>
                        </Form>
                    </Card>
                </TabPane>
            </Tabs>

            {/* 操作按钮 */}
            <Card style={{ marginTop: 24 }}>
                <Row justify="space-between" align="middle">
                    <Col>
                        <Space>
                            <Button
                                type="primary"
                                icon={<SaveOutlined />}
                                onClick={handleSave}
                                loading={loading}
                            >
                                保存设置
                            </Button>
                            <Button
                                icon={<ReloadOutlined />}
                                onClick={handleReset}
                            >
                                重置
                            </Button>
                        </Space>
                    </Col>
                    <Col>
                        <Space>
                            <Button>
                                导入配置
                            </Button>
                            <Button>
                                导出配置
                            </Button>
                        </Space>
                    </Col>
                </Row>
            </Card>
        </div>
    );
};

export default SettingsPage;
