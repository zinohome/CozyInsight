import React, { useState, useEffect, useCallback } from 'react';
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

// Mock API for schedule tasks
const scheduleAPI = {
    list: async (): Promise<ScheduleTask[]> => {
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve([
                    { id: '1', name: '每日邮件报告', type: 'email_report', cronExpr: '0 0 * * *', enabled: true, status: 'active', lastRunTime: Date.now() - 3600000, createTime: Date.now() - 86400000 },
                    { id: '2', name: '数据库快照', type: 'snapshot', cronExpr: '0 2 * * *', enabled: false, status: 'inactive', lastRunTime: 0, createTime: Date.now() - 172800000 },
                    { id: '3', name: '数据同步到BI', type: 'data_sync', cronExpr: '0 */6 * * *', enabled: true, status: 'running', lastRunTime: Date.now() - 7200000, createTime: Date.now() - 259200000 },
                ]);
            }, 500);
        });
    },
    create: async (task: Partial<ScheduleTask>): Promise<ScheduleTask> => {
        return new Promise((resolve) => {
            setTimeout(() => {
                const newTask = { ...task, id: String(Date.now()), status: task.enabled ? 'active' : 'inactive', lastRunTime: 0, createTime: Date.now() } as ScheduleTask;
                resolve(newTask);
            }, 500);
        });
    },
    update: async (_id: string, task: Partial<ScheduleTask>): Promise<ScheduleTask> => {
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve({ ...task, id: _id } as ScheduleTask);
            }, 500);
        });
    },
    delete: async (_id: string): Promise<void> => {
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve();
            }, 500);
        });
    },
    enable: async (_id: string): Promise<void> => {
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve();
            }, 500);
        });
    },
    disable: async (_id: string): Promise<void> => {
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve();
            }, 500);
        });
    },
    execute: async (_id: string): Promise<void> => {
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve();
            }, 500);
        });
    },
};

const ScheduleManagement: React.FC = () => {
    const [tasks, setTasks] = useState<ScheduleTask[]>([]);
    const [loading, setLoading] = useState(false);
    const [modalVisible, setModalVisible] = useState(false);
    const [editingTask, setEditingTask] = useState<ScheduleTask | undefined>();
    const [form] = Form.useForm();

    const loadTasks = useCallback(async () => {
        setLoading(true);
        try {
            const data = await scheduleAPI.list();
            setTasks(data);
        } catch (error) {
            message.error('加载任务失败');
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        loadTasks();
    }, [loadTasks]);

    const handleOpenModal = (task?: ScheduleTask) => {
        setEditingTask(task);
        if (task) {
            form.setFieldsValue(task);
        } else {
            form.resetFields();
            form.setFieldsValue({ enabled: true }); // Default to enabled for new tasks
        }
        setModalVisible(true);
    };

    const handleCloseModal = () => {
        setModalVisible(false);
        setEditingTask(undefined);
        form.resetFields();
    };

    const handleSave = async () => {
        try {
            const values = await form.validateFields();
            if (editingTask) {
                await scheduleAPI.update(editingTask.id, values);
                message.success('更新成功');
            } else {
                await scheduleAPI.create(values);
                message.success('创建成功');
            }
            handleCloseModal();
            loadTasks();
        } catch (error) {
            message.error('操作失败');
        }
    };

    const handleDelete = async (taskId: string) => {
        try {
            await scheduleAPI.delete(taskId);
            message.success('删除成功');
            loadTasks();
        } catch (error) {
            message.error('删除失败');
        }
    };

    const handleToggle = async (task: ScheduleTask) => {
        try {
            if (task.enabled) {
                await scheduleAPI.disable(task.id);
            } else {
                await scheduleAPI.enable(task.id);
            }
            message.success('操作成功');
            loadTasks();
        } catch (error) {
            message.error('操作失败');
        }
    };

    const handleExecute = async (taskId: string) => {
        try {
            await scheduleAPI.execute(taskId);
            message.success('任务已执行');
            loadTasks();
        } catch (error) {
            message.error('执行失败');
        }
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
                        onClick={() => handleToggle(record)}
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
