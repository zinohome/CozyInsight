// 布局相关类型定义

import type { Layout as RGLLayout } from 'react-grid-layout';

// 扩展 react-grid-layout 的 Layout 类型
export interface LayoutItem extends RGLLayout {
    componentType?: ComponentType;  // 组件类型
    componentId?: string;            // 组件ID (如 chartId)
    config?: Record<string, any>;    // 组件配置
}

// 组件类型枚举
export type ComponentType = 'chart' | 'text' | 'image' | 'iframe';

// 布局配置
export interface DashboardLayout {
    layouts: LayoutItem[];
    cols?: number;
    rowHeight?: number;
}
