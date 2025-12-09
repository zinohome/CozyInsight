import request from './request';

export interface ScheduleTask {
    id: string;
    name: string;
    type: string;
    cronExpr: string;
    enabled: boolean;
    status: string;
    config?: any;
    lastRunTime?: number;
    createTime: number;
}

export const scheduleAPI = {
    create: (data: Partial<ScheduleTask>) =>
        request.post<ScheduleTask>('/schedule', data),

    update: (id: string, data: Partial<ScheduleTask>) =>
        request.put<ScheduleTask>(`/schedule/${id}`, data),

    delete: (id: string) =>
        request.delete(`/schedule/${id}`),

    get: (id: string) =>
        request.get<ScheduleTask>(`/schedule/${id}`),

    list: () =>
        request.get<ScheduleTask[]>('/schedule'),

    enable: (id: string) =>
        request.post(`/schedule/${id}/enable`),

    disable: (id: string) =>
        request.post(`/schedule/${id}/disable`),

    execute: (id: string) =>
        request.post(`/schedule/${id}/execute`),
};
