function authHeaders(): Record<string, string> {
  const token = localStorage.getItem('token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export const shareAPI = {
  create: (dashboardId: number) =>
    fetch(`/api/v1/dashboard/${dashboardId}/share`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
    }).then(r => r.json()),

  get: (token: string) =>
    fetch(`/api/v1/share/${token}`).then(r => r.json()),
}
