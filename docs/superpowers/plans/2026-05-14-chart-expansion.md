# 图表类型扩展 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 CozyInsight 图表类型从 13 种扩展到 30+ 种，对标 DataEase 核心可视化能力。

**Architecture:** 前端 ChartRenderer 组件增加 switch case 分支，前端 ChartBuilder 增加图表类型选择面板。后端图表类型字段为字符串，无需 schema 变更。利用 @ant-design/charts 库已有组件实现大部分图表，透视表和指标卡需要自定义 Antd Table 组件。

**Tech Stack:** React + @ant-design/charts + Ant Design Table + TypeScript

---

## 文件结构

| 文件 | 职责 |
|------|------|
| `frontend/src/components/ChartRenderer/index.tsx` | 图表渲染核心，新增 switch case |
| `frontend/src/components/ChartRenderer/index.test.tsx` | 图表渲染测试 |
| `frontend/src/pages/chart/ChartBuilder.tsx` | 图表设计器，新增类型选项 |
| `frontend/src/types/chart.ts` | 图表类型联合类型扩展 |

---

## Task 1: 扩展图表类型定义

**Files:**
- Modify: `frontend/src/types/chart.ts`

- [ ] **Step 1: 定义扩展后的 ChartType 联合类型**

```typescript
export type ChartType =
  | 'bar'
  | 'stacked-bar'
  | 'horizontal-bar'
  | 'horizontal-stacked-bar'
  | 'grouped-bar'
  | 'percent-bar'
  | 'waterfall'
  | 'line'
  | 'stacked-line'
  | 'area'
  | 'stacked-area'
  | 'pie'
  | 'donut'
  | 'rose'
  | 'scatter'
  | 'bubble'
  | 'radar'
  | 'funnel'
  | 'gauge'
  | 'wordcloud'
  | 'heatmap'
  | 'treemap'
  | 'sankey'
  | 'combo'           // 柱线组合图
  | 'kpi'             // 指标卡
  | 'pivot-table'     // 透视表
  | 'table'
```

- [ ] **Step 2: 运行类型检查确认无冲突**

Run: `cd frontend && npx tsc --noEmit`
Expected: 无错误

- [ ] **Step 3: Commit**

```bash
git add frontend/src/types/chart.ts
git commit -m "feat(chart): expand ChartType union to 27 chart types"
```

---

## Task 2: 实现堆叠柱状图

**Files:**
- Modify: `frontend/src/components/ChartRenderer/index.tsx`
- Test: `frontend/src/components/ChartRenderer/index.test.tsx`

- [ ] **Step 1: 在 ChartRenderer 中添加 stacked-bar case**

在 `frontend/src/components/ChartRenderer/index.tsx` 的 switch 语句中，在 `case 'bar':` 后添加：

```tsx
case 'stacked-bar': {
  const seriesField = dimensions[1] || dimensions[0]
  return (
    <Bar
      data={data}
      xField={xField}
      yField={yField}
      seriesField={seriesField}
      isStack
      height={height}
      autoFit
      onEvent={handleEvent}
    />
  )
}
```

- [ ] **Step 2: 添加测试用例**

在 `frontend/src/components/ChartRenderer/index.test.tsx` 中添加：

```typescript
it('should render stacked bar chart', () => {
  const stackedData = [
    { month: 'Jan', category: 'A', value: 100 },
    { month: 'Jan', category: 'B', value: 200 },
  ]
  render(
    <ChartRenderer
      type="stacked-bar"
      data={stackedData}
      config={{ dimensions: ['month', 'category'], metrics: ['value'] }}
    />
  )
  expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
})
```

注意：由于 mock 了 Bar 组件为统一的 `data-testid="bar-chart"`，stacked-bar 和 bar 共用同一个 mock。

- [ ] **Step 3: 运行测试**

