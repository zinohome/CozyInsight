import React from 'react';
import { Pie } from '@ant-design/plots';
import { BaseChart } from '../BaseChart';
import type { ChartConfig } from '../../../types/chart';

interface PieChartProps {
    data: any[];
    config: ChartConfig;
    style?: React.CSSProperties;
}

export const PieChart: React.FC<PieChartProps> = ({ data, config, style }) => {
    const { xAxis, yAxis } = config;
    // 对于饼图，通常用维度作为 colorField，指标作为 angleField
    const colorField = xAxis?.fields[0]?.name || 'type';
    const angleField = yAxis?.fields[0]?.name || 'value';

    const props = {
        appendPadding: 10,
        data,
        angleField,
        colorField,
        radius: 0.8,
        label: {
            type: 'outer',
            content: '{name} {percentage}',
        },
        interactions: [
            {
                type: 'pie-legend-active',
            },
            {
                type: 'element-active',
            },
        ],
    };

    return (
        <BaseChart style={style}>
            <Pie {...props} />
        </BaseChart>
    );
};
