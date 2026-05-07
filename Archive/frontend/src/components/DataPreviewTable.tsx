import React, { useState, useEffect } from 'react';
import { Table, Spin, Alert, Card } from 'antd';
import type { ColumnsType } from 'antd/es/table';

export interface DataPreviewTableProps {
    datasetId: string;
    limit?: number;
}

export interface FieldInfo {
    name: string;
    originName: string;
    type: string;
    deType: number;
    groupType: string;
    description?: string;
    sample?: string;
}

export interface PreviewData {
    fields: FieldInfo[];
    data: any[];
    total: number;
}

/**
 * 数据预览表格组件
 */
export const DataPreviewTable: React.FC<DataPreviewTableProps> = ({
    datasetId,
    limit = 100
}) => {
    const [loading, setLoading] = useState(true);
    const [data, setData] = useState<any[]>([]);
    const [columns, setColumns] = useState<ColumnsType<any>>([]);
    const [error, setError] = useState<string | null>(null);
    const [total, setTotal] = useState(0);

    useEffect(() => {
        fetchPreview();
    }, [datasetId, limit]);

    const fetchPreview = async () => {
        setLoading(true);
        setError(null);

        try {
            // TODO: 替换为实际的 API 调用
            const response = await fetch(
                `/api/v1/dataset/table/${datasetId}/preview?limit=${limit}`,
                {
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('token')}`,
                    },
                }
            );

            if (!response.ok) {
                throw new Error('数据预览失败');
            }

            const result: PreviewData = await response.json();

            // 构建表格列
            const cols: ColumnsType<any> = result.fields.map(field => ({
                title: (
                    <div>
                        <div>{field.name}</div>
                        <div style={{ fontSize: '12px', color: '#999', fontWeight: 'normal' }}>
                            {getTypeLabel(field.deType)} | {field.groupType === 'd' ? '维度' : '指标'}
                        </div>
                    </div>
                ),
                dataIndex: field.name,
                key: field.name,
                width: 150,
                ellipsis: true,
                render: (text: any) => {
                    if (text === null || text === undefined) {
                        return <span style={{ color: '#ccc' }}>NULL</span>;
                    }
                    return String(text);
                },
            }));

            setColumns(cols);
            setData(result.data.map((item, index) => ({ ...item, key: index })));
            setTotal(result.total);
        } catch (err) {
            setError(err instanceof Error ? err.message : '未知错误');
        } finally {
            setLoading(false);
        }
    };

    const getTypeLabel = (deType: number): string => {
        const typeMap: Record<number, string> = {
            0: '文本',
            1: '时间',
            2: '整数',
            3: '小数',
            4: '布尔',
        };
        return typeMap[deType] || '未知';
    };

    if (error) {
        return (
            <Alert
                message="数据加载失败"
                description={error}
                type="error"
                showIcon
                style={{ marginBottom: 16 }}
            />
        );
    }

    return (
        <Card
            title={`数据预览 (共 ${total} 条)`}
            extra={
                <span style={{ fontSize: '14px', color: '#999' }}>
                    显示前 {Math.min(limit, total)} 条
                </span>
            }
        >
            <Table
                columns={columns}
                dataSource={data}
                loading={loading}
                scroll={{ x: 'max-content', y: 500 }}
                pagination={{
                    pageSize: 20,
                    total: total,
                    showTotal: (total) => `共 ${total} 条`,
                    showSizeChanger: true,
                    pageSizeOptions: ['10', '20', '50', '100'],
                }}
                size="small"
            />
        </Card>
    );
};

export default DataPreviewTable;
