import React from 'react';
import { Card, List } from 'antd';
import {
    BarChartOutlined,
    LineChartOutlined,
    PieChartOutlined,
    TableOutlined,
    FileTextOutlined,
    PictureOutlined,
} from '@ant-design/icons';
import type { ComponentType } from './types';

interface ComponentPaletteProps {
    onAddComponent: (type: ComponentType) => void;
}

const componentTypes = [
    { type: 'chart' as ComponentType, icon: <BarChartOutlined />, label: '图表' },
    { type: 'text' as ComponentType, icon: <FileTextOutlined />, label: '文本' },
    { type: 'image' as ComponentType, icon: <PictureOutlined />, label: '图片' },
];

export const ComponentPalette: React.FC<ComponentPaletteProps> = ({ onAddComponent }) => {
    return (
        <Card title="组件库" size="small">
            <List
                dataSource={componentTypes}
                renderItem={(item) => (
                    <List.Item
                        style={{ cursor: 'pointer', padding: '8px 0' }}
                        onClick={() => onAddComponent(item.type)}
                    >
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                            {item.icon}
                            <span>{item.label}</span>
                        </div>
                    </List.Item>
                )}
            />
        </Card>
    );
};
