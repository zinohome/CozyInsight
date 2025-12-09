import React, { useState, useEffect } from 'react';
import { Table, Button, Modal, Form, Input, message, Space, Tag, Popconfirm } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, UserOutlined } from '@ant-design/icons';
import { roleAPI, type Role } from '../../api/permission';
import type { ColumnsType } from 'antd/es/table';

const RoleManagement: React.FC = () => {
    const [roles, setRoles] = useState<Role[]>([]);
    const [loading, setLoading] = useState(false);
    const [modalVisible, setModalVisible] = useState(false);
    const [editingRole, setEditingRole] = useState<Role | null>(null);
    const [form] = Form.useForm();

    // 加载角色列表
    const loadRoles = async () => {
        setLoading(true);
        try {
            const data = await roleAPI.list();
            setRoles(data);
        } catch (error) {
            message.error('加载角色列表失败');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadRoles();
    }, []);

    // 打开新建/编辑对话框
    const handleOpenModal = (role?: Role) => {
        setEditingRole(role || null);
        if (role) {
            form.setFieldsValue(role);
        } else {
            form.resetFields();
        }
        setModalVisible(true);
    };

    // 保存角色
    const handleSave = async () => {
        try {
            const values = await form.validateFields();

            if (editingRole) {
                await roleAPI.update(editingRole.id, values);
                message.success('更新成功');
            } else {
                await roleAPI.create(values);
                message.success('创建成功');
            }

            setModalVisible(false);
            loadRoles();
        } catch (error: any) {
            message.error(error.message || '保存失败');
        }
    };

    // 删除角色
    const handleDelete = async (id: string) => {
        try {
            await roleAPI.delete(id);
            message.success('删除成功');
            loadRoles();
        } catch (error: any) {
            message.error(error.message || '删除失败');
        }
    };

    const columns: ColumnsType<Role> = [
        {
            title: '角色名称',
            dataIndex: 'name',
            key: 'name',
            render: (text: string, record: Role) => (
                <Space>
                    <span>{text}</span>
                    {record.type === 'system' && <Tag color="blue">系统</Tag>}
                </Space>
            ),
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
            render: (_, record: Role) => (
                <Space>
                    <Button
                        type="link"
                        icon={<UserOutlined />}
                        size="small"
                    >
                        分配用户
                    </Button>
                    <Button
                        type="link"
                        icon={<EditOutlined />}
                        size="small"
                        onClick={() => handleOpenModal(record)}
                        disabled={record.type === 'system'}
                    >
                        编辑
                    </Button>
                    <Popconfirm
                        title="确认删除?"
                        onConfirm={() => handleDelete(record.id)}
                        disabled={record.type === 'system'}
                    >
                        <Button
                            type="link"
                            danger
                            icon={<DeleteOutlined />}
                            size="small"
                            disabled={record.type === 'system'}
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
                    新建角色
                </Button>
            </div>

            <Table
                columns={columns}
                dataSource={roles}
                loading={loading}
                rowKey="id"
            />

            <Modal
                title={editingRole ? '编辑角色' : '新建角色'}
                open={modalVisible}
                onOk={handleSave}
                onCancel={() => setModalVisible(false)}
            >
                <Form form={form} layout="vertical">
                    <Form.Item
                        label="角色名称"
                        name="name"
                        rules={[{ required: true, message: '请输入角色名称' }]}
                    >
                        <Input placeholder="请输入角色名称" />
                    </Form.Item>

                    <Form.Item
                        label="描述"
                        name="description"
                    >
                        <Input.TextArea
                            rows={4}
                            placeholder="请输入角色描述"
                        />
                    </Form.Item>
                </Form>
            </Modal>
        </div>
    );
};

export default RoleManagement;
