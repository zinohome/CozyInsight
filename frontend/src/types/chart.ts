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

export interface ChartConfig {
  dimensions: ChartDimension[]
  metrics: ChartMetric[]
  filters: ChartFilter[]
  orders: ChartOrder[]
  limit?: number
}

export interface ChartDataResponse {
  dimensions: string[]
  metrics: string[]
  data: Array<Record<string, unknown>>
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
  }
  height?: number
  onEvent?: (event: ChartEvent) => void
}
