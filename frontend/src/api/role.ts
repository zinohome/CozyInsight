import request from './request'
import type { Role, Menu, CreateRoleRequest } from '@/types/role'

export const roleAPI = {
  list: () => request.get<Role[]>('/role'),
  create: (data: CreateRoleRequest) => request.post<Role>('/role', data),
  get: (id: number) => request.get<Role>(`/role/${id}`),
  update: (id: number, data: Partial<CreateRoleRequest>) => request.put(`/role/${id}`, data),
  remove: (id: number) => request.delete(`/role/${id}`),
  listMenus: () => request.get<Menu[]>('/role/menus'),
  setRoleMenus: (roleId: number, menuIds: number[]) => request.post(`/role/${roleId}/menus`, { menuIds }),
  getRoleMenus: (roleId: number) => request.get<number[]>(`/role/${roleId}/menus`),
  setUserRoles: (userId: number, roleIds: number[]) => request.post(`/user/${userId}/roles`, { roleIds }),
}
