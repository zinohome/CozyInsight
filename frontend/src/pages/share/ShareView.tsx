import { useEffect, useState, useCallback, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { shareAPI } from '@/api/share'
import { chartAPI } from '@/api/chart'
import ChartRenderer from '@/components/ChartRenderer'
import type { Dashboard, ScreenConfig, ScreenItem } from '@/types/dashboard'
import type { Chart, ChartDataResponse } from '@/types/chart'

interface ScreenChartItem extends ScreenItem {
  chart?: Chart
  data?: ChartDataResponse
}

export default function ShareView() {
  const { token } = useParams<{ token: string }>()
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)
  const [items, setItems] = useState<ScreenChartItem[]>([])
  const [canvas, setCanvas] = useState<{ width: number; height: number; bgColor: string; bgImage?: string }>({ width: 1920, height: 1080, bgColor: '#fff' })
  const [scale, setScale] = useState(1)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const containerRef = useRef<HTMLDivElement>(null)

  const fetchData = useCallback(async () => {
    if (!token) return
    try {
      const res = await shareAPI.get(token)
      if (res.code === 200 && res.data) {
        const d = res.data as Dashboard
        setDashboard(d)

        if (d.type === 'screen' && d.config && d.config !== '{}') {
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
      } else {
        setError(res.error || '分享链接无效或已过期')
      }
    } catch {
      setError('加载失败')
    } finally {
      setLoading(false)
    }
  }, [token])

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

  if (loading) return <div style={{ padding: 40, textAlign: 'center' }}>加载中...</div>
  if (error) return <div style={{ padding: 40, textAlign: 'center', color: '#999' }}>{error}</div>
  if (!dashboard) return null

  if (dashboard.type === 'screen') {
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
            <div key={item.instanceId} style={{ position: 'absolute', left: item.x, top: item.y, width: item.width, height: item.height, zIndex: item.zIndex }}>
              {item.data ? (
                <ChartRenderer type={item.chart?.type || 'bar'} data={item.data.data} config={{ dimensions: item.data.dimensions, metrics: item.data.metrics }} height={item.height} />
              ) : (
                <div style={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>数据加载失败</div>
              )}
            </div>
          ))}
        </div>
      </div>
    )
  }

  // Grid dashboard share view
  return (
    <div style={{ padding: 24, maxWidth: 1200, margin: '0 auto' }}>
      <h2>{dashboard.title}</h2>
      <p style={{ color: '#666' }}>通过分享链接查看</p>
    </div>
  )
}
