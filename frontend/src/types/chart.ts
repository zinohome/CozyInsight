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
