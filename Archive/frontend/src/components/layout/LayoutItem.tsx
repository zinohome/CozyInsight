import React from 'react';
import { Card } from 'antd';
import { DragOutlined, DeleteOutlined, SettingOutlined } from '@ant-design/icons';
import type { LayoutItem as LayoutItemType } from './types';

interface LayoutItemProps {
    item: LayoutItemType;
    children: React.ReactNode;
    onDelete?: () => void;
    onConfig?: () => void;
}

export const LayoutItem: React.FC<LayoutItemProps> = ({
    item,
    children,
    onDelete,
    onConfig,
}) => {
    return (
        <Card
            size="small"
            title={
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <DragOutlined style={{ cursor: 'move' }} />
                    <span>{item.componentType || '组件'}</span>
                </div>
            }
            extra={
                <div style={{ display: 'flex', gap: 8 }}>
                    {onConfig && (
                        <SettingOutlined
                            onClick={(e) => {
                                e.stopPropagation();
                                onConfig();
                            }}
                            style={{ cursor: 'pointer' }}
                        />
                    )}
                    {onDelete && (
                        <DeleteOutlined
                            onClick={(e) => {
                                e.stopPropagation();
                                onDelete();
                            }}
                            style={{ cursor: 'pointer', color: '#ff4d4f' }}
                        />
                    )}
                </div>
            }
            style={{ height: '100%', overflow: 'auto' }}
        >
            {children}
        </Card>
    );
};
