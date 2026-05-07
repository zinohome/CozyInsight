import React, { useState, useCallback } from 'react';
import { Responsive, WidthProvider, Layout } from 'react-grid-layout';
import { Card, Button, Space, Modal, Select } from 'antd';
import { PlusOutlined, DeleteOutlined, EditOutlined } from '@ant-design/icons';
import ChartRenderer from './ChartRenderer';
import 'react-grid-layout/css/styles.css';
import 'react-resizable/css/styles.css';

const ResponsiveGridLayout = WidthProvider(Responsive);

export interface DashboardItem {
    i: string; // 组件 ID
    x: number;
    y: number;
    w: number;
    h: number;
    minW?: number;
    minH?: number;
    maxW?: number;
    maxH?: number;
    chartId?: string; // 关联的图表 ID
    type: 'chart' | 'text' | 'image';
    config?: any;
}

export interface DashboardLayoutProps {
    items: DashboardItem[];
    editable?: boolean;
    onLayoutChange?: (layout: Layout[]) => void;
    onItemAdd?: (type: string) => void;
    onItemRemove?: (id: string) => void;
    onItemEdit?: (id: string) => void;
}

/**
 * 仪表板网格布局组件
 * 支持拖拽、调整大小和响应式布局
 */
export const DashboardLayout: React.FC<DashboardLayoutProps> = ({
    items,
    editable = false,
    onLayoutChange,
    onItemAdd,
    onItemRemove,
    onItemEdit,
}) => {
    const [layouts, setLayouts] = useState<{ [key: string]: Layout[] }>({});
    const [addModalVisible, setAddModalVisible] = useState(false);
    const [selectedType, setSelectedType] = useState<string>('chart');

    // 处理布局变化
    const handleLayoutChange = useCallback(
        (currentLayout: Layout[], allLayouts: { [key: string]: Layout[] }) => {
            setLayouts(allLayouts);
            if (onLayoutChange) {
                onLayoutChange(currentLayout);
            }
        },
        [onLayoutChange]
    );

    // 添加组件
    const handleAddItem = () => {
        if (onItemAdd) {
            onItemAdd(selectedType);
        }
        setAddModalVisible(false);
    };

    // 删除组件
    const handleRemoveItem = (id: string) => {
        Modal.confirm({
            title: '确认删除',
            content: '确定要删除这个组件吗？',
            okText: '确定',
            cancelText: '取消',
            onOk: () => {
                if (onItemRemove) {
                    onItemRemove(id);
                }
            },
        });
    };

    // 渲染单个组件
    const renderItem = (item: DashboardItem) => {
        switch (item.type) {
            case 'chart':
                return (
                    <div style={{ height: '100%', position: 'relative' }}>
                        {editable && (
                            <div
                                style={{
                                    position: 'absolute',
                                    top: 8,
                                    right: 8,
                                    zIndex: 10,
                                }}
                            >
                                <Space>
                                    <Button
                                        size="small"
                                        icon={<EditOutlined />}
                                        onClick={() => onItemEdit && onItemEdit(item.i)}
                                    />
                                    <Button
                                        size="small"
                                        danger
                                        icon={<DeleteOutlined />}
                                        onClick={() => handleRemoveItem(item.i)}
                                    />
                                </Space>
                            </div>
                        )}
                        {item.chartId ? (
                            <ChartRenderer
                                config={item.config || { type: 'column', data: [] }}
                                style={{ height: '100%' }}
                            />
                        ) : (
                            <div
                                style={{
                                    height: '100%',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                    color: '#999',
                                }}
                            >
                                点击配置图表
                            </div>
                        )}
                    </div>
                );

            case 'text':
                return (
                    <div style={{ padding: 16, height: '100%', overflow: 'auto' }}>
                        <div
                            dangerouslySetInnerHTML={{
                                __html: item.config?.content || '双击编辑文本',
                            }}
                        />
                    </div>
                );

            case 'image':
                return (
                    <div
                        style={{
                            height: '100%',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                        }}
                    >
                        {item.config?.url ? (
                            <img
                                src={item.config.url}
                                alt="Dashboard"
                                style={{ maxWidth: '100%', maxHeight: '100%' }}
                            />
                        ) : (
                            <div style={{ color: '#999' }}>点击上传图片</div>
                        )}
                    </div>
                );

            default:
                return <div>未知组件类型</div>;
        }
    };

    return (
        <div style={{ background: '#f0f2f5', minHeight: '100vh', padding: '24px' }}>
            {editable && (
                <div style={{ marginBottom: 16 }}>
                    <Button
                        type="primary"
                        icon={<PlusOutlined />}
                        onClick={() => setAddModalVisible(true)}
                    >
                        添加组件
                    </Button>
                </div>
            )}

            <ResponsiveGridLayout
                className="layout"
                layouts={layouts}
                breakpoints={{ lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 }}
                cols={{ lg: 12, md: 10, sm: 6, xs: 4, xxs: 2 }}
                rowHeight={60}
                isDraggable={editable}
                isResizable={editable}
                onLayoutChange={handleLayoutChange}
                draggableHandle=".drag-handle"
            >
                {items.map((item) => (
                    <div key={item.i} data-grid={item}>
                        <Card
                            className={editable ? 'drag-handle' : ''}
                            style={{
                                height: '100%',
                                cursor: editable ? 'move' : 'default',
                            }}
                            bodyStyle={{ height: 'calc(100% - 57px)', padding: 0 }}
                        >
                            {renderItem(item)}
                        </Card>
                    </div>
                ))}
            </ResponsiveGridLayout>

            <Modal
                title="添加组件"
                open={addModalVisible}
                onOk={handleAddItem}
                onCancel={() => setAddModalVisible(false)}
            >
                <Select
                    style={{ width: '100%' }}
                    value={selectedType}
                    onChange={setSelectedType}
                >
                    <Select.Option value="chart">图表</Select.Option>
                    <Select.Option value="text">文本</Select.Option>
                    <Select.Option value="image">图片</Select.Option>
                </Select>
            </Modal>
        </div>
    );
};

export default DashboardLayout;
