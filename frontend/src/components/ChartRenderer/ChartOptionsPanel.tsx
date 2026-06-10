import { Card, Switch, Select, Input, InputNumber, Form, ColorPicker, Divider } from 'antd'
import type { ChartStyleOptions } from '../../types/chart'

interface ChartOptionsPanelProps {
  chartType: string
  options: ChartStyleOptions
  onChange: (next: ChartStyleOptions) => void
}

const labelFormats = [
  { value: 'auto', label: '自动' },
  { value: 'integer', label: '整数' },
  { value: 'percent', label: '百分比' },
  { value: 'currency', label: '货币' },
]

const legendPositions = [
  { value: 'top', label: '顶部' },
  { value: 'right', label: '右侧' },
  { value: 'bottom', label: '底部' },
  { value: 'left', label: '左侧' },
]

const colorSchemes = [
  { value: 'blue', label: '蓝色' },
  { value: 'green', label: '绿色' },
  { value: 'red', label: '红色' },
  { value: 'rainbow', label: '彩虹' },
]

/** 哪些图表类型支持通用样式配置 */
const generalChartTypes = new Set([
  'bar', 'stacked-bar', 'horizontal-bar', 'grouped-bar', 'percent-bar', 'waterfall',
  'line', 'stacked-line', 'area', 'stacked-area',
  'pie', 'donut', 'rose',
  'scatter', 'bubble', 'radar', 'funnel',
])

const axisChartTypes = new Set([
  'bar', 'stacked-bar', 'horizontal-bar', 'grouped-bar', 'percent-bar', 'waterfall',
  'line', 'stacked-line', 'area', 'stacked-area',
  'scatter', 'bubble',
])

const pieChartTypes = new Set(['pie', 'donut', 'rose'])

function pick(o: ChartStyleOptions, key: keyof ChartStyleOptions, fallback: unknown) {
  return o[key] !== undefined ? o[key] : fallback
}

