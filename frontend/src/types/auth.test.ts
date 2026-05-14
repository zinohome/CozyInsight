import { describe, it, expect } from 'vitest'
import type { LoginRequest, LoginResponse, RegisterRequest, UserInfo } from './auth'

describe('auth types', () => {
  it('should allow valid LoginRequest', () => {
    const req: LoginRequest = {
      username: 'admin',
      password: 'password123',
    }
    expect(req.username).toBe('admin')
  })

  it('should allow valid LoginResponse', () => {
    const resp: LoginResponse = {
      token: 'jwt-token',
      userId: 1,
      username: 'admin',
      nickName: 'Admin',
      isAdmin: true,
    }
    expect(resp.token).toBe('jwt-token')
    expect(resp.isAdmin).toBe(true)
  })

  it('should allow valid RegisterRequest', () => {
    const req: RegisterRequest = {
      username: 'newuser',
      password: 'pass123',
      email: 'test@example.com',
      nickName: 'Test User',
    }
    expect(req.email).toBe('test@example.com')
  })

  it('should allow RegisterRequest without optional nickName', () => {
    const req: RegisterRequest = {
      username: 'newuser',
      password: 'pass123',
      email: 'test@example.com',
    }
    expect(req.nickName).toBeUndefined()
  })

  it('should allow valid UserInfo', () => {
    const user: UserInfo = {
      id: 1,
      username: 'admin',
      email: 'admin@example.com',
      nickName: 'Admin',
      avatar: 'https://example.com/avatar.png',
      isAdmin: true,
    }
    expect(user.avatar).toBe('https://example.com/avatar.png')
  })
})
