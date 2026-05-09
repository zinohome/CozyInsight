import { renderHook, act } from '@testing-library/react'
import { useChartLinkage } from './useChartLinkage'

describe('useChartLinkage', () => {
  it('should apply linkage filter', () => {
    const { result } = renderHook(() => useChartLinkage())
    act(() => {
      result.current.applyLinkage('chart-1', 'province', '广东省')
    })
    const filters = result.current.getEffectiveFilters('chart-2')
    expect(filters).toHaveLength(1)
    expect(filters[0]).toEqual({ field: 'province', operator: '=', value: '广东省' })
  })

  it('should clear all linkage filters', () => {
    const { result } = renderHook(() => useChartLinkage())
    act(() => {
      result.current.applyLinkage('chart-1', 'province', '广东省')
      result.current.clearLinkage()
    })
    expect(result.current.getEffectiveFilters('chart-2')).toHaveLength(0)
  })

  it('should clear specific linkage filter', () => {
    const { result } = renderHook(() => useChartLinkage())
    act(() => {
      result.current.applyLinkage('chart-1', 'province', '广东省')
      result.current.applyLinkage('chart-2', 'city', '深圳市')
      result.current.clearLinkage('chart-1')
    })
    expect(result.current.getEffectiveFilters('chart-3')).toHaveLength(1)
    expect(result.current.getEffectiveFilters('chart-3')[0].field).toBe('city')
  })

  it('should apply drill-down dimension', () => {
    const { result } = renderHook(() => useChartLinkage())
    act(() => {
      result.current.applyDrill('chart-1', ['province', 'city', 'district'], 1)
    })
    expect(result.current.drillLevel).toBe(1)
    expect(result.current.drillDimension).toBe('city')
  })

  it('should reset drill-down', () => {
    const { result } = renderHook(() => useChartLinkage())
    act(() => {
      result.current.applyDrill('chart-1', ['province', 'city', 'district'], 1)
      result.current.resetDrill()
    })
    expect(result.current.drillLevel).toBe(0)
    expect(result.current.drillDimension).toBeUndefined()
  })
})
