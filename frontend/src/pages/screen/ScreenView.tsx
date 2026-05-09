import { useEffect, useState, useCallback, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { Button } from 'antd'
import { dashboardAPI } from '@/api/dashboard'
import { chartAPI } from '@/api/chart'
import ChartRenderer from '@/components/ChartRenderer'
import type { ScreenConfig, ScreenItem } from '@/types/dashboard'
import type { Chart, ChartDataResponse } from '@/types/chart'

interface ScreenChartItem extends ScreenItem {
  chart?: Chart
  data?: ChartDataResponse
}

export default function ScreenView() {
  const { id } = useParams<{ id: string }>()
  const [items, setItems] = useState<ScreenChartItem[]>([])
  const [canvas, setCanvas] = useState<{ width: number; height: number; bgColor: string; bgImage?: string }>({ width: 1920, height: 1080, bgColor: '#0a1f44' })
  const [scale, setScale] = useState(1)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const fetchData = useCallback(async () => {
    if (!id) return
    try {
      setLoading(true)
      setError('')
      const d = await dashboardAPI.get(Number(id))
      if (d.type !== 'screen') {
        setError('该资源不是数据大屏')
        return
      }

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
      setError('加载数据大屏失败')
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    let timeoutId: ReturnType<typeof setTimeout>
    const poll = async () => {
      await fetchData()
      timeoutId = setTimeout(poll, 30000)
    }
    poll()
    return () => clearTimeout(timeoutId)
  }, [fetchData])

  useEffect(() => {
    const handleResize = () => {
      if (!containerRef.current) return
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

  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement)
    }
    document.addEventListener('fullscreenchange', handleFullscreenChange)
    return () => document.removeEventListener('fullscreenchange', handleFullscreenChange)
  }, [])

  const toggleFullscreen = () => {
    const el = containerRef.current
    if (!el) return
    if (!document.fullscreenElement) {
      el.requestFullscreen().catch(() => {})
    } else {
      document.exitFullscreen().catch(() => {})
    }
  }

  if (loading) {
    return (
      <div style={{ width: '100vw', height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#000', color: '#fff' }}>
        加载中...
      </div>
    )
  }

  if (error) {
    return (
      <div style={{ width: '100vw', height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#000', color: '#fff' }}>
        {error}
      </div>
    )
  }

  return (
    <div ref={containerRef} style={{ width: '100vw', height: '100vh', overflow: 'hidden', background: '#000', display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
      <Button
        style={{ position: 'fixed', top: 16, right: 16, zIndex: 9999 }}
        onClick={toggleFullscreen}
      >
        {isFullscreen ? '退出全屏' : '全屏'}
      </Button>
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
