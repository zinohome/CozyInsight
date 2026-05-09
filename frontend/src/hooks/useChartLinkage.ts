import { useState, useCallback } from 'react'
import type { ChartFilter } from '@/types/chart'

export interface RuntimeFilterState {
  [sourceChartId: string]: ChartFilter[]
}

export interface UseChartLinkageReturn {
  runtimeFilters: RuntimeFilterState
  drillLevel: number
  drillDimension?: string
  applyLinkage: (sourceChartId: string, dimensionField: string, dimensionValue: string | number) => void
  clearLinkage: (sourceChartId?: string) => void
  applyDrill: (chartId: string, dimensionChain: string[], level: number) => void
  resetDrill: () => void
  getEffectiveFilters: (targetChartId: string) => ChartFilter[]
}

export function useChartLinkage(): UseChartLinkageReturn {
  const [runtimeFilters, setRuntimeFilters] = useState<RuntimeFilterState>({})
  const [drillLevel, setDrillLevel] = useState(0)
  const [drillDimension, setDrillDimension] = useState<string>()

  const applyLinkage = useCallback((sourceChartId: string, dimensionField: string, dimensionValue: string | number) => {
    setRuntimeFilters(prev => ({
      ...prev,
      [sourceChartId]: [{
        field: dimensionField,
        operator: '=',
        value: String(dimensionValue),
      }],
    }))
  }, [])

  const clearLinkage = useCallback((sourceChartId?: string) => {
    if (sourceChartId) {
      setRuntimeFilters(prev => {
        const next = { ...prev }
        delete next[sourceChartId]
        return next
      })
    } else {
      setRuntimeFilters({})
    }
  }, [])

  const getEffectiveFilters = useCallback((targetChartId: string): ChartFilter[] => {
    const filters: ChartFilter[] = []
    // 默认收集所有源图表的过滤器
    // 后续联动规则配置会决定哪些源图表应用到目标图表
    Object.values(runtimeFilters).forEach(sourceFilters => {
      filters.push(...sourceFilters)
    })
    return filters
  }, [runtimeFilters])

  const applyDrill = useCallback((chartId: string, dimensionChain: string[], level: number) => {
    setDrillLevel(level)
    setDrillDimension(dimensionChain[level])
  }, [])

  const resetDrill = useCallback(() => {
    setDrillLevel(0)
    setDrillDimension(undefined)
  }, [])

  return {
    runtimeFilters,
    drillLevel,
    drillDimension,
    applyLinkage,
    clearLinkage,
    applyDrill,
    resetDrill,
    getEffectiveFilters,
  }
}
