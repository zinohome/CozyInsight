import request from './request'
import type { User, CreateUserRequest, ChangePasswordRequest } from '@/types/user'

export const userAPI = {
  list: () => request.get<User[]>('/user'),
  create: (data: CreateUserRequest) => request.post<User>('/user', data),
  get: (id: number) => request.get<User>(`/user/${id}`),
  update: (id: number, data: Partial<CreateUserRequest>) => request.put(`/user/${id}`, data),
  remove: (id: number) => request.delete(`/user/${id}`),
  profile: () => request.get<User>('/user/profile'),
  changePassword: (data: ChangePasswordRequest) => request.post('/user/change-password', data),
}
