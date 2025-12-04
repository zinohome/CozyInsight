import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Form, Input, Button, Card, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { authAPI } from '../../api/auth';
import { useAuthStore } from '../../store/authStore';

const Login = () => {
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false);
    const { login } = useAuthStore();

    const handleLogin = async (values: { username: string; password: string }) => {
        try {
            setLoading(true);
            const response = await authAPI.login(values.username, values.password);

            login(response.token, response.user);
            message.success('登录成功！');
            navigate('/');
        } catch (error: any) {
            message.error(error.message || '登录失败');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{
            minHeight: '100vh',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        }}>
            <Card
                title="CozyInsight 登录"
                style={{ width: 400, boxShadow: '0 4px 12px rgba(0,0,0,0.15)' }}
            >
                <Form
                    name="login"
                    onFinish={handleLogin}
                    autoComplete="off"
                    size="large"
                >
                    <Form.Item
                        name="username"
                        rules={[{ required: true, message: '请输入用户名' }]}
                    >
                        <Input
                            prefix={<UserOutlined />}
                            placeholder="用户名"
                        />
                    </Form.Item>

                    <Form.Item
                        name="password"
                        rules={[{ required: true, message: '请输入密码' }]}
                    >
                        <Input.Password
                            prefix={<LockOutlined />}
                            placeholder="密码"
                        />
                    </Form.Item>

                    <Form.Item>
                        <Button
                            type="primary"
                            htmlType="submit"
                            loading={loading}
                            block
                        >
                            登录
                        </Button>
                    </Form.Item>
                </Form>

                <div style={{ textAlign: 'center', color: '#999' }}>
                    <p>默认用户名/密码: admin/admin</p>
                </div>
            </Card>
        </div>
    );
};

export default Login;
