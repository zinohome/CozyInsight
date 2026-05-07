import React, { useState } from 'react';
import { Steps, Card, Form, Input, Select, Button, message, Radio, Space } from 'antd';
import { useNavigate } from 'react-router-dom';
import { datasourceAPI } from '../../api/datasource';
import { datasetAPI } from '../../api/dataset';

const { Step } = Steps;
const { Option } = Select;
const { TextArea } = Input;

interface CreateDatasetForm {
    name: string;
    description?: string;
    datasourceId: string;
    type: 'db' | 'sql';
    // db类型
    database?: string;
    tableName?: string;
    // sql类型
    sql?: string;
}

const DatasetCreate: React.FC = () => {
    const navigate = useNavigate();
    const [form] = Form.useForm<CreateDatasetForm>();
    const [current, setCurrent] = useState(0);
    const [datasetType, setDatasetType] = useState<'db' | 'sql'>('db');
    const [datasources, setDatasources] = useState<any[]>([]);
    const [databases, setDatabases] = useState<string[]>([]);
    const [tables, setTables] = useState<string[]>([]);
    const [loading, setLoading] = useState(false);

    React.useEffect(() => {
        fetchDatasources();
    }, []);

    const fetchDatasources = async () => {
        try {
            const result = await datasourceAPI.list();
            setDatasources(result);
        } catch (error: any) {
            message.error(error.message || '获取数据源失败');
        }
    };

    const handleDatasourceChange = async (dsId: string) => {
        try {
            const dbs = await datasourceAPI.getDatabases(dsId);
            setDatabases(dbs);
            setTables([]);
            form.setFieldValue('database', undefined);
            form.setFieldValue('tableName', undefined);
        } catch (error: any) {
            message.error(error.message || '获取数据库列表失败');
        }
    };

    const handleDatabaseChange = async (database: string) => {
        const dsId = form.getFieldValue('datasourceId');
        if (!dsId) return;

        try {
            const tbls = await datasourceAPI.getTables(dsId, database);
            setTables(tbls);
            form.setFieldValue('tableName', undefined);
        } catch (error: any) {
            message.error(error.message || '获取表列表失败');
        }
    };

    const next = async () => {
        try {
            if (current === 0) {
                await form.validateFields(['name', 'datasourceId', 'type']);
            } else if (current === 1) {
                if (datasetType === 'db') {
                    await form.validateFields(['database', 'tableName']);
                } else {
                    await form.validateFields(['sql']);
                }
            }
            setCurrent(current + 1);
        } catch (error) {
            console.error('Validation failed:', error);
        }
    };

    const prev = () => {
        setCurrent(current - 1);
    };

    const handleSubmit = async () => {
        try {
            setLoading(true);
            const values = await form.validateFields();

            const tableData: any = {
                name: values.name,
                datasourceId: values.datasourceId,
                type: values.type,
                datasetGroupId: 'root', // 简化处理,实际应选择分组
            };

            if (values.type === 'db') {
                tableData.tableName = values.tableName;
                tableData.info = JSON.stringify({
                    database: values.database,
                    table: values.tableName,
                });
            } else {
                tableData.tableName = `custom_${Date.now()}`;
                tableData.info = JSON.stringify({
                    sql: values.sql,
                });
            }

            await datasetAPI.createTable(tableData);
            message.success('数据集创建成功');
            navigate('/dataset');
        } catch (error: any) {
            message.error(error.message || '创建失败');
        } finally {
            setLoading(false);
        }
    };

    const steps = [
        {
            title: '基本信息',
            content: (
                <>
                    <Form.Item
                        label="数据集名称"
                        name="name"
                        rules={[{ required: true, message: '请输入数据集名称' }]}
                    >
                        <Input placeholder="例如: 销售数据" />
                    </Form.Item>

                    <Form.Item label="描述" name="description">
                        <TextArea rows={2} placeholder="数据集描述信息" />
                    </Form.Item>

                    <Form.Item
                        label="数据源"
                        name="datasourceId"
                        rules={[{ required: true, message: '请选择数据源' }]}
                    >
                        <Select
                            placeholder="选择数据源"
                            onChange={handleDatasourceChange}
                        >
                            {datasources.map(ds => (
                                <Option key={ds.id} value={ds.id}>
                                    {ds.name} ({ds.type})
                                </Option>
                            ))}
                        </Select>
                    </Form.Item>

                    <Form.Item
                        label="数据集类型"
                        name="type"
                        rules={[{ required: true }]}
                    >
                        <Radio.Group onChange={(e) => setDatasetType(e.target.value)}>
                            <Radio value="db">数据库表</Radio>
                            <Radio value="sql">自定义SQL</Radio>
                        </Radio.Group>
                    </Form.Item>
                </>
            ),
        },
        {
            title: '配置数据',
            content: datasetType === 'db' ? (
                <>
                    <Form.Item
                        label="数据库"
                        name="database"
                        rules={[{ required: true, message: '请选择数据库' }]}
                    >
                        <Select
                            placeholder="选择数据库"
                            onchange={handleDatabaseChange}
                        >
                            {databases.map(db => (
                                <Option key={db} value={db}>{db}</Option>
                            ))}
                        </Select>
                    </Form.Item>

                    <Form.Item
                        label="表名"
                        name="tableName"
                        rules={[{ required: true, message: '请选择表' }]}
                    >
                        <Select placeholder="选择表">
                            {tables.map(tbl => (
                                <Option key={tbl} value={tbl}>{tbl}</Option>
                            ))}
                        </Select>
                    </Form.Item>
                </>
            ) : (
                <Form.Item
                    label="SQL语句"
                    name="sql"
                    rules={[{ required: true, message: '请输入SQL语句' }]}
                >
                    <TextArea
                        rows={10}
                        placeholder="SELECT * FROM table WHERE ..."
                        style={{ fontFamily: 'monospace' }}
                    />
                </Form.Item>
            ),
        },
        {
            title: '完成',
            content: (
                <div style={{ textAlign: 'center', padding: '40px 0' }}>
                    <h3>确认创建数据集</h3>
                    <p>名称: {form.getFieldValue('name')}</p>
                    <p>类型: {datasetType === 'db' ? '数据库表' : '自定义SQL'}</p>
                    {datasetType === 'db' ? (
                        <p>表: {form.getFieldValue('database')}.{form.getFieldValue('tableName')}</p>
                    ) : (
                        <p>SQL: {form.getFieldValue('sql')?.substring(0, 50)}...</p>
                    )}
                </div>
            ),
        },
    ];

    return (
        <Card title="创建数据集">
            <Steps current={current} style={{ marginBottom: 24 }}>
                {steps.map(item => (
                    <Step key={item.title} title={item.title} />
                ))}
            </Steps>

            <Form
                form={form}
                layout="vertical"
                initialValues={{ type: 'db' }}
                style={{ maxWidth: 600, margin: '0 auto' }}
            >
                {steps[current].content}
            </Form>

            <div style={{ marginTop: 24, textAlign: 'center' }}>
                <Space>
                    {current > 0 && (
                        <Button onClick={prev}>上一步</Button>
                    )}
                    {current < steps.length - 1 && (
                        <Button type="primary" onClick={next}>
                            下一步
                        </Button>
                    )}
                    {current === steps.length - 1 && (
                        <Button type="primary" onClick={handleSubmit} loading={loading}>
                            创建
                        </Button>
                    )}
                    <Button onClick={() => navigate('/dataset')}>取消</Button>
                </Space>
            </div>
        </Card>
    );
};

export default DatasetCreate;
