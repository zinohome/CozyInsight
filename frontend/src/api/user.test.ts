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
import { userAPI } from './user'

describe('userAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should list users', async () => {
    const mockUsers = [{ id: 1, username: 'admin' }]
    vi.mocked(request.get).mockResolvedValue(mockUsers)

    const result = await userAPI.list()

    expect(request.get).toHaveBeenCalledWith('/user')
    expect(result).toEqual(mockUsers)
  })

  it('should create user', async () => {
    const mockData = { username: 'new', password: 'pass', email: 'a@b.com' }
    vi.mocked(request.post).mockResolvedValue({ id: 2 })

    await userAPI.create(mockData)

    expect(request.post).toHaveBeenCalledWith('/user', mockData)
  })

  it('should get user by id', async () => {
    const mockUser = { id: 1, username: 'admin' }
    vi.mocked(request.get).mockResolvedValue(mockUser)

    const result = await userAPI.get(1)

    expect(request.get).toHaveBeenCalledWith('/user/1')
    expect(result).toEqual(mockUser)
  })

  it('should update user', async () => {
    const mockData = { username: 'updated' }
    vi.mocked(request.put).mockResolvedValue(undefined)

    await userAPI.update(1, mockData)

    expect(request.put).toHaveBeenCalledWith('/user/1', mockData)
  })

  it('should remove user', async () => {
    vi.mocked(request.delete).mockResolvedValue(undefined)

    await userAPI.remove(1)

    expect(request.delete).toHaveBeenCalledWith('/user/1')
  })

  it('should get profile', async () => {
    const mockProfile = { id: 1, username: 'me' }
    vi.mocked(request.get).mockResolvedValue(mockProfile)

    const result = await userAPI.profile()

    expect(request.get).toHaveBeenCalledWith('/user/profile')
    expect(result).toEqual(mockProfile)
  })

  it('should change password', async () => {
    const mockData = { oldPassword: 'old', newPassword: 'new' }
    vi.mocked(request.post).mockResolvedValue(undefined)

    await userAPI.changePassword(mockData)

    expect(request.post).toHaveBeenCalledWith('/user/change-password', mockData)
  })
})
