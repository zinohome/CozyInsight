import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
    Tabs,
    Button,
    Card,
    message,
    Space,
    Spin,
} from 'antd';
import { SaveOutlined, ArrowLeftOutlined } from '@ant-design/icons';
import { dashboardAPI } from '../../api/dashboard';
import { LayoutEditor } from '../../components/layout/LayoutEditor';
import type { UpdateDashboardRequest } from '../../types/dashboard';
import type { LayoutItem } from '../../components/layout/types';

const { TabPane } = Tabs;

const DashboardEditor = () => {
    const navigate = useNavigate();
    const { id } = useParams<{ id: string }>();
    const [loading, setLoading] = useState(false);
    const [dashboardName, setDashboardName] = useState('');
    const [currentLayout, setCurrentLayout] = useState<LayoutItem[]>([]);

    // 加载仪表板详情
    const loadDashboard = async (dashboardId: string) => {
        try {
            setLoading(true);
            const data = await dashboardAPI.get(dashboardId);
            setDashboardName(data.name);

            // 解析布局数据
            if (data.componentData) {
                try {
                    const parsed = JSON.parse(data.componentData);
                    setCurrentLayout(parsed.layouts || []);
                } catch (e) {
                    console.error('Failed to parse componentData', e);
                }
            }
        } catch (error: any) {
            message.error(error.message || '加载仪表板失败');
            navigate('/dashboard');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (id) {
            loadDashboard(id);
        }
    }, [id]);

    // 保存布局
    const handleSaveLayout = async (layout: LayoutItem[]) => {
        if (!id) return;

        try {
            const componentData = JSON.stringify({ layouts: layout });
            const updateData: UpdateDashboardRequest = {
                componentData,
            };

            await dashboardAPI.update(id, updateData);
            message.success('布局已保存');
            setCurrentLayout(layout);
        } catch (error: any) {
            message.error(error.message || '保存失败');
        }
    };

    if (loading) {
        return (
            <div style={{ textAlign: 'center', padding: '100px' }}>
                <Spin size="large" tip="加载中..." />
            </div>
        );
    }

    return (
        <div style={{ padding: '24px' }}>
            <Card
                title={
                    <Space>
                        <Button
                            type="text"
                            icon={<ArrowLeftOutlined />}
                            onClick={() => navigate('/dashboard')}
                        />
                        <span>编辑仪表板: {dashboardName}</span>
                    </Space>
                }
            >
                <Tabs defaultActiveKey="layout">
                    <TabPane tab="布局编辑" key="layout">
                        <LayoutEditor
                            initialLayout={currentLayout}
                            onSave={handleSaveLayout}
                        />
                    </TabPane>
                    <TabPane tab="基础设置" key="settings">
                        <div style={{ padding: 16 }}>
                            <p>ℹ️ 基础设置功能开发中...</p>
                        </div>
                    </TabPane>
                </Tabs>
            </Card>
        </div>
    );
};

export default DashboardEditor;
