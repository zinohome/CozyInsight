import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

vi.mock('./request', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

import { shareAPI } from './share'

describe('shareAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    global.fetch = vi.fn()
  })

  afterEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  it('should create share link', async () => {
    vi.mocked(global.fetch).mockResolvedValue({
      json: () => Promise.resolve({ token: 'abc123' }),
    } as Response)
    localStorage.setItem('token', 'test-token')

    await shareAPI.create(1, 'pass', 24)

    expect(global.fetch).toHaveBeenCalledWith('/api/v1/dashboard/1/share', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: 'Bearer test-token',
      },
      body: JSON.stringify({ password: 'pass', expireHours: 24 }),
    })
  })

  it('should get shared dashboard without password', async () => {
    vi.mocked(global.fetch).mockResolvedValue({
      json: () => Promise.resolve({ dashboard: {} }),
    } as Response)

    await shareAPI.get('token123')

    expect(global.fetch).toHaveBeenCalledWith('/api/v1/share/token123')
  })

  it('should get shared dashboard with password in query', async () => {
    vi.mocked(global.fetch).mockResolvedValue({
      json: () => Promise.resolve({ dashboard: {} }),
    } as Response)

    await shareAPI.get('token123', 'mypass')

    expect(global.fetch).toHaveBeenCalledWith(
      '/api/v1/share/token123?password=mypass'
    )
  })

  it('should list share links with auth header', async () => {
    vi.mocked(global.fetch).mockResolvedValue({
      json: () => Promise.resolve([]),
    } as Response)
    localStorage.setItem('token', 'test-token')

    await shareAPI.list()

    expect(global.fetch).toHaveBeenCalledWith('/api/v1/share-links', {
      headers: { Authorization: 'Bearer test-token' },
    })
  })

  it('should remove share link', async () => {
    vi.mocked(global.fetch).mockResolvedValue({
      json: () => Promise.resolve({}),
    } as Response)
    localStorage.setItem('token', 'test-token')

    await shareAPI.remove(1)

    expect(global.fetch).toHaveBeenCalledWith('/api/v1/dashboard/1/share', {
      method: 'DELETE',
      headers: { Authorization: 'Bearer test-token' },
    })
  })
})
