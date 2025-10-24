import React, { useState } from 'react';
import { Layout, Menu, Button, Avatar, Dropdown, Space, Typography, theme } from 'antd';
import {
    HomeOutlined,
    RobotOutlined,
    DatabaseOutlined,
    HistoryOutlined,
    SettingOutlined,
    UserOutlined,
    LogoutOutlined,
    LoginOutlined,
    MenuFoldOutlined,
    MenuUnfoldOutlined
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';
import LoginModal from '@/components/auth/LoginModal';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

interface MainLayoutProps {
    children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
    const [collapsed, setCollapsed] = useState(false);
    const [loginModalVisible, setLoginModalVisible] = useState(false);
    const navigate = useNavigate();
    const location = useLocation();
    const { user, login, logout, isAuthenticated } = useAuth();
    const {
        token: { colorBgContainer, borderRadiusLG },
    } = theme.useToken();

    const menuItems = [
        {
            key: '/',
            icon: <HomeOutlined />,
            label: '首页',
        },
        {
            key: '/ai-generator',
            icon: <RobotOutlined />,
            label: 'AI图像生成',
        },
        {
            key: '/crawler',
            icon: <DatabaseOutlined />,
            label: '爬虫管理',
        },
        {
            key: '/history',
            icon: <HistoryOutlined />,
            label: '历史记录',
        },
        {
            key: '/settings',
            icon: <SettingOutlined />,
            label: '设置',
        },
    ];

    const userMenuItems = isAuthenticated ? [
        {
            key: 'profile',
            icon: <UserOutlined />,
            label: '个人资料',
        },
        {
            key: 'logout',
            icon: <LogoutOutlined />,
            label: '退出登录',
        },
    ] : [
        {
            key: 'login',
            icon: <LoginOutlined />,
            label: '登录',
        },
    ];

    const handleMenuClick = ({ key }: { key: string }) => {
        navigate(key);
    };

    const handleUserMenuClick = ({ key }: { key: string }) => {
        if (key === 'logout') {
            logout();
        } else if (key === 'profile') {
            // 处理个人资料
            console.log('个人资料');
        } else if (key === 'login') {
            setLoginModalVisible(true);
        }
    };

    const handleLoginSuccess = (userData: any) => {
        login(userData);
        setLoginModalVisible(false);
    };

    return (
        <Layout style={{ minHeight: '100vh' }}>
            <Sider
                trigger={null}
                collapsible
                collapsed={collapsed}
                style={{
                    background: colorBgContainer,
                    boxShadow: '2px 0 8px rgba(0,0,0,0.1)',
                }}
            >
                <div style={{
                    height: 64,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    borderBottom: '1px solid #f0f0f0',
                    marginBottom: 16
                }}>
                    <Text strong style={{ fontSize: collapsed ? 16 : 18, color: '#1890ff' }}>
                        {collapsed ? 'PT' : 'PixivTailor'}
                    </Text>
                </div>

                <Menu
                    mode="inline"
                    selectedKeys={[location.pathname]}
                    items={menuItems}
                    onClick={handleMenuClick}
                    style={{ border: 'none' }}
                />
            </Sider>

            <Layout>
                <Header style={{
                    padding: '0 24px',
                    background: colorBgContainer,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                }}>
                    <Button
                        type="text"
                        icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                        onClick={() => setCollapsed(!collapsed)}
                        style={{ fontSize: '16px', width: 64, height: 64 }}
                    />

                    <Space>
                        <Dropdown
                            menu={{
                                items: userMenuItems,
                                onClick: handleUserMenuClick
                            }}
                            placement="bottomRight"
                        >
                            <Button type="text" style={{ display: 'flex', alignItems: 'center' }}>
                                <Avatar
                                    src={user?.avatar}
                                    icon={<UserOutlined />}
                                    style={{ marginRight: 8 }}
                                />
                                <Text>{user?.username || '游客'}</Text>
                            </Button>
                        </Dropdown>
                    </Space>
                </Header>

                <Content style={{
                    margin: '24px 16px',
                    padding: 24,
                    minHeight: 280,
                    background: colorBgContainer,
                    borderRadius: borderRadiusLG,
                }}>
                    {children}
                </Content>
            </Layout>

            <LoginModal
                visible={loginModalVisible}
                onCancel={() => setLoginModalVisible(false)}
                onSuccess={handleLoginSuccess}
            />
        </Layout>
    );
};

export default MainLayout;
