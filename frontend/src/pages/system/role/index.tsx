import { useState, useEffect } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, message, Tree } from 'antd'
import type { TreeDataNode } from 'antd'
import { roleAPI } from '@/api/role'
import type { Role, Menu } from '@/types/role'

export default function RolePage() {
  const [list, setList] = useState<Role[]>([])
  const [menus, setMenus] = useState<Menu[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [permModal, setPermModal] = useState(false)
  const [currentRole, setCurrentRole] = useState<Role | null>(null)
  const [selectedMenus, setSelectedMenus] = useState<string[]>([])
  const [form] = Form.useForm()

  const fetchList = async () => {
    setLoading(true)
    try {
      const [roleData, menuData] = await Promise.all([roleAPI.list(), roleAPI.listMenus()])
      setList(roleData)
      setMenus(menuData)
    } catch {
      message.error('获取数据失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const handleCreate = async (values: { name: string; code: string; description?: string }) => {
    try {
      await roleAPI.create(values)
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
      await roleAPI.remove(id)
      message.success('删除成功')
      fetchList()
    } catch {
      message.error('删除失败')
    }
  }

  const openPermission = async (role: Role) => {
    setCurrentRole(role)
    try {
      const ids = await roleAPI.getRoleMenus(role.id)
      setSelectedMenus(ids.map(String))
    } catch {
      setSelectedMenus([])
    }
    setPermModal(true)
  }

  const handleSavePermission = async () => {
    if (!currentRole) return
    try {
      await roleAPI.setRoleMenus(currentRole.id, selectedMenus.map(Number))
      message.success('权限设置成功')
      setPermModal(false)
    } catch {
      message.error('权限设置失败')
    }
  }

  const buildTree = (menuList: Menu[]): TreeDataNode[] => {
    const map = new Map<number, TreeDataNode>()
    menuList.forEach((m) => {
      map.set(m.id, { key: String(m.id), title: m.name, children: [] })
    })
    const roots: TreeDataNode[] = []
    menuList.forEach((m) => {
      const node = map.get(m.id)!
      if (m.parentId === 0) {
        roots.push(node)
      } else {
        const parent = map.get(m.parentId)
        if (parent) {
          parent.children = parent.children || []
          parent.children.push(node)
        }
      }
    })
    return roots
  }

  const columns = [
    { title: '角色名', dataIndex: 'name' },
    { title: '编码', dataIndex: 'code' },
    { title: '描述', dataIndex: 'description' },
    { title: '状态', dataIndex: 'status', render: (v: number) => (v === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>) },
    {
      title: '操作',
      render: (_: unknown, record: Role) => (
        <Space>
          <Button type="link" onClick={() => openPermission(record)}>权限</Button>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => setModalVisible(true)}>新建角色</Button>
      </div>
      <Table rowKey="id" columns={columns} dataSource={list} loading={loading} />
      <Modal title="新建角色" open={modalVisible} onCancel={() => setModalVisible(false)} footer={null}>
        <Form form={form} onFinish={handleCreate} layout="vertical">
          <Form.Item name="name" label="角色名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="code" label="编码" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">创建</Button>
          </Form.Item>
        </Form>
      </Modal>
      <Modal title={`设置权限 - ${currentRole?.name || ''}`} open={permModal} onCancel={() => setPermModal(false)} onOk={handleSavePermission}>
        <Tree
          checkable
          treeData={buildTree(menus)}
          checkedKeys={selectedMenus}
          onCheck={(keys) => setSelectedMenus(keys as string[])}
        />
      </Modal>
    </div>
  )
}
