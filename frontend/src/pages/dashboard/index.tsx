import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Table, Button, Space, Tag, Modal, Form, Input, message } from 'antd'
import { dashboardAPI } from '@/api/dashboard'
import type { Dashboard } from '@/types/dashboard'

export default function DashboardPage() {
  const navigate = useNavigate()
  const [list, setList] = useState<Dashboard[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [createType, setCreateType] = useState<'dashboard' | 'screen'>('dashboard')
  const [form] = Form.useForm()

  const fetchList = async () => {
    setLoading(true)
    try {
      const res = await dashboardAPI.list()
      setList(res)
    } catch {
      message.error('获取仪表板列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const handleCreate = async (values: { title: string; config?: string }) => {
    try {
      await dashboardAPI.create({ title: values.title, type: createType, config: values.config || '{}' })
      message.success('创建成功')
      setModalVisible(false)
      form.resetFields()
      fetchList()
    } catch {
      message.error('创建失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await dashboardAPI.remove(id)
      message.success('删除成功')
      fetchList()
    } catch {
      message.error('删除失败')
    }
  }

  const openCreateModal = (type: 'dashboard' | 'screen') => {
    setCreateType(type)
    setModalVisible(true)
  }

  const columns = [
    { title: '标题', dataIndex: 'title' },
    { title: '类型', dataIndex: 'type', render: (type: 'dashboard' | 'screen') => (type === 'screen' ? <Tag color="purple">数据大屏</Tag> : <Tag color="blue">仪表板</Tag>) },
    { title: '状态', dataIndex: 'status', render: (status: number) => (status === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>) },
    { title: '创建时间', dataIndex: 'createdAt' },
    {
      title: '操作',
      render: (_: unknown, record: Dashboard) => (
        <Space>
          <Button type="link" onClick={() => navigate(record.type === 'screen' ? `/screen/designer/${record.id}` : `/dashboard/designer/${record.id}`)}>设计</Button>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => openCreateModal('dashboard')}>新建仪表板</Button>
        <Button type="primary" style={{ marginLeft: 8 }} onClick={() => openCreateModal('screen')}>新建数据大屏</Button>
      </div>
      <Table rowKey="id" columns={columns} dataSource={list} loading={loading} />
      <Modal title={createType === 'screen' ? '新建数据大屏' : '新建仪表板'} open={modalVisible} onCancel={() => setModalVisible(false)} footer={null}>
        <Form form={form} onFinish={handleCreate} layout="vertical">
          <Form.Item name="title" label="标题" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">创建</Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
