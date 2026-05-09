import { renderHook, act } from '@testing-library/react'
import { vi } from 'vitest'
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

  it('should merge filters for same source chart by field', () => {
    const { result } = renderHook(() => useChartLinkage())
    act(() => {
      result.current.applyLinkage('chart-1', 'province', '广东省')
      result.current.applyLinkage('chart-1', 'city', '深圳市')
    })
    const filters = result.current.getEffectiveFilters('chart-2')
    expect(filters).toHaveLength(2)
    expect(filters).toContainEqual({ field: 'province', operator: '=', value: '广东省' })
    expect(filters).toContainEqual({ field: 'city', operator: '=', value: '深圳市' })
  })

  it('should update filter for same source chart and same field', () => {
    const { result } = renderHook(() => useChartLinkage())
    act(() => {
      result.current.applyLinkage('chart-1', 'province', '广东省')
      result.current.applyLinkage('chart-1', 'province', '湖南省')
    })
    const filters = result.current.getEffectiveFilters('chart-2')
    expect(filters).toHaveLength(1)
    expect(filters[0]).toEqual({ field: 'province', operator: '=', value: '湖南省' })
  })

  it('should not include target chart own filters in getEffectiveFilters', () => {
    const { result } = renderHook(() => useChartLinkage())
    act(() => {
      result.current.applyLinkage('chart-1', 'province', '广东省')
      result.current.applyLinkage('chart-2', 'city', '深圳市')
    })
    const filtersForChart1 = result.current.getEffectiveFilters('chart-1')
    expect(filtersForChart1).toHaveLength(1)
    expect(filtersForChart1[0].field).toBe('city')
  })

  it('should apply drill-down dimension', () => {
    const { result } = renderHook(() => useChartLinkage())
    act(() => {
      result.current.applyDrill('chart-1', ['province', 'city', 'district'], 1)
    })
    const drillState = result.current.getDrillState('chart-1')
    expect(drillState.level).toBe(1)
    expect(drillState.dimension).toBe('city')
  })

  it('should reset drill-down for all charts', () => {
    const { result } = renderHook(() => useChartLinkage())
    act(() => {
      result.current.applyDrill('chart-1', ['province', 'city', 'district'], 1)
      result.current.resetDrill()
    })
    const drillState = result.current.getDrillState('chart-1')
    expect(drillState.level).toBe(0)
    expect(drillState.dimension).toBeUndefined()
  })

  it('should reset drill-down for specific chart', () => {
    const { result } = renderHook(() => useChartLinkage())
    act(() => {
      result.current.applyDrill('chart-1', ['province', 'city', 'district'], 1)
      result.current.applyDrill('chart-2', ['province', 'city', 'district'], 2)
      result.current.resetDrill('chart-1')
    })
    const drillState1 = result.current.getDrillState('chart-1')
    expect(drillState1.level).toBe(0)
    expect(drillState1.dimension).toBeUndefined()

    const drillState2 = result.current.getDrillState('chart-2')
    expect(drillState2.level).toBe(2)
    expect(drillState2.dimension).toBe('district')
  })

  it('should not crash when applyDrill with out-of-bounds level', () => {
    const { result } = renderHook(() => useChartLinkage())
    const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})
    act(() => {
      result.current.applyDrill('chart-1', ['province', 'city'], 5)
    })
    const drillState = result.current.getDrillState('chart-1')
    expect(drillState.level).toBe(0)
    expect(drillState.dimension).toBeUndefined()
    expect(consoleSpy).toHaveBeenCalledWith('Drill level 5 out of bounds for chain [province, city]')
    consoleSpy.mockRestore()
  })

  it('should keep per-chart drill state independent', () => {
    const { result } = renderHook(() => useChartLinkage())
    act(() => {
      result.current.applyDrill('chart-1', ['province', 'city', 'district'], 1)
      result.current.applyDrill('chart-2', ['province', 'city', 'district'], 2)
    })
    const drillState1 = result.current.getDrillState('chart-1')
    expect(drillState1.level).toBe(1)
    expect(drillState1.dimension).toBe('city')

    const drillState2 = result.current.getDrillState('chart-2')
    expect(drillState2.level).toBe(2)
    expect(drillState2.dimension).toBe('district')
  })
})
