import request from './request'
import type { Datasource, CreateDatasourceRequest, TestConnectionRequest } from '@/types/datasource'

// 连接测试结果
export interface ConnectionTestResult {
  success: boolean
  message: string
  time?: number // 响应时间(毫秒)
}

export const datasourceAPI = {
  list: () => request.get<Datasource[]>('/datasource'),
  create: (data: CreateDatasourceRequest) => request.post<Datasource>('/datasource', data),
  get: (id: number) => request.get<Datasource>(`/datasource/${id}`),
  update: (id: number, data: Partial<CreateDatasourceRequest>) => request.put(`/datasource/${id}`, data),
  remove: (id: number) => request.delete(`/datasource/${id}`),
  testConnection: (data: TestConnectionRequest) => request.post('/datasource/test', data),
  testConnectionByConfig: (config: string) =>
    request.post<ConnectionTestResult>('/datasource/test-config', { configuration: config }),
  validateDatasource: (id: number) =>
    request.post<{ status: string; message: string }>(`/datasource/${id}/validate`, {}),
  getDatabases: (id: number) => request.get<string[]>(`/datasource/${id}/databases`),
  getTables: (id: number, database: string) =>
    request.get<string[]>(`/datasource/${id}/tables?database=${encodeURIComponent(database)}`),
  upload: (formData: FormData) => request.post<Datasource>('/datasource/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }),
}
