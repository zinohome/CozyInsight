import React, { useState } from 'react';
import { Tabs } from 'antd';
import type { TabsProps } from 'antd';

interface DashboardTabConfig {
    id: string;
    name: string;
    componentIds: string[];
}

interface DashboardTabContainerProps {
    tabs: DashboardTabConfig[];
    components: any[];
    onTabChange?: (activeKey: string) => void;
}

const DashboardTabContainer: React.FC<DashboardTabContainerProps> = ({
    tabs,
    components,
    onTabChange,
}) => {
    const [activeKey, setActiveKey] = useState(tabs[0]?.id || '');

    const handleTabChange = (key: string) => {
        setActiveKey(key);
        onTabChange?.(key);
    };

    // 根据Tab配置过滤组件
    const getTabComponents = (componentIds: string[]) => {
        return components.filter(comp => componentIds.includes(comp.id));
    };

    const tabItems: TabsProps['items'] = tabs.map(tab => ({
        key: tab.id,
        label: tab.name,
        children: (
            <div style={{ padding: '16px 0' }}>
                {getTabComponents(tab.componentIds).map(comp => (
                    <div key={comp.id} style={{ marginBottom: 16 }}>
                        {/* 渲染组件 */}
                        <div>{comp.name}</div>
                    </div>
                ))}
            </div>
        ),
    }));

    return (
        <Tabs
            activeKey={activeKey}
            items={tabItems}
            onChange={handleTabChange}
            type="card"
            style={{ width: '100%' }}
        />
    );
};

export default DashboardTabContainer;
