import request from './request'
import type { Datasource, CreateDatasourceRequest, TestConnectionRequest } from '@/types/datasource'

export const datasourceAPI = {
  list: () => request.get<Datasource[]>('/api/v1/datasource'),
  create: (data: CreateDatasourceRequest) => request.post<Datasource>('/api/v1/datasource', data),
  get: (id: number) => request.get<Datasource>(`/api/v1/datasource/${id}`),
  update: (id: number, data: Partial<CreateDatasourceRequest>) => request.put(`/api/v1/datasource/${id}`, data),
  remove: (id: number) => request.delete(`/api/v1/datasource/${id}`),
  testConnection: (data: TestConnectionRequest) => request.post('/api/v1/datasource/test', data),
  upload: (formData: FormData) => request.post<Datasource>('/api/v1/datasource/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }),
}
