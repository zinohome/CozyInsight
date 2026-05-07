import request from './request'
import type { Dataset, CreateDatasetRequest, PreviewDataResponse } from '@/types/dataset'

export const datasetAPI = {
  list: () => request.get<{ data: Dataset[] }>('/api/v1/dataset'),
  create: (data: CreateDatasetRequest) => request.post<{ data: Dataset }>('/api/v1/dataset', data),
  get: (id: number) => request.get<{ data: Dataset }>(`/api/v1/dataset/${id}`),
  update: (id: number, data: Partial<CreateDatasetRequest>) => request.put(`/api/v1/dataset/${id}`, data),
  remove: (id: number) => request.delete(`/api/v1/dataset/${id}`),
  syncFields: (id: number) => request.post(`/api/v1/dataset/${id}/sync-fields`),
  preview: (id: number, limit?: number) => request.get<{ data: PreviewDataResponse }>(`/api/v1/dataset/${id}/preview`, { params: { limit } }),
}
