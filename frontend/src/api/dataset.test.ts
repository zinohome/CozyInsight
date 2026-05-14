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
import { datasetAPI } from './dataset'

describe('datasetAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should list datasets', async () => {
    vi.mocked(request.get).mockResolvedValue([])
    await datasetAPI.list()
    expect(request.get).toHaveBeenCalledWith('/dataset')
  })

  it('should create dataset', async () => {
    const data = { name: 'ds1', datasourceId: 1, tableName: 't1', type: 'db' }
    vi.mocked(request.post).mockResolvedValue({ id: 1 })
    await datasetAPI.create(data)
    expect(request.post).toHaveBeenCalledWith('/dataset', data)
  })

  it('should get dataset by id', async () => {
    vi.mocked(request.get).mockResolvedValue({ id: 1 })
    await datasetAPI.get(1)
    expect(request.get).toHaveBeenCalledWith('/dataset/1')
  })

  it('should update dataset', async () => {
    vi.mocked(request.put).mockResolvedValue(undefined)
    await datasetAPI.update(1, { name: 'updated' })
    expect(request.put).toHaveBeenCalledWith('/dataset/1', { name: 'updated' })
  })

  it('should remove dataset', async () => {
    vi.mocked(request.delete).mockResolvedValue(undefined)
    await datasetAPI.remove(1)
    expect(request.delete).toHaveBeenCalledWith('/dataset/1')
  })

  it('should sync fields', async () => {
    vi.mocked(request.post).mockResolvedValue(undefined)
    await datasetAPI.syncFields(1)
    expect(request.post).toHaveBeenCalledWith('/dataset/1/sync-fields')
  })

  it('should preview with default limit', async () => {
    vi.mocked(request.get).mockResolvedValue({ columns: [], rows: [] })
    await datasetAPI.preview(1)
    expect(request.get).toHaveBeenCalledWith('/dataset/1/preview', { params: { limit: undefined } })
  })

  it('should preview with custom limit', async () => {
    vi.mocked(request.get).mockResolvedValue({ columns: [], rows: [] })
    await datasetAPI.preview(1, 50)
    expect(request.get).toHaveBeenCalledWith('/dataset/1/preview', { params: { limit: 50 } })
  })
})
