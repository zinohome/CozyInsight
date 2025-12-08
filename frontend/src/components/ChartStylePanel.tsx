import React, { useState } from 'react';
import { Card, Form, Select, InputNumber, Switch, ColorPicker, Divider, Space, Tabs } from 'antd';
import type { Color } from 'antd/es/color-picker';

const { Option } = Select;
const { TabPane } = Tabs;

export interface StyleConfig {
    // 颜色配置
    colors?: string[];
    colorTheme?: string;

    // 图例配置
    legend?: {
        show?: boolean;
        position?: 'top' | 'bottom' | 'left' | 'right';
    };

    // 标题配置
    title?: {
        show?: boolean;
        text?: string;
        fontSize?: number;
        color?: string;
    };

    // 坐标轴配置
    axis?: {
        showXAxis?: boolean;
        showYAxis?: boolean;
        xAxisLabel?: string;
        yAxisLabel?: string;
    };

    // 标签配置
    label?: {
        show?: boolean;
        position?: string;
        fontSize?: number;
    };

    // 背景配置
    background?: {
        color?: string;
    };
}

export interface ChartStylePanelProps {
    chartType: string;
    initialStyle?: StyleConfig;
    onChange: (style: StyleConfig) => void;
}

/**
 * 图表样式配置面板
 */
