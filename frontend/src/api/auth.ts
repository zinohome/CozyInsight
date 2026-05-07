import request from './request'
import type { LoginRequest, LoginResponse, RegisterRequest } from '@/types/auth'

export const authAPI = {
  login: (data: LoginRequest) =>
    request.post<LoginResponse>('/auth/login', data),

  register: (data: RegisterRequest) =>
    request.post('/auth/register', data),
}
