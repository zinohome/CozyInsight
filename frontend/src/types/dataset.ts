export interface Dataset {
  id: number
  name: string
  datasourceId: number
  databaseName: string
  tableName: string
  sql?: string
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

// 数据预览结果(后端 service.DataPreviewResult)
export interface DataPreviewField {
    name: string;
    originName: string;
    type: string;
    deType: number;
    groupType: string;
    description?: string;
    sample?: string;
}

export interface DataPreviewResult {
    fields: DataPreviewField[];
    data: Record<string, unknown>[];
    total: number;
}
