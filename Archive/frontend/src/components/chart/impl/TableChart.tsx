import React from 'react';
import { Table } from 'antd';
import { BaseChart } from '../BaseChart';
import type { ChartConfig } from '../../../types/chart';

interface TableChartProps {
    data: any[];
    config: ChartConfig;
    style?: React.CSSProperties;
}

export const TableChart: React.FC<TableChartProps> = ({ data, config, style }) => {
    const { xAxis, yAxis } = config;

    // 构造列定义
    const columns = [
        ...(xAxis?.fields || []).map(field => ({
            title: field.name,
            dataIndex: field.name,
            key: field.name,
        })),
        ...(yAxis?.fields || []).map(field => ({
            title: field.name,
            dataIndex: field.name,
            key: field.name,
            sorter: (a: any, b: any) => a[field.name] - b[field.name],
        })),
    ];

    return (
        <BaseChart style={style}>
            <Table
                dataSource={data}
                columns={columns}
                pagination={false}
                scroll={{ y: 300 }}
                size="small"
                rowKey={(_, index) => index?.toString() || ''}
            />
        </BaseChart>
    );
};
