import React, { useState } from 'react';
import { Row, Col, Button, message } from 'antd';
import { SaveOutlined } from '@ant-design/icons';
import { GridLayout } from './GridLayout';
import { LayoutItem } from './LayoutItem';
import { ComponentPalette } from './ComponentPalette';
import { ChartRenderer } from '../chart/ChartRenderer';
import type { LayoutItem as LayoutItemType, ComponentType } from './types';
import type { Layout } from 'react-grid-layout';

interface LayoutEditorProps {
    initialLayout?: LayoutItemType[];
    onSave?: (layout: LayoutItemType[]) => void;
}

// Mock 数据生成器
const generateMockData = () => {
    return [
        { type: '分类一', value: 27 },
        { type: '分类二', value: 25 },
        { type: '分类三', value: 18 },
    ];
};

export const LayoutEditor: React.FC<LayoutEditorProps> = ({
    initialLayout = [],
    onSave,
}) => {
    const [layout, setLayout] = useState<LayoutItemType[]>(initialLayout);
    const [nextId, setNextId] = useState(layout.length);

    // 添加组件
    const handleAddComponent = (type: ComponentType) => {
        const newItem: LayoutItemType = {
            i: `component-${nextId}`,
            x: (nextId % 3) * 4,
            y: Math.floor(nextId / 3) * 6,
            w: 4,
            h: 6,
            componentType: type,
            componentId: type === 'chart' ? `chart-${nextId}` : undefined,
            config: {},
        };

        setLayout([...layout, newItem]);
        setNextId(nextId + 1);
        message.success(`已添加${type}组件`);
    };

    // 删除组件
    const handleDeleteComponent = (id: string) => {
        setLayout(layout.filter((item) => item.i !== id));
        message.success('组件已删除');
    };

    // 布局变化
    const handleLayoutChange = (newLayout: Layout[]) => {
        const updatedLayout = layout.map((item) => {
            const updated = newLayout.find((l) => l.i === item.i);
            return updated ? { ...item, ...updated } : item;
        });
        setLayout(updatedLayout);
    };

    // 保存布局
    const handleSave = () => {
        if (onSave) {
            onSave(layout);
            message.success('布局已保存');
        }
    };

    // 渲染组件内容
    const renderComponentContent = (item: LayoutItemType) => {
        switch (item.componentType) {
            case 'chart':
                return (
                    <ChartRenderer
                        type="bar"
                        data={generateMockData()}
                        config={{
                            xAxis: { fields: [{ id: '1', name: 'type' }] },
                            yAxis: { fields: [{ id: '2', name: 'value' }] },
                        }}
                    />
                );
            case 'text':
                return <div style={{ padding: 16 }}>文本组件</div>;
            case 'image':
                return <div style={{ padding: 16 }}>图片组件</div>;
            default:
                return <div style={{ padding: 16 }}>未知组件</div>;
        }
    };

    return (
        <div>
            <Row gutter={16} style={{ marginBottom: 16 }}>
                <Col>
                    <Button type="primary" icon={<SaveOutlined />} onClick={handleSave}>
                        保存布局
                    </Button>
                </Col>
            </Row>

            <Row gutter={16}>
                {/* 左侧组件面板 */}
                <Col span={4}>
                    <ComponentPalette onAddComponent={handleAddComponent} />
                </Col>

                {/* 右侧画布 */}
                <Col span={20}>
                    <div
                        style={{
                            border: '1px dashed #d9d9d9',
                            minHeight: 600,
                            background: '#fafafa',
                            padding: 16,
                        }}
                    >
                        {layout.length === 0 ? (
                            <div
                                style={{
                                    textAlign: 'center',
                                    padding: '100px 0',
                                    color: '#999',
                                }}
                            >
                                从左侧拖拽组件到此处开始设计
                            </div>
                        ) : (
                            <GridLayout layout={layout} onLayoutChange={handleLayoutChange}>
                                {layout.map((item) => (
                                    <div key={item.i}>
                                        <LayoutItem
                                            item={item}
                                            onDelete={() => handleDeleteComponent(item.i)}
                                            onConfig={() => message.info('配置功能开发中')}
                                        >
                                            {renderComponentContent(item)}
                                        </LayoutItem>
                                    </div>
                                ))}
                            </GridLayout>
                        )}
                    </div>
                </Col>
            </Row>
        </div>
    );
};
