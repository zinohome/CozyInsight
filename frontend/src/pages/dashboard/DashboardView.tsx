import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, message, Space } from 'antd'
import { dashboardAPI } from '@/api/dashboard'
import { chartAPI } from '@/api/chart'
import ChartRenderer from '@/components/ChartRenderer'
import DrillBreadcrumb from '@/components/DrillBreadcrumb'
import { useChartLinkage } from '@/hooks/useChartLinkage'
import type { Dashboard } from '@/types/dashboard'
import type { Chart, ChartDataResponse, ChartConfig, ChartEvent } from '@/types/chart'

interface DashboardChartItem {
  instanceId: string
  chartId: number
  layout: { x: number; y: number; w: number; h: number }
  chart?: Chart
  data?: ChartDataResponse
}

export default function DashboardView() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)
  const [items, setItems] = useState<DashboardChartItem[]>([])
  const [charts, setCharts] = useState<Chart[]>([])
  const [loading, setLoading] = useState(true)
  const [isFullscreen, setIsFullscreen] = useState(false)

  const {
    applyLinkage,
    clearLinkage,
    getEffectiveFilters,
    applyDrill,
    resetDrill,
    getDrillState,
  } = useChartLinkage()

  const fetchItemData = useCallback(async (chartId: number) => {
    const filters = getEffectiveFilters(String(chartId))
    const drill = getDrillState(String(chartId))
    const body: { runtimeFilters?: import('@/types/chart').ChartFilter[]; drillDimension?: string } = {}
    if (filters.length > 0) body.runtimeFilters = filters
    if (drill.dimension) body.drillDimension = drill.dimension
    return await chartAPI.getData(chartId, body)
  }, [getEffectiveFilters, getDrillState])

  const refreshAllData = useCallback(async () => {
    const allCharts = await chartAPI.list()
    setCharts(allCharts)

    setItems(prev =>
      prev.map(item => {
        const chart = allCharts.find(c => c.id === item.chartId)
        return { ...item, chart }
      })
    )

    await Promise.all(
      items.map(async item => {
        try {
          const data = await fetchItemData(item.chartId)
          setItems(prev =>
            prev.map(i =>
              i.chartId === item.chartId ? { ...i, data } : i
            )
          )
        } catch {
          // ignore per-chart errors
        }
      })
    )
  }, [items, fetchItemData])

  const fetchDashboard = useCallback(async () => {
    if (!id) return
    const numericId = Number(id)
    if (!Number.isFinite(numericId)) {
      message.error('无效的 ID')
      return
    }
    setLoading(true)
    try {
      const d = await dashboardAPI.get(numericId)
      if (d.type === 'screen') {
        navigate(`/screen/view/${id}`)
        return
      }
      setDashboard(d)
      const allCharts = await chartAPI.list()
      setCharts(allCharts)

      if (d.config) {
        const cfg = JSON.parse(d.config)
        if (cfg.items && Array.isArray(cfg.items)) {
          const restoredItems = await Promise.all(
            cfg.items.map(async (item: DashboardChartItem) => {
              const chart = allCharts.find(c => c.id === item.chartId)
              let data: ChartDataResponse | undefined
              try {
                data = await fetchItemData(item.chartId)
              } catch {
                // ignore
              }
              return { ...item, chart, data }
            })
          )
          setItems(restoredItems)
        }
      }
    } catch {
      message.error('加载仪表板失败')
    } finally {
      setLoading(false)
    }
  }, [id, navigate, fetchItemData])

  useEffect(() => {
    fetchDashboard()
  }, [fetchDashboard])

  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement)
    }
    document.addEventListener('fullscreenchange', handleFullscreenChange)
    return () => document.removeEventListener('fullscreenchange', handleFullscreenChange)
  }, [])

  const handleChartEvent = useCallback((chartId: number, event: ChartEvent) => {
    const chart = charts.find(c => c.id === chartId)
    if (!chart) return
    const chartConfig: ChartConfig | undefined = JSON.parse(chart.config)

    if (chartConfig?.jumpConfig?.enabled) {
      const params = new URLSearchParams()
      chartConfig.jumpConfig.paramsMapping?.forEach(mapping => {
        params.append(mapping.targetParam, String(event.metrics?.[mapping.sourceField]))
      })
      if (chartConfig.jumpConfig.targetType === 'url' && chartConfig.jumpConfig.url) {
        window.open(`${chartConfig.jumpConfig.url}?${params.toString()}`, '_blank')
      } else if (chartConfig.jumpConfig.targetType === 'dashboard' && chartConfig.jumpConfig.targetId) {
        navigate(`/dashboard/view/${chartConfig.jumpConfig.targetId}?${params.toString()}`)
      } else if (chartConfig.jumpConfig.targetType === 'screen' && chartConfig.jumpConfig.targetId) {
        navigate(`/screen/view/${chartConfig.jumpConfig.targetId}?${params.toString()}`)
      }
      return
    }

    if (chartConfig?.drillDown?.enabled && chartConfig.drillDown.dimensions && chartConfig.drillDown.dimensions.length > 1) {
      const current = getDrillState(String(chartId))
      const nextLevel = current.level + 1
      if (chartConfig.drillDown.dimensions && nextLevel < chartConfig.drillDown.dimensions.length) {
        applyDrill(String(chartId), chartConfig.drillDown.dimensions, nextLevel)
        refreshAllData()
        return
      }
    }

    applyLinkage(String(chartId), event.dimensionField, event.dimensionValue)
    refreshAllData()
  }, [charts, applyLinkage, applyDrill, getDrillState, navigate, refreshAllData])

  const toggleFullscreen = () => {
    const el = document.documentElement
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
      <div style={{ height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        加载中...
      </div>
    )
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column', background: '#f5f5f5' }}>
      <div style={{ padding: '8px 16px', borderBottom: '1px solid #f0f0f0', display: 'flex', justifyContent: 'space-between', alignItems: 'center', background: '#fff' }}>
        <h3 style={{ margin: 0 }}>{dashboard?.title || '仪表板'}</h3>
        <Space>
          <Button onClick={() => { clearLinkage(); resetDrill(); refreshAllData(); }}>清除联动</Button>
          <Button onClick={toggleFullscreen}>
            {isFullscreen ? '退出全屏' : '全屏'}
          </Button>
          <Button onClick={() => navigate('/dashboard')}>返回</Button>
        </Space>
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
              {(() => {
                const chartConfig: ChartConfig | undefined = item.chart?.config ? JSON.parse(item.chart.config) : undefined
                const drillConfig = chartConfig?.drillDown
                if (drillConfig?.enabled && drillConfig.dimensions && drillConfig.dimensions.length > 1) {
                  const drill = getDrillState(String(item.chartId))
                  return (
                    <DrillBreadcrumb
                      dimensions={drillConfig.dimensions}
                      currentLevel={drill.level}
                      onDrillUp={(level) => {
                        applyDrill(String(item.chartId), drillConfig.dimensions!, level)
                        refreshAllData()
                      }}
                    />
                  )
                }
                return null
              })()}
              {item.data ? (
                <ChartRenderer
                  type={item.chart?.type || 'bar'}
                  data={item.data.data}
                  config={{ dimensions: item.data.dimensions, metrics: item.data.metrics }}
                  height={item.layout.h * 30}
                  onEvent={(e) => handleChartEvent(item.chartId, e)}
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
