import React from 'react';
import RGL, { WidthProvider, Layout } from 'react-grid-layout';
import 'react-grid-layout/css/styles.css';
import 'react-resizable/css/styles.css';
import type { LayoutItem } from './types';

const ReactGridLayout = WidthProvider(RGL);

interface GridLayoutProps {
    layout: LayoutItem[];
    onLayoutChange?: (layout: Layout[]) => void;
    children: React.ReactNode;
    editable?: boolean;
    cols?: number;
    rowHeight?: number;
}

export const GridLayout: React.FC<GridLayoutProps> = ({
    layout,
    onLayoutChange,
    children,
    editable = true,
    cols = 12,
    rowHeight = 30,
}) => {
    return (
        <ReactGridLayout
            className="layout"
            layout={layout}
            cols={cols}
            rowHeight={rowHeight}
            onLayoutChange={onLayoutChange}
            isDraggable={editable}
            isResizable={editable}
            compactType="vertical"
            preventCollision={false}
        >
            {children}
        </ReactGridLayout>
    );
};
