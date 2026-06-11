import { describe, it, expect, vi, beforeEach } from 'vitest'

// Mock axios before importing request
const mockAxiosInstance = {
  interceptors: {
    request: { use: vi.fn((fn) => fn) },
    response: { use: vi.fn((fn1, fn2) => ({ success: fn1, error: fn2 })) },
  },
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}

vi.mock('axios', () => ({
  default: {
    create: vi.fn(() => mockAxiosInstance),
  },
}))

describe('request', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.stubGlobal('localStorage', {
      getItem: vi.fn(),
      setItem: vi.fn(),
      removeItem: vi.fn(),
    })
  })

  it('should add auth header when token exists', async () => {
    const localStorage = window.localStorage as unknown as { getItem: ReturnType<typeof vi.fn> }
    localStorage.getItem.mockReturnValue('test-token')

    vi.resetModules()
    const request = (await import('./request')).default

    expect(request).toBeDefined()
    expect(mockAxiosInstance.interceptors.request.use).toHaveBeenCalled()

    const requestInterceptor = mockAxiosInstance.interceptors.request.use.mock.calls[0][0]
    const config = { headers: {} }
    const result = requestInterceptor(config)
    expect(result.headers.Authorization).toBe('Bearer test-token')
  })

  it('should not add auth header when no token', async () => {
    const localStorage = window.localStorage as unknown as { getItem: ReturnType<typeof vi.fn> }
    localStorage.getItem.mockReturnValue(null)

    vi.resetModules()
    await import('./request')

    const requestInterceptor = mockAxiosInstance.interceptors.request.use.mock.calls[0][0]
    const config = { headers: {} }
    const result = requestInterceptor(config)
    expect(result.headers.Authorization).toBeUndefined()
  })

  it('should handle successful response', async () => {
    vi.resetModules()
    await import('./request')

    const successInterceptor = mockAxiosInstance.interceptors.response.use.mock.calls[0][0]
    const response = { data: { code: 200, data: { id: 1 } } }
    const result = successInterceptor(response)
    expect(result).toEqual({ id: 1 })
  })

  it('should reject non-200 responses', async () => {
    vi.resetModules()
    await import('./request')

    const successInterceptor = mockAxiosInstance.interceptors.response.use.mock.calls[0][0]
    const response = { data: { code: 500, error: 'Server Error' } }
    const result = successInterceptor(response)
    await expect(result).rejects.toThrow('Server Error')
  })

  it('should reject with default message when no error field', async () => {
    vi.resetModules()
    await import('./request')

    const successInterceptor = mockAxiosInstance.interceptors.response.use.mock.calls[0][0]
    const response = { data: { code: 500 } }
    const result = successInterceptor(response)
    await expect(result).rejects.toThrow('请求失败')
  })

  it('should pass through response error', async () => {
    vi.resetModules()
    await import('./request')

    const errorInterceptor = mockAxiosInstance.interceptors.response.use.mock.calls[0][1]
    const error = new Error('network error')
    const result = errorInterceptor(error)
    await expect(result).rejects.toThrow('network error')
  })

  it('should expose get method', async () => {
    vi.resetModules()
    const request = (await import('./request')).default
    expect(typeof request.get).toBe('function')
  })

  it('should expose post method', async () => {
    vi.resetModules()
    const request = (await import('./request')).default
    expect(typeof request.post).toBe('function')
  })

  it('should expose put method', async () => {
    vi.resetModules()
    const request = (await import('./request')).default
    expect(typeof request.put).toBe('function')
  })

  it('should expose delete method', async () => {
    vi.resetModules()
    const request = (await import('./request')).default
    expect(typeof request.delete).toBe('function')
  })

  it('should call delete method on axios instance', async () => {
    vi.resetModules()
    const request = (await import('./request')).default
    request.delete('/test')
    expect(mockAxiosInstance.delete).toHaveBeenCalledWith('/test', undefined)
  })
})
