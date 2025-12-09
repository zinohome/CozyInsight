import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Input, Button, message, Spin } from 'antd';
import { LockOutlined } from '@ant-design/icons';
import { shareAPI } from '../../api/share';

const SharePage: React.FC = () => {
    const { token } = useParams<{ token: string }>();
    const [loading, setLoading] = useState(true);
    const [needPassword, setNeedPassword] = useState(false);
    const [password, setPassword] = useState('');
    const [share, setShare] = useState<any>(null);
    const [validating, setValidating] = useState(false);

    useEffect(() => {
        validateShare();
    }, [token]);

    const validateShare = async (pwd?: string) => {
        if (!token) return;

        setLoading(true);
        try {
            const result = await shareAPI.validate(token, pwd);
            setShare(result);
            setNeedPassword(false);
            message.success('验证成功');
        } catch (error: any) {
            if (error.message?.includes('password')) {
                setNeedPassword(true);
            } else {
                message.error('分享链接无效或已过期');
            }
        } finally {
            setLoading(false);
        }
    };

    const handlePasswordSubmit = () => {
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
                    loading={validating}
                    style={{ marginTop: 16 }}
                >
                    访问
                </Button>
            </div>
        );
    }

    if (!share) {
        return (
            <div style={{ textAlign: 'center', padding: '100px 0' }}>
                <h2>分享链接无效</h2>
            </div>
        );
    }

    // 渲染分享的内容(Dashboard或Chart)
    return (
        <div style={{ padding: 24 }}>
            <h1>{share.resourceType === 'dashboard' ? '仪表板' : '图表'}分享</h1>
            {/* TODO: 根据share.resourceType渲染对应的内容 */}
            <div>资源ID: {share.resourceId}</div>
        </div>
    );
};

export default SharePage;
