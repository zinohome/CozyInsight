import React from 'react';
import { Area } from '@ant-design/charts';
import type { AreaConfig } from '@ant-design/charts';

interface AreaChartProps {
    data: any[];
    config: {
        xField: string;
        yField: string;
        seriesField?: string;
        smooth?: boolean;
        title?: string;
    };
}

const AreaChart: React.FC<AreaChartProps> = ({ data, config }) => {
    const chartConfig: AreaConfig = {
        data,
        xField: config.xField,
        yField: config.yField,
        seriesField: config.seriesField,
        areaStyle: {
            fillOpacity: 0.6,
        },
        xAxis: {
            type: 'cat',
            tickLine: null,
        },
        yAxis: {
            label: {
                formatter: (v: string) => `${v}`,
            },
        },
        legend: config.seriesField ? {
            position: 'top-right',
        } : false,
        tooltip: {
            showMarkers: true,
        },
    };

    return <Area {...chartConfig} />;
};

export default AreaChart;
