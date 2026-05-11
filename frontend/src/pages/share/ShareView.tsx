import { useState, useCallback, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { shareAPI } from '@/api/share'
import { chartAPI } from '@/api/chart'
import ChartRenderer from '@/components/ChartRenderer'
import { useScreenData } from '@/hooks/useScreenData'
import type { Dashboard } from '@/types/dashboard'
import type { Chart, ChartDataResponse } from '@/types/chart'

export default function ShareView() {
  const { token } = useParams<{ token: string }>()
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)

  const getDashboard = useCallback(async () => {
    const res = await shareAPI.get(token!)
    if (res.code !== 200 || !res.data) {
      throw new Error(res.error || '分享链接无效或已过期')
    }
    const d = res.data as Dashboard
    setDashboard(d)
    return d
  }, [token])

  const { items, canvas, scale, error, loading } = useScreenData(getDashboard)

  if (loading) return <div style={{ padding: 40, textAlign: 'center' }}>加载中...</div>
  if (error) return <div style={{ padding: 40, textAlign: 'center', color: '#999' }}>{error}</div>
  if (!dashboard) return null

  if (dashboard.type === 'screen') {
    return (
      <div style={{ width: '100vw', height: '100vh', overflow: 'hidden', background: '#000', display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
        <div
          style={{
            width: canvas.width,
            height: canvas.height,
            backgroundColor: canvas.bgColor,
            backgroundImage: canvas.bgImage ? `url(${canvas.bgImage})` : undefined,
            backgroundSize: 'cover',
            position: 'relative',
            transform: `scale(${scale})`,
            transformOrigin: 'center center',
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
  return <DashboardShareView dashboard={dashboard} />
}

function DashboardShareView({ dashboard }: { dashboard: Dashboard }) {
  const [items, setItems] = useState<Array<{
    instanceId: string
    chartId: number
    layout: { x: number; y: number; w: number; h: number }
    chart?: Chart
    data?: ChartDataResponse
  }>>([])

  useEffect(() => {
    if (!dashboard.config) return
    const cfg = JSON.parse(dashboard.config)
    if (!cfg.items || !Array.isArray(cfg.items)) return

    let cancelled = false
    const load = async () => {
      const allCharts = await chartAPI.list()
      if (cancelled) return
      const restored = await Promise.all(
        cfg.items.map(async (item: { instanceId: string; chartId: number; layout: { x: number; y: number; w: number; h: number } }) => {
          const chart = allCharts.find((c: Chart) => c.id === item.chartId)
          let data: ChartDataResponse | undefined
          try {
            data = await chartAPI.getData(item.chartId)
          } catch {
            // ignore
          }
          return { ...item, chart, data }
        })
      )
      if (!cancelled) setItems(restored)
    }
    load()
    return () => { cancelled = true }
  }, [dashboard])

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column', background: '#f5f5f5' }}>
      <div style={{ padding: '8px 16px', borderBottom: '1px solid #f0f0f0', background: '#fff' }}>
        <h3 style={{ margin: 0 }}>{dashboard.title || '仪表板'}</h3>
      </div>
      <div style={{ flex: 1, padding: 16, overflow: 'auto' }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(12, 1fr)', gap: 16 }}>
          {items.map(item => (
            <div
              key={item.instanceId}
              style={{
                gridColumn: `span ${item.layout.w}`,
                background: '#fff',
                borderRadius: 4,
                boxShadow: '0 1px 4px rgba(0,0,0,0.1)',
                padding: 12,
              }}
            >
              <div style={{ fontWeight: 'bold', marginBottom: 8 }}>{item.chart?.title || '图表'}</div>
              {item.data ? (
                <ChartRenderer
                  type={item.chart?.type || 'bar'}
                  data={item.data.data}
                  config={{ dimensions: item.data.dimensions, metrics: item.data.metrics }}
                  height={item.layout.h * 30}
                />
              ) : (
                <div style={{ height: 200, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>
                  数据加载失败
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
