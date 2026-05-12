import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Card,
  Statistic,
  Tabs,
  Table,
  Button,
  Space,
  Tag,
  Empty,
  message,
} from 'antd'
import {
  DatabaseOutlined,
  TableOutlined,
  BarChartOutlined,
  LayoutOutlined,
  DesktopOutlined,
  PlusOutlined,
} from '@ant-design/icons'
import { workbenchAPI } from '@/api/workbench'
import { useAuthStore } from '@/store/auth'
import type { WorkbenchStats, RecentViewItem, FavoriteItem } from '@/types/workbench'

const { TabPane } = Tabs

const statCards = [
  { key: 'datasource', label: '数据源', icon: <DatabaseOutlined />, color: '#1890ff', route: '/datasource' },
  { key: 'dataset', label: '数据集', icon: <TableOutlined />, color: '#52c41a', route: '/dataset' },
  { key: 'chart', label: '图表', icon: <BarChartOutlined />, color: '#faad14', route: '/chart' },
  { key: 'dashboard', label: '仪表板', icon: <LayoutOutlined />, color: '#722ed1', route: '/dashboard' },
  { key: 'screen', label: '数据大屏', icon: <DesktopOutlined />, color: '#eb2f96', route: '/dashboard' },
]

export default function WorkbenchPage() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const [stats, setStats] = useState<WorkbenchStats | null>(null)
  const [recentList, setRecentList] = useState<RecentViewItem[]>([])
  const [favList, setFavList] = useState<FavoriteItem[]>([])
  const [loading, setLoading] = useState(false)

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const [s, r, f] = await Promise.all([
        workbenchAPI.getStats(),
        workbenchAPI.getRecent(),
        workbenchAPI.getFavorites(),
      ])
      setStats(s)
      setRecentList(r)
      setFavList(f)
    } catch {
      message.error('获取工作台数据失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleQuickCreate = (key: string) => {
    if (key === 'datasource') navigate('/datasource')
    else if (key === 'dataset') navigate('/dataset')
    else if (key === 'chart') navigate('/chart')
    else if (key === 'dashboard') navigate('/dashboard')
    else if (key === 'screen') navigate('/dashboard')
  }

  const handleStatClick = (route: string) => {
    navigate(route)
  }

  const handleRemoveFavorite = async (type: string, id: number) => {
    try {
      await workbenchAPI.removeFavorite(type, id)
      message.success('取消收藏成功')
      setFavList((prev) => prev.filter((f) => f.id !== id))
    } catch {
      message.error('取消收藏失败')
    }
  }

  const recentColumns = [
    {
      title: '标题',
      dataIndex: 'title',
      render: (text: string, record: RecentViewItem) => (
        <a onClick={() => navigate(record.type === 'screen' ? `/screen/view/${record.id}` : `/dashboard/view/${record.id}`)}>
          {text}
        </a>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      render: (type: string) => (type === 'screen' ? <Tag color="purple">数据大屏</Tag> : <Tag color="blue">仪表板</Tag>),
    },
    {
      title: '最后访问时间',
      dataIndex: 'visitedAt',
      width: 200,
    },
  ]

  const favColumns = [
    {
      title: '标题',
      dataIndex: 'title',
      render: (text: string, record: FavoriteItem) => (
        <a onClick={() => navigate(record.type === 'screen' ? `/screen/view/${record.id}` : `/dashboard/view/${record.id}`)}>
          {text}
        </a>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      render: (type: string) => (type === 'screen' ? <Tag color="purple">数据大屏</Tag> : <Tag color="blue">仪表板</Tag>),
    },
    {
      title: '收藏时间',
      dataIndex: 'createdAt',
      width: 200,
    },
    {
      title: '操作',
      width: 100,
      render: (_: unknown, record: FavoriteItem) => (
        <Button type="link" danger onClick={() => handleRemoveFavorite(record.type, record.id)}>
          取消收藏
        </Button>
      ),
    },
  ]

  const statMap: Record<string, number | undefined> = {
    datasource: stats?.datasourceCount,
    dataset: stats?.datasetCount,
    chart: stats?.chartCount,
    dashboard: stats?.dashboardCount,
    screen: stats?.screenCount,
  }

  return (
    <div style={{ padding: 24 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div>
          <h2 style={{ margin: 0, fontWeight: 500 }}>
            欢迎回来，{user?.nickName || user?.username || '用户'}，祝您开心每一天！
          </h2>
        </div>
        <Space>
          {statCards.map((card) => (
            <Button key={card.key} icon={<PlusOutlined />} onClick={() => handleQuickCreate(card.key)}>
              新建{card.label}
            </Button>
          ))}
        </Space>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: 16, marginBottom: 24 }}>
        {statCards.map((card) => (
          <Card
            key={card.key}
            hoverable
            onClick={() => handleStatClick(card.route)}
            bodyStyle={{ textAlign: 'center', padding: 24 }}
          >
            <div style={{ fontSize: 32, color: card.color, marginBottom: 8 }}>{card.icon}</div>
            <Statistic
              title={card.label}
              value={statMap[card.key] ?? 0}
              valueStyle={{ color: card.color, fontSize: 28, fontWeight: 600 }}
            />
          </Card>
        ))}
      </div>

      <Card loading={loading}>
        <Tabs defaultActiveKey="recent">
          <TabPane tab="最近访问" key="recent">
            <Table
              rowKey="id"
              columns={recentColumns}
              dataSource={recentList}
              pagination={false}
              locale={{ emptyText: <Empty description="暂无最近访问记录" /> }}
            />
          </TabPane>
          <TabPane tab="我的收藏" key="favorites">
            <Table
              rowKey="id"
              columns={favColumns}
              dataSource={favList}
              pagination={false}
              locale={{ emptyText: <Empty description="暂无收藏资源" /> }}
            />
          </TabPane>
        </Tabs>
      </Card>
    </div>
  )
}
