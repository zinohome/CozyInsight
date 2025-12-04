export interface DatasetGroup {
    id: string;
    name: string;
    pid?: string;
    level: number;
    nodeType: 'folder' | 'dataset';
    type?: string;
    createTime?: number;
    createBy?: string;
    children?: DatasetGroup[];
}

export interface DatasetTable {
    id: string;
    name: string;
    tableName?: string;
    datasourceId: string;
    datasetGroupId: string;
    type: 'db' | 'sql';
    info: string; // JSON string
    createTime?: number;
    updateTime?: number;
    createBy?: string;
}

export interface DatasetTableField {
    id: string;
    datasourceId: string;
    datasetTableId: string;
    datasetGroupId: string;
    originName: string;
    name: string;
    type: string;
    deType: number;
    groupType: 'd' | 'q'; // dimension | measure
    checked: boolean;
    columnIndex: number;
}

export interface CreateGroupRequest {
    name: string;
    pid?: string;
    level: number;
    nodeType: 'folder' | 'dataset';
    type?: string;
}

export interface CreateTableRequest {
    name: string;
    tableName?: string;
    datasourceId: string;
    datasetGroupId: string;
    type: 'db' | 'sql';
    info: string;
}
