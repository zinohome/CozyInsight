import { create } from 'zustand'
import { userAPI } from '@/api/user'
import type { UserInfo } from '@/types/auth'

interface AuthState {
  token: string | null
  user: UserInfo | null
  isAuthenticated: boolean
  setToken: (token: string) => void
  setUser: (user: UserInfo) => void
  fetchUser: () => Promise<void>
  logout: () => void
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: localStorage.getItem('token'),
  user: null,
  isAuthenticated: !!localStorage.getItem('token'),

  setToken: (token) => {
    localStorage.setItem('token', token)
    set({ token, isAuthenticated: true })
    // Auto fetch user info after login
    get().fetchUser().catch(() => {})
  },

  setUser: (user) => {
    set({ user })
  },

  fetchUser: async () => {
    if (!get().isAuthenticated) return
    try {
      const user = await userAPI.profile()
      set({ user })
    } catch {
      // ignore: user info fetch failure shouldn't break the app
    }
  },

  logout: () => {
    localStorage.removeItem('token')
    set({ token: null, user: null, isAuthenticated: false })
  },
}))
