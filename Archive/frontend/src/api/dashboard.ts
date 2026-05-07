import request from './request';
import type { Dashboard, CreateDashboardRequest, UpdateDashboardRequest, DashboardListParams } from '../types/dashboard';

/**
 * Dashboard API 接口封装
 */
export const dashboardAPI = {
    /**
     * 创建仪表板/文件夹
     */
    create: (data: CreateDashboardRequest) => {
        return request.post<any, Dashboard>('/dashboard', data);
    },

    /**
     * 更新仪表板
     */
    update: (id: string, data: UpdateDashboardRequest) => {
        return request.put<any, Dashboard>(`/dashboard/${id}`, data);
    },

    /**
     * 删除仪表板
     */
    delete: (id: string) => {
        return request.delete<any, void>(`/dashboard/${id}`);
    },

    /**
     * 获取仪表板详情
     */
    get: (id: string) => {
        return request.get<any, Dashboard>(`/dashboard/${id}`);
    },

    /**
     * 获取仪表板列表
     */
    list: (params?: DashboardListParams) => {
        return request.get<any, Dashboard[]>('/dashboard', { params });
    },
};

export default dashboardAPI;
