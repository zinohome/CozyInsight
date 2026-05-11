import { Breadcrumb, Button } from 'antd'

interface DrillBreadcrumbProps {
  dimensions: string[]
  currentLevel: number
  onDrillUp: (level: number) => void
}

export default function DrillBreadcrumb({ dimensions, currentLevel, onDrillUp }: DrillBreadcrumbProps) {
  if (dimensions.length <= 1) return null

  const items = dimensions.slice(0, currentLevel + 1).map((dim, index) => ({
    title: index === currentLevel ? dim : (
      <Button type="link" size="small" onClick={() => onDrillUp(index)}>{dim}</Button>
    ),
  }))

  return <Breadcrumb items={items} style={{ marginBottom: 8 }} />
}
