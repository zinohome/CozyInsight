// Chart 图表视图类型定义

export interface ChartView {
    id: string;
    name: string;
    sceneId?: string;  // 场景ID/仪表板ID
    tableId: string;   // 数据集ID
    type: ChartType;   // 图表类型
    title?: string;    // 图表标题
    xAxis?: string;    // X轴配置 (JSON)
    yAxis?: string;    // Y轴配置 (JSON)
    customAttr?: string;   // 自定义属性 (JSON)
    customStyle?: string;  // 自定义样式 (JSON)
    snapshot?: string;     // 快照/缩略图
    createTime?: number;
    updateTime?: number;
    createBy?: string;
}

// 图表类型枚举
export type ChartType =
    | 'bar'           // 柱状图
    | 'line'          // 折线图
    | 'pie'           // 饼图
    | 'scatter'       // 散点图
    | 'area'          // 面积图
    | 'table'         // 表格
    | 'map'           // 地图
    | 'gauge'         // 仪表盘
    | 'radar'         // 雷达图
    | 'funnel'        // 漏斗图
    | 'wordcloud';    // 词云

// 创建图表请求
export interface CreateChartRequest {
    name: string;
    tableId: string;
    type: ChartType;
    sceneId?: string;
    title?: string;
    xAxis?: string;
    yAxis?: string;
    customAttr?: string;
    customStyle?: string;
}

// 更新图表请求
export interface UpdateChartRequest {
    name?: string;
    tableId?: string;
    type?: ChartType;
    sceneId?: string;
    title?: string;
    xAxis?: string;
    yAxis?: string;
    customAttr?: string;
    customStyle?: string;
    snapshot?: string;
}

// 图表列表查询参数
export interface ChartListParams {
    sceneId?: string;
}

// 图表配置 - XAxis/YAxis 的结构
export interface ChartAxisConfig {
    fields: ChartFieldConfig[];
}

export interface ChartFieldConfig {
    id: string;
    name: string;
    dataType?: string;
    aggregate?: 'sum' | 'avg' | 'count' | 'max' | 'min';
    sort?: 'asc' | 'desc';
}

// 自定义属性结构（简化版）
export interface ChartCustomAttr {
    color?: {
        colors?: string[];
        alpha?: number;
    };
    size?: {
        width?: number;
        height?: number;
    };
    label?: {
        show?: boolean;
        position?: string;
    };
}

// 自定义样式结构（简化版）
export interface ChartCustomStyle {
    title?: {
        show?: boolean;
        text?: string;
        fontSize?: number;
        color?: string;
    };
    legend?: {
        show?: boolean;
        position?: 'top' | 'bottom' | 'left' | 'right';
    };
    background?: {
        color?: string;
    };
}
// 统一的图表配置接口
export interface ChartConfig {
    xAxis?: ChartAxisConfig;
    yAxis?: ChartAxisConfig;
    customAttr?: ChartCustomAttr;
    customStyle?: ChartCustomStyle;
}
