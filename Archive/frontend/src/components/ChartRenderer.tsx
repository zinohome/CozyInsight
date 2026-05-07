import React from 'react';
import { Column, Line, Pie, Bar, Scatter, Area, Radar } from '@ant-design/charts';

export interface ChartConfig {
  type: string;
  data: any[];
  xField?: string;
  yField?: string;
  seriesField?: string;
  angleField?: string;
  colorField?: string;

  // 样式配置
  color?: string | string[];
  legend?: any;
  title?: any;
  xAxis?: any;
  yAxis?: any;
  label?: any;

  [key: string]: any;
}

export interface ChartRendererProps {
  config: ChartConfig;
  loading?: boolean;
  style?: React.CSSProperties;
}

/**
 * 图表渲染器组件
 * 支持多种图表类型的统一渲染和样式配置
 */
export const ChartRenderer: React.FC<ChartRendererProps> = ({
  config,
  loading = false,
  style
}) => {
  if (loading) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100%',
        ...style
      }}>
        加载中...
      </div>
    );
  }

  if (!config || !config.data || config.data.length === 0) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100%',
        color: '#999',
        ...style
      }}>
        暂无数据
      </div>
    );
  }

  // 构建通用配置
  const commonConfig = {
    ...config,
    color: config.color,
    legend: config.legend !== false ? {
      position: config.legend?.position || 'bottom',
      ...config.legend,
    } : false,
    label: config.label?.show ? {
      position: config.label?.position || 'top',
      style: {
        fontSize: config.label?.fontSize || 12,
      },
    } : false,
  };

  const renderChart = () => {
    switch (config.type) {
      case 'column':
      case 'bar-vertical':
        return <Column {...commonConfig} style={style} />;

      case 'bar':
      case 'bar-horizontal':
        return <Bar {...commonConfig} style={style} />;

      case 'line':
        return <Line {...commonConfig} style={style} />;

      case 'pie':
        return <Pie {...commonConfig} style={style} />;

      case 'scatter':
        return <Scatter {...commonConfig} style={style} />;

      case 'area':
        return <Area {...commonConfig} style={style} />;

      case 'radar':
        return <Radar {...commonConfig} style={style} />;

      default:
        return (
          <div style={{
            padding: '20px',
            textAlign: 'center',
            color: '#999',
            ...style
          }}>
            不支持的图表类型: {config.type}
          </div>
        );
    }
  };

  return (
    <div style={{ width: '100%', height: '100%', ...style }}>
      {renderChart()}
    </div>
  );
};

export default ChartRenderer;
