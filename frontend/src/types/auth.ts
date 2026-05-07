export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  userId: number
  username: string
  nickName: string
  isAdmin: boolean
}

export interface RegisterRequest {
  username: string
  password: string
  email: string
  nickName?: string
}

export interface UserInfo {
  id: number
  username: string
  email: string
  nickName: string
  avatar: string
  isAdmin: boolean
}
