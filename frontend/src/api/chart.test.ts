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
import { chartAPI } from './chart'

describe('chartAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should list charts with correct path', async () => {
    const mockCharts = [{ id: 1, title: 'chart1' }]
    vi.mocked(request.get).mockResolvedValue(mockCharts)

    await chartAPI.list()

    expect(request.get).toHaveBeenCalledWith('/chart')
  })

  it('should create chart with correct path', async () => {
    const mockData = { title: 'New Chart', type: 'bar', datasetId: 1, config: '{}' }
    vi.mocked(request.post).mockResolvedValue({ id: 1 })

    await chartAPI.create(mockData)

    expect(request.post).toHaveBeenCalledWith('/chart', mockData)
  })

  it('should get chart by id with correct path', async () => {
    vi.mocked(request.get).mockResolvedValue({ id: 1 })

    await chartAPI.get(1)

    expect(request.get).toHaveBeenCalledWith('/chart/1')
  })

  it('should update chart with correct path', async () => {
    const mockData = { title: 'Updated' }
    vi.mocked(request.put).mockResolvedValue(undefined)

    await chartAPI.update(1, mockData)

    expect(request.put).toHaveBeenCalledWith('/chart/1', mockData)
  })

  it('should remove chart with correct path', async () => {
    vi.mocked(request.delete).mockResolvedValue(undefined)

    await chartAPI.remove(1)

    expect(request.delete).toHaveBeenCalledWith('/chart/1')
  })

  it('should get chart data with empty body by default', async () => {
    vi.mocked(request.post).mockResolvedValue({ data: [] })

    await chartAPI.getData(1)

    expect(request.post).toHaveBeenCalledWith('/chart/1/data', {})
  })

  it('should get chart data with runtime filters and drill dimension', async () => {
    const mockBody = {
      runtimeFilters: [{ field: 'province', operator: '=', value: '广东' }],
      drillDimension: 'city',
    }
    vi.mocked(request.post).mockResolvedValue({ data: [] })

    await chartAPI.getData(1, mockBody)

    expect(request.post).toHaveBeenCalledWith('/chart/1/data', mockBody)
  })
})
