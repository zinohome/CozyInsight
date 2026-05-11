import request from './request'
import type { Chart, CreateChartRequest, ChartDataResponse } from '@/types/chart'

export const chartAPI = {
  list: () => request.get<Chart[]>('/api/v1/chart'),
  create: (data: CreateChartRequest) => request.post<Chart>('/api/v1/chart', data),
  get: (id: number) => request.get<Chart>(`/api/v1/chart/${id}`),
  update: (id: number, data: Partial<CreateChartRequest>) => request.put(`/api/v1/chart/${id}`, data),
  remove: (id: number) => request.delete(`/api/v1/chart/${id}`),
  getData: (id: number, body?: { runtimeFilters?: import('@/types/chart').ChartFilter[]; drillDimension?: string }) =>
    request.post<ChartDataResponse>(`/api/v1/chart/${id}/data`, body || {}),
}
