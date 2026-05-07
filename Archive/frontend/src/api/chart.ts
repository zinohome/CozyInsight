import request from './request';
import type { ChartView, CreateChartRequest, UpdateChartRequest, ChartListParams } from '../types/chart';

/**
 * Chart API 接口封装
 */
export const chartAPI = {
    /**
     * 创建图表
     */
    create: (data: CreateChartRequest) => {
        return request.post<any, ChartView>('/chart', data);
    },

    /**
     * 更新图表
     */
    update: (id: string, data: UpdateChartRequest) => {
        return request.put<any, ChartView>(`/chart/${id}`, data);
    },

    /**
     * 删除图表
     */
    delete: (id: string) => {
        return request.delete<any, void>(`/chart/${id}`);
    },

    /**
     * 获取图表详情
     */
    get: (id: string) => {
        return request.get<any, ChartView>(`/chart/${id}`);
    },

    /**
     * 获取图表列表
     */
    list: (params?: ChartListParams) => {
        return request.get<any, ChartView[]>('/chart', { params });
    },

    /**
     * 获取图表数据
     */
    getData: (id: string) => {
        return request.get<any[]>(`/chart/${id}/data`);
    },
};

export default chartAPI;
