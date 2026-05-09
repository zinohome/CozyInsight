import { useCallback } from 'react'
import { Bar, Line, Pie, Area, Scatter, Radar, Funnel, WordCloud, Sankey, Heatmap, Treemap, Gauge } from '@ant-design/charts'
import { Table } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import type { ChartRendererProps, ChartEvent } from '../../types/chart'

export default function ChartRenderer({ type, data, config, height = 300, onEvent }: ChartRendererProps) {
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

  const handleEvent = useCallback(
    (_chart: unknown, event: Record<string, unknown>) => {
      if (event.type === 'element:click' && onEvent) {
        const record = (event.data as Record<string, unknown> | undefined)?.data as Record<string, unknown> | undefined
        if (!record) return

        const dimensionField = config.dimensions[0]
        if (!dimensionField) return

        const chartEvent: ChartEvent = {
          type: 'element:click',
          dimensionField,
          dimensionValue: record[dimensionField] as string | number,
          metrics: record,
        }
        onEvent(chartEvent)
      }
    },
    [config.dimensions, onEvent]
  )

  switch (type) {
    case 'bar':
      return (
        <Bar
          data={data}
          xField={xField}
          yField={yField}
          height={height}
          autoFit
          onEvent={handleEvent}
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
          onEvent={handleEvent}
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
          onEvent={handleEvent}
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
          onEvent={handleEvent}
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
          onEvent={handleEvent}
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
          onEvent={handleEvent}
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
          onEvent={handleEvent}
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
          onEvent={handleEvent}
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
          onEvent={handleEvent}
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
          onEvent={handleEvent}
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
          onEvent={handleEvent}
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
          onEvent={handleEvent}
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
      return (
        <Table
          columns={cols}
          dataSource={data}
          rowKey={(_, idx) => idx!}
          pagination={false}
          onRow={(record) => ({
            onClick: () => {
              if (!onEvent) return
              const dimensionField = config.dimensions[0]
              if (!dimensionField) return
              onEvent({
                type: 'element:click',
                dimensionField,
                dimensionValue: record[dimensionField] as string | number,
                metrics: record,
              })
            },
          })}
        />
      )
    }
    default:
      return (
        <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>
          不支持的图表类型
        </div>
      )
  }
}
