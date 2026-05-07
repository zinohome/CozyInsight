import request from '../utils/request';

export interface Share {
    id: string;
    resourceType: string;
    resourceId: string;
    token: string;
    password?: string;
    expireTime: number;
    createTime: number;
    createBy: string;
}

export const shareAPI = {
    create: (data: Partial<Share>) => request.post<Share>('/share', data),
    get: (id: string) => request.get<Share>(`/share/${id}`),
    delete: (id: string) => request.delete(`/share/${id}`),
    list: (resourceType: string, resourceId: string) =>
        request.get<Share[]>(`/share?resourceType=${resourceType}&resourceId=${resourceId}`),
    validate: (token: string, password?: string) =>
        request.get<Share>(`/share/validate/${token}${password ? `?password=${password}` : ''}`),
};
