import React from 'react';
import { Scatter } from '@ant-design/charts';
import type { ScatterConfig } from '@ant-design/charts';

interface ScatterChartProps {
    data: any[];
    config: {
        xField: string;
        yField: string;
        sizeField?: string;
        colorField?: string;
        title?: string;
    };
}

const ScatterChart: React.FC<ScatterChartProps> = ({ data, config }) => {
    const chartConfig: ScatterConfig = {
        data,
        xField: config.xField,
        yField: config.yField,
        colorField: config.colorField,
        sizeField: config.sizeField,
        size: config.sizeField ? [4, 30] : 4,
        shape: 'circle',
        pointStyle: {
            fillOpacity: 0.8,
            stroke: '#bbb',
        },
        xAxis: {
            nice: true,
            line: {
                style: {
                    stroke: '#aaa',
                },
            },
        },
        yAxis: {
            nice: true,
            line: {
                style: {
                    stroke: '#aaa',
                },
            },
        },
        tooltip: {
            showTitle: true,
            showMarkers: true,
        },
        legend: config.colorField ? {
            position: 'top-right',
        } : false,
    };

    return <Scatter {...chartConfig} />;
};

export default ScatterChart;
