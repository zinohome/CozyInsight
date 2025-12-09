import React, { useState, useEffect } from 'react';
import { Table, Button, Modal, Form, Input, Switch, Select, message, Space, Tag, Popconfirm } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, PlayCircleOutlined, PauseCircleOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';

interface ScheduleTask {
    id: string;
    name: string;
    type: string;
    cronExpr: string;
    enabled: boolean;
    status: string;
    lastRunTime: number;
    createTime: number;
}

const ScheduleManagement: React.FC = () => {
    const [tasks, setTasks] = useState<ScheduleTask[]>([]);
    const [loading, setLoading] = useState(false);
    const [modalVisible, setModalVisible] = useState(false);
    const [editingTask, setEditingTask] = useState<ScheduleTask | null>(null);
    const [form] = Form.useForm();

    // Mock数据加载
    const loadTasks = async () => {
        setLoading(true);
        // TODO: 调用API
        setLoading(false);
    };

    useEffect(() => {
        loadTasks();
    }, []);

    const handleOpenModal = (task?: ScheduleTask) => {
        setEditingTask(task || null);
        if (task) {
            form.setFieldsValue(task);
        } else {
            form.resetFields();
        }
        setModalVisible(true);
    };

    const handleSave = async () => {
        try {
            const values = await form.validateFields();
            message.success(editingTask ? '更新成功' : '创建成功');
            setModalVisible(false);
            loadTasks();
        } catch (error: any) {
            message.error(error.message || '保存失败');
        }
    };

    const handleDelete = async (id: string) => {
        message.success('删除成功');
        loadTasks();
    };

    const handleToggle = async (id: string, enabled: boolean) => {
        message.success(enabled ? '已启用' : '已禁用');
        loadTasks();
    };

    const handleExecute = async (id: string) => {
        message.success('任务已执行');
    };

    const columns: ColumnsType<ScheduleTask> = [
        {
            title: '任务名称',
            dataIndex: 'name',
            key: 'name',
        },
        {
            title: 'Cron表达式',
            dataIndex: 'cronExpr',
            key: 'cronExpr',
            render: (text: string) => <code>{text}</code>,
        },
        {
            title: '类型',
            dataIndex: 'type',
            key: 'type',
            render: (type: string) => {
                const typeMap: Record<string, string> = {
                    email_report: '邮件报告',
                    snapshot: '快照',
                    data_sync: '数据同步',
                };
                return typeMap[type] || type;
            },
        },
        {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            render: (status: string) => {
                const statusMap: Record<string, { color: string; text: string }> = {
                    active: { color: 'green', text: '运行中' },
                    inactive: { color: 'default', text: '未启用' },
                    running: { color: 'blue', text: '执行中' },
                };
                const s = statusMap[status] || { color: 'default', text: status };
                return <Tag color={s.color}>{s.text}</Tag>;
            },
        },
        {
            title: '上次执行',
            dataIndex: 'lastRunTime',
            key: 'lastRunTime',
            render: (time: number) => time ? new Date(time).toLocaleString() : '-',
        },
        {
            title: '操作',
            key: 'action',
            render: (_, record: ScheduleTask) => (
                <Space>
                    <Button
                        type="link"
                        icon={record.enabled ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
                        size="small"
                        onClick={() => handleToggle(record.id, !record.enabled)}
                    >
                        {record.enabled ? '禁用' : '启用'}
                    </Button>
                    <Button
                        type="link"
                        icon={<PlayCircleOutlined />}
                        size="small"
                        onClick={() => handleExecute(record.id)}
                    >
                        立即执行
                    </Button>
                    <Button
                        type="link"
                        icon={<EditOutlined />}
                        size="small"
                        onClick={() => handleOpenModal(record)}
                    >
                        编辑
                    </Button>
                    <Popconfirm
                        title="确认删除?"
                        onConfirm={() => handleDelete(record.id)}
                    >
                        <Button
                            type="link"
                            danger
                            icon={<DeleteOutlined />}
                            size="small"
                        >
                            删除
                        </Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <div style={{ padding: 24 }}>
            <div style={{ marginBottom: 16 }}>
                <Button
                    type="primary"
                    icon={<PlusOutlined />}
                    onClick={() => handleOpenModal()}
                >
                    新建任务
                </Button>
            </div>

            <Table
                columns={columns}
                dataSource={tasks}
                loading={loading}
                rowKey="id"
            />

            <Modal
                title={editingTask ? '编辑任务' : '新建任务'}
                open={modalVisible}
                onOk={handleSave}
                onCancel={() => setModalVisible(false)}
                width={600}
            >
                <Form form={form} layout="vertical">
                    <Form.Item
                        label="任务名称"
                        name="name"
                        rules={[{ required: true, message: '请输入任务名称' }]}
                    >
                        <Input placeholder="请输入任务名称" />
                    </Form.Item>

                    <Form.Item
                        label="任务类型"
                        name="type"
                        rules={[{ required: true, message: '请选择任务类型' }]}
                    >
                        <Select placeholder="请选择任务类型">
                            <Select.Option value="email_report">邮件报告</Select.Option>
                            <Select.Option value="snapshot">快照</Select.Option>
                            <Select.Option value="data_sync">数据同步</Select.Option>
                        </Select>
                    </Form.Item>

                    <Form.Item
                        label="Cron表达式"
                        name="cronExpr"
                        rules={[{ required: true, message: '请输入Cron表达式' }]}
                        extra="例如: 0 0 * * * (每天0点执行)"
                    >
                        <Input placeholder="0 0 * * *" />
                    </Form.Item>

                    <Form.Item
                        label="启用"
                        name="enabled"
                        valuePropName="checked"
                    >
                        <Switch />
                    </Form.Item>
                </Form>
            </Modal>
        </div>
    );
};

export default ScheduleManagement;
