export interface Datasource {
    id: string;
    name: string;
    description?: string;
    type: string;
    pid?: string;
    editType?: number;
    configuration: string;
    status?: string;
    createTime?: number;
    updateTime?: number;
    createBy?: string;
    qrtzInstance?: string;
    taskStatus?: string;
}

export interface CreateDatasourceRequest {
    name: string;
    description?: string;
    type: string;
    configuration: string;
}

export interface UpdateDatasourceRequest {
    id: string;
    name: string;
    description?: string;
    type: string;
    configuration: string;
}
