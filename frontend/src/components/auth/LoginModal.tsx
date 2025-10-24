import React, { useState } from 'react';
import { Modal, Form, Input, Button, Typography, message, Divider } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons';

const { Title, Text } = Typography;

interface LoginModalProps {
    visible: boolean;
    onCancel: () => void;
    onSuccess: (user: any) => void;
}

const LoginModal: React.FC<LoginModalProps> = ({ visible, onCancel, onSuccess }) => {
    const [form] = Form.useForm();
    const [loading, setLoading] = useState(false);
    const [isLogin, setIsLogin] = useState(true);

    const handleSubmit = async (values: any) => {
        setLoading(true);
        try {
            // 模拟API调用
            await new Promise(resolve => setTimeout(resolve, 1000));

            if (isLogin) {
                // 模拟登录
                const user = {
                    id: '1',
                    username: values.username,
                    email: values.email || `${values.username}@example.com`,
                    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=' + values.username
                };
                onSuccess(user);
                message.success('登录成功');
            } else {
                // 模拟注册
                message.success('注册成功，请登录');
                setIsLogin(true);
                form.resetFields();
            }
        } catch (error) {
            message.error(isLogin ? '登录失败' : '注册失败');
        } finally {
            setLoading(false);
        }
    };

    const handleCancel = () => {
        form.resetFields();
        setIsLogin(true);
        onCancel();
    };

    return (
        <Modal
            title={
                <div style={{ textAlign: 'center' }}>
                    <Title level={3} style={{ margin: 0 }}>
                        {isLogin ? '欢迎回来' : '创建账户'}
                    </Title>
                    <Text type="secondary">
                        {isLogin ? '登录到 PixivTailor' : '开始使用 PixivTailor'}
                    </Text>
                </div>
            }
            open={visible}
            onCancel={handleCancel}
            footer={null}
            width={400}
            centered
        >
            <Form
                form={form}
                layout="vertical"
                onFinish={handleSubmit}
                autoComplete="off"
            >
                <Form.Item
                    name="username"
                    rules={[
                        { required: true, message: '请输入用户名' },
                        { min: 3, message: '用户名至少3个字符' }
                    ]}
                >
                    <Input
                        prefix={<UserOutlined />}
                        placeholder="用户名"
                        size="large"
                    />
                </Form.Item>

                {!isLogin && (
                    <Form.Item
                        name="email"
                        rules={[
                            { required: true, message: '请输入邮箱' },
                            { type: 'email', message: '请输入有效的邮箱地址' }
                        ]}
                    >
                        <Input
                            prefix={<MailOutlined />}
                            placeholder="邮箱地址"
                            size="large"
                        />
                    </Form.Item>
                )}

                <Form.Item
                    name="password"
                    rules={[
                        { required: true, message: '请输入密码' },
                        { min: 6, message: '密码至少6个字符' }
                    ]}
                >
                    <Input.Password
                        prefix={<LockOutlined />}
                        placeholder="密码"
                        size="large"
                    />
                </Form.Item>

                {!isLogin && (
                    <Form.Item
                        name="confirmPassword"
                        dependencies={['password']}
                        rules={[
                            { required: true, message: '请确认密码' },
                            ({ getFieldValue }) => ({
                                validator(_, value) {
                                    if (!value || getFieldValue('password') === value) {
                                        return Promise.resolve();
                                    }
                                    return Promise.reject(new Error('两次输入的密码不一致'));
                                },
                            }),
                        ]}
                    >
                        <Input.Password
                            prefix={<LockOutlined />}
                            placeholder="确认密码"
                            size="large"
                        />
                    </Form.Item>
                )}

                <Form.Item>
                    <Button
                        type="primary"
                        htmlType="submit"
                        loading={loading}
                        size="large"
                        block
                        style={{ height: 48 }}
                    >
                        {isLogin ? '登录' : '注册'}
                    </Button>
                </Form.Item>

                <Divider>
                    <Text type="secondary">
                        {isLogin ? '还没有账户？' : '已有账户？'}
                    </Text>
                </Divider>

                <Button
                    type="link"
                    onClick={() => setIsLogin(!isLogin)}
                    block
                    size="large"
                >
                    {isLogin ? '立即注册' : '立即登录'}
                </Button>
            </Form>
        </Modal>
    );
};

export default LoginModal;
