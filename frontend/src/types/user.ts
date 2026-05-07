export interface User {
  id: number
  username: string
  email: string
  nickName: string
  avatar: string
  phone: string
  status: number
  isAdmin: boolean
  lastLoginAt?: string
  createdAt: string
}

export interface CreateUserRequest {
  username: string
  password: string
  email: string
  nickName?: string
  phone?: string
  status?: number
  isAdmin?: boolean
}

export interface ChangePasswordRequest {
  oldPassword: string
  newPassword: string
}