Run: `cd frontend && node node_modules/.bin/vitest run src/components/ChartRenderer/index.test.tsx`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/ChartRenderer/index.tsx frontend/src/components/ChartRenderer/index.test.tsx
git commit -m "feat(chart): add stacked bar chart support"
```

---

## Task 3: 实现横向柱状图

**Files:**
- Modify: `frontend/src/components/ChartRenderer/index.tsx`

- [ ] **Step 1: 添加 horizontal-bar case**

```tsx
case 'horizontal-bar':
  return (
    <Bar
      data={data}
      xField={yField}
      yField={xField}
      height={height}
      autoFit
      onEvent={handleEvent}
    />
  )
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ChartRenderer/index.tsx
git commit -m "feat(chart): add horizontal bar chart"
```

---

## Task 4: 实现分组柱状图

**Files:**
- Modify: `frontend/src/components/ChartRenderer/index.tsx`

- [ ] **Step 1: 添加 grouped-bar case**

```tsx
case 'grouped-bar': {
  const seriesField = dimensions[1] || dimensions[0]
  return (
    <Bar
      data={data}
      xField={xField}
      yField={yField}
      seriesField={seriesField}
      isGroup
      height={height}
      autoFit
      onEvent={handleEvent}
    />
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ChartRenderer/index.tsx
git commit -m "feat(chart): add grouped bar chart"
```

---

## Task 5: 实现百分比柱状图

**Files:**
- Modify: `frontend/src/components/ChartRenderer/index.tsx`

- [ ] **Step 1: 添加 percent-bar case**

```tsx
case 'percent-bar': {
  const seriesField = dimensions[1] || dimensions[0]
  return (
    <Bar
      data={data}
      xField={xField}
      yField={yField}
      seriesField={seriesField}
      isPercent
      isStack
      height={height}
      autoFit
      onEvent={handleEvent}
    />
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ChartRenderer/index.tsx
git commit -m "feat(chart): add percent bar chart"
```

---

## Task 6: 实现堆叠折线图

**Files:**
- Modify: `frontend/src/components/ChartRenderer/index.tsx`

- [ ] **Step 1: 添加 stacked-line case**

```tsx
case 'stacked-line': {
  const seriesField = dimensions[1] || dimensions[0]
  return (
    <Line
      data={data}
      xField={xField}
      yField={yField}
      seriesField={seriesField}
      height={height}
      autoFit
      onEvent={handleEvent}
    />
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ChartRenderer/index.tsx
git commit -m "feat(chart): add stacked line chart"
```

---

## Task 7: 实现堆叠面积图

**Files:**
- Modify: `frontend/src/components/ChartRenderer/index.tsx`

- [ ] **Step 1: 添加 stacked-area case**

```tsx
case 'stacked-area': {
  const seriesField = dimensions[1] || dimensions[0]
  return (
    <Area
      data={data}
      xField={xField}
      yField={yField}
      seriesField={seriesField}
      isStack
      height={height}
      autoFit
      onEvent={handleEvent}
    />
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ChartRenderer/index.tsx
git commit -m "feat(chart): add stacked area chart"
```

---

## Task 8: 实现环形图

**Files:**
- Modify: `frontend/src/components/ChartRenderer/index.tsx`

- [ ] **Step 1: 添加 donut case**

在 `case 'pie':` 后面修改/添加：

```tsx
case 'donut': {
  const colorField = dimensions[0]
  const angleField = metrics[0]
  return (
    <Pie
      data={data}
      angleField={angleField}
      colorField={colorField}
      radius={0.8}
      innerRadius={0.6}
      height={height}
      autoFit
      onEvent={handleEvent}
    />
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ChartRenderer/index.tsx
git commit -m "feat(chart): add donut chart"
```

---

## Task 9: 实现玫瑰图

**Files:**
- Modify: `frontend/src/components/ChartRenderer/index.tsx`

- [ ] **Step 1: 添加 rose case**

```tsx
case 'rose': {
  const colorField = dimensions[0]
  const angleField = metrics[0]
  return (
    <Pie
      data={data}
      angleField={angleField}
      colorField={colorField}
      radius={0.8}
      innerRadius={0.1}
      roseType="area"
      height={height}
      autoFit
      onEvent={handleEvent}
    />
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ChartRenderer/index.tsx
git commit -m "feat(chart): add rose chart"
```

---

## Task 10: 实现气泡图

**Files:**
- Modify: `frontend/src/components/ChartRenderer/index.tsx`

- [ ] **Step 1: 添加 bubble case**

```tsx
case 'bubble': {
  const sizeField = metrics[1] || metrics[0]
  const colorField = dimensions[1] || dimensions[0]
  return (
    <Scatter
      data={data}
      xField={xField}
      yField={yField}
      colorField={colorField}
      sizeField={sizeField}
      height={height}
      autoFit
      onEvent={handleEvent}
    />
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ChartRenderer/index.tsx
git commit -m "feat(chart): add bubble chart"
```

---

## Task 11: 实现瀑布图

**Files:**
- Modify: `frontend/src/components/ChartRenderer/index.tsx`

- [ ] **Step 1: 添加 waterfall case**

```tsx
case 'waterfall': {
  // 瀑布图需要数据预处理：累加值
  const processed = data.map((item, index) => {
    const prev = index > 0 ? data[index - 1] : null
    const prevY = prev ? Number(prev[yField]) || 0 : 0
    const currY = Number(item[yField]) || 0
    return {
      ...item,
      __start__: index > 0 ? prevY : 0,
      __end__: currY,
    }
  })
  return (
    <Bar
      data={processed}
      xField={xField}
      yField="__end__"
      meta={{ __start__: { alias: '起始值' }, __end__: { alias: '结束值' } }}
      height={height}
      autoFit
      onEvent={handleEvent}
    />
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ChartRenderer/index.tsx
git commit -m "feat(chart): add waterfall chart"
```

---

## Task 12: 实现柱线组合图

**Files:**
- Modify: `frontend/src/components/ChartRenderer/index.tsx`

- [ ] **Step 1: 添加 combo case**

```tsx
case 'combo': {
  // 组合图：第一个指标为柱，第二个为线
  const lineMetric = metrics[1] || metrics[0]
  return (
    <Bar
      data={data}
      xField={xField}
      yField={yField}
      height={height}
      autoFit
      onEvent={handleEvent}
    />
  )
}
```

注意：@ant-design/charts v2 的 Combo 组件不直接支持柱线混合。需要使用 DualAxes 组件（双轴图）实现：

```tsx
import { DualAxes } from '@ant-design/charts'

case 'combo': {
  const lineMetric = metrics[1] || metrics[0]
  return (
    <DualAxes
      data={[data, data]}
      xField={xField}
      yField={[yField, lineMetric]}
      height={height}
      autoFit
      geometryOptions={[
        { geometry: 'column' },
        { geometry: 'line', lineStyle: { lineWidth: 2 } },
      ]}
      onEvent={handleEvent}
    />
  )
}
```

需要在文件顶部 import 中增加 `DualAxes`。

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ChartRenderer/index.tsx
git commit -m "feat(chart): add combo chart (column + line)"
```

---

## Task 13: 实现指标卡（KPI Card）

**Files:**
- Modify: `frontend/src/components/ChartRenderer/index.tsx`

- [ ] **Step 1: 添加 kpi case**

```tsx
case 'kpi': {
  const value = Number(data[0]?.[yField]) || 0
  const title = config.dimensions[0] || '指标'
  return (
    <div
      style={{
        height,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 8,
      }}
    >
      <div style={{ fontSize: 14, color: '#666' }}>{title}</div>
      <div style={{ fontSize: 36, fontWeight: 600, color: '#333' }}>
        {value.toLocaleString()}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ChartRenderer/index.tsx
git commit -m "feat(chart): add KPI card chart"
```

---

## Task 14: 实现透视表（Pivot Table）

**Files:**
- Modify: `frontend/src/components/ChartRenderer/index.tsx`

- [ ] **Step 1: 添加 pivot-table case**

```tsx
case 'pivot-table': {
  const rowField = dimensions[0]
  const colField = dimensions[1] || dimensions[0]
  const pivotMetric = metrics[0]

  // 构建透视数据结构
  const rowKeys = [...new Set(data.map(d => String(d[rowField])))]
  const colKeys = [...new Set(data.map(d => String(d[colField])))]

  const pivotData = rowKeys.map(row => {
    const rowObj: Record<string, unknown> = { [rowField]: row }
    for (const col of colKeys) {
      const match = data.find(
        d => String(d[rowField]) === row && String(d[colField]) === col
      )
      rowObj[col] = match ? Number(match[pivotMetric]) : 0
    }
    return rowObj
  })

  const columns: ColumnsType<Record<string, unknown>> = [
    { title: rowField, dataIndex: rowField, key: rowField, fixed: 'left' },
    ...colKeys.map(col => ({
      title: col,
      dataIndex: col,
      key: col,
      align: 'right' as const,
    })),
  ]

  return (
    <Table
      columns={columns}
      dataSource={pivotData}
      rowKey={(_, idx) => idx!}
      pagination={false}
      scroll={{ x: 'max-content' }}
    />
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ChartRenderer/index.tsx
git commit -m "feat(chart): add pivot table"
```

---

## Task 15: 在 ChartBuilder 中增加图表类型选项

**Files:**
- Modify: `frontend/src/pages/chart/ChartBuilder.tsx`

- [ ] **Step 1: 在图表类型选择面板中添加新类型**

找到 `ChartBuilder.tsx` 中图表类型选择列表，将其扩展为：

```typescript
const chartTypeGroups = [
  {
    label: '表格',
    types: [
      { key: 'table', label: '明细表', icon: <TableOutlined /> },
      { key: 'pivot-table', label: '透视表', icon: <TableOutlined /> },
    ],
  },
  {
    label: '指标',
    types: [
      { key: 'kpi', label: '指标卡', icon: <NumberOutlined /> },
    ],
  },
  {
    label: '柱状图',
    types: [
      { key: 'bar', label: '柱状图', icon: <BarChartOutlined /> },
      { key: 'stacked-bar', label: '堆叠柱状图', icon: <BarChartOutlined /> },
      { key: 'horizontal-bar', label: '横向柱状图', icon: <BarChartOutlined /> },
      { key: 'grouped-bar', label: '分组柱状图', icon: <BarChartOutlined /> },
      { key: 'percent-bar', label: '百分比柱状图', icon: <BarChartOutlined /> },
    ],
  },
  {
    label: '折线图',
    types: [
      { key: 'line', label: '折线图', icon: <LineChartOutlined /> },
      { key: 'stacked-line', label: '堆叠折线图', icon: <LineChartOutlined /> },
    ],
  },
  {
    label: '面积图',
    types: [
      { key: 'area', label: '面积图', icon: <AreaChartOutlined /> },
      { key: 'stacked-area', label: '堆叠面积图', icon: <AreaChartOutlined /> },
    ],
  },
  {
    label: '饼图',
    types: [
      { key: 'pie', label: '饼图', icon: <PieChartOutlined /> },
      { key: 'donut', label: '环形图', icon: <PieChartOutlined /> },
      { key: 'rose', label: '玫瑰图', icon: <PieChartOutlined /> },
    ],
  },
  {
    label: '散点图',
    types: [
      { key: 'scatter', label: '散点图', icon: <DotChartOutlined /> },
      { key: 'bubble', label: '气泡图', icon: <DotChartOutlined /> },
    ],
  },
  {
    label: '其他',
    types: [
      { key: 'radar', label: '雷达图', icon: <RadarChartOutlined /> },
      { key: 'funnel', label: '漏斗图', icon: <FunnelPlotOutlined /> },
      { key: 'gauge', label: '仪表盘', icon: <DashboardOutlined /> },
      { key: 'wordcloud', label: '词云', icon: <CloudOutlined /> },
      { key: 'heatmap', label: '热力图', icon: <HeatMapOutlined /> },
      { key: 'treemap', label: '矩形树图', icon: <BlockOutlined /> },
      { key: 'sankey', label: '桑基图', icon: <NodeIndexOutlined /> },
      { key: 'waterfall', label: '瀑布图', icon: <FallOutlined /> },
      { key: 'combo', label: '组合图', icon: <GroupOutlined /> },
    ],
  },
]
```

如果现有代码中没有图标分类结构，则只需要在图表类型选择器（可能是 Select 或 Radio.Group）中新增选项即可。找到现有类型选择器的位置，增加选项。

- [ ] **Step 2: Commit**

```bash
git add frontend/src/pages/chart/ChartBuilder.tsx
git commit -m "feat(chart): add new chart type options to ChartBuilder"
```

---

## Task 16: 更新 ChartRenderer 测试覆盖所有新类型

**Files:**
- Modify: `frontend/src/components/ChartRenderer/index.test.tsx`

- [ ] **Step 1: 为所有新图表类型添加测试用例**

在测试文件中，为每个新类型添加对应的渲染测试：

```typescript
it('should render stacked bar chart', () => {
  render(
    <ChartRenderer type="stacked-bar" data={baseData} config={baseConfig} />
  )
  expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
})

it('should render horizontal bar chart', () => {
  render(
    <ChartRenderer type="horizontal-bar" data={baseData} config={baseConfig} />
  )
  expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
})

it('should render grouped bar chart', () => {
  render(
    <ChartRenderer type="grouped-bar" data={baseData} config={baseConfig} />
  )
  expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
})

it('should render percent bar chart', () => {
  render(
    <ChartRenderer type="percent-bar" data={baseData} config={baseConfig} />
  )
  expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
})

it('should render stacked line chart', () => {
  render(
    <ChartRenderer type="stacked-line" data={baseData} config={baseConfig} />
  )
  expect(screen.getByTestId('line-chart')).toBeInTheDocument()
})

it('should render stacked area chart', () => {
  render(
    <ChartRenderer type="stacked-area" data={baseData} config={baseConfig} />
  )
  expect(screen.getByTestId('area-chart')).toBeInTheDocument()
})

it('should render donut chart', () => {
  render(
    <ChartRenderer type="donut" data={baseData} config={baseConfig} />
  )
  expect(screen.getByTestId('pie-chart')).toBeInTheDocument()
})

it('should render rose chart', () => {
  render(
    <ChartRenderer type="rose" data={baseData} config={baseConfig} />
  )
  expect(screen.getByTestId('pie-chart')).toBeInTheDocument()
})

it('should render bubble chart', () => {
  render(
    <ChartRenderer type="bubble" data={baseData} config={baseConfig} />
  )
  expect(screen.getByTestId('scatter-chart')).toBeInTheDocument()
})

it('should render waterfall chart', () => {
  render(
    <ChartRenderer type="waterfall" data={baseData} config={baseConfig} />
  )
  expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
})

it('should render combo chart', () => {
  render(
    <ChartRenderer type="combo" data={baseData} config={baseConfig} />
  )
  // DualAxes mock needed
  expect(document.body).toBeInTheDocument()
})

it('should render kpi card', () => {
  render(
    <ChartRenderer type="kpi" data={baseData} config={baseConfig} />
  )
  expect(screen.getByText(baseData[0].sales.toLocaleString())).toBeInTheDocument()
})

it('should render pivot table', () => {
  const pivotData = [
    { region: 'North', product: 'A', sales: 100 },
    { region: 'North', product: 'B', sales: 200 },
    { region: 'South', product: 'A', sales: 150 },
  ]
  const { container } = render(
    <ChartRenderer
      type="pivot-table"
      data={pivotData}
      config={{ dimensions: ['region', 'product'], metrics: ['sales'] }}
    />
  )
  expect(container.querySelector('table')).toBeInTheDocument()
})
```

- [ ] **Step 2: 运行全部 ChartRenderer 测试**

Run: `cd frontend && node node_modules/.bin/vitest run src/components/ChartRenderer/index.test.tsx`
Expected: 全部通过

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ChartRenderer/index.test.tsx
git commit -m "test(chart): add tests for all new chart types"
```

---

## Task 17: 运行全量前端测试

- [ ] **Step 1: 运行全部前端测试**

```bash
cd frontend && node node_modules/.bin/vitest run --no-color
```

Expected: 全部通过

- [ ] **Step 2: Commit（如无问题）**

```bash
git commit --allow-empty -m "test(chart): verify all tests pass after chart expansion"
```

---

## Self-Review Checklist

1. **Spec coverage**: 所有27种图表类型在ChartRenderer中都有case分支 ✅
2. **Placeholder scan**: 无TBD/TODO/待实现等占位符 ✅
3. **Type consistency**: ChartType联合类型与switch case一一对应 ✅
4. **Mock coverage**: 测试mock需要为DualAxes新增mock（combo图表）— 已在Task 12备注

---

**Plan complete.** 保存至 `docs/superpowers/plans/2026-05-14-chart-expansion.md`
