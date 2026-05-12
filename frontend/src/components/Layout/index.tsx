import { useState } from 'react'
import { Layout, Menu, Avatar, Dropdown, theme } from 'antd'
import {
  DashboardOutlined,
  DatabaseOutlined,
  TableOutlined,
  BarChartOutlined,
  LayoutOutlined,
  SettingOutlined,
  UserOutlined,
  TeamOutlined,
  FileTextOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/store/auth'
import AppBreadcrumb from '@/components/Breadcrumb'
import MessageCenter from '@/components/MessageCenter'

const { Header, Sider, Content } = Layout

const allMenuItems = [
  { key: '/', icon: <DashboardOutlined />, label: '工作台' },
  { key: '/datasource', icon: <DatabaseOutlined />, label: '数据源' },
  { key: '/dataset', icon: <TableOutlined />, label: '数据集' },
  { key: '/chart', icon: <BarChartOutlined />, label: '图表' },
  { key: '/dashboard', icon: <LayoutOutlined />, label: '仪表板' },
  {
    key: 'system',
    icon: <SettingOutlined />,
    label: '系统管理',
    children: [
      { key: '/system/user', icon: <UserOutlined />, label: '用户管理' },
      { key: '/system/role', icon: <TeamOutlined />, label: '角色管理' },
      { key: '/system/log', icon: <FileTextOutlined />, label: '操作日志' },
    ],
  },
]

export default function MainLayout({ children }: { children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const logout = useAuthStore((s) => s.logout)
  const user = useAuthStore((s) => s.user)
  const {
    token: { colorBgContainer },
  } = theme.useToken()

  const handleMenuClick = ({ key }: { key: string }) => {
    if (key !== 'system') {
      navigate(key)
    }
  }

  const menuItems = user?.isAdmin ? allMenuItems : allMenuItems.filter(item => item.key !== 'system')

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed}>
        <div style={{ height: 32, margin: 16, background: 'rgba(255,255,255,0.2)', borderRadius: 6 }} />
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          defaultOpenKeys={['system']}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>
      <Layout>
        <Header style={{ padding: 0, background: colorBgContainer, display: 'flex', justifyContent: 'flex-end', alignItems: 'center', paddingRight: 24, gap: 16 }}>
          <MessageCenter />
          <span style={{ color: '#666' }}>{user?.nickName || user?.username || '用户'}</span>
          <Dropdown
            menu={{
              items: [
                { key: 'profile', icon: <UserOutlined />, label: '个人中心', onClick: () => navigate('/profile') },
                { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', onClick: logout },
              ],
            }}
          >
            <Avatar icon={<UserOutlined />} style={{ cursor: 'pointer' }} />
          </Dropdown>
        </Header>
        <Content style={{ margin: 16, background: '#fff', borderRadius: 8, padding: 16 }}>
          <AppBreadcrumb />
          {children}
        </Content>
      </Layout>
    </Layout>
  )
}
