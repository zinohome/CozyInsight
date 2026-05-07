import React, { useEffect } from 'react';
import { Modal, Form, Input, Select, message } from 'antd';
import type { Datasource, CreateDatasourceRequest, UpdateDatasourceRequest } from '../../../types/datasource';
import { createDatasource, updateDatasource } from '../../../api/datasource';

interface DatasourceModalProps {
    open: boolean;
    onCancel: () => void;
    onSuccess: () => void;
    editingDatasource?: Datasource;
}

const DatasourceModal: React.FC<DatasourceModalProps> = ({
    open,
    onCancel,
    onSuccess,
    editingDatasource,
}) => {
    const [form] = Form.useForm();
    const [loading, setLoading] = React.useState(false);

    useEffect(() => {
        if (open) {
            if (editingDatasource) {
                form.setFieldsValue({
                    ...editingDatasource,
                    // configuration is JSON string, might need parsing if we have specific fields
                    // For now, we treat it as a string or handle specific types
                });
            } else {
                form.resetFields();
            }
        }
    }, [open, editingDatasource, form]);

    const handleSubmit = async () => {
        try {
            const values = await form.validateFields();
            setLoading(true);

            // Simple configuration handling for now
            const payload = {
                ...values,
                configuration: values.configuration || '{}',
            };

            if (editingDatasource) {
                await updateDatasource(editingDatasource.id, payload as UpdateDatasourceRequest);
                message.success('更新成功');
            } else {
                await createDatasource(payload as CreateDatasourceRequest);
                message.success('创建成功');
            }
            onSuccess();
        } catch (error: any) {
            // message.error handled by request interceptor usually, but safe to add
        } finally {
            setLoading(false);
        }
    };

    return (
        <Modal
            title={editingDatasource ? '编辑数据源' : '新建数据源'}
            open={open}
            onCancel={onCancel}
            onOk={handleSubmit}
            confirmLoading={loading}
        >
            <Form form={form} layout="vertical">
                <Form.Item
                    name="name"
                    label="名称"
                    rules={[{ required: true, message: '请输入名称' }]}
                >
                    <Input placeholder="请输入数据源名称" />
                </Form.Item>
                <Form.Item
                    name="type"
                    label="类型"
                    rules={[{ required: true, message: '请选择类型' }]}
                >
                    <Select placeholder="请选择数据源类型">
                        <Select.Option value="mysql">MySQL</Select.Option>
                        <Select.Option value="postgresql">PostgreSQL</Select.Option>
                        <Select.Option value="oracle">Oracle</Select.Option>
                        <Select.Option value="sqlserver">SQL Server</Select.Option>
                    </Select>
                </Form.Item>
                <Form.Item name="description" label="描述">
                    <Input.TextArea placeholder="请输入描述" />
                </Form.Item>
                <Form.Item
                    name="configuration"
                    label="配置 (JSON)"
                    rules={[{ required: true, message: '请输入配置' }]}
                    initialValue="{}"
                >
                    <Input.TextArea rows={4} placeholder='{"host": "localhost", ...}' />
                </Form.Item>
            </Form>
        </Modal>
    );
};

export default DatasourceModal;
