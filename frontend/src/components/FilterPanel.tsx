import { useState } from 'react'
import { Button, Select, Input, Space, Tag } from 'antd'
import { PlusOutlined, CloseOutlined } from '@ant-design/icons'

export interface DashboardFilter {
  id: string
  field: string
  operator: string
  value: string
}

interface FilterPanelProps {
  fields: string[]
  filters: DashboardFilter[]
  onChange: (filters: DashboardFilter[]) => void
}

const operators = [
  { value: '=', label: '等于' },
  { value: '!=', label: '不等于' },
  { value: '>', label: '大于' },
  { value: '<', label: '小于' },
  { value: '>=', label: '大于等于' },
  { value: '<=', label: '小于等于' },
  { value: 'LIKE', label: '包含' },
]

export default function FilterPanel({ fields, filters, onChange }: FilterPanelProps) {
  const addFilter = () => {
    onChange([
      ...filters,
      { id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`, field: fields[0] || '', operator: '=', value: '' },
    ])
  }

  const updateFilter = (id: string, updates: Partial<DashboardFilter>) => {
    onChange(filters.map(f => (f.id === id ? { ...f, ...updates } : f)))
  }

  const removeFilter = (id: string) => {
    onChange(filters.filter(f => f.id !== id))
  }

  return (
    <div style={{ padding: '8px 0', borderBottom: '1px solid #f0f0f0', marginBottom: 8 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
        <span style={{ fontWeight: 500, marginRight: 8 }}>筛选条件:</span>
        {filters.map(filter => (
          <Tag
            key={filter.id}
            closable
            onClose={() => removeFilter(filter.id)}
            style={{ display: 'flex', alignItems: 'center', gap: 4 }}
          >
            <Space size={4}>
              <Select
                size="small"
                value={filter.field}
                options={fields.map(f => ({ value: f, label: f }))}
                onChange={(v) => updateFilter(filter.id, { field: v })}
                style={{ width: 120 }}
                variant="borderless"
              />
              <Select
                size="small"
                value={filter.operator}
                options={operators}
                onChange={(v) => updateFilter(filter.id, { operator: v })}
                style={{ width: 100 }}
                variant="borderless"
              />
              <Input
                size="small"
                value={filter.value}
                onChange={(e) => updateFilter(filter.id, { value: e.target.value })}
                placeholder="值"
                style={{ width: 120 }}
                variant="borderless"
              />
            </Space>
          </Tag>
        ))}
        <Button type="dashed" size="small" icon={<PlusOutlined />} onClick={addFilter}>
          添加筛选
        </Button>
      </div>
    </div>
  )
}
