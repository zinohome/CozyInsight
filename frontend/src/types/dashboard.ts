export interface Dashboard {
  id: number
  title: string
  type: 'dashboard' | 'screen'
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

export interface ScreenConfig {
  mode: 'screen'
  canvas: {
    width: number
    height: number
    bgColor: string
    bgImage?: string
  }
  items: ScreenItem[]
}

export interface ScreenItem {
  instanceId: string
  chartId: number
  x: number
  y: number
  width: number
  height: number
  zIndex: number
}

export interface CreateDashboardRequest {
  title: string
  type?: 'dashboard' | 'screen'
  config: string
}
