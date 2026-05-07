import React, { useEffect, useState } from 'react';
import { Table, Button, Space, message, Popconfirm, Tag } from 'antd';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';
import { datasourceAPI } from '../../api/datasource';
import type { Datasource } from '../../types/datasource';

const DatasourceList: React.FC = () => {
    const navigate = useNavigate();
    const [data, setData] = useState<Datasource[]>([]);
    const [loading, setLoading] = useState(false);

    const fetchData = async () => {
        try {
            setLoading(true);
            const result = await datasourceAPI.list();
            setData(result);
        } catch (error: any) {
            message.error(error.message || '加载失败');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
    }, []);

    const handleDelete = async (id: string) => {
        try {
            await datasourceAPI.delete(id);
            message.success('删除成功');
            fetchData();
        } catch (error: any) {
            message.error(error.message || '删除失败');
        }
    };

    const handleTest = async (id: string) => {
        try {
            const result = await datasourceAPI.testConnection(id);
            if (result.success) {
                message.success(result.message);
            } else {
                message.error(result.message);
            }
        } catch (error: any) {
            message.error(error.message || '连接测试失败');
        }
    };

    const columns: ColumnsType<Datasource> = [
        {
            title: '名称',
            dataIndex: 'name',
            key: 'name',
        },
        {
            title: '类型',
            dataIndex: 'type',
            key: 'type',
            render: (type: string) => {
                const colorMap: Record<string, string> = {
                    mysql: 'blue',
                    postgresql: 'green',
                    clickhouse: 'purple',
                    oracle: 'orange',
                    sqlserver: 'red',
                };
                return <Tag color={colorMap[type] || 'default'}>{type.toUpperCase()}</Tag>;
            },
        },
        {
            title: '描述',
            dataIndex: 'description',
            key: 'description',
        },
        {
            title: '创建时间',
            dataIndex: 'createTime',
            key: 'createTime',
            render: (time: number) => new Date(time).toLocaleString(),
        },
        {
            title: '操作',
            key: 'action',
            render: (_, record) => (
                <Space size="small">
                    <Button type="link" size="small" onClick={() => handleTest(record.id!)}>
                        测试连接
                    </Button>
                    <Button type="link" size="small" onClick={() => navigate(`/datasource/edit/${record.id}`)}>
                        编辑
                    </Button>
                    <Popconfirm
                        title="确定要删除吗?"
                        onConfirm={() => handleDelete(record.id!)}
                        okText="确定"
                        cancelText="取消"
                    >
                        <Button type="link" size="small" danger>
                            删除
                        </Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <div>
            <div style={{ marginBottom: 16 }}>
                <Button type="primary" onClick={() => navigate('/datasource/create')}>
                    新建数据源
                </Button>
            </div>
            <Table
                columns={columns}
                dataSource={data}
                rowKey="id"
                loading={loading}
            />
        </div>
    );
};

export default DatasourceList;
