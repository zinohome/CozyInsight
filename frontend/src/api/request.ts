import axios from 'axios'

const instance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

instance.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

instance.interceptors.response.use(
  (response) => {
    const { data } = response
    if (data.code !== 200) {
      return Promise.reject(new Error(data.error || '请求失败'))
    }
    return data.data
  },
  (error) => {
    return Promise.reject(error)
  },
)

const request = {
  get: <T>(url: string, config?: Parameters<typeof instance.get>[1]) =>
    instance.get<unknown, T>(url, config),
  post: <T>(url: string, data?: unknown, config?: Parameters<typeof instance.post>[2]) =>
    instance.post<unknown, T>(url, data, config),
  put: <T>(url: string, data?: unknown, config?: Parameters<typeof instance.put>[2]) =>
    instance.put<unknown, T>(url, data, config),
  delete: <T>(url: string, config?: Parameters<typeof instance.delete>[1]) =>
    instance.delete<unknown, T>(url, config),
}

export default request
