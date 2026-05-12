import request from './request'
import type { WorkbenchStats, RecentViewItem, FavoriteItem, RecordVisitRequest } from '@/types/workbench'

export const workbenchAPI = {
  getStats: () => request.get<WorkbenchStats>('/workbench/stats'),
  getRecent: () => request.get<RecentViewItem[]>('/workbench/recent'),
  recordVisit: (data: RecordVisitRequest) => request.post<void>('/workbench/recent', data),
  getFavorites: () => request.get<FavoriteItem[]>('/workbench/favorites'),
  addFavorite: (type: string, resourceId: number) =>
    request.post<void>('/workbench/favorites', { resourceType: type, resourceId }),
  removeFavorite: (type: string, resourceId: number) =>
    request.delete<void>(`/workbench/favorites/${type}/${resourceId}`),
}
