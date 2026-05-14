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
import { workbenchAPI } from './workbench'

describe('workbenchAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should get stats', async () => {
    vi.mocked(request.get).mockResolvedValue({ chartCount: 10, dashboardCount: 5 })
    await workbenchAPI.getStats()
    expect(request.get).toHaveBeenCalledWith('/workbench/stats')
  })

  it('should get recent views', async () => {
    vi.mocked(request.get).mockResolvedValue([])
    await workbenchAPI.getRecent()
    expect(request.get).toHaveBeenCalledWith('/workbench/recent')
  })

  it('should record visit', async () => {
    const data = { resourceType: 'dashboard', resourceId: 1 }
    vi.mocked(request.post).mockResolvedValue(undefined)
    await workbenchAPI.recordVisit(data)
    expect(request.post).toHaveBeenCalledWith('/workbench/recent', data)
  })

  it('should get favorites', async () => {
    vi.mocked(request.get).mockResolvedValue([])
    await workbenchAPI.getFavorites()
    expect(request.get).toHaveBeenCalledWith('/workbench/favorites')
  })

  it('should add favorite', async () => {
    vi.mocked(request.post).mockResolvedValue(undefined)
    await workbenchAPI.addFavorite('dashboard', 1)
    expect(request.post).toHaveBeenCalledWith('/workbench/favorites', { resourceType: 'dashboard', resourceId: 1 })
  })

  it('should remove favorite', async () => {
    vi.mocked(request.delete).mockResolvedValue(undefined)
    await workbenchAPI.removeFavorite('dashboard', 1)
    expect(request.delete).toHaveBeenCalledWith('/workbench/favorites/dashboard/1')
  })
})
