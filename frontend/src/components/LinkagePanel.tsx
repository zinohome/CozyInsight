import { Select, Button, Space, Card } from 'antd'
import type { ChartLinkageRule } from '@/types/chart'

interface LinkagePanelProps {
  charts: Array<{ id: number; title: string }>
  rules: ChartLinkageRule[]
  onChange: (rules: ChartLinkageRule[]) => void
}

export default function LinkagePanel({ charts, rules, onChange }: LinkagePanelProps) {
  const addRule = () => {
    onChange([...rules, { sourceChartId: 0, targetChartId: 0, sourceField: '', targetField: '' }])
  }

  const removeRule = (index: number) => {
    onChange(rules.filter((_, i) => i !== index))
  }

  const updateRule = (index: number, patch: Partial<ChartLinkageRule>) => {
    const next = rules.map((r, i) => (i === index ? { ...r, ...patch } : r))
    onChange(next)
  }

  return (
    <Card title="图表联动配置" size="small">
      {rules.map((rule, index) => (
        <Space key={index} style={{ marginBottom: 8, display: 'flex' }} align="start">
          <Select
            placeholder="源图表"
            style={{ width: 140 }}
            options={charts.map(c => ({ value: c.id, label: c.title }))}
            value={rule.sourceChartId || undefined}
            onChange={v => updateRule(index, { sourceChartId: v })}
          />
          <Select
            placeholder="源字段"
            style={{ width: 120 }}
            value={rule.sourceField || undefined}
            onChange={v => updateRule(index, { sourceField: v })}
            options={[{ value: rule.sourceField, label: rule.sourceField }]}
            showSearch
          />
          <span style={{ color: '#999' }}>→</span>
          <Select
            placeholder="目标图表"
            style={{ width: 140 }}
            options={charts.map(c => ({ value: c.id, label: c.title }))}
            value={rule.targetChartId || undefined}
            onChange={v => updateRule(index, { targetChartId: v })}
          />
          <Select
            placeholder="目标字段"
            style={{ width: 120 }}
            value={rule.targetField || undefined}
            onChange={v => updateRule(index, { targetField: v })}
            options={[{ value: rule.targetField, label: rule.targetField }]}
            showSearch
          />
          <Button danger size="small" onClick={() => removeRule(index)}>删除</Button>
        </Space>
      ))}
      <Button type="dashed" size="small" onClick={addRule}>添加联动规则</Button>
    </Card>
  )
}
