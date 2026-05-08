import { Bar, Line, Pie, Area, Scatter, Radar, Funnel, WordCloud, Sankey, Heatmap, Treemap, Gauge } from '@ant-design/charts'
import { Table } from 'antd'
import type { ColumnsType } from 'antd/es/table'

interface ChartRendererProps {
  type: string
  data: Array<Record<string, unknown>>
  config: {
    dimensions: string[]
    metrics: string[]
  }
  height?: number
}

export default function ChartRenderer({ type, data, config, height = 300 }: ChartRendererProps) {
  if (!data || data.length === 0) {
    return (
      <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>
        暂无数据
      </div>
    )
  }

  const { dimensions, metrics } = config
  if (dimensions.length === 0 || metrics.length === 0) {
    return (
      <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>
        配置不完整
      </div>
    )
  }

  const xField = dimensions[0]
  const yField = metrics[0]

  switch (type) {
    case 'bar':
      return (
        <Bar
          data={data}
          xField={xField}
          yField={yField}
          height={height}
          autoFit
        />
      )
    case 'line':
      return (
        <Line
          data={data}
          xField={xField}
          yField={yField}
          height={height}
          autoFit
        />
      )
    case 'pie': {
      const colorField = dimensions[0]
      const angleField = metrics[0]
      return (
        <Pie
          data={data}
          angleField={angleField}
          colorField={colorField}
          height={height}
          autoFit
        />
      )
    }
    case 'area':
      return (
        <Area
          data={data}
          xField={xField}
          yField={yField}
          height={height}
          autoFit
        />
      )
    case 'scatter':
      return (
        <Scatter
          data={data}
          xField={xField}
          yField={yField}
          height={height}
          autoFit
        />
      )
    case 'radar':
      return (
        <Radar
          data={data}
          xField={xField}
          yField={yField}
          height={height}
          autoFit
        />
      )
    case 'funnel':
      return (
        <Funnel
          data={data}
          xField={xField}
          yField={yField}
          height={height}
          autoFit
        />
      )
    case 'wordcloud':
      return (
        <WordCloud
          data={data}
          wordField={xField}
          weightField={yField}
          height={height}
          autoFit
        />
      )
    case 'sankey':
      return (
        <Sankey
          data={data}
          sourceField={xField}
          targetField={yField}
          weightField={metrics[0]}
          height={height}
          autoFit
        />
      )
    case 'heatmap':
      return (
        <Heatmap
          data={data}
          xField={xField}
          yField={dimensions[1] || yField}
          colorField={yField}
          height={height}
          autoFit
        />
      )
    case 'treemap':
      return (
        <Treemap
          data={data}
          colorField={xField}
          valueField={yField}
          height={height}
          autoFit
        />
      )
    case 'gauge': {
      let percent = Number(data[0]?.[yField]) || 0
      percent = Math.max(0, Math.min(1, percent))
      return (
        <Gauge
          percent={percent}
          height={height}
          autoFit
        />
      )
    }
    case 'table': {
      const cols: ColumnsType<Record<string, unknown>> = []
      for (const d of dimensions) {
        cols.push({ title: d, dataIndex: d, key: d })
      }
      for (const m of metrics) {
        cols.push({ title: m, dataIndex: m, key: m })
      }
      return <Table columns={cols} dataSource={data} rowKey={(_, idx) => idx!} pagination={false} />
    }
    default:
      return (
        <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>
          不支持的图表类型
        </div>
      )
  }
}
