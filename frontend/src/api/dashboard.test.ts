import { describe, it, expect, vi } from 'vitest'

vi.mock('./request', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

import request from './request'
import { dashboardAPI } from './dashboard'

describe('dashboardAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should list dashboards', async () => {
    vi.mocked(request.get).mockResolvedValue([])
    await dashboardAPI.list()
    expect(request.get).toHaveBeenCalledWith('/dashboard')
  })

  it('should create dashboard', async () => {
    const data = { name: 'db1', type: 'dashboard' }
    vi.mocked(request.post).mockResolvedValue({ id: 1 })
    await dashboardAPI.create(data)
    expect(request.post).toHaveBeenCalledWith('/dashboard', data)
  })

  it('should get dashboard by id', async () => {
    vi.mocked(request.get).mockResolvedValue({ id: 1 })
    await dashboardAPI.get(1)
    expect(request.get).toHaveBeenCalledWith('/dashboard/1')
  })

  it('should update dashboard', async () => {
    const data = { name: 'updated' }
    vi.mocked(request.put).mockResolvedValue(undefined)
    await dashboardAPI.update(1, data)
    expect(request.put).toHaveBeenCalledWith('/dashboard/1', data)
  })

  it('should remove dashboard', async () => {
    vi.mocked(request.delete).mockResolvedValue(undefined)
    await dashboardAPI.remove(1)
    expect(request.delete).toHaveBeenCalledWith('/dashboard/1')
  })

  it('should add chart to dashboard with position config', async () => {
    const data = { chartId: 1, positionX: 0, positionY: 0, width: 6, height: 4, config: '{}' }
    vi.mocked(request.post).mockResolvedValue(undefined)
    await dashboardAPI.addChart(1, data)
    expect(request.post).toHaveBeenCalledWith('/dashboard/1/charts', data)
  })

  it('should get dashboard charts', async () => {
    vi.mocked(request.get).mockResolvedValue([])
    await dashboardAPI.getCharts(1)
    expect(request.get).toHaveBeenCalledWith('/dashboard/1/charts')
  })

  it('should remove chart from dashboard', async () => {
    vi.mocked(request.delete).mockResolvedValue(undefined)
    await dashboardAPI.removeChart(1, 2)
    expect(request.delete).toHaveBeenCalledWith('/dashboard/1/charts/2')
  })
})
