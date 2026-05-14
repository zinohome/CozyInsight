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
import { logAPI } from './log'

describe('logAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should list logs with default limit', async () => {
    vi.mocked(request.get).mockResolvedValue([])
    await logAPI.list()
    expect(request.get).toHaveBeenCalledWith('/operation-log', { params: { limit: undefined } })
  })

  it('should list logs with custom limit', async () => {
    vi.mocked(request.get).mockResolvedValue([])
    await logAPI.list(50)
    expect(request.get).toHaveBeenCalledWith('/operation-log', { params: { limit: 50 } })
  })
})
