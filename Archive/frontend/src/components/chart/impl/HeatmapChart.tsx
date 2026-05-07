import React from 'react';
import { Heatmap } from '@ant-design/charts';
import type { HeatmapConfig } from '@ant-design/charts';

interface HeatmapChartProps {
    data: any[];
    config: {
        xField: string;
        yField: string;
        colorField: string;
        title?: string;
    };
}

const HeatmapChart: React.FC<HeatmapChartProps> = ({ data, config }) => {
    const chartConfig: HeatmapConfig = {
        data,
        xField: config.xField,
        yField: config.yField,
        colorField: config.colorField,
        color: ['#174c83', '#7eb6d4', '#efefeb', '#efa759', '#9b4d16'],
        meta: {
            [config.xField]: {
                type: 'cat',
            },
            [config.yField]: {
                type: 'cat',
            },
        },
        label: {
            style: {
                fill: '#fff',
                shadowBlur: 2,
                shadowColor: 'rgba(0, 0, 0, .45)',
            },
        },
        xAxis: {
            line: null,
            tickLine: null,
        },
        yAxis: {
            line: null,
            tickLine: null,
        },
        tooltip: {
            showMarkers: false,
        },
        interactions: [
            {
                type: 'element-active',
            },
        ],
    };

    return <Heatmap {...chartConfig} />;
};

export default HeatmapChart;
