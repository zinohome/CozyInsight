import { useState, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { shareAPI } from '@/api/share'
import ChartRenderer from '@/components/ChartRenderer'
import { useScreenData } from '@/hooks/useScreenData'
import type { Dashboard } from '@/types/dashboard'

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

  // Grid dashboard share view (placeholder — can be enhanced later)
  return (
    <div style={{ padding: 24, maxWidth: 1200, margin: '0 auto' }}>
      <h2>{dashboard.title}</h2>
      <p style={{ color: '#666' }}>通过分享链接查看</p>
    </div>
  )
}
