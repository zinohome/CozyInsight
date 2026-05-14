import request from './request'
import type { Dashboard, DashboardChart, CreateDashboardRequest } from '@/types/dashboard'

export const dashboardAPI = {
  list: () => request.get<Dashboard[]>('/dashboard'),
  create: (data: CreateDashboardRequest) => request.post<Dashboard>('/dashboard', data),
  get: (id: number) => request.get<Dashboard>(`/dashboard/${id}`),
  update: (id: number, data: Partial<CreateDashboardRequest>) => request.put(`/dashboard/${id}`, data),
  remove: (id: number) => request.delete(`/dashboard/${id}`),
  addChart: (dashboardId: number, data: { chartId: number; positionX: number; positionY: number; width: number; height: number; config?: string }) => request.post(`/dashboard/${dashboardId}/charts`, data),
  getCharts: (dashboardId: number) => request.get<DashboardChart[]>(`/dashboard/${dashboardId}/charts`),
  removeChart: (dashboardId: number, chartId: number) => request.delete(`/dashboard/${dashboardId}/charts/${chartId}`),
}
