import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Table,
    Button,
    Space,
    message,
    Popconfirm,
    Card,
    Breadcrumb,
    Modal,
    Form,
    Input,
    Radio,
} from 'antd';
import {
    PlusOutlined,
    EditOutlined,
    DeleteOutlined,
    FolderOutlined,
    FolderAddOutlined,
    DashboardOutlined,
    FundProjectionScreenOutlined,
    HomeOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { dashboardAPI } from '../../api/dashboard';
import type { Dashboard, CreateDashboardRequest } from '../../types/dashboard';

interface BreadcrumbItem {
    id: string;
    name: string;
}

const DashboardList = () => {
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false);
    const [dashboards, setDashboards] = useState<Dashboard[]>([]);
    const [currentPid, setCurrentPid] = useState('0');
    const [breadcrumbs, setBreadcrumbs] = useState<BreadcrumbItem[]>([
        { id: '0', name: '根目录' },
    ]);

    // 模态框状态
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [modalType, setModalType] = useState<'folder' | 'dashboard'>('dashboard');
    const [form] = Form.useForm();
    const [submitting, setSubmitting] = useState(false);

    // 加载仪表板列表
    const loadDashboards = async (pid: string) => {
        try {
            setLoading(true);
            const data = await dashboardAPI.list({ pid });
            setDashboards(data);
        } catch (error: any) {
            message.error(error.message || '加载列表失败');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadDashboards(currentPid);
    }, [currentPid]);

    // 处理文件夹点击
    const handleFolderClick = (record: Dashboard) => {
        setCurrentPid(record.id);
        setBreadcrumbs([...breadcrumbs, { id: record.id, name: record.name }]);
    };

    // 处理面包屑点击
    const handleBreadcrumbClick = (item: BreadcrumbItem, index: number) => {
        setCurrentPid(item.id);
        setBreadcrumbs(breadcrumbs.slice(0, index + 1));
    };

    // 删除仪表板/文件夹
    const handleDelete = async (id: string) => {
        try {
            await dashboardAPI.delete(id);
            message.success('删除成功');
            loadDashboards(currentPid);
        } catch (error: any) {
            message.error(error.message || '删除失败');
        }
    };

    // 打开创建模态框
    const showCreateModal = (type: 'folder' | 'dashboard') => {
        setModalType(type);
        form.resetFields();
        form.setFieldsValue({
            nodeType: type,
            type: type === 'dashboard' ? 'dashboard' : undefined,
        });
        setIsModalOpen(true);
    };

    // 提交创建
    const handleCreate = async () => {
        try {
            const values = await form.validateFields();
            setSubmitting(true);

            const data: CreateDashboardRequest = {
                name: values.name,
                pid: currentPid,
                nodeType: modalType,
                type: values.type,
            };

            await dashboardAPI.create(data);
            message.success('创建成功');
            setIsModalOpen(false);
            loadDashboards(currentPid);
        } catch (error: any) {
            message.error(error.message || '创建失败');
        } finally {
            setSubmitting(false);
        }
    };

    // 表格列定义
    const columns: ColumnsType<Dashboard> = [
        {
            title: '名称',
            dataIndex: 'name',
            key: 'name',
            render: (text: string, record: Dashboard) => (
                <Space>
                    {record.nodeType === 'folder' ? (
                        <FolderOutlined style={{ color: '#1890ff' }} />
                    ) : record.type === 'dataV' ? (
                        <FundProjectionScreenOutlined style={{ color: '#722ed1' }} />
                    ) : (
                        <DashboardOutlined style={{ color: '#52c41a' }} />
                    )}
                    {record.nodeType === 'folder' ? (
                        <a onClick={() => handleFolderClick(record)}>{text}</a>
                    ) : (
                        <span>{text}</span>
                    )}
                </Space>
            ),
        },
        {
            title: '类型',
            dataIndex: 'type',
            key: 'type',
            width: 120,
            render: (text: string, record: Dashboard) => {
                if (record.nodeType === 'folder') return '文件夹';
                return text === 'dataV' ? '数据大屏' : '仪表板';
            },
        },
        {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            width: 100,
            render: (status: number) => (status === 1 ? '已发布' : '未发布'),
        },
        {
            title: '创建时间',
            dataIndex: 'createTime',
            key: 'createTime',
            width: 180,
            render: (time: number) =>
                time ? new Date(time).toLocaleString('zh-CN') : '-',
        },
        {
            title: '操作',
            key: 'action',
            width: 200,
            render: (_, record) => (
                <Space>
                    {record.nodeType === 'dashboard' && (
                        <Button
                            type="link"
                            size="small"
                            icon={<EditOutlined />}
                            onClick={() => navigate(`/dashboard/edit/${record.id}`)}
                        >
                            编辑
                        </Button>
                    )}
                    <Button
                        type="link"
                        size="small"
                        icon={<EditOutlined />}
                        onClick={() => {
                            // TODO: 重命名功能
                            message.info('重命名功能开发中');
                        }}
                    >
                        重命名
                    </Button>
                    <Popconfirm
                        title="确认删除"
                        description={`确定要删除这个${record.nodeType === 'folder' ? '文件夹' : '仪表板'
                            }吗？`}
                        onConfirm={() => handleDelete(record.id)}
                        okText="确定"
                        cancelText="取消"
                    >
                        <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                            删除
                        </Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <div style={{ padding: '24px' }}>
            <Card>
                {/* 工具栏 */}
                <div
                    style={{
                        marginBottom: 16,
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                    }}
                >
                    {/* 面包屑导航 */}
                    <Breadcrumb>
                        {breadcrumbs.map((item, index) => (
                            <Breadcrumb.Item key={item.id}>
                                {index === breadcrumbs.length - 1 ? (
                                    item.name === '根目录' ? <HomeOutlined /> : item.name
                                ) : (
                                    <a onClick={() => handleBreadcrumbClick(item, index)}>
                                        {item.name === '根目录' ? <HomeOutlined /> : item.name}
                                    </a>
                                )}
                            </Breadcrumb.Item>
                        ))}
                    </Breadcrumb>

                    <Space>
                        <Button
                            icon={<FolderAddOutlined />}
                            onClick={() => showCreateModal('folder')}
                        >
                            新建文件夹
                        </Button>
                        <Button
                            type="primary"
                            icon={<PlusOutlined />}
                            onClick={() => showCreateModal('dashboard')}
                        >
                            新建仪表板
                        </Button>
                    </Space>
                </div>

                {/* 列表 */}
                <Table
                    columns={columns}
                    dataSource={dashboards}
                    rowKey="id"
                    loading={loading}
                    pagination={false}
                />
            </Card>

            {/* 创建模态框 */}
            <Modal
                title={modalType === 'folder' ? '新建文件夹' : '新建仪表板'}
                open={isModalOpen}
                onOk={handleCreate}
                onCancel={() => setIsModalOpen(false)}
                confirmLoading={submitting}
            >
                <Form form={form} layout="vertical">
                    <Form.Item
                        label="名称"
                        name="name"
                        rules={[{ required: true, message: '请输入名称' }]}
                    >
                        <Input placeholder="请输入名称" />
                    </Form.Item>

                    {modalType === 'dashboard' && (
                        <Form.Item
                            label="类型"
                            name="type"
                            rules={[{ required: true, message: '请选择类型' }]}
                            initialValue="dashboard"
                        >
                            <Radio.Group>
                                <Radio value="dashboard">仪表板</Radio>
                                <Radio value="dataV">数据大屏</Radio>
                            </Radio.Group>
                        </Form.Item>
                    )}
                </Form>
            </Modal>
        </div>
    );
};

export default DashboardList;
