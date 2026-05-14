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
import { datasourceAPI } from './datasource'

describe('datasourceAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should list datasources', async () => {
    vi.mocked(request.get).mockResolvedValue([])
    await datasourceAPI.list()
    expect(request.get).toHaveBeenCalledWith('/datasource')
  })

  it('should create datasource', async () => {
    const data = { name: 'mysql1', type: 'mysql', config: '{}' }
    vi.mocked(request.post).mockResolvedValue({ id: 1 })
    await datasourceAPI.create(data)
    expect(request.post).toHaveBeenCalledWith('/datasource', data)
  })

  it('should get datasource by id', async () => {
    vi.mocked(request.get).mockResolvedValue({ id: 1 })
    await datasourceAPI.get(1)
    expect(request.get).toHaveBeenCalledWith('/datasource/1')
  })

  it('should update datasource', async () => {
    vi.mocked(request.put).mockResolvedValue(undefined)
    await datasourceAPI.update(1, { name: 'updated' })
    expect(request.put).toHaveBeenCalledWith('/datasource/1', { name: 'updated' })
  })

  it('should remove datasource', async () => {
    vi.mocked(request.delete).mockResolvedValue(undefined)
    await datasourceAPI.remove(1)
    expect(request.delete).toHaveBeenCalledWith('/datasource/1')
  })

  it('should test connection', async () => {
    const data = { type: 'mysql', config: { host: 'localhost' } }
    vi.mocked(request.post).mockResolvedValue({ success: true })
    await datasourceAPI.testConnection(data)
    expect(request.post).toHaveBeenCalledWith('/datasource/test', data)
  })

  it('should upload file with multipart header', async () => {
    const formData = new FormData()
    formData.append('file', new Blob(['test']))
    vi.mocked(request.post).mockResolvedValue({ id: 1 })
    await datasourceAPI.upload(formData)
    expect(request.post).toHaveBeenCalledWith(
      '/datasource/upload',
      formData,
      { headers: { 'Content-Type': 'multipart/form-data' } }
    )
  })
})
