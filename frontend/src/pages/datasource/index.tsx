import { useState, useEffect } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, Select, message, Upload, Tabs } from 'antd'
import { UploadOutlined } from '@ant-design/icons'
import { datasourceAPI } from '@/api/datasource'
import type { Datasource } from '@/types/datasource'

export default function DatasourcePage() {
  const [list, setList] = useState<Datasource[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [modalTab, setModalTab] = useState('database')
  const [form] = Form.useForm()

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
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => setModalVisible(true)}>新建数据源</Button>
      </div>
      <Table rowKey="id" columns={columns} dataSource={list} loading={loading} />
      <Modal title="新建数据源" open={modalVisible} onCancel={() => { setModalVisible(false); form.resetFields() }} footer={null}>
        <Tabs activeKey={modalTab} onChange={setModalTab}>
          <Tabs.TabPane tab="数据库" key="database">
            <Form form={form} onFinish={handleCreate} layout="vertical">
              <Form.Item name="name" label="名称" rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="type" label="类型" rules={[{ required: true }]}>
                <Select options={[{ value: 'mysql', label: 'MySQL' }, { value: 'postgresql', label: 'PostgreSQL' }]} />
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
    </div>
  )
}
