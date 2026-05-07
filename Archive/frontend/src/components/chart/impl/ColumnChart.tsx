import React from 'react';
import { Column } from '@ant-design/charts';
import type { ColumnConfig } from '@ant-design/charts';

interface ColumnChartProps {
    data: any[];
    config: {
        xField: string;
        yField: string;
        seriesField?: string;
        title?: string;
    };
}

const ColumnChart: React.FC<ColumnChartProps> = ({ data, config }) => {
    const chartConfig: ColumnConfig = {
        data,
        xField: config.xField,
        yField: config.yField,
        seriesField: config.seriesField,
        columnStyle: {
            radius: [8, 8, 0, 0],
        },
        label: {
            position: 'top',
            style: {
                fill: '#000000',
                opacity: 0.6,
            },
        },
        xAxis: {
            label: {
                autoHide: true,
                autoRotate: false,
            },
        },
        legend: config.seriesField ? {
            position: 'top-right',
        } : false,
        tooltip: {
            showMarkers: true,
        },
    };

    return <Column {...chartConfig} />;
};

export default ColumnChart;
