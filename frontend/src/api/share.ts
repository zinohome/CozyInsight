import request from './request';

export interface ShareLink {
  id: number
  token: string
  resourceType: string
  resourceId: number
  expireAt?: string
  password?: string
  status: number
  createdAt: string
}

export const shareAPI = {
  create: (dashboardId: number, password?: string, expireHours?: number) =>
    request.post(`/dashboard/${dashboardId}/share`, { password, expireHours }),

  get: (token: string, password?: string) => {
    const url = password
      ? `/share/${token}?password=${encodeURIComponent(password)}`
      : `/share/${token}`
    return request.get(url)
  },

  list: () => request.get('/share-links'),

  remove: (id: number) => request.delete(`/dashboard/${id}/share`),
}
