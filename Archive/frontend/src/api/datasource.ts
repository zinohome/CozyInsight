import request from './request';
import type { Datasource, CreateDatasourceRequest, UpdateDatasourceRequest } from '../types/datasource';

// 连接测试结果
export interface ConnectionTestResult {
    success: boolean;
    message: string;
    time?: number; // 响应时间(毫秒)
}

export const getDatasourceList = () => {
    return request.get<any, Datasource[]>('/datasource');
};

export const getDatasource = (id: string) => {
    return request.get<any, Datasource>(`/datasource/${id}`);
};

export const createDatasource = (data: CreateDatasourceRequest) => {
    return request.post<any, Datasource>('/datasource', data);
};

export const datasourceAPI = {
    list: () => request.get<Datasource[]>('/datasource'),
    get: (id: string) => request.get<Datasource>(`/datasource/${id}`),
    create: (data: Partial<Datasource>) => request.post('/datasource', data),
    update: (id: string, data: Partial<Datasource>) => request.put(`/datasource/${id}`, data),
    delete: (id: string) => request.delete(`/datasource/${id}`),
    testConnection: (id: string) => request.post<ConnectionTestResult>(`/datasource/${id}/test`, {}),
    testConnectionByConfig: (config: string) => request.post<ConnectionTestResult>('/datasource/test-config', { configuration: config }),
    getDatabases: (id: string) => request.get<string[]>(`/datasource/${id}/databases`),
    getTables: (id: string, database: string) => request.get<string[]>(`/datasource/${id}/tables?database=${database}`),
};

export const updateDatasource = (id: string, data: UpdateDatasourceRequest) => {
    return request.put<any, Datasource>(`/datasource/${id}`, data);
};

export const deleteDatasource = (id: string) => {
    return request.delete(`/datasource/${id}`);
};

// 测试连接
export const validateDatasource = (id: string) => {
    return request.post<{ status: string; message: string }>(`/datasource/${id}/validate`);
};

// 获取表列表
export const getDatasourceTables = (id: string) => {
    return request.get<string[]>(`/datasource/${id}/tables`);
};

// 获取字段列表
export const getDatasourceFields = (id: string, tableName: string) => {
    return request.get<any[]>(`/datasource/${id}/fields`, {
        params: { table: tableName }
    });
};
