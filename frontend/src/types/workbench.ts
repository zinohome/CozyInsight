export interface WorkbenchStats {
  datasourceCount: number
  datasetCount: number
  chartCount: number
  dashboardCount: number
  screenCount: number
}

export interface RecentViewItem {
  id: number
  title: string
  type: 'dashboard' | 'screen'
  visitedAt: string
}

export interface FavoriteItem {
  id: number
  title: string
  type: string
  createdAt: string
}

export interface RecordVisitRequest {
  resourceType: 'dashboard' | 'screen'
  resourceId: number
}
