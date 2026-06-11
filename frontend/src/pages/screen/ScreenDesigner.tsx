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
import type { Chart, ChartConfig, ChartEvent } from '@/types/chart'
import ChartRenderer from '@/components/ChartRenderer'
import DrillBreadcrumb from '@/components/DrillBreadcrumb'
import { useChartLinkage } from '@/hooks/useChartLinkage'

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
  const chartDataCacheRef = useRef<ChartDataCache>({})
  const chartInfoCacheRef = useRef<ChartInfoCache>({})
  const [cacheVersion, setCacheVersion] = useState(0)
  // eslint-disable-next-line @typescript-eslint/no-unused-expressions
  cacheVersion

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

  const [addModalOpen, setAddModalOpen] = useState(false)
  const [saving, setSaving] = useState(false)

  // Fetch dashboard and load config
  const fetchDashboard = useCallback(async () => {
    if (!id) return
    setLoading(true)
    const numericId = Number(id)
    if (!Number.isFinite(numericId)) {
      setError('无效的 ID')
      setLoading(false)
      return
    }
    try {
      const d = await dashboardAPI.get(numericId)
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

  const refreshAllData = useCallback(async () => {
    let updated = false
    for (const item of items) {
      try {
        const response = await fetchItemData(item.chartId)
        chartDataCacheRef.current = {
          ...chartDataCacheRef.current,
          [item.chartId]: {
            data: response.data,
            dimensions: response.dimensions,
            metrics: response.metrics,
          },
        }
        updated = true
      } catch {
        // per-chart error shown inline
      }
    }
    if (updated) {
      setCacheVersion((v) => v + 1)
    }
  }, [items, fetchItemData])

  const handleChartEvent = useCallback((chartId: number, event: ChartEvent) => {
    const chart = chartInfoCacheRef.current[chartId]
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
  }, [applyLinkage, applyDrill, getDrillState, navigate, refreshAllData])

  // Load chart data and info for items
  useEffect(() => {
    const loadChartData = async () => {
      let updated = false

      for (const item of items) {
        if (!chartDataCacheRef.current[item.chartId]) {
          try {
            const response = await fetchItemData(item.chartId)
            chartDataCacheRef.current = {
              ...chartDataCacheRef.current,
              [item.chartId]: {
                data: response.data,
                dimensions: response.dimensions,
                metrics: response.metrics,
              },
            }
            updated = true
          } catch {
            // per-chart error shown inline
          }
        }

        if (!chartInfoCacheRef.current[item.chartId]) {
          try {
            const chart = await chartAPI.get(item.chartId)
            chartInfoCacheRef.current = {
              ...chartInfoCacheRef.current,
              [item.chartId]: chart,
            }
            updated = true
          } catch {
            // per-chart error shown inline
          }
        }
      }

      if (updated) {
        setCacheVersion((v) => v + 1)
      }
    }

    if (items.length > 0) {
      loadChartData()
    }
  }, [items, fetchItemData])

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
      const instanceId = `item_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`
      setItems((prev) => {
        const maxZ = prev.reduce((max, i) => Math.max(max, i.zIndex), 0)
        const newItem: ScreenItem = {
          instanceId,
          chartId,
          x: 50,
          y: 50,
          width: 400,
          height: 300,
          zIndex: maxZ + 1,
        }
        return [...prev, newItem]
      })
      setSelectedItemId(instanceId)
      setAddModalOpen(false)
    },
    []
  )

  const handleDragStop = useCallback(
    (instanceId: string, _e: unknown, d: { x: number; y: number }) => {
      updateItem(instanceId, { x: d.x, y: d.y })
    },
    [updateItem]
  )

  const handleResizeStop = useCallback(
    (
      instanceId: string,
      _e: unknown,
      _dir: unknown,
      ref: { style: { width: string; height: string } },
      _delta: unknown,
      pos: { x: number; y: number }
    ) => {
      updateItem(instanceId, {
        width: parseInt(ref.style.width),
        height: parseInt(ref.style.height),
        x: pos.x,
        y: pos.y,
      })
    },
    [updateItem]
  )

  const handleSelect = useCallback(
    (instanceId: string, e: React.MouseEvent) => {
      e.stopPropagation()
      setSelectedItemId(instanceId)
    },
    []
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
  const chartColumns = useMemo(
    () => [
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
    ],
    [addChart]
  )

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

  // per-chart errors are handled inline in each Rnd component

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
          <Button onClick={() => navigate(`/screen/view/${id}`)}>预览</Button>
          <Button type="primary" loading={saving} onClick={handleSave}>
            保存
          </Button>
          <Button onClick={() => { clearLinkage(); resetDrill(); refreshAllData(); }}>清除联动</Button>
          <Button onClick={() => navigate('/dashboard')}>返回</Button>
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
            {/* Reading ref.current in render is intentional: cacheVersion state */}
            {/* is bumped to trigger re-render, then we look up the latest cache. */}
            {/* eslint-disable react-hooks/refs */}
            {items.map((item) => {
              const chartInfo = chartInfoCacheRef.current[item.chartId]
              const chartData = chartDataCacheRef.current[item.chartId]
              const isSelected = item.instanceId === selectedItemId

              return (
                <Rnd
                  key={item.instanceId}
                  position={{ x: item.x, y: item.y }}
                  size={{ width: item.width, height: item.height }}
                  onDragStop={(e, d) => handleDragStop(item.instanceId, e, d)}
                  onResizeStop={(e, dir, ref, delta, pos) =>
                    handleResizeStop(item.instanceId, e, dir, ref, delta, pos)
                  }
                  onClick={(e) => handleSelect(item.instanceId, e)}
                  z={item.zIndex}
                  style={{
                    border: isSelected
                      ? '2px solid #1890ff'
                      : '1px dashed rgba(255,255,255,0.3)',
                    boxSizing: 'border-box',
                  }}
                  // Note: bounds disabled because parent has CSS scale() which causes react-rnd coordinate drift
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
                    {(() => {
                      const chartConfig: ChartConfig | undefined = chartInfo?.config ? JSON.parse(chartInfo.config) : undefined
                      const drillConfig = chartConfig?.drillDown
                      if (drillConfig?.enabled && drillConfig.dimensions && drillConfig.dimensions.length > 1 && chartData) {
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
                    {chartInfo && chartData ? (
                      <ChartRenderer
                        type={chartInfo.type}
                        data={chartData.data}
                        config={{
                          dimensions: chartData.dimensions,
                          metrics: chartData.metrics,
                        }}
                        height={item.height}
                        onEvent={(e) => handleChartEvent(item.chartId, e)}
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
            {/* eslint-enable react-hooks/refs */}
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
                  {/* eslint-disable-next-line react-hooks/refs */}
                  {chartInfoCacheRef.current[selectedItem.chartId]?.title || `图表 #${selectedItem.chartId}`}
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
