import React, { useState } from 'react';
import { Form, Input, Select, Button, message, Card } from 'antd';
import { useNavigate } from 'react-router-dom';
import { datasourceAPI } from '../../api/datasource';

const { Option } = Select;
const { TextArea } = Input;

interface FormValues {
    name: string;
    description: string;
    type: string;
    configuration: {
        host: string;
        port: number;
        username: string;
        password: string;
        database: string;
        charset?: string;
    };
}

const DatasourceCreate: React.FC = () => {
    const [form] = Form.useForm<FormValues>();
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false);
    const [testing, setTesting] = useState(false);
    const [dsType, setDsType] = useState<string>('mysql');

    const handleTest = async () => {
        try {
            await form.validateFields();
            const values = form.getFieldsValue();
            setTesting(true);

            const config = JSON.stringify(values.configuration);
            const result = await datasourceAPI.testConnectionByConfig(config);

            if (result.success) {
                message.success(result.message);
            } else {
                message.error(result.message);
            }
        } catch (error: any) {
            message.error(error.message || '测试连接失败');
        } finally {
            setTesting(false);
        }
    };

    const handleSubmit = async (values: FormValues) => {
        try {
            setLoading(true);

            const datasource = {
                name: values.name,
                description: values.description,
                type: values.type,
                configuration: JSON.stringify(values.configuration),
            };

            await datasourceAPI.create(datasource);
            message.success('数据源创建成功');
            navigate('/datasource');
        } catch (error: any) {
            message.error(error.message || '创建失败');
        } finally {
            setLoading(false);
        }
    };

    const getDefaultPort = (type: string): number => {
        switch (type) {
            case 'mysql':
                return 3306;
            case 'postgresql':
                return 5432;
            case 'clickhouse':
                return 8123;
            case 'oracle':
                return 1521;
            case 'sqlserver':
                return 1433;
            default:
                return 3306;
        }
    };

    const handleTypeChange = (value: string) => {
        setDsType(value);
        form.setFieldValue(['configuration', 'port'], getDefaultPort(value));
        if (value === 'mysql') {
            form.setFieldValue(['configuration', 'charset'], 'utf8mb4');
        }
    };

    return (
        <Card title="创建数据源" extra={
            <Button onClick={() => navigate('/datasource')}>返回</Button>
        }>
            <Form
                form={form}
                layout="vertical"
                onFinish={handleSubmit}
                initialValues={{
                    type: 'mysql',
                    configuration: {
                        port: 3306,
                        charset: 'utf8mb4',
                    },
                }}
                style={{ maxWidth: 600 }}
            >
                <Form.Item
                    label="数据源名称"
                    name="name"
                    rules={[{ required: true, message: '请输入数据源名称' }]}
                >
                    <Input placeholder="例如: 生产环境MySQL" />
                </Form.Item>

                <Form.Item
                    label="描述"
                    name="description"
                >
                    <TextArea rows={2} placeholder="数据源描述信息" />
                </Form.Item>

                <Form.Item
                    label="数据源类型"
                    name="type"
                    rules={[{ required: true, message: '请选择数据源类型' }]}
                >
                    <Select onChange={handleTypeChange}>
                        <Option value="mysql">MySQL</Option>
                        <Option value="postgresql">PostgreSQL</Option>
                        <Option value="clickhouse">ClickHouse</Option>
                        <Option value="oracle">Oracle</Option>
                        <Option value="sqlserver">SQL Server</Option>
                    </Select>
                </Form.Item>

                <Form.Item
                    label="主机地址"
                    name={['configuration', 'host']}
                    rules={[{ required: true, message: '请输入主机地址' }]}
                >
                    <Input placeholder="例如: localhost 或 192.168.1.100" />
                </Form.Item>

                <Form.Item
                    label="端口"
                    name={['configuration', 'port']}
                    rules={[{ required: true, message: '请输入端口' }]}
                >
                    <Input type="number" placeholder="默认端口" />
                </Form.Item>

                <Form.Item
                    label="用户名"
                    name={['configuration', 'username']}
                    rules={[{ required: true, message: '请输入用户名' }]}
                >
                    <Input placeholder="数据库用户名" />
                </Form.Item>

                <Form.Item
                    label="密码"
                    name={['configuration', 'password']}
                    rules={[{ required: true, message: '请输入密码' }]}
                >
                    <Input.Password placeholder="数据库密码" />
                </Form.Item>

                <Form.Item
                    label="数据库名"
                    name={['configuration', 'database']}
                    rules={[{ required: true, message: '请输入数据库名' }]}
                >
                    <Input placeholder="例如: mydb" />
                </Form.Item>

                {dsType === 'mysql' && (
                    <Form.Item
                        label="字符集"
                        name={['configuration', 'charset']}
                    >
                        <Select>
                            <Option value="utf8mb4">utf8mb4</Option>
                            <Option value="utf8">utf8</Option>
                            <Option value="latin1">latin1</Option>
                        </Select>
                    </Form.Item>
                )}

                <Form.Item>
                    <Button
                        onClick={handleTest}
                        loading={testing}
                        style={{ marginRight: 8 }}
                    >
                        测试连接
                    </Button>
                    <Button
                        type="primary"
                        htmlType="submit"
                        loading={loading}
                    >
                        创建
                    </Button>
                </Form.Item>
            </Form>
        </Card>
    );
};

export default DatasourceCreate;
