import request from './request';

export interface SystemSetting {
    id: string;
    type: string;
    key: string;
    value: string;
    updateTime: number;
    updateBy: string;
}

export const settingAPI = {
    get: (key: string) =>
        request.get<SystemSetting>(`/setting/${key}`),

    set: (type: string, key: string, value: string) =>
        request.post('/setting', { type, key, value }),

    listByType: (type: string) =>
        request.get<SystemSetting[]>(`/setting/type/${type}`),

    delete: (key: string) =>
        request.delete(`/setting/${key}`),
};
