import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
    Form,
    Input,
    Button,
    Card,
    Select,
    Row,
    Col,
    message,
    Space,
    Spin,
} from 'antd';
import { SaveOutlined, ArrowLeftOutlined } from '@ant-design/icons';
import { chartAPI } from '../../api/chart';
import { datasetAPI } from '../../api/dataset';
import { ChartRenderer } from '../../components/chart/ChartRenderer';
import type { CreateChartRequest, UpdateChartRequest, ChartType, ChartConfig } from '../../types/chart';
import type { DatasetTable } from '../../types/dataset';

const { Option } = Select;

const ChartEditor = () => {
    const navigate = useNavigate();
    const { id } = useParams<{ id: string }>();
    const [form] = Form.useForm();
    const [loading, setLoading] = useState(false);
    const [saving, setSaving] = useState(false);
    const [datasets, setDatasets] = useState<DatasetTable[]>([]);

    // 预览状态
    const [previewType, setPreviewType] = useState<ChartType>('bar');
    const [previewData, setPreviewData] = useState<any[]>([]);

    const isEditMode = !!id;

    // 图表类型选项
    const chartTypes: { value: ChartType; label: string }[] = [
        { value: 'bar', label: '柱状图' },
        { value: 'line', label: '折线图' },
        { value: 'pie', label: '饼图' },
        { value: 'table', label: '表格' },
    ];

    // 加载数据集列表
    const loadDatasets = async () => {
        try {
            const data = await datasetAPI.listTables();
            setDatasets(data);
        } catch (error: any) {
            message.error(error.message || '加载数据集列表失败');
        }
    };

    // 生成 Mock 数据
    const generateMockData = (type: ChartType) => {
        const mockData = [
            { type: '分类一', value: 27 },
            { type: '分类二', value: 25 },
            { type: '分类三', value: 18 },
            { type: '分类四', value: 15 },
            { type: '分类五', value: 10 },
            { type: '其他', value: 5 },
        ];

        if (type === 'line') {
            return [
                { x: '2021-01', y: 10 },
                { x: '2021-02', y: 15 },
                { x: '2021-03', y: 8 },
                { x: '2021-04', y: 20 },
                { x: '2021-05', y: 25 },
            ];
        }

        return mockData;
    };

    // 加载图表真实数据
    const loadChartData = async (chartId: string) => {
        try {
            const data = await chartAPI.getData(chartId);
            setPreviewData(data);
        } catch (error: any) {
            console.error('Failed to load chart data:', error);
            // 如果加载失败，使用 Mock 数据
            setPreviewData(generateMockData(previewType));
        }
    };

    // 加载图表配置（编辑模式）
    const loadChart = async (chartId: string) => {
        try {
            setLoading(true);
            const data = await chartAPI.get(chartId);
            form.setFieldsValue({
                name: data.name,
                title: data.title,
                type: data.type,
                tableId: data.tableId,
                sceneId: data.sceneId,
            });
            setPreviewType(data.type);

            // 加载真实数据
            await loadChartData(chartId);
        } catch (error: any) {
            message.error(error.message || '加载图表失败');
            navigate('/chart');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadDatasets();
        if (isEditMode && id) {
            loadChart(id);
        } else {
            // 默认预览数据
            setPreviewData(generateMockData('bar'));
        }
    }, [id]);

    const handleValuesChange = (changedValues: any) => {
        if (changedValues.type) {
            setPreviewType(changedValues.type);
            setPreviewData(generateMockData(changedValues.type));
        }
    };

    const handleSave = async (values: any) => {
        try {
            setSaving(true);
            const data = {
                ...values,
                // 暂时使用默认配置
                xAxis: JSON.stringify({ fields: [{ name: 'type' }] }),
                yAxis: JSON.stringify({ fields: [{ name: 'value' }] }),
            };

            if (isEditMode) {
                await chartAPI.update(id as string, data as UpdateChartRequest);
                message.success('更新成功');
            } else {
                await chartAPI.create(data as CreateChartRequest);
                message.success('创建成功');
            }
            navigate('/chart');
        } catch (error: any) {
            message.error(error.message || '保存失败');
        } finally {
            setSaving(false);
        }
    };

    // 构造预览配置
    const previewConfig: ChartConfig = {
        xAxis: { fields: [{ id: '1', name: previewType === 'line' ? 'x' : 'type' }] },
        yAxis: { fields: [{ id: '2', name: previewType === 'line' ? 'y' : 'value' }] },
    };

    if (loading) {
        return (
            <div style={{ textAlign: 'center', padding: '100px' }}>
                <Spin size="large" tip="加载中..." />
            </div>
        );
    }

    return (
        <div style={{ padding: '24px' }}>
            <Row gutter={24}>
                {/* 左侧编辑区 */}
                <Col span={8}>
                    <Card
                        title={
                            <Space>
                                <Button
                                    type="text"
                                    icon={<ArrowLeftOutlined />}
                                    onClick={() => navigate('/chart')}
                                />
                                <span>{isEditMode ? '编辑图表' : '新建图表'}</span>
                            </Space>
                        }
                    >
                        <Form
                            form={form}
                            layout="vertical"
                            onFinish={handleSave}
                            onValuesChange={handleValuesChange}
                            initialValues={{ type: 'bar' }}
                        >
                            <Form.Item
                                label="名称"
                                name="name"
                                rules={[{ required: true, message: '请输入名称' }]}
                            >
                                <Input placeholder="请输入图表名称" />
                            </Form.Item>

                            <Form.Item label="标题" name="title">
                                <Input placeholder="请输入图表标题" />
                            </Form.Item>

                            <Form.Item
                                label="图表类型"
                                name="type"
                                rules={[{ required: true, message: '请选择图表类型' }]}
                            >
                                <Select>
                                    {chartTypes.map((type) => (
                                        <Option key={type.value} value={type.value}>
                                            {type.label}
                                        </Option>
                                    ))}
                                </Select>
                            </Form.Item>

                            <Form.Item
                                label="数据集"
                                name="tableId"
                                rules={[{ required: true, message: '请选择数据集' }]}
                            >
                                <Select
                                    placeholder="请选择数据集"
                                    showSearch
                                    optionFilterProp="children"
                                >
                                    {datasets.map((dataset) => (
                                        <Option key={dataset.id} value={dataset.id}>
                                            {dataset.name}
                                        </Option>
                                    ))}
                                </Select>
                            </Form.Item>

                            <Form.Item label="所属场景/仪表板" name="sceneId">
                                <Input placeholder="请输入场景ID" />
                            </Form.Item>

                            <Form.Item>
                                <Space>
                                    <Button
                                        type="primary"
                                        htmlType="submit"
                                        icon={<SaveOutlined />}
                                        loading={saving}
                                    >
                                        保存
                                    </Button>
                                    <Button onClick={() => navigate('/chart')}>取消</Button>
                                </Space>
                            </Form.Item>
                        </Form>
                    </Card>
                </Col>

                {/* 右侧预览区 */}
                <Col span={16}>
                    <Card title="图表预览" style={{ height: '100%', minHeight: 600 }}>
                        <div style={{ height: 500 }}>
                            <ChartRenderer
                                type={previewType}
                                data={previewData}
                                config={previewConfig}
                            />
                        </div>
                    </Card>
                </Col>
            </Row>
        </div>
    );
};

export default ChartEditor;
