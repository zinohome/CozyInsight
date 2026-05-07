import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button, Space, message, Input, Modal } from 'antd';
import { SaveOutlined, EyeOutlined, PlusOutlined } from '@ant-design/icons';
import DashboardLayout, { DashboardItem } from '../components/DashboardLayout';
import type { Layout } from 'react-grid-layout';

/**
 * 仪表板编辑器页面
 */
const DashboardEditor: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();

    const [dashboardName, setDashboardName] = useState('新建仪表板');
    const [items, setItems] = useState<DashboardItem[]>([]);
    const [editMode, setEditMode] = useState(true);
    const [saving, setSaving] = useState(false);

    // 示例数据 - 实际应从 API 加载
    useEffect(() => {
        if (id) {
            loadDashboard(id);
        } else {
            // 新建仪表板，添加一些示例组件
            setItems([
                {
                    i: '1',
                    x: 0,
                    y: 0,
                    w: 6,
                    h: 4,
                    type: 'chart',
                    chartId: 'chart1',
                    config: {
                        type: 'column',
                        data: [
                            { category: '一月', value: 3500 },
                            { category: '二月', value: 4200 },
                            { category: '三月', value: 3800 },
                        ],
                        xField: 'category',
                        yField: 'value',
                    },
                },
                {
                    i: '2',
                    x: 6,
                    y: 0,
                    w: 6,
                    h: 4,
                    type: 'chart',
                    chartId: 'chart2',
                    config: {
                        type: 'pie',
                        data: [
                            { type: '分类一', value: 27 },
                            { type: '分类二', value: 25 },
                            { type: '分类三', value: 18 },
                            { type: '分类四', value: 15 },
                        ],
                        angleField: 'value',
                        colorField: 'type',
                    },
                },
            ]);
        }
    }, [id]);

    const loadDashboard = async (dashboardId: string) => {
        try {
            // TODO: 实际 API 调用
            // const data = await dashboardApi.get(dashboardId);
            // setDashboardName(data.name);
            // setItems(JSON.parse(data.layout));
        } catch (error) {
            message.error('加载仪表板失败');
        }
    };

    const handleLayoutChange = (layout: Layout[]) => {
        // 更新布局信息
        setItems((prevItems) =>
            prevItems.map((item) => {
                const layoutItem = layout.find((l) => l.i === item.i);
                if (layoutItem) {
                    return {
                        ...item,
                        x: layoutItem.x,
                        y: layoutItem.y,
                        w: layoutItem.w,
                        h: layoutItem.h,
                    };
                }
                return item;
            })
        );
    };

    const handleAddItem = (type: string) => {
        const newId = `item-${Date.now()}`;
        const newItem: DashboardItem = {
            i: newId,
            x: 0,
            y: Infinity, // 添加到最后
            w: 6,
            h: 4,
            type: type as any,
        };

        setItems([...items, newItem]);
        message.success('组件添加成功');
    };

    const handleRemoveItem = (id: string) => {
        setItems(items.filter((item) => item.i !== id));
        message.success('组件删除成功');
    };

    const handleEditItem = (id: string) => {
        // 打开组件编辑弹窗
        Modal.info({
            title: '编辑组件',
            content: '组件编辑功能开发中...',
        });
    };

    const handleSave = async () => {
        setSaving(true);
        try {
            const dashboardData = {
                name: dashboardName,
                layout: JSON.stringify(items),
            };

            // TODO: 实际 API 调用
            // if (id) {
            //   await dashboardApi.update(id, dashboardData);
            // } else {
            //   const newDashboard = await dashboardApi.create(dashboardData);
            //   navigate(`/dashboard/edit/${newDashboard.id}`);
            // }

            message.success('保存成功');
        } catch (error) {
            message.error('保存失败');
        } finally {
            setSaving(false);
        }
    };

    const toggleEditMode = () => {
        setEditMode(!editMode);
    };

    return (
        <div style={{ height: '100vh', display: 'flex', flexDirection: 'column' }}>
            {/* 顶部工具栏 */}
            <div
                style={{
                    background: '#fff',
                    padding: '16px 24px',
                    borderBottom: '1px solid #f0f0f0',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                }}
            >
                <Input
                    value={dashboardName}
                    onChange={(e) => setDashboardName(e.target.value)}
                    style={{ width: 300 }}
                    placeholder="仪表板名称"
                />

                <Space>
                    <Button onClick={() => navigate('/dashboard')}>取消</Button>
                    <Button onClick={toggleEditMode}>
                        {editMode ? '预览' : '编辑'}
                    </Button>
                    <Button
                        type="primary"
                        icon={<SaveOutlined />}
                        loading={saving}
                        onClick={handleSave}
                    >
                        保存
                    </Button>
                </Space>
            </div>

            {/* 仪表板画布 */}
            <div style={{ flex: 1, overflow: 'auto' }}>
                <DashboardLayout
                    items={items}
                    editable={editMode}
                    onLayoutChange={handleLayoutChange}
                    onItemAdd={handleAddItem}
                    onItemRemove={handleRemoveItem}
                    onItemEdit={handleEditItem}
                />
            </div>
        </div>
    );
};

export default DashboardEditor;
