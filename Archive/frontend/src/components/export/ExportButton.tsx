import React from 'react';
import { Dropdown, Button, message } from 'antd';
import { DownloadOutlined } from '@ant-design/icons';
import type { MenuProps } from 'antd';
import { exportAPI } from '../../api/export';

interface ExportButtonProps {
    type: 'dataset' | 'chart';
    id: string;
    disabled?: boolean;
}

const ExportButton: React.FC<ExportButtonProps> = ({ type, id, disabled }) => {
    const handleExport = async (format: 'excel' | 'csv') => {
        try {
            message.loading({ content: '正在导出...', key: 'export' });

            if (type === 'dataset') {
                await exportAPI.exportDataset(id, format);
            } else {
                await exportAPI.exportChartData(id, format);
            }

            message.success({ content: '导出成功', key: 'export' });
        } catch (error) {
            message.error({ content: '导出失败', key: 'export' });
        }
    };

    const menuItems: MenuProps['items'] = [
        {
            key: 'excel',
            label: '导出为 Excel',
            onClick: () => handleExport('excel'),
        },
        {
            key: 'csv',
            label: '导出为 CSV',
            onClick: () => handleExport('csv'),
        },
    ];

    return (
        <Dropdown menu={{ items: menuItems }} disabled={disabled}>
            <Button icon={<DownloadOutlined />} disabled={disabled}>
                导出
            </Button>
        </Dropdown>
    );
};

export default ExportButton;
