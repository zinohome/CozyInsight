import { describe, it, expect } from 'vitest'
import type { Chart, ChartConfig, ChartDimension, ChartMetric, ChartFilter, ChartDataResponse } from './chart'

describe('chart types', () => {
  it('should allow valid Chart object', () => {
    const chart: Chart = {
      id: 1,
      title: 'Sales Chart',
      type: 'bar',
      datasetId: 1,
      config: '{}',
      status: 1,
      createdBy: 1,
      createdAt: '2024-01-01',
    }
    expect(chart.title).toBe('Sales Chart')
    expect(chart.type).toBe('bar')
  })

  it('should allow valid ChartConfig', () => {
    const config: ChartConfig = {
      dimensions: [{ field: 'month', sort: 'asc' }],
      metrics: [{ field: 'sales', aggregation: 'SUM', alias: 'Total Sales' }],
      filters: [{ field: 'region', operator: '=', value: 'North' }],
      orders: [{ field: 'sales', direction: 'desc' }],
      limit: 100,
    }
    expect(config.dimensions).toHaveLength(1)
    expect(config.metrics[0].aggregation).toBe('SUM')
  })

  it('should allow empty ChartConfig', () => {
    const config: ChartConfig = {
      dimensions: [],
      metrics: [],
      filters: [],
      orders: [],
    }
    expect(config.dimensions).toEqual([])
  })

  it('should allow ChartDimension with optional sort', () => {
    const dim: ChartDimension = { field: 'category' }
    expect(dim.sort).toBeUndefined()
  })

  it('should allow ChartMetric with optional alias', () => {
    const metric: ChartMetric = { field: 'amount', aggregation: 'COUNT' }
    expect(metric.alias).toBeUndefined()
  })

  it('should allow ChartDataResponse', () => {
    const response: ChartDataResponse = {
      dimensions: ['month'],
      metrics: ['sales'],
      data: [{ month: 'Jan', sales: 100 }],
    }
    expect(response.data).toHaveLength(1)
  })

  it('should allow ChartFilter with all operators', () => {
    const filters: ChartFilter[] = [
      { field: 'a', operator: '=', value: '1' },
      { field: 'a', operator: '!=', value: '1' },
      { field: 'a', operator: '>', value: '1' },
      { field: 'a', operator: '<', value: '1' },
      { field: 'a', operator: '>=', value: '1' },
      { field: 'a', operator: '<=', value: '1' },
      { field: 'a', operator: 'LIKE', value: '%a%' },
      { field: 'a', operator: 'IN', value: '1,2,3' },
    ]
    expect(filters).toHaveLength(8)
  })

  it('should allow ChartDrillConfig', () => {
    const drill = {
      enabled: true,
      dimensions: ['province', 'city', 'district'],
    }
    expect(drill.dimensions).toHaveLength(3)
  })

  it('should allow ChartJumpConfig', () => {
    const jump = {
      enabled: true,
      targetType: 'dashboard' as const,
      targetId: 1,
      url: 'https://example.com',
      paramsMapping: [{ sourceField: 'a', targetParam: 'b' }],
    }
    expect(jump.targetType).toBe('dashboard')
  })

  it('should allow ChartLinkageRule', () => {
    const rule = {
      sourceChartId: 1,
      targetChartId: 2,
      sourceField: 'province',
      targetField: 'region',
    }
    expect(rule.sourceChartId).toBe(1)
    expect(rule.targetChartId).toBe(2)
  })
})
