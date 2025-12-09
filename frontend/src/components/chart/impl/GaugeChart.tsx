import React from 'react';
import { Gauge } from '@ant-design/charts';
import type { GaugeConfig } from '@ant-design/charts';

interface GaugeChartProps {
    data: any[];
    config: {
        valueField?: string;
        min?: number;
        max?: number;
        title?: string;
    };
}

const GaugeChart: React.FC<GaugeChartProps> = ({ data, config }) => {
    // 从数据中提取值
    const value = data.length > 0 && config.valueField
        ? data[0][config.valueField]
        : (data.length > 0 ? Object.values(data[0])[0] as number : 0);

    const chartConfig: GaugeConfig = {
        percent: value / (config.max || 100),
        range: {
            color: 'l(0) 0:#30BF78 0.5:#FAAD14 1:#F4664A',
        },
        indicator: {
            pointer: {
                style: {
                    stroke: '#D0D0D0',
                },
            },
            pin: {
                style: {
                    stroke: '#D0D0D0',
                },
            },
        },
        axis: {
            label: {
                formatter: (v: string) => {
                    const num = Number(v) * (config.max || 100);
                    return num.toFixed(0);
                },
            },
            subTickLine: {
                count: 3,
            },
        },
        statistic: {
            content: {
                formatter: () => {
                    return `${value.toFixed(2)}`;
                },
                style: {
                    fontSize: '36px',
                    lineHeight: '36px',
                },
            },
        },
    };

    return <Gauge {...chartConfig} />;
};

export default GaugeChart;
