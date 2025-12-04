import React, { useEffect, useState } from 'react';
import { Table, Button, Space, Card, Popconfirm, message } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import type { Datasource } from '../../types/datasource';
import { getDatasourceList, deleteDatasource } from '../../api/datasource';
import DatasourceModal from './components/DatasourceModal';
import dayjs from 'dayjs';

const DatasourceList: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const [data, setData] = useState<Datasource[]>([]);
    const [modalOpen, setModalOpen] = useState(false);
    const [editingDatasource, setEditingDatasource] = useState<Datasource | undefined>();

    const fetchData = async () => {
        try {
            setLoading(true);
            const list = await getDatasourceList();
            setData(list);
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
    }, []);

    const handleDelete = async (id: string) => {
        try {
            await deleteDatasource(id);
            message.success('删除成功');
            fetchData();
        } catch (error) {
            console.error(error);
        }
    };

    const handleEdit = (record: Datasource) => {
        setEditingDatasource(record);
        setModalOpen(true);
    };

    const handleCreate = () => {
        setEditingDatasource(undefined);
        setModalOpen(true);
    };

    const columns = [
        {
            title: '名称',
            dataIndex: 'name',
            key: 'name',
        },
        {
            title: '类型',
            dataIndex: 'type',
            key: 'type',
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
            render: (text: number) => text ? dayjs(text).format('YYYY-MM-DD HH:mm:ss') : '-',
        },
        {
            title: '操作',
            key: 'action',
            render: (_: any, record: Datasource) => (
                <Space size="middle">
                    <Button type="link" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
                        编辑
                    </Button>
                    <Popconfirm
                        title="确定删除吗?"
                        onConfirm={() => handleDelete(record.id)}
                        okText="确定"
                        cancelText="取消"
                    >
                        <Button type="link" danger icon={<DeleteOutlined />}>
                            删除
                        </Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <Card
            title="数据源管理"
            extra={
                <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
                    新建数据源
                </Button>
            }
        >
            <Table
                columns={columns}
                dataSource={data}
                rowKey="id"
                loading={loading}
            />
            <DatasourceModal
                open={modalOpen}
                editingDatasource={editingDatasource}
                onCancel={() => setModalOpen(false)}
                onSuccess={() => {
                    setModalOpen(false);
                    fetchData();
                }}
            />
        </Card>
    );
};

export default DatasourceList;
