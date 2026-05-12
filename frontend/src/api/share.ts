function authHeaders(): Record<string, string> {
  const token = localStorage.getItem('token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

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
    fetch(`/api/v1/dashboard/${dashboardId}/share`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
      body: JSON.stringify({ password, expireHours }),
    }).then(r => r.json()),

  get: (token: string, password?: string) => {
    const url = password
      ? `/api/v1/share/${token}?password=${encodeURIComponent(password)}`
      : `/api/v1/share/${token}`
    return fetch(url).then(r => r.json())
  },

  list: () =>
    fetch('/api/v1/share-links', {
      headers: authHeaders(),
    }).then(r => r.json()),

  remove: (id: number) =>
    fetch(`/api/v1/dashboard/${id}/share`, {
      method: 'DELETE',
      headers: authHeaders(),
    }).then(r => r.json()),
}
