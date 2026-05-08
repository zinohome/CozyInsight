import { useState, useEffect, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { Button, Select, Card, Space, Tag, InputNumber, Radio, Input, message } from 'antd'
import { chartAPI } from '@/api/chart'
import { datasetAPI } from '@/api/dataset'
import ChartRenderer from '@/components/ChartRenderer'
import type { Chart, ChartConfig, ChartFilter, ChartDataResponse } from '@/types/chart'
import type { PreviewDataResponse } from '@/types/dataset'

export default function ChartBuilder() {
  const { id } = useParams<{ id: string }>()
  const [chart, setChart] = useState<Chart | null>(null)
  const [fields, setFields] = useState<PreviewDataResponse['fields']>([])
  const [config, setConfig] = useState<ChartConfig>({ dimensions: [], metrics: [], filters: [], orders: [], limit: 100 })
  const [chartType, setChartType] = useState<string>('bar')
  const [previewData, setPreviewData] = useState<ChartDataResponse | null>(null)
  const [loading, setLoading] = useState(false)

  const fetchChart = useCallback(async () => {
    if (!id) return
    try {
      const c = await chartAPI.get(Number(id))
      setChart(c)
      setChartType(c.type)
      if (c.config) {
        const parsed: ChartConfig = JSON.parse(c.config)
        setConfig({
          dimensions: parsed.dimensions || [],
          metrics: parsed.metrics || [],
          filters: parsed.filters || [],
          orders: parsed.orders || [],
          limit: parsed.limit || 100,
        })
      }
      // Fetch dataset fields for field panel
      const preview = await datasetAPI.preview(c.datasetId, 0)
      setFields(preview.fields)
    } catch {
      message.error('加载图表失败')
    }
  }, [id])

  useEffect(() => {
    fetchChart()
  }, [fetchChart])

  const handlePreview = async () => {
    if (!chart) return
    setLoading(true)
    try {
      // Save current config to chart first
      await chartAPI.update(chart.id, {
        config: JSON.stringify(config),
        type: chartType,
      })
      const data = await chartAPI.getData(chart.id)
      setPreviewData(data)
    } catch (e) {
      message.error('预览失败: ' + (e instanceof Error ? e.message : '未知错误'))
    } finally {
      setLoading(false)
    }
  }

  const toggleDimension = (field: string) => {
    setConfig(prev => {
      const exists = prev.dimensions.find(d => d.field === field)
      if (exists) {
        return { ...prev, dimensions: prev.dimensions.filter(d => d.field !== field) }
      }
      return { ...prev, dimensions: [...prev.dimensions, { field }] }
    })
  }

  const toggleMetric = (field: string, aggregation: string) => {
    setConfig(prev => {
      const exists = prev.metrics.find(m => m.field === field)
      if (exists) {
        return { ...prev, metrics: prev.metrics.filter(m => m.field !== field) }
      }
      return { ...prev, metrics: [...prev.metrics, { field, aggregation }] }
    })
  }

  const addFilter = (field: string) => {
    setConfig(prev => ({
      ...prev,
      filters: [...prev.filters, { field, operator: '=', value: '' }],
    }))
  }

  const updateFilter = (idx: number, patch: Partial<ChartFilter>) => {
    setConfig(prev => {
      const filters = [...prev.filters]
      filters[idx] = { ...filters[idx], ...patch }
      return { ...prev, filters }
    })
  }

  const removeFilter = (idx: number) => {
    setConfig(prev => ({
      ...prev,
      filters: prev.filters.filter((_, i) => i !== idx),
    }))
  }

  const textFields = fields.filter(f => f.deType === 0)
  const timeFields = fields.filter(f => f.deType === 1)
  const numFields = fields.filter(f => f.deType === 2)

  return (
    <div style={{ display: 'flex', height: 'calc(100vh - 64px)' }}>
      {/* Left: Field Panel */}
      <div style={{ width: 280, borderRight: '1px solid #f0f0f0', padding: 16, overflow: 'auto', background: '#fafafa' }}>
        <h4>字段列表</h4>
        <div style={{ marginBottom: 12 }}>
          <div style={{ fontWeight: 'bold', marginBottom: 4 }}>文本</div>
          {textFields.map(f => (
            <Tag key={f.id} style={{ margin: 2, cursor: 'pointer' }}
              color={config.dimensions.find(d => d.field === f.name) ? 'blue' : 'default'}
              onClick={() => toggleDimension(f.name)}>
              {f.name}
            </Tag>
          ))}
        </div>
        <div style={{ marginBottom: 12 }}>
          <div style={{ fontWeight: 'bold', marginBottom: 4 }}>时间</div>
          {timeFields.map(f => (
            <Tag key={f.id} style={{ margin: 2, cursor: 'pointer' }}
              color={config.dimensions.find(d => d.field === f.name) ? 'blue' : 'default'}
              onClick={() => toggleDimension(f.name)}>
              {f.name}
            </Tag>
          ))}
        </div>
        <div>
          <div style={{ fontWeight: 'bold', marginBottom: 4 }}>数值</div>
          {numFields.map(f => (
            <div key={f.id} style={{ marginBottom: 4 }}>
              <Tag style={{ cursor: 'pointer' }}
                color={config.metrics.find(m => m.field === f.name) ? 'green' : 'default'}
                onClick={() => toggleMetric(f.name, 'SUM')}>
                {f.name}
              </Tag>
              {config.metrics.find(m => m.field === f.name) && (
                <Select size="small" style={{ width: 80 }}
                  value={config.metrics.find(m => m.field === f.name)?.aggregation || 'SUM'}
                  onChange={v => {
                    setConfig(prev => ({
                      ...prev,
                      metrics: prev.metrics.map(m => m.field === f.name ? { ...m, aggregation: v } : m),
                    }))
                  }}
                  options={['SUM', 'COUNT', 'AVG', 'MAX', 'MIN'].map(a => ({ value: a, label: a }))}
                />
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Middle: Config Panel */}
      <div style={{ flex: 1, padding: 16, overflow: 'auto' }}>
        <Card title="图表配置" style={{ marginBottom: 16 }}>
          <Space direction="vertical" style={{ width: '100%' }}>
            <div>
              <span style={{ marginRight: 8 }}>图表类型:</span>
              <Radio.Group value={chartType} onChange={e => setChartType(e.target.value)}>
                <Radio.Button value="bar">柱状图</Radio.Button>
                <Radio.Button value="line">折线图</Radio.Button>
                <Radio.Button value="area">面积图</Radio.Button>
                <Radio.Button value="pie">饼图</Radio.Button>
                <Radio.Button value="scatter">散点图</Radio.Button>
                <Radio.Button value="radar">雷达图</Radio.Button>
                <Radio.Button value="table">表格</Radio.Button>
              </Radio.Group>
            </div>
            <div>
              <span style={{ marginRight: 8 }}>维度 ({config.dimensions.length}):</span>
              {config.dimensions.map(d => <Tag key={d.field} color="blue">{d.field}</Tag>)}
            </div>
            <div>
              <span style={{ marginRight: 8 }}>指标 ({config.metrics.length}):</span>
              {config.metrics.map(m => <Tag key={m.field} color="green">{m.aggregation}({m.field})</Tag>)}
            </div>
            <div>
              <span style={{ marginRight: 8 }}>数据量限制:</span>
              <InputNumber min={1} max={10000} value={config.limit} onChange={v => setConfig(prev => ({ ...prev, limit: v || 100 }))} />
            </div>
            <div>
              <span style={{ marginRight: 8 }}>过滤条件:</span>
              {config.filters.map((f, idx) => (
                <Space key={idx} style={{ marginBottom: 4, display: 'flex' }}>
                  <span>{f.field}</span>
                  <Select value={f.operator} style={{ width: 100 }}
                    options={['=', '!=', '>', '<', '>=', '<=', 'LIKE'].map(o => ({ value: o, label: o }))}
                    onChange={v => updateFilter(idx, { operator: v })}
                  />
                  <Input value={f.value} style={{ width: 120 }} onChange={e => updateFilter(idx, { value: e.target.value })} />
                  <Button size="small" danger onClick={() => removeFilter(idx)}>删除</Button>
                </Space>
              ))}
              <Select style={{ width: 120 }} placeholder="添加过滤"
                options={fields.map(f => ({ value: f.name, label: f.name }))}
                onChange={v => addFilter(v)}
              />
            </div>
            <Button type="primary" onClick={handlePreview} loading={loading}>预览</Button>
          </Space>
        </Card>

        {previewData && (
          <Card title="预览">
            <ChartRenderer
              type={chartType}
              data={previewData.data}
              config={{ dimensions: previewData.dimensions, metrics: previewData.metrics }}
              height={400}
            />
          </Card>
        )}
      </div>
    </div>
  )
}
