import { create } from 'zustand'
import type { UserInfo } from '@/types/auth'

interface AuthState {
  token: string | null
  user: UserInfo | null
  isAuthenticated: boolean
  setToken: (token: string) => void
  setUser: (user: UserInfo) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem('token'),
  user: null,
  isAuthenticated: !!localStorage.getItem('token'),

  setToken: (token) => {
    localStorage.setItem('token', token)
    set({ token, isAuthenticated: true })
  },

  setUser: (user) => {
    set({ user })
  },

  logout: () => {
    localStorage.removeItem('token')
    set({ token: null, user: null, isAuthenticated: false })
  },
}))
