import request from './request'
import type { Chart, CreateChartRequest, ChartDataResponse, ChartFilter } from '@/types/chart'

export const chartAPI = {
  list: () => request.get<Chart[]>('/chart'),
  create: (data: CreateChartRequest) => request.post<Chart>('/chart', data),
  get: (id: number) => request.get<Chart>(`/chart/${id}`),
  update: (id: number, data: Partial<CreateChartRequest>) => request.put(`/chart/${id}`, data),
  remove: (id: number) => request.delete(`/chart/${id}`),
  getData: (id: number, body?: { runtimeFilters?: ChartFilter[]; drillDimension?: string }) =>
    request.post<ChartDataResponse>(`/chart/${id}/data`, body || {}),
}
