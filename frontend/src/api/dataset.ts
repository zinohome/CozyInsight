import request from './request'
import type { Dataset, CreateDatasetRequest, PreviewDataResponse } from '@/types/dataset'

export const datasetAPI = {
  list: () => request.get<Dataset[]>('/dataset'),
  create: (data: CreateDatasetRequest) => request.post<Dataset>('/dataset', data),
  get: (id: number) => request.get<Dataset>(`/dataset/${id}`),
  update: (id: number, data: Partial<CreateDatasetRequest>) => request.put(`/dataset/${id}`, data),
  remove: (id: number) => request.delete(`/dataset/${id}`),
  syncFields: (id: number) => request.post(`/dataset/${id}/sync-fields`),
  preview: (id: number, limit?: number) => request.get<PreviewDataResponse>(`/dataset/${id}/preview`, { params: { limit } }),
}
