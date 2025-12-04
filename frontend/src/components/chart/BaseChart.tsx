import React from 'react';
import { Spin, Empty } from 'antd';

interface BaseChartProps {
    loading?: boolean;
    empty?: boolean;
    children: React.ReactNode;
    style?: React.CSSProperties;
    className?: string;
}

export const BaseChart: React.FC<BaseChartProps> = ({
    loading = false,
    empty = false,
    children,
    style,
    className,
}) => {
    if (loading) {
        return (
            <div
                style={{
                    display: 'flex',
                    justifyContent: 'center',
                    alignItems: 'center',
                    height: '100%',
                    width: '100%',
                    minHeight: 200,
                    ...style,
                }}
                className={className}
            >
                <Spin size="large" />
            </div>
        );
    }

    if (empty) {
        return (
            <div
                style={{
                    display: 'flex',
                    justifyContent: 'center',
                    alignItems: 'center',
                    height: '100%',
                    width: '100%',
                    minHeight: 200,
                    ...style,
                }}
                className={className}
            >
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据" />
            </div>
        );
    }

    return (
        <div style={{ height: '100%', width: '100%', minHeight: 200, ...style }} className={className}>
            {children}
        </div>
    );
};
