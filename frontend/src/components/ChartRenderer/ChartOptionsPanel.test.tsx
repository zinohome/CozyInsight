import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import ChartOptionsPanel from './ChartOptionsPanel'
import type { ChartStyleOptions } from '../../types/chart'

describe('ChartOptionsPanel', () => {
  const baseOptions: ChartStyleOptions = { showLegend: true, showLabel: false }

  it('renders general options for bar chart', () => {
    const onChange = vi.fn()
    render(
      <ChartOptionsPanel
        chartType="bar"
        options={baseOptions}
        onChange={onChange}
      />
    )
    expect(screen.getByText('图表样式')).toBeInTheDocument()
    expect(screen.getByText('标题')).toBeInTheDocument()
    expect(screen.getByText('显示图例')).toBeInTheDocument()
  })

  it('renders pie-specific options for donut', () => {
    const onChange = vi.fn()
    render(
      <ChartOptionsPanel
        chartType="donut"
        options={{}}
        onChange={onChange}
      />
    )
    expect(screen.getByText('内径比例')).toBeInTheDocument()
  })

  it('renders rose type selector for rose chart', () => {
    const onChange = vi.fn()
    render(
      <ChartOptionsPanel
        chartType="rose"
        options={{}}
        onChange={onChange}
      />
    )
    expect(screen.getByText('玫瑰图模式')).toBeInTheDocument()
  })

  it('renders gauge min/max', () => {
    const onChange = vi.fn()
    render(
      <ChartOptionsPanel
        chartType="gauge"
        options={{}}
        onChange={onChange}
      />
    )
    expect(screen.getByText('最小值')).toBeInTheDocument()
    expect(screen.getByText('最大值')).toBeInTheDocument()
  })

  it('renders KPI prefix/suffix/thresholds', () => {
    const onChange = vi.fn()
    render(
      <ChartOptionsPanel
        chartType="kpi"
        options={{}}
        onChange={onChange}
      />
    )
    expect(screen.getByText('前缀')).toBeInTheDocument()
    expect(screen.getByText('后缀')).toBeInTheDocument()
    expect(screen.getByText('阈值颜色（按值从小到大）')).toBeInTheDocument()
  })

  it('renders heatmap color scheme', () => {
    const onChange = vi.fn()
    render(
      <ChartOptionsPanel
        chartType="heatmap"
        options={{}}
        onChange={onChange}
      />
    )
    expect(screen.getByText('色阶')).toBeInTheDocument()
  })

  it('shows fallback message for unsupported chart type', () => {
    const onChange = vi.fn()
    render(
      <ChartOptionsPanel
        chartType="sankey"
        options={{}}
        onChange={onChange}
      />
    )
    expect(screen.getByText('该图表类型暂无可配置项')).toBeInTheDocument()
  })

  it('calls onChange when title is typed', () => {
    const onChange = vi.fn()
    render(
      <ChartOptionsPanel
        chartType="bar"
        options={{}}
        onChange={onChange}
      />
    )
    const input = screen.getByPlaceholderText('可选：图表标题')
    fireEvent.change(input, { target: { value: '我的图表' } })
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ title: '我的图表' }))
  })

  it('toggles showLegend', () => {
    const onChange = vi.fn()
    render(
      <ChartOptionsPanel
        chartType="bar"
        options={{ showLegend: true }}
        onChange={onChange}
      />
    )
    // Ant Design Switch 渲染为 button role with aria-checked
    const switchEl = screen.getAllByRole('switch')[0]
    fireEvent.click(switchEl)
    expect(onChange).toHaveBeenCalled()
  })

  it('adds KPI threshold', () => {
    const onChange = vi.fn()
    render(
      <ChartOptionsPanel
        chartType="kpi"
        options={{}}
        onChange={onChange}
      />
    )
    const addBtn = screen.getByText('+ 添加阈值')
    fireEvent.click(addBtn)
    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({
        thresholds: [{ value: 0, color: '#1677ff' }],
      })
    )
  })
})
