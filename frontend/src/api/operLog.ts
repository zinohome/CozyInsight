import request from './request';

export interface OperLog {
    id: string;
    userId: string;
    username: string;
    module: string;
    action: string;
    detail: string;
    resourceId: string;
    ip: string;
    status: number;
    createTime: number;
}

export const operLogAPI = {
    list: (params?: {
        userId?: string;
        module?: string;
        startTime?: number;
        endTime?: number;
        page?: number;
        pageSize?: number;
    }) =>
        request.get<{
            data: OperLog[];
            total: number;
            page: number;
            pageSize: number;
        }>('/log', { params }),

    cleanOld: (beforeDays: number = 90) =>
        request.post('/log/clean', null, { params: { beforeDays } }),
};
