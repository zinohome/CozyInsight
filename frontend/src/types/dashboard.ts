export interface Dashboard {
  id: number
  title: string
  config: string
  status: number
  createdBy: number
  createdAt: string
}

export interface DashboardChart {
  id: number
  dashboardId: number
  chartId: number
  positionX: number
  positionY: number
  width: number
  height: number
  config: string
}

export interface CreateDashboardRequest {
  title: string
  config: string
}
