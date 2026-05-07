export interface Dataset {
  id: number
  name: string
  datasourceId: number
  databaseName: string
  tableName: string
  type: string
  mode: number
  status: number
  createdBy: number
  createdAt: string
}

export interface DatasetField {
  id: number
  datasetId: number
  name: string
  type: string
  deType: number
  length: number
  precision: number
  scale: number
  originName: string
}

export interface CreateDatasetRequest {
  name: string
  datasourceId: number
  databaseName: string
  tableName: string
  type: string
  mode: number
}

export interface PreviewDataResponse {
  fields: Array<{
    id: number
    name: string
    type: string
    deType: number
    length: number
    precision: number
    scale: number
    originName: string
  }>
  data: Array<Record<string, unknown>>
}
