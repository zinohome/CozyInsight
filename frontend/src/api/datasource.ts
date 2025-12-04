import request from './request';
import type { Datasource, CreateDatasourceRequest, UpdateDatasourceRequest } from '../types/datasource';

export const getDatasourceList = () => {
    return request.get<any, Datasource[]>('/datasource');
};

export const getDatasource = (id: string) => {
    return request.get<any, Datasource>(`/datasource/${id}`);
};

export const createDatasource = (data: CreateDatasourceRequest) => {
    return request.post<any, Datasource>('/datasource', data);
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
