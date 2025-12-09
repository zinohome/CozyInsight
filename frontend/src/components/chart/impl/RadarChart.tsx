import React from 'react';
import { Radar } from '@ant-design/charts';
import type { RadarConfig } from '@ant-design/charts';

interface RadarChartProps {
    data: any[];
    config: {
        xField: string;
        yField: string;
        seriesField?: string;
        title?: string;
    };
}

const RadarChart: React.FC<RadarChartProps> = ({ data, config }) => {
    const chartConfig: RadarConfig = {
        data,
        xField: config.xField,
        yField: config.yField,
        seriesField: config.seriesField,
        point: {
            size: 3,
        },
        area: {
            style: {
                fillOpacity: 0.2,
            },
        },
        xAxis: {
            line: null,
            tickLine: null,
            grid: {
                line: {
                    style: {
                        lineDash: null,
                    },
                },
            },
        },
        yAxis: {
            line: null,
            tickLine: null,
            grid: {
                line: {
                    type: 'line',
                    style: {
                        lineDash: null,
                    },
                },
                alternateColor: 'rgba(0, 0, 0, 0.04)',
            },
        },
        legend: config.seriesField ? {
            position: 'top-right',
        } : false,
    };

    return <Radar {...chartConfig} />;
};

export default RadarChart;
