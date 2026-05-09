import { useEffect, useState, useCallback, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { dashboardAPI } from '@/api/dashboard'
import { chartAPI } from '@/api/chart'
import ChartRenderer from '@/components/ChartRenderer'
import type { Dashboard, ScreenConfig, ScreenItem } from '@/types/dashboard'
import type { Chart, ChartDataResponse } from '@/types/chart'

interface ScreenChartItem extends ScreenItem {
  chart?: Chart
  data?: ChartDataResponse
}

export default function ScreenView() {
  const { id } = useParams<{ id: string }>()
  const [, setDashboard] = useState<Dashboard | null>(null)
  const [items, setItems] = useState<ScreenChartItem[]>([])
  const [canvas, setCanvas] = useState<{ width: number; height: number; bgColor: string; bgImage?: string }>({ width: 1920, height: 1080, bgColor: '#0a1f44' })
  const containerRef = useRef<HTMLDivElement>(null)

  const fetchData = useCallback(async () => {
    if (!id) return
    try {
      const d = await dashboardAPI.get(Number(id))
      if (d.type !== 'screen') return
      setDashboard(d)

      if (d.config && d.config !== '{}') {
        const cfg: ScreenConfig = JSON.parse(d.config)
        if (cfg.canvas) setCanvas(cfg.canvas)

        if (cfg.items && Array.isArray(cfg.items)) {
          const charts = await chartAPI.list()
          const restored = await Promise.all(
            cfg.items.map(async (item: ScreenItem) => {
              const chart = charts.find(c => c.id === item.chartId)
              let data: ChartDataResponse | undefined
              try { data = await chartAPI.getData(item.chartId) } catch { /* ignore */ }
              return { ...item, chart, data }
            })
          )
          setItems(restored)
        }
      }
    } catch {
      /* ignore */
    }
  }, [id])

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 30000)
    return () => clearInterval(interval)
  }, [fetchData])

  // Request fullscreen on mount
  useEffect(() => {
    const el = containerRef.current
    if (!el) return
    const requestFullscreen = () => {
      if (el.requestFullscreen) el.requestFullscreen()
      else if ((el as any).webkitRequestFullscreen) (el as any).webkitRequestFullscreen()
    }
    requestFullscreen()
  }, [])

  const scale = containerRef.current
    ? Math.min(window.innerWidth / canvas.width, window.innerHeight / canvas.height)
    : 1

  return (
    <div ref={containerRef} style={{ width: '100vw', height: '100vh', overflow: 'hidden', background: '#000', display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
      <div
        style={{
          width: canvas.width,
          height: canvas.height,
          backgroundColor: canvas.bgColor,
          backgroundImage: canvas.bgImage ? `url(${canvas.bgImage})` : undefined,
          backgroundSize: 'cover',
          position: 'relative',
          transform: `scale(${scale})`,
          transformOrigin: 'top left',
        }}
      >
        {items.map(item => (
          <div
            key={item.instanceId}
            style={{
              position: 'absolute',
              left: item.x,
              top: item.y,
              width: item.width,
              height: item.height,
              zIndex: item.zIndex,
            }}
          >
            {item.data ? (
              <ChartRenderer
                type={item.chart?.type || 'bar'}
                data={item.data.data}
                config={{ dimensions: item.data.dimensions, metrics: item.data.metrics }}
                height={item.height}
              />
            ) : (
              <div style={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>
                数据加载失败
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
