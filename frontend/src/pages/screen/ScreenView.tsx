import { useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import { Button, message } from 'antd'
import { dashboardAPI } from '@/api/dashboard'
import { workbenchAPI } from '@/api/workbench'
import ChartRenderer from '@/components/ChartRenderer'
import { useScreenData } from '@/hooks/useScreenData'

export default function ScreenView() {
  const { id } = useParams<{ id: string }>()
  const containerRef = useRef<HTMLDivElement>(null)
  const [isFullscreen, setIsFullscreen] = useState(false)

  const numericId = Number(id)
  const isValidId = Number.isFinite(numericId)

  const { items, canvas, scale, error, loading } = useScreenData(
    () => isValidId ? dashboardAPI.get(numericId) : Promise.reject(new Error('无效的 ID'))
  )

  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement)
    }
    document.addEventListener('fullscreenchange', handleFullscreenChange)
    return () => document.removeEventListener('fullscreenchange', handleFullscreenChange)
  }, [])

  useEffect(() => {
    if (isValidId) {
      workbenchAPI.recordVisit({ resourceType: 'screen', resourceId: numericId }).catch(() => {})
    }
  }, [numericId, isValidId])

  const toggleFullscreen = () => {
    const el = containerRef.current
    if (!el) return
    if (!document.fullscreenElement) {
      el.requestFullscreen().catch(() => {
        message.error('全屏模式不可用')
      })
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
          transformOrigin: 'center center',
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
