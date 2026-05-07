// Dashboard 仪表板类型定义

export interface Dashboard {
    id: string;
    name: string;
    pid: string;          // 父ID，'0'表示根节点
    nodeType: 'folder' | 'dashboard';
    type?: 'dashboard' | 'dataV'; // 仪表板 | 数据大屏
    canvasStyleData?: string;     // 画布样式 (JSON)
    componentData?: string;       // 组件数据 (JSON)
    status: number;               // 0=未发布 1=已发布
    sort: number;
    createTime?: number;
    updateTime?: number;
    createBy?: string;
    children?: Dashboard[];       // 用于树形展示
}

// 创建仪表板请求
export interface CreateDashboardRequest {
    name: string;
    pid: string;
    nodeType: 'folder' | 'dashboard';
    type?: 'dashboard' | 'dataV';
    canvasStyleData?: string;
    componentData?: string;
}

// 更新仪表板请求
export interface UpdateDashboardRequest {
    name?: string;
    pid?: string;
    canvasStyleData?: string;
    componentData?: string;
    status?: number;
    sort?: number;
}

// 仪表板列表查询参数
export interface DashboardListParams {
    pid?: string;
}
