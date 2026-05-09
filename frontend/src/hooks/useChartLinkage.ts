import { useState, useCallback, useMemo } from 'react'
import type { ChartFilter } from '@/types/chart'

export interface RuntimeFilterState {
  [sourceChartId: string]: ChartFilter[]
}

export interface DrillState {
  level: number
  dimension?: string
}

export interface UseChartLinkageReturn {
  runtimeFilters: RuntimeFilterState
  drillStates: Record<string, DrillState>
  applyLinkage: (sourceChartId: string, dimensionField: string, dimensionValue: string | number) => void
  clearLinkage: (sourceChartId?: string) => void
  applyDrill: (chartId: string, dimensionChain: string[], level: number) => void
  resetDrill: (chartId?: string) => void
  getEffectiveFilters: (targetChartId: string) => ChartFilter[]
  getDrillState: (chartId: string) => DrillState
}

export function useChartLinkage(): UseChartLinkageReturn {
  const [runtimeFilters, setRuntimeFilters] = useState<RuntimeFilterState>({})
  const [drillStates, setDrillStates] = useState<Record<string, DrillState>>({})

  const applyLinkage = useCallback((sourceChartId: string, dimensionField: string, dimensionValue: string | number) => {
    setRuntimeFilters(prev => {
      const existing = prev[sourceChartId] || []
      const others = existing.filter(f => f.field !== dimensionField)
      return {
        ...prev,
        [sourceChartId]: [...others, { field: dimensionField, operator: '=', value: String(dimensionValue) }],
      }
    })
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
    Object.entries(runtimeFilters).forEach(([sourceChartId, sourceFilters]) => {
      if (sourceChartId !== targetChartId) {
        filters.push(...sourceFilters)
      }
    })
    return filters
  }, [runtimeFilters])

  const applyDrill = useCallback((chartId: string, dimensionChain: string[], level: number) => {
    if (level < 0 || level >= dimensionChain.length) {
      console.warn(`Drill level ${level} out of bounds for chain [${dimensionChain.join(', ')}]`)
      return
    }
    setDrillStates(prev => ({
      ...prev,
      [chartId]: { level, dimension: dimensionChain[level] },
    }))
  }, [])

  const resetDrill = useCallback((chartId?: string) => {
    if (chartId) {
      setDrillStates(prev => {
        const next = { ...prev }
        delete next[chartId]
        return next
      })
    } else {
      setDrillStates({})
    }
  }, [])

  const getDrillState = useCallback((chartId: string) => {
    return drillStates[chartId] || { level: 0, dimension: undefined }
  }, [drillStates])

  return useMemo(() => ({
    runtimeFilters,
    drillStates,
    applyLinkage,
    clearLinkage,
    applyDrill,
    resetDrill,
    getEffectiveFilters,
    getDrillState,
  }), [runtimeFilters, drillStates, applyLinkage, clearLinkage, applyDrill, resetDrill, getEffectiveFilters, getDrillState])
}
