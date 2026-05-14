import request from './request'
import type { Datasource, CreateDatasourceRequest, TestConnectionRequest } from '@/types/datasource'

export const datasourceAPI = {
  list: () => request.get<Datasource[]>('/datasource'),
  create: (data: CreateDatasourceRequest) => request.post<Datasource>('/datasource', data),
  get: (id: number) => request.get<Datasource>(`/datasource/${id}`),
  update: (id: number, data: Partial<CreateDatasourceRequest>) => request.put(`/datasource/${id}`, data),
  remove: (id: number) => request.delete(`/datasource/${id}`),
  testConnection: (data: TestConnectionRequest) => request.post('/datasource/test', data),
  upload: (formData: FormData) => request.post<Datasource>('/datasource/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }),
}
