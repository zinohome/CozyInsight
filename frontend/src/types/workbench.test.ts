import { describe, it, expect } from 'vitest'
import type { WorkbenchStats, RecentViewItem, FavoriteItem, RecordVisitRequest } from './workbench'

describe('workbench types', () => {
  it('should allow valid WorkbenchStats', () => {
    const stats: WorkbenchStats = {
      datasourceCount: 10,
      datasetCount: 20,
      chartCount: 30,
      dashboardCount: 5,
      screenCount: 2,
    }
    expect(stats.chartCount).toBe(30)
    expect(stats.screenCount).toBe(2)
  })

  it('should allow valid RecentViewItem', () => {
    const item: RecentViewItem = {
      id: 1,
      title: 'Sales Dashboard',
      type: 'dashboard',
      visitedAt: '2024-01-01',
    }
    expect(item.type).toBe('dashboard')
  })

  it('should allow RecentViewItem with screen type', () => {
    const item: RecentViewItem = {
      id: 2,
      title: 'Big Screen',
      type: 'screen',
      visitedAt: '2024-01-01',
    }
    expect(item.type).toBe('screen')
  })

  it('should allow valid FavoriteItem', () => {
    const item: FavoriteItem = {
      id: 1,
      title: 'My Chart',
      type: 'chart',
      createdAt: '2024-01-01',
    }
    expect(item.type).toBe('chart')
  })

  it('should allow valid RecordVisitRequest', () => {
    const req: RecordVisitRequest = {
      resourceType: 'dashboard',
      resourceId: 1,
    }
    expect(req.resourceType).toBe('dashboard')
  })
})
