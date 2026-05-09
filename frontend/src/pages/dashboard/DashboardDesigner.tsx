import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, message, Modal, Select, Space } from 'antd'
import { Responsive, WidthProvider } from 'react-grid-layout'
import { dashboardAPI } from '@/api/dashboard'
import { chartAPI } from '@/api/chart'
import { exportAPI } from '@/api/export'
import { shareAPI } from '@/api/share'
import ChartRenderer from '@/components/ChartRenderer'
import type { Dashboard } from '@/types/dashboard'
import type { Chart, ChartDataResponse } from '@/types/chart'
import 'react-grid-layout/css/styles.css'
import 'react-resizable/css/styles.css'

const ResponsiveGridLayout = WidthProvider(Responsive)

interface LayoutItem {
  i: string
  x: number
  y: number
  w: number
  h: number
}

interface DashboardChartItem {
  instanceId: string
  chartId: number
  layout: LayoutItem
  data?: ChartDataResponse
  chart?: Chart
}

export default function DashboardDesigner() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)
  const [items, setItems] = useState<DashboardChartItem[]>([])
  const [charts, setCharts] = useState<Chart[]>([])
  const [addModalOpen, setAddModalOpen] = useState(false)
  const [selectedChartId, setSelectedChartId] = useState<number | null>(null)

  const fetchDashboard = useCallback(async () => {
    if (!id) return
    try {
      const d = await dashboardAPI.get(Number(id))
      if (d.type === 'screen') {
        navigate(`/screen/designer/${id}`)
        return
      }
      setDashboard(d)
      const allCharts = await chartAPI.list()
      setCharts(allCharts)
      if (d.config) {
        const cfg = JSON.parse(d.config)
        if (cfg.items && Array.isArray(cfg.items)) {
          // Restore chart info and fetch data for each item
          const restoredItems = await Promise.all(
            cfg.items.map(async (item: DashboardChartItem) => {
              const chart = allCharts.find(c => c.id === item.chartId)
              let data: ChartDataResponse | undefined
              try {
                data = await chartAPI.getData(item.chartId)
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
    }
  }, [id])

  useEffect(() => {
    fetchDashboard()
  }, [fetchDashboard])

  const handleLayoutChange = (layout: LayoutItem[]) => {
    setItems(prev => prev.map(item => {
      const l = layout.find(x => x.i === item.instanceId)
      if (l) {
        return { ...item, layout: { i: l.i, x: l.x, y: l.y, w: l.w, h: l.h } }
      }
      return item
    }))
  }

  const handleAddChart = async () => {
    if (!selectedChartId) return
    const chart = charts.find(c => c.id === selectedChartId)
    if (!chart) return

    let data: ChartDataResponse | undefined
    try {
      data = await chartAPI.getData(selectedChartId)
    } catch {
      message.warning('图表数据加载失败，将只显示占位')
    }

    const instanceId = `${selectedChartId}-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    const newItem: DashboardChartItem = {
      instanceId,
      chartId: selectedChartId,
      layout: { i: instanceId, x: 0, y: 0, w: 6, h: 8 },
      chart,
      data,
    }
    setItems(prev => [...prev, newItem])
    setAddModalOpen(false)
    setSelectedChartId(null)
  }

  const handleRemoveChart = (instanceId: string) => {
    setItems(prev => prev.filter(i => i.instanceId !== instanceId))
  }

  const handleSave = async () => {
    if (!dashboard) return
    const config = JSON.stringify({
      items: items.map(({ instanceId, chartId, layout }) => ({ instanceId, chartId, layout }))
    })
    try {
      await dashboardAPI.update(dashboard.id, { config })
      message.success('保存成功')
    } catch {
      message.error('保存失败')
    }
  }

  const handleShare = async () => {
    if (!dashboard) return
    try {
      const res = await shareAPI.create(dashboard.id)
      if (res.code === 200) {
        const shareUrl = `${window.location.origin}/share/${res.data}`
        await navigator.clipboard.writeText(shareUrl)
        message.success(`分享链接已复制: ${shareUrl}`)
      } else {
        message.error(res.error || '分享失败')
      }
    } catch {
      message.error('分享失败')
    }
  }

  return (
    <div style={{ height: 'calc(100vh - 64px)', display: 'flex', flexDirection: 'column' }}>
      <div style={{ padding: '8px 16px', borderBottom: '1px solid #f0f0f0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h3 style={{ margin: 0 }}>{dashboard?.title || '仪表板设计器'}</h3>
        <Space>
          <Button onClick={() => setAddModalOpen(true)}>添加图表</Button>
          {dashboard?.type === 'screen' && <Button onClick={() => navigate(`/screen/view/${id}`)}>预览</Button>}
          <Button onClick={handleShare}>分享</Button>
          <Button type="primary" onClick={handleSave}>保存</Button>
          <Button onClick={() => navigate('/dashboard')}>返回</Button>
        </Space>
      </div>
      <div style={{ flex: 1, padding: 16, overflow: 'auto', background: '#f5f5f5' }}>
        <ResponsiveGridLayout
          className="layout"
          layouts={{ lg: items.map(i => i.layout) }}
          breakpoints={{ lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 }}
          cols={{ lg: 12, md: 10, sm: 6, xs: 4, xxs: 2 }}
          rowHeight={30}
          onLayoutChange={handleLayoutChange}
          isDraggable
          isResizable
        >
          {items.map(item => (
            <div key={item.instanceId} style={{ background: '#fff', borderRadius: 4, boxShadow: '0 1px 4px rgba(0,0,0,0.1)' }}>
              <div style={{ padding: '8px 12px', borderBottom: '1px solid #f0f0f0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ fontWeight: 'bold' }}>{item.chart?.title || '图表'}</span>
                <Space>
                  <Button type="text" size="small" onClick={() => exportAPI.downloadCSV(item.chartId)}>导出 CSV</Button>
                  <Button type="text" size="small" danger onClick={() => handleRemoveChart(item.instanceId)}>移除</Button>
                </Space>
              </div>
              <div style={{ padding: 8, height: 'calc(100% - 40px)' }}>
                {item.data ? (
                  <ChartRenderer
                    type={item.chart?.type || 'bar'}
                    data={item.data.data}
                    config={{ dimensions: item.data.dimensions, metrics: item.data.metrics }}
                    height={item.layout.h * 30 - 60}
                  />
                ) : (
                  <div style={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>
                    数据加载失败
                  </div>
                )}
              </div>
            </div>
          ))}
        </ResponsiveGridLayout>
      </div>

      <Modal title="添加图表" open={addModalOpen} onOk={handleAddChart} onCancel={() => setAddModalOpen(false)}>
        <Select
          style={{ width: '100%' }}
          placeholder="选择图表"
          options={charts.map(c => ({ value: c.id, label: c.title }))}
          onChange={v => setSelectedChartId(v)}
        />
      </Modal>
    </div>
  )
}
