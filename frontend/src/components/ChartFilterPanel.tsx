import React, { useState } from 'react';
import { Card, Form, Select, Input, Button, Space, DatePicker } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import type { DatasetField } from '../types/chart';

const { RangePicker } = DatePicker;

export interface ChartFilter {
    field: string;
    operator: 'eq' | 'ne' | 'gt' | 'lt' | 'gte' | 'lte' | 'in' | 'like' | 'between';
    value: any;
}

export interface ChartFilterPanelProps {
    fields: DatasetField[];
    filters: ChartFilter[];
    onChange: (filters: ChartFilter[]) => void;
}

/**
 * 图表筛选器配置面板
 */
export const ChartFilterPanel: React.FC<ChartFilterPanelProps> = ({
    fields,
    filters,
    onChange,
}) => {
    const [localFilters, setLocalFilters] = useState<ChartFilter[]>(filters);

    const handleAdd = () => {
        const newFilters = [
            ...localFilters,
            { field: '', operator: 'eq' as const, value: '' },
        ];
        setLocalFilters(newFilters);
        onChange(newFilters);
    };

    const handleRemove = (index: number) => {
        const newFilters = localFilters.filter((_, i) => i !== index);
        setLocalFilters(newFilters);
        onChange(newFilters);
    };

    const handleChange = (index: number, key: keyof ChartFilter, value: any) => {
        const newFilters = [...localFilters];
        newFilters[index] = { ...newFilters[index], [key]: value };
        setLocalFilters(newFilters);
        onChange(newFilters);
    };

    const getOperatorOptions = (fieldType?: string) => {
        const commonOps = [
            { value: 'eq', label: '等于' },
            { value: 'ne', label: '不等于' },
        ];

        if (fieldType === 'number' || fieldType === 'INTEGER' || fieldType === 'FLOAT') {
            return [
                ...commonOps,
                { value: 'gt', label: '大于' },
                { value: 'lt', label: '小于' },
                { value: 'gte', label: '大于等于' },
                { value: 'lte', label: '小于等于' },
                { value: 'between', label: '介于' },
            ];
        }

        if (fieldType === 'string' || fieldType === 'TEXT') {
            return [
                ...commonOps,
                { value: 'like', label: '包含' },
                { value: 'in', label: '在列表中' },
            ];
        }

        if (fieldType === 'time' || fieldType === 'TIME') {
            return [
                ...commonOps,
                { value: 'gt', label: '晚于' },
                { value: 'lt', label: '早于' },
                { value: 'between', label: '时间范围' },
            ];
        }

        return commonOps;
    };

    const renderValueInput = (filter: ChartFilter, index: number) => {
        const field = fields.find((f) => f.name === filter.field);

        if (filter.operator === 'in') {
            return (
                <Select
                    mode="tags"
                    style={{ width: '100%' }}
                    placeholder="输入值后按回车"
                    value={Array.isArray(filter.value) ? filter.value : []}
                    onChange={(value) => handleChange(index, 'value', value)}
                />
            );
        }

        if (filter.operator === 'between') {
            if (field?.deType === 1) {
                // 时间类型
                return (
                    <RangePicker
                        style={{ width: '100%' }}
                        onChange={(dates) => handleChange(index, 'value', dates)}
                    />
                );
            }
            // 数值类型
            return (
                <Space.Compact style={{ width: '100%' }}>
                    <Input
                        placeholder="最小值"
                        onChange={(e) =>
                            handleChange(index, 'value', [
                                e.target.value,
                                Array.isArray(filter.value) ? filter.value[1] : '',
                            ])
                        }
                    />
                    <Input
                        placeholder="最大值"
                        onChange={(e) =>
                            handleChange(index, 'value', [
                                Array.isArray(filter.value) ? filter.value[0] : '',
                                e.target.value,
                            ])
                        }
                    />
                </Space.Compact>
            );
        }

        return (
            <Input
                placeholder="筛选值"
                value={filter.value}
                onChange={(e) => handleChange(index, 'value', e.target.value)}
            />
        );
    };

    return (
        <Card title="筛选条件" size="small">
            <Space direction="vertical" style={{ width: '100%' }}>
                {localFilters.map((filter, index) => {
                    const field = fields.find((f) => f.name === filter.field);
                    const fieldType = field?.deType;

                    return (
                        <Space key={index} style={{ width: '100%' }} align="start">
                            <Select
                                style={{ width: 150 }}
                                placeholder="选择字段"
                                value={filter.field || undefined}
                                onChange={(value) => handleChange(index, 'field', value)}
                            >
                                {fields.map((f) => (
                                    <Select.Option key={f.name} value={f.name}>
                                        {f.name}
                                    </Select.Option>
                                ))}
                            </Select>

                            <Select
                                style={{ width: 120 }}
                                placeholder="条件"
                                value={filter.operator}
                                onChange={(value) => handleChange(index, 'operator', value)}
                            >
                                {getOperatorOptions(
                                    fieldType === 0 ? 'string' :
                                        fieldType === 1 ? 'time' :
                                            fieldType === 2 || fieldType === 3 ? 'number' : 'string'
                                ).map((op) => (
                                    <Select.Option key={op.value} value={op.value}>
                                        {op.label}
                                    </Select.Option>
                                ))}
                            </Select>

                            <div style={{ flex: 1 }}>{renderValueInput(filter, index)}</div>

                            <Button
                                danger
                                icon={<DeleteOutlined />}
                                onClick={() => handleRemove(index)}
                            />
                        </Space>
                    );
                })}

                <Button
                    type="dashed"
                    icon={<PlusOutlined />}
                    onClick={handleAdd}
                    block
                >
                    添加筛选条件
                </Button>
            </Space>
        </Card>
    );
};

export default ChartFilterPanel;
