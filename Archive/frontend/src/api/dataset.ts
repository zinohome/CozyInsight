import request from './request';
import type { DatasetGroup, DatasetTable, CreateGroupRequest, CreateTableRequest } from '../types/dataset';

export const datasetAPI = {
    listGroups: () => {
        return request.get<any, DatasetGroup[]>('/dataset/group');
    },

    createGroup: (data: CreateGroupRequest) => {
        return request.post<any, DatasetGroup>('/dataset/group', data);
    },

    listTables: () => {
        return request.get<any, DatasetTable[]>('/dataset/table');
    },

    createTable: (data: CreateTableRequest) => {
        return request.post<any, DatasetTable>('/dataset/table', data);
    },

    previewTable: (id: string) => {
        return request.post<any, any[]>(`/dataset/table/${id}/preview`);
    },
};

// 保留向后兼容的导出
export const getDatasetGroups = datasetAPI.listGroups;
export const createDatasetGroup = datasetAPI.createGroup;
export const createDatasetTable = datasetAPI.createTable;
export const previewDatasetTable = datasetAPI.previewTable;
