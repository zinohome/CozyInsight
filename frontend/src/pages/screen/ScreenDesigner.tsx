import { useParams, useNavigate } from 'react-router-dom'
import {
  message,
  Button,
  Modal,
  Table,
  Input,
  InputNumber,
  Form,
  ColorPicker,
  Space,
  Typography,
  Card,
} from 'antd'
import { useState, useEffect, useCallback, useRef, useMemo } from 'react'
import { Rnd } from 'react-rnd'
import { dashboardAPI } from '@/api/dashboard'
import { chartAPI } from '@/api/chart'
import type { Dashboard, ScreenConfig, ScreenItem } from '@/types/dashboard'
import type { Chart } from '@/types/chart'
import ChartRenderer from '@/components/ChartRenderer'

interface ChartDataCache {
  [chartId: number]: {
    data: Array<Record<string, unknown>>
    dimensions: string[]
    metrics: string[]
  }
}

interface ChartInfoCache {
  [chartId: number]: Chart
}

export default function ScreenDesigner() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const canvasContainerRef = useRef<HTMLDivElement>(null)

  const [dashboard, setDashboard] = useState<Dashboard | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  const [canvas, setCanvas] = useState<ScreenConfig['canvas']>({
    width: 1920,
    height: 1080,
    bgColor: '#0d1b2a',
  })

  const [items, setItems] = useState<ScreenItem[]>([])
  const [selectedItemId, setSelectedItemId] = useState<string | null>(null)
  const [scale, setScale] = useState(1)

  const [chartList, setChartList] = useState<Chart[]>([])
  const [chartDataCache, setChartDataCache] = useState<ChartDataCache>({})
  const [chartInfoCache, setChartInfoCache] = useState<ChartInfoCache>({})

  const [addModalOpen, setAddModalOpen] = useState(false)
  const [saving, setSaving] = useState(false)

  // Fetch dashboard and load config
  const fetchDashboard = useCallback(async () => {
    if (!id) return
    setLoading(true)
    try {
      const d = await dashboardAPI.get(Number(id))
      if (d.type !== 'screen') {
        setError('该资源不是数据大屏')
        setLoading(false)
        return
      }
      setDashboard(d)

      // Parse config
      let config: ScreenConfig | null = null
      if (d.config && d.config !== '{}') {
        try {
          config = JSON.parse(d.config) as ScreenConfig
        } catch {
          message.warning('大屏配置解析失败，将使用默认配置')
        }
      }

      if (config && config.mode === 'screen') {
        setCanvas(config.canvas || { width: 1920, height: 1080, bgColor: '#0d1b2a' })
        setItems(config.items || [])
      } else {
        setCanvas({ width: 1920, height: 1080, bgColor: '#0d1b2a' })
        setItems([])
      }
    } catch {
      message.error('加载数据大屏失败')
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    fetchDashboard()
  }, [fetchDashboard])

  // Compute scale when canvas or container changes
  useEffect(() => {
    const computeScale = () => {
      if (!canvasContainerRef.current) return
      const containerWidth = canvasContainerRef.current.clientWidth
      const containerHeight = canvasContainerRef.current.clientHeight
      const newScale = Math.min(
        containerWidth / canvas.width,
        containerHeight / canvas.height,
        1
      )
      setScale(newScale)
    }

    computeScale()
    window.addEventListener('resize', computeScale)
    return () => window.removeEventListener('resize', computeScale)
  }, [canvas.width, canvas.height])

  // Load chart list for add modal
  const fetchChartList = useCallback(async () => {
    try {
      const charts = await chartAPI.list()
      setChartList(charts)
    } catch {
      message.error('加载图表列表失败')
    }
  }, [])

  useEffect(() => {
    fetchChartList()
  }, [fetchChartList])

  // Load chart data and info for items
  useEffect(() => {
    const loadChartData = async () => {
      const newDataCache: ChartDataCache = { ...chartDataCache }
      const newInfoCache: ChartInfoCache = { ...chartInfoCache }
      let updated = false

      for (const item of items) {
        if (!newDataCache[item.chartId]) {
          try {
            const response = await chartAPI.getData(item.chartId)
            newDataCache[item.chartId] = {
              data: response.data,
              dimensions: response.dimensions,
              metrics: response.metrics,
            }
            updated = true
          } catch {
            // Leave placeholder
          }
        }

        if (!newInfoCache[item.chartId]) {
          try {
            const chart = await chartAPI.get(item.chartId)
            newInfoCache[item.chartId] = chart
            updated = true
          } catch {
            // Leave placeholder
          }
        }
      }

      if (updated) {
        setChartDataCache(newDataCache)
        setChartInfoCache(newInfoCache)
      }
    }

    if (items.length > 0) {
      loadChartData()
    }
  }, [items])

  const updateCanvas = useCallback((patch: Partial<ScreenConfig['canvas']>) => {
    setCanvas((prev) => ({ ...prev, ...patch }))
  }, [])

  const updateItem = useCallback((instanceId: string, patch: Partial<ScreenItem>) => {
    setItems((prev) =>
      prev.map((item) => (item.instanceId === instanceId ? { ...item, ...patch } : item))
    )
  }, [])

  const removeItem = useCallback((instanceId: string) => {
    setItems((prev) => prev.filter((item) => item.instanceId !== instanceId))
    setSelectedItemId((prev) => (prev === instanceId ? null : prev))
  }, [])

  const addChart = useCallback(
    (chartId: number) => {
      const newItem: ScreenItem = {
        instanceId: `item_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`,
        chartId,
        x: 50,
        y: 50,
        width: 400,
        height: 300,
        zIndex: items.length + 1,
      }
      setItems((prev) => [...prev, newItem])
      setSelectedItemId(newItem.instanceId)
      setAddModalOpen(false)
    },
    [items.length]
  )

  const handleSave = useCallback(async () => {
    if (!dashboard) return
    setSaving(true)
    try {
      const config: ScreenConfig = {
        mode: 'screen',
        canvas,
        items: items.map(({ instanceId, chartId, x, y, width, height, zIndex }) => ({
          instanceId,
          chartId,
          x,
          y,
          width,
          height,
          zIndex,
        })),
      }
      await dashboardAPI.update(dashboard.id, { config: JSON.stringify(config) })
      message.success('保存成功')
    } catch {
      message.error('保存失败')
    } finally {
      setSaving(false)
    }
  }, [dashboard, canvas, items])

  const selectedItem = useMemo(
    () => items.find((item) => item.instanceId === selectedItemId) || null,
    [items, selectedItemId]
  )

  // Add modal columns
  const chartColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 100,
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: unknown, record: Chart) => (
        <Button type="link" onClick={() => addChart(record.id)}>
          添加
        </Button>
      ),
    },
  ]

  if (loading) {
    return (
      <div style={{ height: 'calc(100vh - 64px)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        加载中...
      </div>
    )
  }

  if (error) {
    return (
      <div style={{ padding: 40, textAlign: 'center', color: '#999' }}>
        {error}
      </div>
    )
  }

  return (
    <div style={{ height: 'calc(100vh - 64px)', display: 'flex', flexDirection: 'column' }}>
      {/* Toolbar */}
      <div
        style={{
          height: 48,
          borderBottom: '1px solid #303030',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0 16px',
          background: '#141414',
        }}
      >
        <Typography.Text strong style={{ color: '#fff' }}>
          {dashboard?.title || '数据大屏设计器'}
        </Typography.Text>
        <Space>
          <Button onClick={() => setAddModalOpen(true)}>添加图表</Button>
          <Button onClick={() => navigate(`/screen/${id}/preview`)}>预览</Button>
          <Button type="primary" loading={saving} onClick={handleSave}>
            保存
          </Button>
          <Button onClick={() => navigate('/screen')}>返回</Button>
        </Space>
      </div>

      {/* Main area */}
      <div style={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
        {/* Canvas container */}
        <div
          ref={canvasContainerRef}
          style={{
            flex: 1,
            background: '#1a1a1a',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            overflow: 'hidden',
            position: 'relative',
          }}
        >
          {/* Scaled canvas */}
          <div
            style={{
              width: canvas.width,
              height: canvas.height,
              transform: `scale(${scale})`,
              transformOrigin: 'center center',
              backgroundColor: canvas.bgColor,
              backgroundImage: canvas.bgImage ? `url(${canvas.bgImage})` : undefined,
              backgroundSize: 'cover',
              backgroundPosition: 'center',
              position: 'relative',
              boxShadow: '0 0 20px rgba(0,0,0,0.5)',
            }}
            onClick={() => setSelectedItemId(null)}
          >
            {items.map((item) => {
              const chartInfo = chartInfoCache[item.chartId]
              const chartData = chartDataCache[item.chartId]
              const isSelected = item.instanceId === selectedItemId

              return (
                <Rnd
                  key={item.instanceId}
                  position={{ x: item.x, y: item.y }}
                  size={{ width: item.width, height: item.height }}
                  onDragStop={(_e: unknown, d: { x: number; y: number }) => {
                    updateItem(item.instanceId, { x: d.x, y: d.y })
                  }}
                  onResizeStop={(_e: unknown, _dir: unknown, ref: { style: { width: string; height: string } }, _delta: unknown, pos: { x: number; y: number }) => {
                    updateItem(item.instanceId, {
                      width: parseInt(ref.style.width),
                      height: parseInt(ref.style.height),
                      x: pos.x,
                      y: pos.y,
                    })
                  }}
                  onClick={(e: React.MouseEvent) => {
                    e.stopPropagation()
                    setSelectedItemId(item.instanceId)
                  }}
                  style={{
                    zIndex: item.zIndex,
                    border: isSelected
                      ? '2px solid #1890ff'
                      : '1px dashed rgba(255,255,255,0.3)',
                    boxSizing: 'border-box',
                  }}
                  bounds="parent"
                >
                  <div
                    style={{
                      width: '100%',
                      height: '100%',
                      background: 'rgba(0,0,0,0.3)',
                      overflow: 'hidden',
                      position: 'relative',
                    }}
                  >
                    {chartInfo && chartData ? (
                      <ChartRenderer
                        type={chartInfo.type}
                        data={chartData.data}
                        config={{
                          dimensions: chartData.dimensions,
                          metrics: chartData.metrics,
                        }}
                        height={item.height}
                      />
                    ) : chartInfo && !chartData ? (
                      <div
                        style={{
                          width: '100%',
                          height: '100%',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          color: '#999',
                          fontSize: 14,
                        }}
                      >
                        数据加载失败
                      </div>
                    ) : (
                      <div
                        style={{
                          width: '100%',
                          height: '100%',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          color: '#999',
                          fontSize: 14,
                        }}
                      >
                        加载中...
                      </div>
                    )}
                  </div>
                </Rnd>
              )
            })}
          </div>
        </div>

        {/* Property panel */}
        <div
          style={{
            width: 280,
            borderLeft: '1px solid #303030',
            background: '#141414',
            overflow: 'auto',
            padding: 16,
          }}
        >
          <Typography.Title level={5} style={{ color: '#fff', marginTop: 0 }}>
            画布设置
          </Typography.Title>
          <Form layout="vertical" style={{ marginBottom: 24 }}>
            <Form.Item label={<span style={{ color: '#ccc' }}>宽度</span>}>
              <InputNumber
                min={800}
                max={3840}
                value={canvas.width}
                onChange={(v) => updateCanvas({ width: v || 1920 })}
                style={{ width: '100%' }}
              />
            </Form.Item>
            <Form.Item label={<span style={{ color: '#ccc' }}>高度</span>}>
              <InputNumber
                min={600}
                max={2160}
                value={canvas.height}
                onChange={(v) => updateCanvas({ height: v || 1080 })}
                style={{ width: '100%' }}
              />
            </Form.Item>
            <Form.Item label={<span style={{ color: '#ccc' }}>背景颜色</span>}>
              <ColorPicker
                value={canvas.bgColor}
                onChange={(color) => updateCanvas({ bgColor: color.toHexString() })}
                showText
              />
            </Form.Item>
            <Form.Item label={<span style={{ color: '#ccc' }}>背景图片 URL</span>}>
              <Input
                value={canvas.bgImage || ''}
                onChange={(e) => updateCanvas({ bgImage: e.target.value })}
                placeholder="输入图片 URL"
              />
            </Form.Item>
          </Form>

          {selectedItem && (
            <>
              <Typography.Title level={5} style={{ color: '#fff', marginTop: 0 }}>
                选中组件
              </Typography.Title>
              <Card
                size="small"
                style={{ background: '#1f1f1f', border: '1px solid #303030', marginBottom: 16 }}
                bodyStyle={{ padding: 12 }}
              >
                <Typography.Text style={{ color: '#ccc' }}>
                  {chartInfoCache[selectedItem.chartId]?.title || `图表 #${selectedItem.chartId}`}
                </Typography.Text>
                <Button
                  danger
                  type="link"
                  size="small"
                  onClick={() => removeItem(selectedItem.instanceId)}
                  style={{ float: 'right' }}
                >
                  删除
                </Button>
              </Card>
              <Form layout="vertical">
                <Form.Item label={<span style={{ color: '#ccc' }}>X</span>}>
                  <InputNumber
                    value={selectedItem.x}
                    onChange={(v) => updateItem(selectedItem.instanceId, { x: v || 0 })}
                    style={{ width: '100%' }}
                  />
                </Form.Item>
                <Form.Item label={<span style={{ color: '#ccc' }}>Y</span>}>
                  <InputNumber
                    value={selectedItem.y}
                    onChange={(v) => updateItem(selectedItem.instanceId, { y: v || 0 })}
                    style={{ width: '100%' }}
                  />
                </Form.Item>
                <Form.Item label={<span style={{ color: '#ccc' }}>宽度</span>}>
                  <InputNumber
                    min={50}
                    value={selectedItem.width}
                    onChange={(v) => updateItem(selectedItem.instanceId, { width: v || 100 })}
                    style={{ width: '100%' }}
                  />
                </Form.Item>
                <Form.Item label={<span style={{ color: '#ccc' }}>高度</span>}>
                  <InputNumber
                    min={50}
                    value={selectedItem.height}
                    onChange={(v) => updateItem(selectedItem.instanceId, { height: v || 100 })}
                    style={{ width: '100%' }}
                  />
                </Form.Item>
                <Form.Item label={<span style={{ color: '#ccc' }}>层级</span>}>
                  <InputNumber
                    value={selectedItem.zIndex}
                    onChange={(v) => updateItem(selectedItem.instanceId, { zIndex: v || 1 })}
                    style={{ width: '100%' }}
                  />
                </Form.Item>
              </Form>
            </>
          )}
        </div>
      </div>

      {/* Add chart modal */}
      <Modal
        title="添加图表"
        open={addModalOpen}
        onCancel={() => setAddModalOpen(false)}
        footer={null}
        width={700}
        bodyStyle={{ maxHeight: 500, overflow: 'auto', padding: '12px 0' }}
      >
        <Table
          dataSource={chartList}
          columns={chartColumns}
          rowKey="id"
          size="small"
          pagination={{ pageSize: 10 }}
        />
      </Modal>
    </div>
  )
}
