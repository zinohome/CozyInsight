import { describe, it, expect } from 'vitest'
import type { Role, Menu, CreateRoleRequest } from './role'

describe('role types', () => {
  it('should allow valid Role', () => {
    const role: Role = {
      id: 1,
      name: '管理员',
      code: 'admin',
      description: '系统管理员',
      status: 1,
      createdAt: '2024-01-01',
    }
    expect(role.code).toBe('admin')
    expect(role.status).toBe(1)
  })

  it('should allow valid Menu', () => {
    const menu: Menu = {
      id: 1,
      parentId: 0,
      name: '工作台',
      path: '/',
      component: 'Workbench',
      icon: 'DashboardOutlined',
      sort: 1,
      status: 1,
    }
    expect(menu.parentId).toBe(0)
    expect(menu.sort).toBe(1)
  })

  it('should allow valid CreateRoleRequest', () => {
    const req: CreateRoleRequest = {
      name: '新角色',
      code: 'new_role',
      description: '新角色描述',
    }
    expect(req.code).toBe('new_role')
  })

  it('should allow CreateRoleRequest without description', () => {
    const req: CreateRoleRequest = {
      name: '新角色',
      code: 'new_role',
    }
    expect(req.description).toBeUndefined()
  })
})
