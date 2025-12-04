import React from 'react';
import { Column } from '@ant-design/plots';
import { BaseChart } from '../BaseChart';
import type { ChartConfig } from '../../../types/chart';

interface BarChartProps {
    data: any[];
    config: ChartConfig;
    style?: React.CSSProperties;
}

export const BarChart: React.FC<BarChartProps> = ({ data, config, style }) => {
    const { xAxis, yAxis } = config;
    const xField = xAxis?.fields[0]?.name || 'x';
    const yField = yAxis?.fields[0]?.name || 'y';

    const props = {
        data,
        xField,
        yField,
        label: {
            position: 'middle',
            style: {
                fill: '#FFFFFF',
                opacity: 0.6,
            },
        },
        xAxis: {
            label: {
                autoHide: true,
                autoRotate: false,
            },
        },
        meta: {
            [xField]: {
                alias: xAxis?.fields[0]?.name,
            },
            [yField]: {
                alias: yAxis?.fields[0]?.name,
            },
        },
    };

    return (
        <BaseChart style={style}>
            <Column {...props} />
        </BaseChart>
    );
};
