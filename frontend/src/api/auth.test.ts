import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

// Mock request module
vi.mock('./request', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

import request from './request'
import { authAPI } from './auth'

describe('authAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    localStorage.clear()
  })

  it('should call correct endpoint for login', async () => {
    const mockData = { username: 'admin', password: '123456' }
    const mockResponse = { token: 'jwt-token' }
    vi.mocked(request.post).mockResolvedValue(mockResponse)

    const result = await authAPI.login(mockData)

    expect(request.post).toHaveBeenCalledWith('/auth/login', mockData)
    expect(result).toEqual(mockResponse)
  })

  it('should call correct endpoint for register', async () => {
    const mockData = { username: 'newuser', password: 'pass', email: 'a@b.com' }
    vi.mocked(request.post).mockResolvedValue(undefined)

    await authAPI.register(mockData)

    expect(request.post).toHaveBeenCalledWith('/auth/register', mockData)
  })
})