export default function ChartOptionsPanel({ chartType, options, onChange }: ChartOptionsPanelProps) {
  const set = (patch: Partial<ChartStyleOptions>) => onChange({ ...options, ...patch })

  const showGeneral = generalChartTypes.has(chartType)
  const showAxis = axisChartTypes.has(chartType)
  const showPie = pieChartTypes.has(chartType)
  const showGauge = chartType === 'gauge'
  const showKPI = chartType === 'kpi'
  const showHeatmap = chartType === 'heatmap'

  return (
    <Card size="small" title="图表样式" style={{ marginTop: 12 }}>
      {showGeneral && (
        <>
          <Form layout="vertical" size="small">
            <Form.Item label="标题">
              <Input
                value={pick(options, 'title', '') as string}
                onChange={e => set({ title: e.target.value })}
                placeholder="可选：图表标题"
              />
            </Form.Item>
            <Form.Item label="显示图例">
              <Switch
                checked={pick(options, 'showLegend', true) as boolean}
                onChange={v => set({ showLegend: v })}
              />
            </Form.Item>
            {Boolean(pick(options, 'showLegend', true)) && (
              <Form.Item label="图例位置">
                <Select
                  value={pick(options, 'legendPosition', 'top')}
                  options={legendPositions}
                  onChange={v => set({ legendPosition: v as ChartStyleOptions['legendPosition'] })}
                />
              </Form.Item>
            )}
            <Form.Item label="显示标签">
              <Switch
                checked={pick(options, 'showLabel', false) as boolean}
                onChange={v => set({ showLabel: v })}
              />
            </Form.Item>
            <Form.Item label="数值格式">
              <Select
                value={pick(options, 'labelFormat', 'auto')}
                options={labelFormats}
                onChange={v => set({ labelFormat: v as ChartStyleOptions['labelFormat'] })}
              />
            </Form.Item>
          </Form>
          <Divider style={{ margin: '8px 0' }} />
        </>
      )}

      {showAxis && (
        <>
          <Form layout="vertical" size="small">
            <Form.Item label="平滑曲线（折线）">
              <Switch
                checked={pick(options, 'smooth', false) as boolean}
                onChange={v => set({ smooth: v })}
                disabled={chartType !== 'line' && chartType !== 'stacked-line'}
              />
            </Form.Item>
            <Form.Item label="柱条圆角">
              <InputNumber
                min={0}
                max={20}
                value={pick(options, 'radius', 0) as number}
                onChange={v => set({ radius: Number(v) || 0 })}
                style={{ width: '100%' }}
                disabled={!chartType.startsWith('bar') && chartType !== 'waterfall'}
              />
            </Form.Item>
          </Form>
          <Divider style={{ margin: '8px 0' }} />
        </>
      )}

      {showPie && (
        <>
          <Form layout="vertical" size="small">
            <Form.Item label="内径比例">
              <InputNumber
                min={0}
                max={0.95}
                step={0.05}
                value={pick(options, 'innerRadius', chartType === 'donut' ? 0.6 : 0) as number}
                onChange={v => set({ innerRadius: Number(v) || 0 })}
                style={{ width: '100%' }}
              />
            </Form.Item>
            {chartType === 'rose' && (
              <Form.Item label="玫瑰图模式">
                <Select
                  value={pick(options, 'roseType', 'area')}
                  options={[
                    { value: 'radius', label: '半径模式' },
                    { value: 'area', label: '面积模式' },
                  ]}
                  onChange={v => set({ roseType: v as ChartStyleOptions['roseType'] })}
                />
              </Form.Item>
            )}
          </Form>
          <Divider style={{ margin: '8px 0' }} />
        </>
      )}

      {showGauge && (
        <Form layout="vertical" size="small">
          <Form.Item label="最小值">
            <InputNumber
              value={pick(options, 'min', 0) as number}
              onChange={v => set({ min: Number(v) || 0 })}
              style={{ width: '100%' }}
            />
          </Form.Item>
          <Form.Item label="最大值">
            <InputNumber
              value={pick(options, 'max', 1) as number}
              onChange={v => set({ max: Number(v) || 1 })}
              style={{ width: '100%' }}
            />
          </Form.Item>
        </Form>
      )}

      {showKPI && (
        <Form layout="vertical" size="small">
          <Form.Item label="前缀">
            <Input
              value={pick(options, 'prefix', '') as string}
              onChange={e => set({ prefix: e.target.value })}
              placeholder="如 ¥、%"
            />
          </Form.Item>
          <Form.Item label="后缀">
            <Input
              value={pick(options, 'suffix', '') as string}
              onChange={e => set({ suffix: e.target.value })}
              placeholder="如 元、件"
            />
          </Form.Item>
          <Form.Item label="阈值颜色（按值从小到大）">
            {(options.thresholds || []).map((t, idx) => (
              <div key={idx} style={{ display: 'flex', gap: 8, marginBottom: 4 }}>
                <InputNumber
                  value={t.value}
                  onChange={v => {
                    const next = [...(options.thresholds || [])]
                    next[idx] = { ...t, value: Number(v) || 0 }
                    set({ thresholds: next })
                  }}
                  style={{ flex: 1 }}
                />
                <ColorPicker
                  value={t.color}
                  onChange={(_, hex) => {
                    const next = [...(options.thresholds || [])]
                    next[idx] = { ...t, color: hex }
                    set({ thresholds: next })
                  }}
                />
                <a onClick={() => {
                  const next = (options.thresholds || []).filter((_, i) => i !== idx)
                  set({ thresholds: next })
                }}>删除</a>
              </div>
            ))}
            <a onClick={() => set({ thresholds: [...(options.thresholds || []), { value: 0, color: '#1677ff' }] })}>
              + 添加阈值
            </a>
          </Form.Item>
        </Form>
      )}

      {showHeatmap && (
        <Form layout="vertical" size="small">
          <Form.Item label="色阶">
            <Select
              value={pick(options, 'colorScheme', 'blue')}
              options={colorSchemes}
              onChange={v => set({ colorScheme: v as ChartStyleOptions['colorScheme'] })}
            />
          </Form.Item>
        </Form>
      )}

      {!showGeneral && !showGauge && !showKPI && !showHeatmap && (
        <div style={{ color: '#999', fontSize: 12 }}>该图表类型暂无可配置项</div>
      )}
    </Card>
  )
}
