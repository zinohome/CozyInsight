import { describe, it, expect } from 'vitest'
import type { Dashboard, DashboardChart, ScreenConfig, ScreenItem, CreateDashboardRequest } from './dashboard'

describe('dashboard types', () => {
  it('should allow valid Dashboard', () => {
    const db: Dashboard = {
      id: 1,
      title: 'Sales Dashboard',
      type: 'dashboard',
      config: '{}',
      status: 1,
      createdBy: 1,
      createdAt: '2024-01-01',
    }
    expect(db.type).toBe('dashboard')
  })

  it('should allow screen type Dashboard', () => {
    const screen: Dashboard = {
      id: 2,
      title: 'Big Screen',
      type: 'screen',
      config: JSON.stringify({ canvas: { width: 1920, height: 1080, bgColor: '#000' } }),
      status: 1,
      createdBy: 1,
      createdAt: '2024-01-01',
    }
    expect(screen.type).toBe('screen')
  })

  it('should allow valid DashboardChart', () => {
    const dc: DashboardChart = {
      id: 1,
      dashboardId: 1,
      chartId: 2,
      positionX: 0,
      positionY: 0,
      width: 6,
      height: 4,
      config: '{}',
    }
    expect(dc.positionX).toBe(0)
    expect(dc.width).toBe(6)
  })

  it('should allow valid ScreenConfig', () => {
    const cfg: ScreenConfig = {
      mode: 'screen',
      canvas: {
        width: 1920,
        height: 1080,
        bgColor: '#0a1f44',
        bgImage: 'https://example.com/bg.png',
      },
      items: [
        { instanceId: 'i1', chartId: 1, x: 0, y: 0, width: 300, height: 200, zIndex: 1 },
      ],
    }
    expect(cfg.canvas.bgImage).toBe('https://example.com/bg.png')
  })

  it('should allow valid ScreenItem', () => {
    const item: ScreenItem = {
      instanceId: 'i1',
      chartId: 1,
      x: 100,
      y: 50,
      width: 300,
      height: 200,
      zIndex: 2,
    }
    expect(item.zIndex).toBe(2)
  })

  it('should allow valid CreateDashboardRequest', () => {
    const req: CreateDashboardRequest = {
      title: 'New Dashboard',
      type: 'dashboard',
      config: '{}',
    }
    expect(req.title).toBe('New Dashboard')
  })

  it('should allow CreateDashboardRequest without optional type', () => {
    const req: CreateDashboardRequest = {
      title: 'New Dashboard',
      config: '{}',
    }
    expect(req.type).toBeUndefined()
  })
})
