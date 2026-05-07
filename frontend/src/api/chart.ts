import request from './request'
import type { Chart, CreateChartRequest } from '@/types/chart'

export const chartAPI = {
  list: () => request.get<{ data: Chart[] }>('/api/v1/chart'),
  create: (data: CreateChartRequest) => request.post<{ data: Chart }>('/api/v1/chart', data),
  get: (id: number) => request.get<{ data: Chart }>(`/api/v1/chart/${id}`),
  update: (id: number, data: Partial<CreateChartRequest>) => request.put(`/api/v1/chart/${id}`, data),
  remove: (id: number) => request.delete(`/api/v1/chart/${id}`),
}
