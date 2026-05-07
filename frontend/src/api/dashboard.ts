import request from './request'
import type { Dashboard, DashboardChart, CreateDashboardRequest } from '@/types/dashboard'

export const dashboardAPI = {
  list: () => request.get<{ data: Dashboard[] }>('/api/v1/dashboard'),
  create: (data: CreateDashboardRequest) => request.post<{ data: Dashboard }>('/api/v1/dashboard', data),
  get: (id: number) => request.get<{ data: Dashboard }>(`/api/v1/dashboard/${id}`),
  update: (id: number, data: Partial<CreateDashboardRequest>) => request.put(`/api/v1/dashboard/${id}`, data),
  remove: (id: number) => request.delete(`/api/v1/dashboard/${id}`),
  addChart: (dashboardId: number, data: { chartId: number; positionX: number; positionY: number; width: number; height: number; config?: string }) => request.post(`/api/v1/dashboard/${dashboardId}/charts`, data),
  getCharts: (dashboardId: number) => request.get<{ data: DashboardChart[] }>(`/api/v1/dashboard/${dashboardId}/charts`),
  removeChart: (dashboardId: number, chartId: number) => request.delete(`/api/v1/dashboard/${dashboardId}/charts/${chartId}`),
}
