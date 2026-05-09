import { useState, useEffect, useCallback, useRef } from 'react'
import { chartAPI } from '@/api/chart'
import type { Dashboard, ScreenConfig, ScreenItem } from '@/types/dashboard'
import type { Chart, ChartDataResponse } from '@/types/chart'

interface ScreenChartItem extends ScreenItem {
  chart?: Chart
  data?: ChartDataResponse
}

export interface UseScreenDataResult {
  items: ScreenChartItem[]
  canvas: { width: number; height: number; bgColor: string; bgImage?: string }
  scale: number
  error: string
  loading: boolean
  refetch: (silent?: boolean) => Promise<void>
}

export function useScreenData(
  getDashboard: () => Promise<Dashboard>
): UseScreenDataResult {
  const [items, setItems] = useState<ScreenChartItem[]>([])
  const [canvas, setCanvas] = useState({ width: 1920, height: 1080, bgColor: '#0a1f44' })
  const [scale, setScale] = useState(1)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  // Keep latest fetcher reference without recreating refetch on every render
  const getDashboardRef = useRef(getDashboard)
  getDashboardRef.current = getDashboard
  const chartListRef = useRef<Chart[]>([])

  // eslint-disable-next-line react-hooks/exhaustive-deps
  const refetch = useCallback(async (silent = false) => {
    try {
      if (!silent) setLoading(true)
      if (!silent) setError('')
      const d = await getDashboardRef.current()
      if (d.type !== 'screen') {
        if (!silent) setError('该资源不是数据大屏')
        return
      }

      let newItems: ScreenChartItem[] = []
      if (d.config && d.config !== '{}') {
        const cfg: ScreenConfig = JSON.parse(d.config)
        if (cfg.canvas) setCanvas(cfg.canvas)

        if (cfg.items && Array.isArray(cfg.items) && cfg.items.length > 0) {
          if (chartListRef.current.length === 0) {
            try {
              chartListRef.current = await chartAPI.list()
            } catch {
              /* ignore */
            }
          }
          const charts = chartListRef.current
          newItems = await Promise.all(
            cfg.items.map(async (item: ScreenItem) => {
              const chart = charts.find((c) => c.id === item.chartId)
              let data: ChartDataResponse | undefined
              try {
                data = await chartAPI.getData(item.chartId)
              } catch {
                /* ignore per-chart errors */
              }
              return { ...item, chart, data }
            })
          )
        }
      }
      setItems(newItems)
    } catch {
      if (!silent) setError('加载数据大屏失败')
    } finally {
      if (!silent) setLoading(false)
    }
  }, [])

  useEffect(() => {
    let timeoutId: ReturnType<typeof setTimeout>
    const poll = async () => {
      await refetch(true)
      timeoutId = setTimeout(poll, 30000)
    }
    poll()
    return () => clearTimeout(timeoutId)
  }, [refetch])

  useEffect(() => {
    const handleResize = () => {
      const newScale = Math.min(
        window.innerWidth / canvas.width,
        window.innerHeight / canvas.height
      )
      setScale(newScale)
    }
    handleResize()
    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [canvas.width, canvas.height])

  return { items, canvas, scale, error, loading, refetch }
}
