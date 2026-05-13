import { Breadcrumb } from 'antd'
import { useLocation, useNavigate } from 'react-router-dom'

const routeNames: Record<string, string> = {
  '/': '工作台',
  '/profile': '个人中心',
  '/datasource': '数据源',
  '/dataset': '数据集',
  '/chart': '图表',
  '/chart/builder': '图表设计器',
  '/dashboard': '仪表板',
  '/dashboard/designer': '仪表板设计器',
  '/dashboard/view': '仪表板预览',
  '/screen/designer': '数据大屏设计器',
  '/screen/view': '数据大屏预览',
  '/system/user': '用户管理',
  '/system/role': '角色管理',
  '/system/log': '操作日志',
}

export default function AppBreadcrumb() {
  const location = useLocation()
  const navigate = useNavigate()
  const path = location.pathname

  // Build breadcrumb items from path segments
  const parts = path.split('/').filter(Boolean)
  const items: Array<{ title: string; onClick?: () => void }> = [
    { title: '工作台', onClick: () => navigate('/') },
  ]

  let currentPath = ''
  for (const part of parts) {
    currentPath += `/${part}`
    const name = routeNames[currentPath] || routeNames[`${currentPath.split('/').slice(0, -1).join('/')}/${part}`] || part
    const isDynamic = !isNaN(Number(part))
    if (!isDynamic) {
      items.push({ title: name, onClick: () => navigate(currentPath) })
    } else {
      items.push({ title: name })
    }
  }

  return (
    <Breadcrumb style={{ marginBottom: 16 }}>
      {items.map((item, index) => (
        <Breadcrumb.Item key={index}>
          {item.onClick ? (
            <a onClick={item.onClick} style={{ cursor: 'pointer' }}>{item.title}</a>
          ) : (
            item.title
          )}
        </Breadcrumb.Item>
      ))}
    </Breadcrumb>
  )
}
