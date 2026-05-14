import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import LinkagePanel from './LinkagePanel'
import type { ChartLinkageRule } from '@/types/chart'

describe('LinkagePanel', () => {
  const charts = [
    { id: 1, title: 'Chart A' },
    { id: 2, title: 'Chart B' },
  ]

  it('should render empty rules with add button', () => {
    render(
      <LinkagePanel charts={charts} rules={[]} onChange={vi.fn()} />
    )
    expect(screen.getByText('图表联动配置')).toBeInTheDocument()
    expect(screen.getByText('添加联动规则')).toBeInTheDocument()
  })

  it('should add new rule when clicking add button', () => {
    const onChange = vi.fn()
    render(
      <LinkagePanel charts={charts} rules={[]} onChange={onChange} />
    )

    fireEvent.click(screen.getByText('添加联动规则'))
    expect(onChange).toHaveBeenCalledTimes(1)
    expect(onChange.mock.calls[0][0]).toHaveLength(1)
    expect(onChange.mock.calls[0][0][0]).toEqual({
      sourceChartId: 0,
      targetChartId: 0,
      sourceField: '',
      targetField: '',
    })
  })

  it('should render existing rules', () => {
    const rules: ChartLinkageRule[] = [
      { sourceChartId: 1, targetChartId: 2, sourceField: 'province', targetField: 'province' },
    ]
    render(
      <LinkagePanel charts={charts} rules={rules} onChange={vi.fn()} />
    )
    expect(screen.getByText('→')).toBeInTheDocument()
    expect(screen.getAllByRole('combobox').length).toBeGreaterThan(0)
  })

  it('should remove rule when clicking delete', () => {
    const onChange = vi.fn()
    const rules: ChartLinkageRule[] = [
      { sourceChartId: 1, targetChartId: 2, sourceField: 'a', targetField: 'b' },
      { sourceChartId: 2, targetChartId: 1, sourceField: 'c', targetField: 'd' },
    ]
    const { container } = render(
      <LinkagePanel charts={charts} rules={rules} onChange={onChange} />
    )
    const deleteBtn = container.querySelector('.ant-btn-dangerous') || container.querySelector('button')
    expect(deleteBtn).toBeDefined()
    if (deleteBtn) fireEvent.click(deleteBtn)
    expect(onChange).toHaveBeenCalledTimes(1)
    expect(onChange.mock.calls[0][0]).toHaveLength(1)
    expect(onChange.mock.calls[0][0][0]).toEqual(rules[1])
  })

  it('should update rule source chart', async () => {
    const onChange = vi.fn()
    const rules: ChartLinkageRule[] = [
      { sourceChartId: 0, targetChartId: 0, sourceField: '', targetField: '' },
    ]
    const { container } = render(
      <LinkagePanel charts={charts} rules={rules} onChange={onChange} />
    )

    const selects = container.querySelectorAll('.ant-select')
    fireEvent.mouseDown(selects[0])

    await waitFor(() => {
      const option = document.querySelector('.ant-select-dropdown')?.querySelector('[title="Chart A"]')
      if (option) fireEvent.click(option)
    })

    expect(onChange).toHaveBeenCalled()
  })

  it('should update rule target chart', async () => {
    const onChange = vi.fn()
    const rules: ChartLinkageRule[] = [
      { sourceChartId: 0, targetChartId: 0, sourceField: '', targetField: '' },
    ]
    const { container } = render(
      <LinkagePanel charts={charts} rules={rules} onChange={onChange} />
    )

    const selects = container.querySelectorAll('.ant-select')
    // Third select is target chart
    fireEvent.mouseDown(selects[2])

    await waitFor(() => {
      const option = document.querySelector('.ant-select-dropdown')?.querySelector('[title="Chart B"]')
      if (option) fireEvent.click(option)
    })

    expect(onChange).toHaveBeenCalled()
  })

  it('should render arrow between source and target selects', () => {
    const rules: ChartLinkageRule[] = [
      { sourceChartId: 1, targetChartId: 2, sourceField: 'a', targetField: 'b' },
    ]
    render(<LinkagePanel charts={charts} rules={rules} onChange={vi.fn()} />)
    const arrow = screen.getByText('→')
    expect(arrow).toBeInTheDocument()
    expect(arrow.tagName).toBe('SPAN')
  })

  it('should handle empty charts array', () => {
    render(
      <LinkagePanel charts={[]} rules={[]} onChange={vi.fn()} />
    )
    expect(screen.getByText('图表联动配置')).toBeInTheDocument()
  })

  it('should render multiple rules', () => {
    const rules: ChartLinkageRule[] = [
      { sourceChartId: 1, targetChartId: 2, sourceField: 'a', targetField: 'b' },
      { sourceChartId: 2, targetChartId: 1, sourceField: 'c', targetField: 'd' },
    ]
    const { container } = render(
      <LinkagePanel charts={charts} rules={rules} onChange={vi.fn()} />
    )
    expect(container.textContent).toContain('Chart A')
    expect(container.textContent).toContain('Chart B')
  })
})