export const ChartStylePanel: React.FC<ChartStylePanelProps> = ({
    chartType,
    initialStyle = {},
    onChange,
}) => {
    const [form] = Form.useForm();
    const [currentStyle, setCurrentStyle] = useState<StyleConfig>(initialStyle);

    const colorThemes = [
        { value: 'default', label: '默认主题', colors: ['#5B8FF9', '#5AD8A6', '#5D7092', '#F6BD16', '#E86452'] },
        { value: 'blue', label: '蓝色系', colors: ['#1890FF', '#36CFC9', '#40A9FF', '#69C0FF', '#91D5FF'] },
        { value: 'green', label: '绿色系', colors: ['#52C41A', '#73D13D', '#95DE64', '#B7EB8F', '#D9F7BE'] },
        { value: 'purple', label: '紫色系', colors: ['#722ED1', '#9254DE', '#B37FEB', '#D3ADF7', '#EFDBFF'] },
        { value: 'orange', label: '橙色系', colors: ['#FA8C16', '#FFA940', '#FFC069', '#FFD591', '#FFE7BA'] },
    ];

    const handleValuesChange = (changedValues: any, allValues: any) => {
        const newStyle: StyleConfig = {
            colorTheme: allValues.colorTheme,
            colors: colorThemes.find(t => t.value === allValues.colorTheme)?.colors,
            legend: {
                show: allValues.legendShow,
                position: allValues.legendPosition,
            },
            title: {
                show: allValues.titleShow,
                text: allValues.titleText,
                fontSize: allValues.titleFontSize,
                color: allValues.titleColor,
            },
            axis: {
                showXAxis: allValues.showXAxis,
                showYAxis: allValues.showYAxis,
                xAxisLabel: allValues.xAxisLabel,
                yAxisLabel: allValues.yAxisLabel,
            },
            label: {
                show: allValues.labelShow,
                position: allValues.labelPosition,
                fontSize: allValues.labelFontSize,
            },
            background: {
                color: allValues.backgroundColor,
            },
        };

        setCurrentStyle(newStyle);
        onChange(newStyle);
    };

    return (
        <Card title="样式配置" size="small">
            <Form
                form={form}
                layout="vertical"
                initialValues={{
                    colorTheme: initialStyle.colorTheme || 'default',
                    legendShow: initialStyle.legend?.show !== false,
                    legendPosition: initialStyle.legend?.position || 'bottom',
                    titleShow: initialStyle.title?.show !== false,
                    titleText: initialStyle.title?.text || '图表标题',
                    titleFontSize: initialStyle.title?.fontSize || 16,
                    titleColor: initialStyle.title?.color || '#000000',
                    showXAxis: initialStyle.axis?.showXAxis !== false,
                    showYAxis: initialStyle.axis?.showYAxis !== false,
                    xAxisLabel: initialStyle.axis?.xAxisLabel || '',
                    yAxisLabel: initialStyle.axis?.yAxisLabel || '',
                    labelShow: initialStyle.label?.show || false,
                    labelPosition: initialStyle.label?.position || 'top',
                    labelFontSize: initialStyle.label?.fontSize || 12,
                    backgroundColor: initialStyle.background?.color || '#ffffff',
                }}
                onValuesChange={handleValuesChange}
            >
                <Tabs defaultActiveKey="color" size="small">
                    <TabPane tab="颜色" key="color">
                        <Form.Item label="配色方案" name="colorTheme">
                            <Select>
                                {colorThemes.map(theme => (
                                    <Option key={theme.value} value={theme.value}>
                                        <Space>
                                            <span>{theme.label}</span>
                                            <div style={{ display: 'flex', gap: 2 }}>
                                                {theme.colors.map((color, i) => (
                                                    <div
                                                        key={i}
                                                        style={{
                                                            width: 12,
                                                            height: 12,
                                                            backgroundColor: color,
                                                            borderRadius: 2,
                                                        }}
                                                    />
                                                ))}
                                            </div>
                                        </Space>
                                    </Option>
                                ))}
                            </Select>
                        </Form.Item>

                        <Form.Item label="背景颜色" name="backgroundColor">
                            <Select>
                                <Option value="#ffffff">白色</Option>
                                <Option value="#f5f5f5">浅灰</Option>
                                <Option value="#fafafa">极浅灰</Option>
                                <Option value="transparent">透明</Option>
                            </Select>
                        </Form.Item>
                    </TabPane>

                    <TabPane tab="图例" key="legend">
                        <Form.Item label="显示图例" name="legendShow" valuePropName="checked">
                            <Switch />
                        </Form.Item>

                        <Form.Item label="图例位置" name="legendPosition">
                            <Select>
                                <Option value="top">顶部</Option>
                                <Option value="bottom">底部</Option>
                                <Option value="left">左侧</Option>
                                <Option value="right">右侧</Option>
                            </Select>
                        </Form.Item>
                    </TabPane>

                    <TabPane tab="标题" key="title">
                        <Form.Item label="显示标题" name="titleShow" valuePropName="checked">
                            <Switch />
                        </Form.Item>

                        <Form.Item label="标题内容" name="titleText">
                            <input
                                type="text"
                                className="ant-input"
                                placeholder="请输入标题"
                            />
                        </Form.Item>

                        <Form.Item label="字体大小" name="titleFontSize">
                            <InputNumber min={12} max={32} style={{ width: '100%' }} />
                        </Form.Item>

                        <Form.Item label="字体颜色" name="titleColor">
                            <Select>
                                <Option value="#000000">黑色</Option>
                                <Option value="#666666">深灰</Option>
                                <Option value="#999999">灰色</Option>
                                <Option value="#1890ff">蓝色</Option>
                            </Select>
                        </Form.Item>
                    </TabPane>

                    {(chartType === 'column' || chartType === 'bar' || chartType === 'line') && (
                        <TabPane tab="坐标轴" key="axis">
                            <Form.Item label="显示 X 轴" name="showXAxis" valuePropName="checked">
                                <Switch />
                            </Form.Item>

                            <Form.Item label="X 轴标签" name="xAxisLabel">
                                <input
                                    type="text"
                                    className="ant-input"
                                    placeholder="X 轴标签"
                                />
                            </Form.Item>

                            <Form.Item label="显示 Y 轴" name="showYAxis" valuePropName="checked">
                                <Switch />
                            </Form.Item>

                            <Form.Item label="Y 轴标签" name="yAxisLabel">
                                <input
                                    type="text"
                                    className="ant-input"
                                    placeholder="Y 轴标签"
                                />
                            </Form.Item>
                        </TabPane>
                    )}

                    <TabPane tab="标签" key="label">
                        <Form.Item label="显示数据标签" name="labelShow" valuePropName="checked">
                            <Switch />
                        </Form.Item>

                        <Form.Item label="标签位置" name="labelPosition">
                            <Select>
                                <Option value="top">顶部</Option>
                                <Option value="middle">中间</Option>
                                <Option value="bottom">底部</Option>
                            </Select>
                        </Form.Item>

                        <Form.Item label="标签字体大小" name="labelFontSize">
                            <InputNumber min={8} max={20} style={{ width: '100%' }} />
                        </Form.Item>
                    </TabPane>
                </Tabs>
            </Form>
        </Card>
    );
};

export default ChartStylePanel;
