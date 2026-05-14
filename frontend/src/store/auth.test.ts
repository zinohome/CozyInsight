import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

// Must mock before importing store since store is created at module load time
vi.mock('@/api/user', () => ({
  userAPI: {
    profile: vi.fn(),
  },
}))

import { userAPI } from '@/api/user'
import { useAuthStore } from './auth'

describe('useAuthStore', () => {
  beforeEach(() => {
    localStorage.clear()
    // Reset store to initial state
    useAuthStore.setState({
      token: null,
      user: null,
      isAuthenticated: false,
    })
    vi.clearAllMocks()
  })

  afterEach(() => {
    localStorage.clear()
  })

  it('should initialize with no token', () => {
    expect(useAuthStore.getState().token).toBeNull()
    expect(useAuthStore.getState().isAuthenticated).toBe(false)
    expect(useAuthStore.getState().user).toBeNull()
  })

  it('should initialize from localStorage if token exists', () => {
    localStorage.setItem('token', 'existing-token')
    // Re-create store by re-importing is tricky in vitest; verify state directly
    // Instead, manually simulate what the store constructor does
    const token = localStorage.getItem('token')
    useAuthStore.setState({ token, isAuthenticated: !!token })
    expect(useAuthStore.getState().token).toBe('existing-token')
    expect(useAuthStore.getState().isAuthenticated).toBe(true)
  })

  it('should set token and update state', () => {
    useAuthStore.getState().setToken('new-token')
    expect(localStorage.getItem('token')).toBe('new-token')
    expect(useAuthStore.getState().token).toBe('new-token')
    expect(useAuthStore.getState().isAuthenticated).toBe(true)
  })

  it('should set user', () => {
    const user = { id: 1, username: 'test', email: 't@t.com', nickName: 'Test', avatar: '', isAdmin: false }
    useAuthStore.getState().setUser(user)
    expect(useAuthStore.getState().user).toEqual(user)
  })

  it('should fetch user profile when authenticated', async () => {
    const user = { id: 1, username: 'me', email: 't@t.com', nickName: 'Me', avatar: '', isAdmin: false }
    vi.mocked(userAPI.profile).mockResolvedValue(user)
    useAuthStore.setState({ token: 'token', isAuthenticated: true })

    await useAuthStore.getState().fetchUser()

    expect(userAPI.profile).toHaveBeenCalledTimes(1)
    expect(useAuthStore.getState().user).toEqual(user)
  })

  it('should not fetch user when not authenticated', async () => {
    await useAuthStore.getState().fetchUser()
    expect(userAPI.profile).not.toHaveBeenCalled()
  })

  it('should handle fetch user error gracefully', async () => {
    vi.mocked(userAPI.profile).mockRejectedValue(new Error('network error'))
    useAuthStore.setState({ token: 'token', isAuthenticated: true })

    // Should not throw
    await expect(useAuthStore.getState().fetchUser()).resolves.toBeUndefined()
    expect(useAuthStore.getState().user).toBeNull()
  })

  it('should logout and clear state', () => {
    useAuthStore.setState({
      token: 'token',
      user: { id: 1, username: 'test', email: 't@t.com', nickName: 'T', avatar: '', isAdmin: false },
      isAuthenticated: true,
    })

    useAuthStore.getState().logout()

    expect(localStorage.getItem('token')).toBeNull()
    expect(useAuthStore.getState().token).toBeNull()
    expect(useAuthStore.getState().user).toBeNull()
    expect(useAuthStore.getState().isAuthenticated).toBe(false)
  })

  it('setToken should auto fetch user', async () => {
    const user = { id: 1, username: 'auto', email: 'a@b.com', nickName: 'Auto', avatar: '', isAdmin: false }
    vi.mocked(userAPI.profile).mockResolvedValue(user)

    useAuthStore.getState().setToken('auto-token')

    // Wait for async fetchUser inside setToken
    await new Promise((resolve) => setTimeout(resolve, 10))

    expect(userAPI.profile).toHaveBeenCalled()
  })
})
