import React, { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { Input, Button, message, Spin } from 'antd';
import { LockOutlined } from '@ant-design/icons';

// Mock shareAPI since the actual API file may not exist yet
const shareAPI = {
    validate: async (token: string, password?: string): Promise<{ resourceType: string; resourceId: string }> => {
        return new Promise((resolve, reject) => {
            setTimeout(() => {
                if (password && password !== 'test123') {
                    reject(new Error('Invalid password'));
                } else {
                    resolve({ resourceType: 'dashboard', resourceId: token });
                }
            }, 500);
        });
    }
};

const SharePage: React.FC = () => {
    const { token } = useParams<{ token: string }>();
    const [loading, setLoading] = useState(true);
    const [needPassword, setNeedPassword] = useState(false);
    const [password, setPassword] = useState('');
    const [content, setContent] = useState<{ resourceType: string; resourceId: string } | null>(null);

    const validateShare = useCallback(async (pwd?: string) => {
        if (!token) return;

        setLoading(true);
        try {
            const result = await shareAPI.validate(token, pwd);
            setContent(result);
            setNeedPassword(false);
            message.success('验证成功');
        } catch (error) {
            const err = error as Error;
            if (err.message?.includes('password')) {
                setNeedPassword(true);
            } else {
                message.error('分享链接无效或已过期');
            }
        } finally {
            setLoading(false);
        }
    }, [token]);

    useEffect(() => {
        validateShare();
    }, [validateShare]);

    const handlePasswordSubmit = async () => {
        if (!password) {
            message.error('请输入密码');
            return;
        }
        validateShare(password);
    };

    if (loading) {
        return (
            <div style={{ textAlign: 'center', padding: '100px 0' }}>
                <Spin size="large" />
                <div style={{ marginTop: 16 }}>正在加载...</div>
            </div>
        );
    }

    if (needPassword) {
        return (
            <div style={{
                maxWidth: 400,
                margin: '100px auto',
                padding: 24,
                background: '#fff',
                borderRadius: 8,
                boxShadow: '0 2px 8px rgba(0,0,0,0.15)'
            }}>
                <h2 style={{ textAlign: 'center', marginBottom: 24 }}>
                    <LockOutlined /> 请输入访问密码
                </h2>
                <Input.Password
                    placeholder="请输入密码"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    onPressEnter={handlePasswordSubmit}
                    size="large"
                />
                <Button
                    type="primary"
                    block
                    size="large"
                    onClick={handlePasswordSubmit}
                    style={{ marginTop: 16 }}
                >
                    访问
                </Button>
            </div>
        );
    }

    if (!content) {
        return (
            <div style={{ textAlign: 'center', padding: '100px 0' }}>
                <h2>分享链接无效</h2>
            </div>
        );
    }

    // 渲染分享的内容(Dashboard或Chart)
    return (
        <div style={{ padding: 24 }}>
            <h1>{content.resourceType === 'dashboard' ? '仪表板' : '图表'}分享</h1>
            <div>资源ID: {content.resourceId}</div>
            {/* TODO: 根据content.resourceType渲染对应的内容 */}
        </div>
    );
};

export default SharePage;
