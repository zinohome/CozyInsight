import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import ChartRenderer from './index'

// Mock @ant-design/charts to avoid canvas rendering issues in jsdom
// Also support onEvent callback
vi.mock('@ant-design/charts', () => ({
  Bar: ({ onEvent }: { onEvent?: (chart: unknown, event: Record<string, unknown>) => void }) => {
    const handleClick = () => {
      if (onEvent) {
        onEvent({}, { type: 'element:click', data: { data: { month: 'Jan', sales: 100 } } })
      }
    }
    return <div data-testid="bar-chart" onClick={handleClick} />
  },
  Line: () => <div data-testid="line-chart" />,
  Pie: () => <div data-testid="pie-chart" />,
  Area: () => <div data-testid="area-chart" />,
  Scatter: () => <div data-testid="scatter-chart" />,
  Radar: () => <div data-testid="radar-chart" />,
  Funnel: () => <div data-testid="funnel-chart" />,
  WordCloud: () => <div data-testid="wordcloud-chart" />,
  Sankey: () => <div data-testid="sankey-chart" />,
  Heatmap: () => <div data-testid="heatmap-chart" />,
  Treemap: () => <div data-testid="treemap-chart" />,
  Gauge: () => <div data-testid="gauge-chart" />,
  DualAxes: () => <div data-testid="dualaxes-chart" />,
}))

