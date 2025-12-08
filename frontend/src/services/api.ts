import axios from 'axios';
import type { ChartView, CreateChartRequest, UpdateChartRequest } from '../types/chart';

const API_BASE_URL = '/api/v1';

// 获取 token
const getToken = () => localStorage.getItem('token');

// 创建 axios 实例
const apiClient = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
});

// 请求拦截器 - 添加 token
apiClient.interceptors.request.use(
    (config) => {
        const token = getToken();
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => Promise.reject(error)
);

// 响应拦截器 - 处理错误
apiClient.interceptors.response.use(
    (response) => response.data,
    (error) => {
        if (error.response?.status === 401) {
            // Token 过期，跳转登录
            localStorage.removeItem('token');
            window.location.href = '/login';
        }
        return Promise.reject(error);
    }
);

/**
 * 图表 API 服务
 */
export const chartApi = {
    /**
     * 获取图表列表
     */
    list: async (): Promise<ChartView[]> => {
        return apiClient.get('/chart');
    },

    /**
     * 获取图表详情
     */
    get: async (id: string): Promise<ChartView> => {
        return apiClient.get(`/chart/${id}`);
    },

    /**
     * 创建图表
     */
    create: async (data: CreateChartRequest): Promise<ChartView> => {
        return apiClient.post('/chart', data);
    },

    /**
     * 更新图表
     */
    update: async (id: string, data: UpdateChartRequest): Promise<ChartView> => {
        return apiClient.put(`/chart/${id}`, data);
    },

    /**
     * 删除图表
     */
    delete: async (id: string): Promise<void> => {
        return apiClient.delete(`/chart/${id}`);
    },

    /**
     * 获取图表数据
     */
    getData: async (id: string): Promise<any[]> => {
        return apiClient.get(`/chart/${id}/data`);
    },
};

/**
 * 数据集 API 服务
 */
export const datasetApi = {
    /**
     * 获取数据集列表
     */
    listTables: async (): Promise<any[]> => {
        return apiClient.get('/dataset/table');
    },

    /**
     * 预览数据
     */
    preview: async (id: string, limit = 100): Promise<any> => {
        return apiClient.get(`/dataset/table/${id}/preview?limit=${limit}`);
    },

    /**
     * 获取字段信息
     */
    getFields: async (id: string): Promise<any> => {
        return apiClient.get(`/dataset/table/${id}/fields`);
    },

    /**
     * 同步字段
     */
    syncFields: async (id: string): Promise<void> => {
        return apiClient.post(`/dataset/table/${id}/fields/sync`);
    },
};

/**
 * 数据源 API 服务
 */
export const datasourceApi = {
    /**
     * 获取数据源列表
     */
    list: async (): Promise<any[]> => {
        return apiClient.get('/datasource');
    },

    /**
     * 测试连接
     */
    testConnection: async (id: string): Promise<any> => {
        return apiClient.post(`/datasource/${id}/test`);
    },

    /**
     * 获取数据库列表
     */
    getDatabases: async (id: string): Promise<any> => {
        return apiClient.get(`/datasource/${id}/databases`);
    },

    /**
     * 获取表列表
     */
    getTables: async (id: string, database?: string): Promise<any> => {
        const params = database ? `?database=${database}` : '';
        return apiClient.get(`/datasource/${id}/tables${params}`);
    },

    /**
     * 获取表结构
     */
    getTableSchema: async (id: string, database: string, table: string): Promise<any> => {
        return apiClient.get(`/datasource/${id}/schema?database=${database}&table=${table}`);
    },
};

export default {
    chart: chartApi,
    dataset: datasetApi,
    datasource: datasourceApi,
};
