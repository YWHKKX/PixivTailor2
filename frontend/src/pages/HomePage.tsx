import React from 'react';
import { Row, Col, Card, Typography, Button, Space, Statistic, Progress } from 'antd';
import {
    RobotOutlined,
    DatabaseOutlined,
    HistoryOutlined,
    ArrowRightOutlined,
    ThunderboltOutlined,
    CloudOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const { Title, Paragraph, Text } = Typography;

const HomePage: React.FC = () => {
    const navigate = useNavigate();

    const features = [
        {
            title: 'AI图像生成',
            description: '基于先进AI模型，快速生成高质量图像',
            icon: <RobotOutlined style={{ fontSize: 48, color: '#1890ff' }} />,
            path: '/ai-generator',
            color: '#1890ff'
        },
        {
            title: '爬虫管理',
            description: '智能爬取Pixiv数据，支持批量处理',
            icon: <DatabaseOutlined style={{ fontSize: 48, color: '#52c41a' }} />,
            path: '/crawler',
            color: '#52c41a'
        },
        {
            title: '历史记录',
            description: '完整记录所有操作，便于管理和回溯',
            icon: <HistoryOutlined style={{ fontSize: 48, color: '#faad14' }} />,
            path: '/history',
            color: '#faad14'
        }
    ];

    const stats = [
        { title: '生成图像', value: 1234, suffix: '张' },
        { title: '爬取数据', value: 5678, suffix: '条' },
        { title: '活跃任务', value: 12, suffix: '个' },
        { title: '系统运行', value: 99.9, suffix: '%' }
    ];

    return (
        <div>
            {/* 欢迎区域 */}
            <div style={{
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                padding: '60px 40px',
                borderRadius: 16,
                marginBottom: 32,
                color: 'white',
                textAlign: 'center'
            }}>
                <Title level={1} style={{ color: 'white', marginBottom: 16 }}>
                    🎨 欢迎使用 PixivTailor
                </Title>
                <Paragraph style={{ fontSize: 18, color: 'rgba(255,255,255,0.9)', marginBottom: 32 }}>
                    专业的AI图像生成与数据爬取平台，为您提供高效、智能的创作工具
                </Paragraph>
                <Space size="large">
                    <Button
                        type="primary"
                        size="large"
                        icon={<RobotOutlined />}
                        onClick={() => navigate('/ai-generator')}
                        style={{ height: 48, paddingLeft: 24, paddingRight: 24 }}
                    >
                        开始生成
                    </Button>
                    <Button
                        size="large"
                        icon={<DatabaseOutlined />}
                        onClick={() => navigate('/crawler')}
                        style={{ height: 48, paddingLeft: 24, paddingRight: 24, color: 'white', borderColor: 'white' }}
                    >
                        数据爬取
                    </Button>
                </Space>
            </div>

            {/* 统计信息 */}
            <Row gutter={[16, 16]} style={{ marginBottom: 32 }}>
                {stats.map((stat, index) => (
                    <Col xs={24} sm={12} lg={6} key={index}>
                        <Card>
                            <Statistic
                                title={stat.title}
                                value={stat.value}
                                suffix={stat.suffix}
                                valueStyle={{ color: '#1890ff' }}
                            />
                        </Card>
                    </Col>
                ))}
            </Row>

            {/* 功能模块 */}
            <Title level={2} style={{ marginBottom: 24 }}>核心功能</Title>
            <Row gutter={[24, 24]}>
                {features.map((feature, index) => (
                    <Col xs={24} lg={8} key={index}>
                        <Card
                            hoverable
                            style={{
                                height: '100%',
                                border: `2px solid ${feature.color}20`,
                                borderRadius: 12
                            }}
                            styles={{
                                body: { padding: 32, textAlign: 'center' }
                            }}
                        >
                            <div style={{ marginBottom: 24 }}>
                                {feature.icon}
                            </div>
                            <Title level={3} style={{ marginBottom: 16 }}>
                                {feature.title}
                            </Title>
                            <Paragraph style={{ color: '#666', marginBottom: 24 }}>
                                {feature.description}
                            </Paragraph>
                            <Button
                                type="primary"
                                icon={<ArrowRightOutlined />}
                                onClick={() => navigate(feature.path)}
                                style={{
                                    background: feature.color,
                                    borderColor: feature.color,
                                    borderRadius: 8
                                }}
                            >
                                立即使用
                            </Button>
                        </Card>
                    </Col>
                ))}
            </Row>

            {/* 系统状态 */}
            <Card style={{ marginTop: 32 }}>
                <Title level={3} style={{ marginBottom: 24 }}>
                    <ThunderboltOutlined style={{ marginRight: 8, color: '#52c41a' }} />
                    系统状态
                </Title>
                <Row gutter={[24, 24]}>
                    <Col xs={24} md={12}>
                        <div style={{ marginBottom: 16 }}>
                            <Text strong>CPU 使用率</Text>
                            <Progress percent={30} status="active" />
                        </div>
                        <div style={{ marginBottom: 16 }}>
                            <Text strong>内存使用率</Text>
                            <Progress percent={65} status="active" />
                        </div>
                        <div>
                            <Text strong>磁盘使用率</Text>
                            <Progress percent={45} status="active" />
                        </div>
                    </Col>
                    <Col xs={24} md={12}>
                        <div style={{
                            background: '#f5f5f5',
                            padding: 24,
                            borderRadius: 8,
                            textAlign: 'center'
                        }}>
                            <CloudOutlined style={{ fontSize: 48, color: '#1890ff', marginBottom: 16 }} />
                            <Title level={4} style={{ marginBottom: 8 }}>服务状态</Title>
                            <Text type="success" strong>运行正常</Text>
                        </div>
                    </Col>
                </Row>
            </Card>
        </div>
    );
};

export default HomePage;
