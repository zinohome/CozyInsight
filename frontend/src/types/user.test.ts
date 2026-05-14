import { describe, it, expect } from 'vitest'
import type { User, CreateUserRequest, ChangePasswordRequest } from './user'

describe('user types', () => {
  it('should allow valid User', () => {
    const user: User = {
      id: 1,
      username: 'admin',
      email: 'admin@example.com',
      nickName: 'Admin',
      avatar: 'https://example.com/avatar.png',
      phone: '13800138000',
      status: 1,
      isAdmin: true,
      lastLoginAt: '2024-01-01',
      createdAt: '2024-01-01',
    }
    expect(user.isAdmin).toBe(true)
    expect(user.status).toBe(1)
  })

  it('should allow User without optional lastLoginAt', () => {
    const user: User = {
      id: 1,
      username: 'newuser',
      email: 'new@example.com',
      nickName: 'New',
      avatar: '',
      phone: '',
      status: 1,
      isAdmin: false,
      createdAt: '2024-01-01',
    }
    expect(user.lastLoginAt).toBeUndefined()
  })

  it('should allow valid CreateUserRequest', () => {
    const req: CreateUserRequest = {
      username: 'test',
      password: 'password123',
      email: 'test@example.com',
      nickName: 'Test',
      phone: '13800138000',
      status: 1,
      isAdmin: false,
    }
    expect(req.username).toBe('test')
  })

  it('should allow CreateUserRequest with minimal fields', () => {
    const req: CreateUserRequest = {
      username: 'test',
      password: 'pass',
      email: 'test@example.com',
    }
    expect(req.nickName).toBeUndefined()
    expect(req.isAdmin).toBeUndefined()
  })

  it('should allow valid ChangePasswordRequest', () => {
    const req: ChangePasswordRequest = {
      oldPassword: 'old123',
      newPassword: 'new123',
    }
    expect(req.oldPassword).toBe('old123')
    expect(req.newPassword).toBe('new123')
  })
})
