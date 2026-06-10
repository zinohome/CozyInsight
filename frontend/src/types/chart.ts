export type ChartType =
  | 'bar'
  | 'stacked-bar'
  | 'horizontal-bar'
  | 'horizontal-stacked-bar'
  | 'grouped-bar'
  | 'percent-bar'
  | 'waterfall'
  | 'line'
  | 'stacked-line'
  | 'area'
  | 'stacked-area'
  | 'pie'
  | 'donut'
  | 'rose'
  | 'scatter'
  | 'bubble'
  | 'radar'
  | 'funnel'
  | 'gauge'
  | 'wordcloud'
  | 'heatmap'
  | 'treemap'
  | 'sankey'
  | 'combo'
  | 'kpi'
  | 'pivot-table'
  | 'table'

export interface Chart {
  id: number
  title: string
  type: string
  datasetId: number
  config: string
  status: number
  createdBy: number
  createdAt: string
}

export interface CreateChartRequest {
  title: string
  type: string
  datasetId: number
  config: string
}

export interface ChartDimension {
  field: string
  sort?: string
}

export interface ChartMetric {
  field: string
  aggregation: string
  alias?: string
}

export interface ChartFilter {
  field: string
  operator: string
  value: string
}

export interface ChartOrder {
  field: string
  direction: string
}

export interface ChartDrillConfig {
  enabled?: boolean
  dimensions?: string[]
}

export interface ChartJumpConfig {
  enabled?: boolean
  targetType?: 'dashboard' | 'screen' | 'url'
  targetId?: number
  url?: string
  paramsMapping?: Array<{ sourceField: string; targetParam: string }>
}

/** 图表样式选项 — 通用配置（标题/图例/标签） */
export interface ChartStyleOptions {
  // 通用
  title?: string
  showLegend?: boolean
  legendPosition?: 'top' | 'right' | 'bottom' | 'left'
  showLabel?: boolean
  labelFormat?: 'auto' | 'integer' | 'percent' | 'currency'
  // 坐标轴类（bar/line/area）
  smooth?: boolean
  isStack?: boolean
  radius?: number
  // 饼图类
  innerRadius?: number
  roseType?: 'radius' | 'area'
  // gauge
  min?: number
  max?: number
  // kpi
  prefix?: string
  suffix?: string
  thresholds?: Array<{ value: number; color: string }>
  // heatmap
  colorScheme?: 'blue' | 'green' | 'red' | 'rainbow'
}

export interface ChartConfig {
  dimensions: ChartDimension[]
  metrics: ChartMetric[]
  filters: ChartFilter[]
  orders: ChartOrder[]
  limit?: number
  drillDown?: ChartDrillConfig
  jumpConfig?: ChartJumpConfig
  options?: ChartStyleOptions
}

export interface ChartDataResponse {
  dimensions: string[]
  metrics: string[]
  data: Array<Record<string, unknown>>
}

export interface ChartLinkageRule {
  sourceChartId: number
  targetChartId: number
  sourceField: string
  targetField: string
}

export interface ChartEvent {
  type: 'element:click'
  dimensionField: string
  dimensionValue: string | number
  metrics?: Record<string, unknown>
}

export interface ChartRendererProps {
  type: string
  data: Array<Record<string, unknown>>
  config: {
    dimensions: string[]
    metrics: string[]
    options?: ChartStyleOptions
  }
  height?: number
  onEvent?: (event: ChartEvent) => void
  loading?: boolean
  error?: string | null
  onRetry?: () => void
}
