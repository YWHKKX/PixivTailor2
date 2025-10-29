import React, { useState } from 'react';
import { Layout, Menu, Button, Avatar, Dropdown, Space, Typography, theme, Badge, MenuProps } from 'antd';
import {
    HomeOutlined,
    RobotOutlined,
    DatabaseOutlined,
    TagsOutlined,
    HistoryOutlined,
    FileTextOutlined,
    UserOutlined,
    LogoutOutlined,
    LoginOutlined,
    MenuFoldOutlined,
    MenuUnfoldOutlined,
    BellOutlined
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';
import LoginModal from '@/components/auth/LoginModal';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

// ==================== 类型定义 ====================
interface MainLayoutProps {
    children: React.ReactNode;
}

type MenuItem = Required<MenuProps>['items'][number];

// ==================== 主布局组件 ====================
const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
    // ==================== 状态管理 ====================
    const [collapsed, setCollapsed] = useState(false);
    const [loginModalVisible, setLoginModalVisible] = useState(false);

    // ==================== 路由和认证 ====================
    const navigate = useNavigate();
    const location = useLocation();
    const { user, login, logout, isAuthenticated } = useAuth();

    // ==================== 主题 ====================
    const {
        token: { colorBgContainer, borderRadiusLG },
    } = theme.useToken();

    // ==================== 菜单配置 ====================
    const menuItems: MenuItem[] = [
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
            key: '/config-manager',
            icon: <FileTextOutlined />,
            label: '配置管理',
        },
        {
            key: '/crawler',
            icon: <DatabaseOutlined />,
            label: '爬虫管理',
        },
        {
            key: '/tagger',
            icon: <TagsOutlined />,
            label: '图像标签',
        },
        {
            key: '/history',
            icon: <HistoryOutlined />,
            label: '历史记录',
        },
        // 已移除系统设置菜单项
    ];

    // ==================== 事件处理 ====================
    const handleMenuClick = ({ key }: { key: string }) => {
        navigate(key);
    };

    const handleLogin = async (user: any) => {
        try {
            // 直接使用传入的用户对象
            login(user);
            setLoginModalVisible(false);
        } catch (error) {
            console.error('登录失败:', error);
        }
    };

    const handleLogout = () => {
        logout();
    };

    const handleToggleCollapsed = () => {
        setCollapsed(!collapsed);
    };

    // ==================== 用户菜单 ====================
    const userMenuItems = [
        {
            key: 'profile',
            icon: <UserOutlined />,
            label: '个人资料',
        },
        // 已移除账户设置菜单项
        {
            type: 'divider' as const,
        },
        {
            key: 'logout',
            icon: <LogoutOutlined />,
            label: '退出登录',
            onClick: handleLogout,
        },
    ];

    // ==================== 渲染函数 ====================
    const renderUserSection = () => {
        if (isAuthenticated && user) {
            return (
                <Dropdown
                    menu={{ items: userMenuItems }}
                    placement="bottomRight"
                    arrow
                >
                    <Space style={{ cursor: 'pointer' }}>
                        <Avatar size="small" icon={<UserOutlined />} />
                        <Text strong>{user.username}</Text>
                    </Space>
                </Dropdown>
            );
        }

        return (
            <Button
                type="primary"
                icon={<LoginOutlined />}
                onClick={() => setLoginModalVisible(true)}
            >
                登录
            </Button>
        );
    };

    const renderHeader = () => (
        <Header
            style={{
                padding: '0 16px',
                background: colorBgContainer,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                borderBottom: '1px solid #f0f0f0',
            }}
        >
            <Space>
                <Button
                    type="text"
                    icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                    onClick={handleToggleCollapsed}
                    style={{ fontSize: '16px', width: 40, height: 40 }}
                />
                <Typography.Title level={4} style={{ margin: 0, color: '#1890ff' }}>
                    PixivTailor
                </Typography.Title>
            </Space>

            <Space>
                <Badge count={0} size="small">
                    <Button
                        type="text"
                        icon={<BellOutlined />}
                        style={{ fontSize: '16px' }}
                    />
                </Badge>
                {renderUserSection()}
            </Space>
        </Header>
    );

    const renderSidebar = () => (
        <Sider
            trigger={null}
            collapsible
            collapsed={collapsed}
            style={{
                background: colorBgContainer,
                borderRight: '1px solid #f0f0f0',
            }}
        >
            <Menu
                mode="inline"
                selectedKeys={[location.pathname]}
                items={menuItems}
                onClick={handleMenuClick}
                style={{
                    height: '100%',
                    borderRight: 0,
                }}
            />
        </Sider>
    );

    const renderContent = () => (
        <Content
            style={{
                margin: '16px',
                padding: 24,
                minHeight: 280,
                background: colorBgContainer,
                borderRadius: borderRadiusLG,
                overflow: 'auto',
            }}
        >
            {children}
        </Content>
    );

    const renderLoginModal = () => (
        <LoginModal
            visible={loginModalVisible}
            onCancel={() => setLoginModalVisible(false)}
            onSuccess={handleLogin}
        />
    );

    // ==================== 主渲染 ====================
    return (
        <Layout style={{ minHeight: '100vh' }}>
            {renderSidebar()}
            <Layout>
                {renderHeader()}
                {renderContent()}
            </Layout>
            {renderLoginModal()}
        </Layout>
    );
};

export default MainLayout;