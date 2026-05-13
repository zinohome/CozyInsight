import { useState, useEffect } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, Select, message, Upload, Tabs } from 'antd'
import { UploadOutlined, EditOutlined, ApiOutlined } from '@ant-design/icons'
import { datasourceAPI } from '@/api/datasource'
import type { Datasource } from '@/types/datasource'

export default function DatasourcePage() {
  const [list, setList] = useState<Datasource[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [editRecord, setEditRecord] = useState<Datasource | null>(null)
  const [searchText, setSearchText] = useState('')
  const [modalTab, setModalTab] = useState('database')
  const [form] = Form.useForm()
  const [editForm] = Form.useForm()

  const fetchList = async () => {
    setLoading(true)
    try {
      const res = await datasourceAPI.list()
      setList(res)
    } catch {
      message.error('获取数据源列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const handleCreate = async (values: { name: string; type: string; config: string }) => {
    try {
      await datasourceAPI.create(values)
      message.success('创建成功')
      setModalVisible(false)
      form.resetFields()
      fetchList()
    } catch {
      message.error('创建失败')
    }
  }

  const handleUpdate = async (values: { name: string; type: string; config: string }) => {
    if (!editRecord) return
    try {
      await datasourceAPI.update(editRecord.id, values)
      message.success('更新成功')
      setEditModalVisible(false)
      setEditRecord(null)
      editForm.resetFields()
      fetchList()
    } catch {
      message.error('更新失败')
    }
  }

  const handleTestConnection = async (record: Datasource) => {
    try {
      const config = JSON.parse(record.config)
      await datasourceAPI.testConnection({ type: record.type, config })
      message.success('连接测试成功')
    } catch {
      message.error('连接测试失败')
    }
  }

  const openEdit = (record: Datasource) => {
    setEditRecord(record)
    try {
      const parsedConfig = JSON.parse(record.config)
      editForm.setFieldsValue({
        name: record.name,
        type: record.type,
        config: JSON.stringify(parsedConfig, null, 2),
      })
    } catch {
      editForm.setFieldsValue({
        name: record.name,
        type: record.type,
        config: record.config,
      })
    }
    setEditModalVisible(true)
  }

  const handleUpload = async (file: File, type: 'excel' | 'csv') => {
    try {
      const formData = new FormData()
      formData.append('file', file)
      formData.append('type', type)
      await datasourceAPI.upload(formData)
      message.success('上传成功')
      setModalVisible(false)
      fetchList()
    } catch {
      message.error('上传失败')
    }
  }

  const filteredList = list.filter(d => !searchText || d.name.toLowerCase().includes(searchText.toLowerCase()))

  const handleDelete = async (id: number) => {
    try {
      await datasourceAPI.remove(id)
      message.success('删除成功')
      fetchList()
    } catch {
      message.error('删除失败')
    }
  }

  const columns = [
    { title: '名称', dataIndex: 'name' },
    { title: '类型', dataIndex: 'type', render: (type: string) => <Tag>{type}</Tag> },
    { title: '状态', dataIndex: 'status', render: (status: number) => (status === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>) },
    { title: '创建时间', dataIndex: 'createdAt' },
    {
      title: '操作',
      render: (_: unknown, record: Datasource) => (
        <Space>
          <Button type="link" icon={<ApiOutlined />} onClick={() => handleTestConnection(record)}>测试连接</Button>
          <Button type="link" icon={<EditOutlined />} onClick={() => openEdit(record)}>编辑</Button>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Button type="primary" onClick={() => setModalVisible(true)}>新建数据源</Button>
        <Input.Search
          placeholder="搜索名称"
          allowClear
          style={{ width: 250 }}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
        />
      </div>
      <Table rowKey="id" columns={columns} dataSource={filteredList} loading={loading} pagination={{ pageSize: 10, showSizeChanger: true }} />

      {/* Create Modal */}
      <Modal title="新建数据源" open={modalVisible} onCancel={() => { setModalVisible(false); form.resetFields() }} footer={null}>
        <Tabs activeKey={modalTab} onChange={setModalTab}>
          <Tabs.TabPane tab="数据库" key="database">
            <Form form={form} onFinish={handleCreate} layout="vertical">
              <Form.Item name="name" label="名称" rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="type" label="类型" rules={[{ required: true }]}>
                <Select options={[
                  { value: 'mysql', label: 'MySQL' },
                  { value: 'postgresql', label: 'PostgreSQL' },
                  { value: 'sqlserver', label: 'SQL Server' },
                  { value: 'sqlite', label: 'SQLite' },
                  { value: 'clickhouse', label: 'ClickHouse' },
                ]} />
              </Form.Item>
              <Form.Item name="config" label="配置 (JSON)" rules={[{ required: true }]}>
                <Input.TextArea rows={4} placeholder='{"host":"localhost","port":3306,...}' />
              </Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit">创建</Button>
              </Form.Item>
            </Form>
          </Tabs.TabPane>
          <Tabs.TabPane tab="文件上传" key="file">
            <Upload
              accept=".xlsx,.xls,.csv"
              beforeUpload={(file) => {
                const dot = file.name.lastIndexOf('.')
                if (dot <= 0) {
                  message.error('文件缺少扩展名')
                  return false
                }
                const ext = file.name.slice(dot).toLowerCase()
                if (!['.xlsx', '.xls', '.csv'].includes(ext)) {
                  message.error('仅支持 .xlsx、.xls、.csv 格式')
                  return false
                }
                const type = ext === '.csv' ? 'csv' : 'excel'
                handleUpload(file, type as 'excel' | 'csv')
                return false
              }}
              showUploadList={false}
            >
              <Button icon={<UploadOutlined />}>点击上传 Excel/CSV</Button>
            </Upload>
            <p style={{ marginTop: 8, color: '#888' }}>支持 .xlsx、.xls、.csv 格式</p>
          </Tabs.TabPane>
        </Tabs>
      </Modal>

      {/* Edit Modal */}
      <Modal
        title="编辑数据源"
        open={editModalVisible}
        onCancel={() => { setEditModalVisible(false); setEditRecord(null); editForm.resetFields() }}
        footer={null}
      >
        <Form form={editForm} onFinish={handleUpdate} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select options={[
              { value: 'mysql', label: 'MySQL' },
              { value: 'postgresql', label: 'PostgreSQL' },
              { value: 'sqlserver', label: 'SQL Server' },
              { value: 'sqlite', label: 'SQLite' },
              { value: 'clickhouse', label: 'ClickHouse' },
            ]} />
          </Form.Item>
          <Form.Item name="config" label="配置 (JSON)" rules={[{ required: true }]}>
            <Input.TextArea rows={6} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">保存</Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
