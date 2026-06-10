import { useState, useEffect, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { Button, Select, Card, Space, Tag, InputNumber, Radio, Input, message } from 'antd'
import { chartAPI } from '@/api/chart'
import { datasetAPI } from '@/api/dataset'
import ChartRenderer from '@/components/ChartRenderer'
import ChartOptionsPanel from '@/components/ChartRenderer/ChartOptionsPanel'
import type { Chart, ChartConfig, ChartFilter, ChartDataResponse, ChartStyleOptions } from '@/types/chart'
import type { PreviewDataResponse } from '@/types/dataset'

export default function ChartBuilder() {
  const { id } = useParams<{ id: string }>()
  const [chart, setChart] = useState<Chart | null>(null)
  const [fields, setFields] = useState<PreviewDataResponse['fields']>([])
  const [config, setConfig] = useState<ChartConfig>({ dimensions: [], metrics: [], filters: [], orders: [], limit: 100 })
  const [chartType, setChartType] = useState<string>('bar')
  const [previewData, setPreviewData] = useState<ChartDataResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [previewError, setPreviewError] = useState<string | null>(null)
  const [showJumpConfig, setShowJumpConfig] = useState(false)
  const [showDrillConfig, setShowDrillConfig] = useState(false)
  const [showStyleConfig, setShowStyleConfig] = useState(false)

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
    setPreviewError(null)
    try {
      // Save current config to chart first
      await chartAPI.update(chart.id, {
        config: JSON.stringify(config),
        type: chartType,
      })
      const data = await chartAPI.getData(chart.id)
      setPreviewData(data)
    } catch (e) {
      const msg = e instanceof Error ? e.message : '未知错误'
      setPreviewError(msg)
      message.error('预览失败: ' + msg)
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
                <Radio.Button value="table">表格</Radio.Button>
                <Radio.Button value="pivot-table">透视表</Radio.Button>
                <Radio.Button value="kpi">指标卡</Radio.Button>
                <Radio.Button value="bar">柱状图</Radio.Button>
                <Radio.Button value="stacked-bar">堆叠柱状图</Radio.Button>
                <Radio.Button value="horizontal-bar">横向柱状图</Radio.Button>
                <Radio.Button value="grouped-bar">分组柱状图</Radio.Button>
                <Radio.Button value="percent-bar">百分比柱状图</Radio.Button>
                <Radio.Button value="waterfall">瀑布图</Radio.Button>
                <Radio.Button value="line">折线图</Radio.Button>
                <Radio.Button value="stacked-line">堆叠折线图</Radio.Button>
                <Radio.Button value="area">面积图</Radio.Button>
                <Radio.Button value="stacked-area">堆叠面积图</Radio.Button>
                <Radio.Button value="pie">饼图</Radio.Button>
                <Radio.Button value="donut">环形图</Radio.Button>
                <Radio.Button value="rose">玫瑰图</Radio.Button>
                <Radio.Button value="scatter">散点图</Radio.Button>
                <Radio.Button value="bubble">气泡图</Radio.Button>
                <Radio.Button value="radar">雷达图</Radio.Button>
                <Radio.Button value="funnel">漏斗图</Radio.Button>
                <Radio.Button value="gauge">仪表盘</Radio.Button>
                <Radio.Button value="wordcloud">词云</Radio.Button>
                <Radio.Button value="heatmap">热力图</Radio.Button>
                <Radio.Button value="treemap">矩形树图</Radio.Button>
                <Radio.Button value="sankey">桑基图</Radio.Button>
                <Radio.Button value="combo">组合图</Radio.Button>
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
            <Button onClick={() => setShowDrillConfig(v => !v)}>{showDrillConfig ? '隐藏' : '配置'}下钻</Button>
            <Button onClick={() => setShowJumpConfig(v => !v)}>{showJumpConfig ? '隐藏' : '配置'}跳转</Button>
            <Button onClick={() => setShowStyleConfig(v => !v)}>{showStyleConfig ? '隐藏' : '配置'}样式</Button>
          </Space>

          {showDrillConfig && (
            <Card title="下钻配置" size="small" style={{ marginTop: 12 }}>
              <Space direction="vertical" style={{ width: '100%' }}>
                <div>
                  <span>启用下钻: </span>
                  <Radio.Group
                    value={config.drillDown?.enabled ? 'yes' : 'no'}
                    onChange={e => setConfig(prev => ({
                      ...prev,
                      drillDown: { ...prev.drillDown, enabled: e.target.value === 'yes', dimensions: prev.drillDown?.dimensions || [] },
                    }))}
                  >
                    <Radio.Button value="yes">启用</Radio.Button>
                    <Radio.Button value="no">禁用</Radio.Button>
                  </Radio.Group>
                </div>
                {config.drillDown?.enabled && (
                  <div>
                    <span>下钻维度链（逗号分隔）: </span>
                    <Input
                      style={{ width: 300 }}
                      value={config.drillDown?.dimensions?.join(',') || ''}
                      onChange={e => setConfig(prev => ({
                        ...prev,
                        drillDown: {
                          ...prev.drillDown,
                          dimensions: e.target.value.split(',').map(s => s.trim()).filter(Boolean),
                        },
                      }))}
                      placeholder="province,city,district"
                    />
                  </div>
                )}
              </Space>
            </Card>
          )}

          {showJumpConfig && (
            <Card title="跳转配置" size="small" style={{ marginTop: 12 }}>
              <Space direction="vertical" style={{ width: '100%' }}>
                <div>
                  <span>启用跳转: </span>
                  <Radio.Group
                    value={config.jumpConfig?.enabled ? 'yes' : 'no'}
                    onChange={e => setConfig(prev => ({
                      ...prev,
                      jumpConfig: { ...prev.jumpConfig, enabled: e.target.value === 'yes' },
                    }))}
                  >
                    <Radio.Button value="yes">启用</Radio.Button>
                    <Radio.Button value="no">禁用</Radio.Button>
                  </Radio.Group>
                </div>
                {config.jumpConfig?.enabled && (
                  <>
                    <div>
                      <span>跳转目标类型: </span>
                      <Select
                        style={{ width: 120 }}
                        value={config.jumpConfig?.targetType || 'dashboard'}
                        onChange={v => setConfig(prev => ({
                          ...prev,
                          jumpConfig: { ...prev.jumpConfig, targetType: v },
                        }))}
                        options={[
                          { value: 'dashboard', label: '仪表板' },
                          { value: 'screen', label: '大屏' },
                          { value: 'url', label: '外部链接' },
                        ]}
                      />
                    </div>
                    {config.jumpConfig?.targetType === 'url' ? (
                      <div>
                        <span>URL: </span>
                        <Input
                          style={{ width: 300 }}
                          value={config.jumpConfig?.url || ''}
                          onChange={e => setConfig(prev => ({
                            ...prev,
                            jumpConfig: { ...prev.jumpConfig, url: e.target.value },
                          }))}
                          placeholder="https://example.com"
                        />
                      </div>
                    ) : (
                      <div>
                        <span>目标ID: </span>
                        <InputNumber
                          style={{ width: 120 }}
                          value={config.jumpConfig?.targetId || undefined}
                          onChange={v => setConfig(prev => ({
                            ...prev,
                            jumpConfig: { ...prev.jumpConfig, targetId: v || undefined },
                          }))}
                        />
                      </div>
                    )}
                    <div>
                      <span>参数映射: </span>
                      {(config.jumpConfig?.paramsMapping || []).map((mapping, idx) => (
                        <Space key={idx} style={{ marginBottom: 4, display: 'flex' }}>
                          <Input
                            placeholder="源字段"
                            style={{ width: 120 }}
                            value={mapping.sourceField}
                            onChange={e => {
                              const next = [...(config.jumpConfig?.paramsMapping || [])]
                              next[idx] = { ...next[idx], sourceField: e.target.value }
                              setConfig(prev => ({ ...prev, jumpConfig: { ...prev.jumpConfig, paramsMapping: next } }))
                            }}
                          />
                          <span>→</span>
                          <Input
                            placeholder="目标参数"
                            style={{ width: 120 }}
                            value={mapping.targetParam}
                            onChange={e => {
                              const next = [...(config.jumpConfig?.paramsMapping || [])]
                              next[idx] = { ...next[idx], targetParam: e.target.value }
                              setConfig(prev => ({ ...prev, jumpConfig: { ...prev.jumpConfig, paramsMapping: next } }))
                            }}
                          />
                          <Button size="small" danger onClick={() => {
                            const next = (config.jumpConfig?.paramsMapping || []).filter((_, i) => i !== idx)
                            setConfig(prev => ({ ...prev, jumpConfig: { ...prev.jumpConfig, paramsMapping: next } }))
                          }}>删除</Button>
                        </Space>
                      ))}
                      <Button size="small" onClick={() => {
                        const next = [...(config.jumpConfig?.paramsMapping || []), { sourceField: '', targetParam: '' }]
                        setConfig(prev => ({ ...prev, jumpConfig: { ...prev.jumpConfig, paramsMapping: next } }))
                      }}>添加映射</Button>
                    </div>
                  </>
                )}
              </Space>
            </Card>
          )}

          {showStyleConfig && (
            <ChartOptionsPanel
              chartType={chartType}
              options={config.options || {}}
              onChange={(next: ChartStyleOptions) =>
                setConfig(prev => ({ ...prev, options: next }))
              }
            />
          )}
        </Card>

        {previewData && (
          <Card title="预览">
            <ChartRenderer
              type={chartType}
              data={previewData.data}
              config={{
                dimensions: previewData.dimensions,
                metrics: previewData.metrics,
                options: config.options,
              }}
              height={400}
              loading={loading}
              error={previewError}
              onRetry={handlePreview}
            />
          </Card>
        )}
      </div>
    </div>
  )
}
