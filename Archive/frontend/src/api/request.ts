import axios from 'axios';

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
});

// 请求拦截器 - 自动附加 Token
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器 - 提取数据和错误处理
request.interceptors.response.use(
  (response) => {
    return response.data;
  },
  (error) => {
    // Token 过期或无效，跳转登录页
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      if (window.location.pathname !== '/login') {
        window.location.href = '/login';
      }
    }

    const message = error.response?.data?.error || error.message || '请求失败';
    return Promise.reject(new Error(message));
  }
);

export default request;
