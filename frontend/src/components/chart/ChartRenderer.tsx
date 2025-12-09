import React from 'react';
import { Empty } from 'antd';
import { BarChart } from './impl/BarChart';
import { LineChart } from './impl/LineChart';
import { PieChart } from './impl/PieChart';
import { TableChart } from './impl/TableChart';
import ScatterChart from './impl/ScatterChart';
import RadarChart from './impl/RadarChart';
import HeatmapChart from './impl/HeatmapChart';
import AreaChart from './impl/AreaChart';
import FunnelChart from './impl/FunnelChart';
import GaugeChart from './impl/GaugeChart';
import WordCloudChart from './impl/WordCloudChart';
import ColumnChart from './impl/ColumnChart';
import type { ChartType, ChartConfig } from '../../types/chart';

interface ChartRendererProps {
    type: ChartType;
    data: any[];
    config: ChartConfig;
    style?: React.CSSProperties;
    loading?: boolean;
}

export const ChartRenderer: React.FC<ChartRendererProps> = ({
    type,
    data,
    config,
    style,
    loading,
}) => {
    // 如果没有数据且不在加载中，显示空状态
    if (!loading && (!data || data.length === 0)) {
        return (
            <div style={{ ...style, display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据" />
            </div>
        );
    }

    const commonProps = {
        data,
        config,
        style,
    };

    switch (type) {
        case 'bar':
            return <BarChart {...commonProps} />;
        case 'column':
            return <ColumnChart {...commonProps} />;
        case 'line':
            return <LineChart {...commonProps} />;
        case 'pie':
            return <PieChart {...commonProps} />;
        case 'table':
            return <TableChart {...commonProps} />;
        case 'scatter':
            return <ScatterChart {...commonProps} />;
        case 'radar':
            return <RadarChart {...commonProps} />;
        case 'heatmap':
            return <HeatmapChart {...commonProps} />;
        case 'area':
            return <AreaChart {...commonProps} />;
        case 'funnel':
            return <FunnelChart {...commonProps} />;
        case 'gauge':
            return <GaugeChart {...commonProps} />;
        case 'wordcloud':
            return <WordCloudChart {...commonProps} />;
        default:
            return (
                <div style={{ ...style, display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
                    <Empty description={`暂不支持的图表类型: ${type}`} />
                </div>
            );
    }
};
