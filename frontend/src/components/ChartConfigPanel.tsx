import React, { useState } from 'react';
import { Card, Form, Select, Button, Space, Divider } from 'antd';
import type { ChartConfig } from '../types/chart';

export interface ChartConfigPanelProps {
    fields: Array<{
        name: string;
        type: string;
        deType: number;
        groupType: string;
    }>;
    onConfigChange: (config: Partial<ChartConfig>) => void;
    initialConfig?: Partial<ChartConfig>;
}

const { Option } = Select;

/**
 * 图表配置面板
 */
export const ChartConfigPanel: React.FC<ChartConfigPanelProps> = ({
    fields,
    onConfigChange,
    initialConfig = {},
}) => {
    const [form] = Form.useForm();
    const [chartType, setChartType] = useState(initialConfig.type || 'column');

    // 获取维度字段（用于 X 轴）
    const dimensionFields = fields.filter(f => f.groupType === 'd');

    // 获取指标字段（用于 Y 轴）
    const measureFields = fields.filter(f => f.groupType === 'q');

    const handleValuesChange = (changedValues: any, allValues: any) => {
        const config: Partial<ChartConfig> = {
            type: allValues.chartType,
            xField: allValues.xField,
            yField: allValues.yField,
            seriesField: allValues.seriesField,
        };

        onConfigChange(config);
    };

    const chartTypes = [
        { value: 'column', label: '柱状图', icon: '📊' },
        { value: 'bar', label: '条形图', icon: '📈' },
        { value: 'line', label: '折线图', icon: '📉' },
        { value: 'area', label: '面积图', icon: '📊' },
        { value: 'pie', label: '饼图', icon: '🥧' },
        { value: 'scatter', label: '散点图', icon: '⚫' },
        { value: 'radar', label: '雷达图', icon: '🕸️' },
    ];

    return (
        <Card title="图表配置" size="small">
            <Form
                form={form}
                layout="vertical"
                initialValues={{
                    chartType: initialConfig.type || 'column',
                    xField: initialConfig.xField,
                    yField: initialConfig.yField,
                    seriesField: initialConfig.seriesField,
                }}
                onValuesChange={handleValuesChange}
            >
                <Form.Item
                    label="图表类型"
                    name="chartType"
                    rules={[{ required: true, message: '请选择图表类型' }]}
                >
                    <Select
                        placeholder="选择图表类型"
                        onChange={(value) => setChartType(value)}
                    >
                        {chartTypes.map(type => (
                            <Option key={type.value} value={type.value}>
                                <Space>
                                    <span>{type.icon}</span>
                                    <span>{type.label}</span>
                                </Space>
                            </Option>
                        ))}
                    </Select>
                </Form.Item>

                <Divider />

                {(chartType === 'column' || chartType === 'bar' || chartType === 'line') && (
                    <>
                        <Form.Item
                            label="X 轴字段（维度）"
                            name="xField"
                            rules={[{ required: true, message: '请选择 X 轴字段' }]}
                        >
                            <Select placeholder="选择维度字段">
                                {dimensionFields.map(field => (
                                    <Option key={field.name} value={field.name}>
                                        {field.name} ({field.type})
                                    </Option>
                                ))}
                            </Select>
                        </Form.Item>

                        <Form.Item
                            label="Y 轴字段（指标）"
                            name="yField"
                            rules={[{ required: true, message: '请选择 Y 轴字段' }]}
                        >
                            <Select placeholder="选择指标字段">
                                {measureFields.map(field => (
                                    <Option key={field.name} value={field.name}>
                                        {field.name} ({field.type})
                                    </Option>
                                ))}
                            </Select>
                        </Form.Item>

                        {chartType === 'line' && (
                            <Form.Item
                                label="分组字段（可选）"
                                name="seriesField"
                            >
                                <Select placeholder="选择分组字段" allowClear>
                                    {dimensionFields.map(field => (
                                        <Option key={field.name} value={field.name}>
                                            {field.name} ({field.type})
                                        </Option>
                                    ))}
                                </Select>
                            </Form.Item>
                        )}
                    </>
                )}

                {chartType === 'pie' && (
                    <>
                        <Form.Item
                            label="分类字段"
                            name="angleField"
                            rules={[{ required: true, message: '请选择分类字段' }]}
                        >
                            <Select placeholder="选择分类字段">
                                {dimensionFields.map(field => (
                                    <Option key={field.name} value={field.name}>
                                        {field.name} ({field.type})
                                    </Option>
                                ))}
                            </Select>
                        </Form.Item>

                        <Form.Item
                            label="数值字段"
                            name="colorField"
                            rules={[{ required: true, message: '请选择数值字段' }]}
                        >
                            <Select placeholder="选择数值字段">
                                {measureFields.map(field => (
                                    <Option key={field.name} value={field.name}>
                                        {field.name} ({field.type})
                                    </Option>
                                ))}
                            </Select>
                        </Form.Item>
                    </>
                )}
            </Form>
        </Card>
    );
};

export default ChartConfigPanel;
