export interface Datasource {
  id: number
  name: string
  type: string
  config: string
  status: number
  createdBy: number
  createdAt: string
}

export interface CreateDatasourceRequest {
  name: string
  type: string
  config: string
}

export interface TestConnectionRequest {
  type: string
  config: Record<string, unknown>
}
