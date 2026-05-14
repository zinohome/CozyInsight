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
import { roleAPI } from './role'

describe('roleAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should list roles', async () => {
    vi.mocked(request.get).mockResolvedValue([])
    await roleAPI.list()
    expect(request.get).toHaveBeenCalledWith('/role')
  })

  it('should create role', async () => {
    const data = { name: 'admin' }
    vi.mocked(request.post).mockResolvedValue({ id: 1 })
    await roleAPI.create(data)
    expect(request.post).toHaveBeenCalledWith('/role', data)
  })

  it('should get role by id', async () => {
    vi.mocked(request.get).mockResolvedValue({ id: 1 })
    await roleAPI.get(1)
    expect(request.get).toHaveBeenCalledWith('/role/1')
  })

  it('should update role', async () => {
    vi.mocked(request.put).mockResolvedValue(undefined)
    await roleAPI.update(1, { name: 'updated' })
    expect(request.put).toHaveBeenCalledWith('/role/1', { name: 'updated' })
  })

  it('should remove role', async () => {
    vi.mocked(request.delete).mockResolvedValue(undefined)
    await roleAPI.remove(1)
    expect(request.delete).toHaveBeenCalledWith('/role/1')
  })

  it('should list menus', async () => {
    vi.mocked(request.get).mockResolvedValue([])
    await roleAPI.listMenus()
    expect(request.get).toHaveBeenCalledWith('/role/menus')
  })

  it('should set role menus', async () => {
    vi.mocked(request.post).mockResolvedValue(undefined)
    await roleAPI.setRoleMenus(1, [1, 2, 3])
    expect(request.post).toHaveBeenCalledWith('/role/1/menus', { menuIds: [1, 2, 3] })
  })

  it('should get role menus', async () => {
    vi.mocked(request.get).mockResolvedValue([1, 2])
    await roleAPI.getRoleMenus(1)
    expect(request.get).toHaveBeenCalledWith('/role/1/menus')
  })

  it('should set user roles', async () => {
    vi.mocked(request.post).mockResolvedValue(undefined)
    await roleAPI.setUserRoles(1, [2, 3])
    expect(request.post).toHaveBeenCalledWith('/user/1/roles', { roleIds: [2, 3] })
  })
})
