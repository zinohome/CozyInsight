import axios, { type AxiosRequestConfig, type AxiosResponse } from 'axios';

// 响应拦截器会从 AxiosResponse<T> 中提取 data 为 T,所以方法直接返回 T。
// 这里用一个轻量包装避免每个调用点都写 .data。
type Unwrap<T> = T extends AxiosResponse<infer U> ? U : T;

const instance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器 - 自动附加 Token
instance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error),
);

// 响应拦截器 - 提取数据和错误处理
instance.interceptors.response.use(
  (response) => {
    const { data } = response;
    if (data.code !== 200) {
      return Promise.reject(new Error(data.error || '请求失败'));
    }
    return data.data;
  },
  (error) => Promise.reject(error),
);

function makeRequest<TData = unknown>(
  config: AxiosRequestConfig,
): Promise<Unwrap<TData>> {
  return instance.request<Unwrap<TData>>(config) as unknown as Promise<Unwrap<TData>>;
}

const request = {
  get: <TData = unknown>(url: string, config?: AxiosRequestConfig) =>
    makeRequest<TData>({ ...config, method: 'GET', url }),
  post: <TData = unknown>(
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig,
  ) => makeRequest<TData>({ ...config, method: 'POST', url, data }),
  put: <TData = unknown>(
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig,
  ) => makeRequest<TData>({ ...config, method: 'PUT', url, data }),
  delete: <TData = unknown>(url: string, config?: AxiosRequestConfig) =>
    makeRequest<TData>({ ...config, method: 'DELETE', url }),
};

export default request;