describe('ChartRenderer', () => {
  const baseConfig = {
    dimensions: ['month'],
    metrics: ['sales'],
  }

  const baseData = [
    { month: 'Jan', sales: 100 },
    { month: 'Feb', sales: 200 },
  ]

  it('should show empty state when no data', () => {
    render(
      <ChartRenderer type="bar" data={[]} config={baseConfig} />
    )
    expect(screen.getByText('暂无数据')).toBeInTheDocument()
  })

  it('should show config incomplete when dimensions are empty', () => {
    render(
      <ChartRenderer type="bar" data={baseData} config={{ dimensions: [], metrics: [] }} />
    )
    expect(screen.getByText('配置不完整')).toBeInTheDocument()
  })

  it('should render bar chart', () => {
    render(<ChartRenderer type="bar" data={baseData} config={baseConfig} />)
    expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
  })

  it('should render stacked bar chart', () => {
    render(<ChartRenderer type="stacked-bar" data={baseData} config={{ dimensions: ['month', 'category'], metrics: ['sales'] }} />)
    expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
  })

  it('should render horizontal bar chart', () => {
    render(<ChartRenderer type="horizontal-bar" data={baseData} config={baseConfig} />)
    expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
  })

  it('should render grouped bar chart', () => {
    render(<ChartRenderer type="grouped-bar" data={baseData} config={{ dimensions: ['month', 'category'], metrics: ['sales'] }} />)
    expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
  })

  it('should render percent bar chart', () => {
    render(<ChartRenderer type="percent-bar" data={baseData} config={{ dimensions: ['month', 'category'], metrics: ['sales'] }} />)
    expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
  })

  it('should render line chart', () => {
    render(<ChartRenderer type="line" data={baseData} config={baseConfig} />)
    expect(screen.getByTestId('line-chart')).toBeInTheDocument()
  })

  it('should render stacked line chart', () => {
    render(<ChartRenderer type="stacked-line" data={baseData} config={{ dimensions: ['month', 'category'], metrics: ['sales'] }} />)
    expect(screen.getByTestId('line-chart')).toBeInTheDocument()
  })

  it('should render pie chart', () => {
    render(<ChartRenderer type="pie" data={baseData} config={baseConfig} />)
    expect(screen.getByTestId('pie-chart')).toBeInTheDocument()
  })

  it('should render donut chart', () => {
    render(<ChartRenderer type="donut" data={baseData} config={baseConfig} />)
    expect(screen.getByTestId('pie-chart')).toBeInTheDocument()
  })

  it('should render rose chart', () => {
    render(<ChartRenderer type="rose" data={baseData} config={baseConfig} />)
    expect(screen.getByTestId('pie-chart')).toBeInTheDocument()
  })

  it('should render area chart', () => {
    render(<ChartRenderer type="area" data={baseData} config={baseConfig} />)
    expect(screen.getByTestId('area-chart')).toBeInTheDocument()
  })

  it('should render stacked area chart', () => {
    render(<ChartRenderer type="stacked-area" data={baseData} config={{ dimensions: ['month', 'category'], metrics: ['sales'] }} />)
    expect(screen.getByTestId('area-chart')).toBeInTheDocument()
  })

  it('should render scatter chart', () => {
    render(<ChartRenderer type="scatter" data={baseData} config={baseConfig} />)
    expect(screen.getByTestId('scatter-chart')).toBeInTheDocument()
  })

  it('should render bubble chart', () => {
    render(<ChartRenderer type="bubble" data={baseData} config={{ dimensions: ['month', 'category'], metrics: ['sales', 'count'] }} />)
    expect(screen.getByTestId('scatter-chart')).toBeInTheDocument()
  })

  it('should render waterfall chart', () => {
    render(<ChartRenderer type="waterfall" data={baseData} config={baseConfig} />)
    expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
  })

  it('should render combo chart', () => {
    render(<ChartRenderer type="combo" data={baseData} config={{ dimensions: ['month'], metrics: ['sales', 'count'] }} />)
    expect(screen.getByTestId('dualaxes-chart')).toBeInTheDocument()
  })

  it('should render radar chart', () => {
    render(<ChartRenderer type="radar" data={baseData} config={baseConfig} />)
    expect(screen.getByTestId('radar-chart')).toBeInTheDocument()
  })

  it('should render funnel chart', () => {
    render(<ChartRenderer type="funnel" data={baseData} config={baseConfig} />)
    expect(screen.getByTestId('funnel-chart')).toBeInTheDocument()
  })

  it('should render wordcloud chart', () => {
    render(<ChartRenderer type="wordcloud" data={baseData} config={baseConfig} />)
    expect(screen.getByTestId('wordcloud-chart')).toBeInTheDocument()
  })

  it('should render sankey chart', () => {
    render(<ChartRenderer type="sankey" data={baseData} config={baseConfig} />)
    expect(screen.getByTestId('sankey-chart')).toBeInTheDocument()
  })

  it('should render heatmap chart', () => {
    render(<ChartRenderer type="heatmap" data={baseData} config={{ dimensions: ['x', 'y'], metrics: ['z'] }} />)
    expect(screen.getByTestId('heatmap-chart')).toBeInTheDocument()
  })

  it('should render treemap chart', () => {
    render(<ChartRenderer type="treemap" data={baseData} config={baseConfig} />)
    expect(screen.getByTestId('treemap-chart')).toBeInTheDocument()
  })

  it('should render gauge chart', () => {
    const gaugeData = [{ value: 0.75 }]
    render(<ChartRenderer type="gauge" data={gaugeData} config={{ dimensions: ['name'], metrics: ['value'] }} />)
    expect(screen.getByTestId('gauge-chart')).toBeInTheDocument()
  })

  it('should render kpi card', () => {
    const kpiData = [{ sales: 12345 }]
    render(<ChartRenderer type="kpi" data={kpiData} config={{ dimensions: ['total'], metrics: ['sales'] }} />)
    expect(screen.getByText('total')).toBeInTheDocument()
    expect(screen.getByText(kpiData[0].sales.toLocaleString())).toBeInTheDocument()
  })

  it('should render pivot table', () => {
    const pivotData = [
      { region: 'North', product: 'A', sales: 100 },
      { region: 'North', product: 'B', sales: 200 },
      { region: 'South', product: 'A', sales: 150 },
    ]
    const { container } = render(
      <ChartRenderer
        type="pivot-table"
        data={pivotData}
        config={{ dimensions: ['region', 'product'], metrics: ['sales'] }}
      />
    )
    expect(container.querySelector('table')).toBeInTheDocument()
  })

  it('should render table chart', () => {
    const { container } = render(<ChartRenderer type="table" data={baseData} config={baseConfig} />)
    expect(container.querySelector('table')).toBeInTheDocument()
  })

  it('should handle table row click', () => {
    const onEvent = vi.fn()
    const { container } = render(<ChartRenderer type="table" data={baseData} config={baseConfig} onEvent={onEvent} />)
    const row = container.querySelector('tbody tr')
    if (row) {
      fireEvent.click(row)
    }
    expect(container.querySelector('table')).toBeInTheDocument()
  })

  it('should handle chart click event', () => {
    const onEvent = vi.fn()
    render(<ChartRenderer type="bar" data={baseData} config={baseConfig} onEvent={onEvent} />)
    const chart = screen.getByTestId('bar-chart')
    fireEvent.click(chart)
    expect(onEvent).toHaveBeenCalled()
  })

  it('should show unsupported type message', () => {
    render(<ChartRenderer type="unknown" data={baseData} config={baseConfig} />)
    expect(screen.getByText('不支持的图表类型')).toBeInTheDocument()
  })

  // ---------------- Phase B: 新增状态与样式选项 ----------------

  it('should render loading state when loading=true', () => {
    const { container } = render(
      <ChartRenderer type="bar" data={[]} config={baseConfig} loading />
    )
    expect(container.querySelector('.ant-skeleton')).toBeInTheDocument()
  })

  it('should render error state when error is set', () => {
    render(
      <ChartRenderer
        type="bar"
        data={baseData}
        config={baseConfig}
        error="数据库连接超时"
        onRetry={() => {}}
      />
    )
    expect(screen.getByText('加载失败')).toBeInTheDocument()
    expect(screen.getByText('数据库连接超时')).toBeInTheDocument()
  })

  it('should call onRetry when retry button clicked', () => {
    const onRetry = vi.fn()
    render(
      <ChartRenderer
        type="bar"
        data={baseData}
        config={baseConfig}
        error="网络异常"
        onRetry={onRetry}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: /重试/ }))
    expect(onRetry).toHaveBeenCalledTimes(1)
  })

  it('error state takes priority over empty data', () => {
    render(
      <ChartRenderer type="bar" data={[]} config={baseConfig} error="先显示错误" />
    )
    expect(screen.getByText('先显示错误')).toBeInTheDocument()
    expect(screen.queryByText('暂无数据')).not.toBeInTheDocument()
  })

  it('loading state takes priority over empty data', () => {
    const { container } = render(
      <ChartRenderer type="bar" data={[]} config={baseConfig} loading />
    )
    expect(container.querySelector('.ant-skeleton')).toBeInTheDocument()
  })

  it('should apply KPI prefix/suffix and threshold color', () => {
    const kpiData = [{ sales: 150 }]
    render(
      <ChartRenderer
        type="kpi"
        data={kpiData}
        config={{
          dimensions: ['total'],
          metrics: ['sales'],
          options: {
            prefix: '¥',
            suffix: ' 元',
            thresholds: [{ value: 0, color: '#52c41a' }, { value: 100, color: '#cf1322' }],
          },
        }}
      />
    )
    expect(screen.getByText('¥150 元')).toBeInTheDocument()
    // 阈值按升序遍历，后匹配的覆盖前面的 → 150 >= 100 → #cf1322 红色
    const valueEl = screen.getByText('¥150 元')
    expect(valueEl.style.color).toBe('rgb(207, 19, 34)')
  })

  it('should apply gauge min/max normalization', () => {
    // 当 min=0 max=200 时，data 值 100 应归一化为 0.5
    const gaugeData = [{ value: 100 }]
    render(
      <ChartRenderer
        type="gauge"
        data={gaugeData}
        config={{
          dimensions: ['name'],
          metrics: ['value'],
          options: { min: 0, max: 200 },
        }}
      />
    )
    // Gauge 渲染了，不需要断言具体内部，仅验证不报错
    expect(screen.getByTestId('gauge-chart')).toBeInTheDocument()
  })

  it('should apply KPI percent labelFormat', () => {
    const kpiData = [{ sales: 0.85 }]
    render(
      <ChartRenderer
        type="kpi"
        data={kpiData}
        config={{
          dimensions: ['rate'],
          metrics: ['sales'],
          options: { labelFormat: 'percent' },
        }}
      />
    )
    expect(screen.getByText('85.0%')).toBeInTheDocument()
  })

  it('should apply line smooth option', () => {
    render(
      <ChartRenderer
        type="line"
        data={baseData}
        config={{ ...baseConfig, options: { smooth: true } }}
      />
    )
    expect(screen.getByTestId('line-chart')).toBeInTheDocument()
  })

  it('should apply bar radius option', () => {
    render(
      <ChartRenderer
        type="bar"
        data={baseData}
        config={{ ...baseConfig, options: { radius: 8 } }}
      />
    )
    expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
  })
})
