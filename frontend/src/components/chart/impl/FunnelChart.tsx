import React from 'react';
import { Funnel } from '@ant-design/charts';
import type { FunnelConfig } from '@ant-design/charts';

interface FunnelChartProps {
    data: any[];
    config: {
        xField: string;
        yField: string;
        title?: string;
    };
}

const FunnelChart: React.FC<FunnelChartProps> = ({ data, config }) => {
    const chartConfig: FunnelConfig = {
        data,
        xField: config.xField,
        yField: config.yField,
        legend: false,
        label: {
            formatter: (datum: any) => {
                return `${datum[config.xField]}: ${datum[config.yField]}`;
            },
        },
        conversionTag: {
            formatter: (datum: any) => {
                return `转化率: ${(datum.$$percentage$$ * 100).toFixed(2)}%`;
            },
        },
        tooltip: {
            showTitle: true,
            formatter: (datum: any) => {
                return {
                    name: datum[config.xField],
                    value: datum[config.yField],
                };
            },
        },
    };

    return <Funnel {...chartConfig} />;
};

export default FunnelChart;
