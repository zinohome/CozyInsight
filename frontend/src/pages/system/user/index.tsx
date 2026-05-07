import { useState, useEffect } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, Switch, message } from 'antd'
import { userAPI } from '@/api/user'
import type { User } from '@/types/user'

export default function UserPage() {
  const [list, setList] = useState<User[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [form] = Form.useForm()

  const fetchList = async () => {
    setLoading(true)
    try {
      const data = await userAPI.list()
      setList(data)
    } catch {
      message.error('获取用户列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const handleCreate = async (values: { username: string; password: string; email: string; nickName?: string; isAdmin?: boolean }) => {
    try {
      await userAPI.create({ ...values, status: 1 })
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
      await userAPI.remove(id)
      message.success('删除成功')
      fetchList()
    } catch {
      message.error('删除失败')
    }
  }

  const columns = [
    { title: '用户名', dataIndex: 'username' },
    { title: '邮箱', dataIndex: 'email' },
    { title: '昵称', dataIndex: 'nickName' },
    { title: '管理员', dataIndex: 'isAdmin', render: (v: boolean) => (v ? <Tag color="blue">是</Tag> : <Tag>否</Tag>) },
    { title: '状态', dataIndex: 'status', render: (v: number) => (v === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>) },
    {
      title: '操作',
      render: (_: unknown, record: User) => (
        <Space>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => setModalVisible(true)}>新建用户</Button>
      </div>
      <Table rowKey="id" columns={columns} dataSource={list} loading={loading} />
      <Modal title="新建用户" open={modalVisible} onCancel={() => setModalVisible(false)} footer={null}>
        <Form form={form} onFinish={handleCreate} layout="vertical">
          <Form.Item name="username" label="用户名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, min: 6 }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ required: true, type: 'email' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="nickName" label="昵称">
            <Input />
          </Form.Item>
          <Form.Item name="isAdmin" label="管理员" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">创建</Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
