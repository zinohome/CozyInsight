import { useState, useEffect } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, Select, message } from 'antd'
import { datasetAPI } from '@/api/dataset'
import { datasourceAPI } from '@/api/datasource'
import type { Dataset, PreviewDataResponse } from '@/types/dataset'
import type { Datasource } from '@/types/datasource'

export default function DatasetPage() {
  const [list, setList] = useState<Dataset[]>([])
  const [datasources, setDatasources] = useState<Datasource[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [previewModal, setPreviewModal] = useState(false)
  const [previewData, setPreviewData] = useState<PreviewDataResponse | null>(null)
  const [form] = Form.useForm()

  const fetchList = async () => {
    setLoading(true)
    try {
      const [dsRes, listRes] = await Promise.all([datasourceAPI.list(), datasetAPI.list()])
      setDatasources(dsRes.data.data)
      setList(listRes.data.data)
    } catch {
      message.error('获取数据失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const handleCreate = async (values: { name: string; datasourceId: number; databaseName: string; tableName: string; type: string }) => {
    try {
      await datasetAPI.create({ ...values, mode: 0 })
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
      await datasetAPI.remove(id)
      message.success('删除成功')
      fetchList()
    } catch {
      message.error('删除失败')
    }
  }

  const handlePreview = async (id: number) => {
    try {
      const res = await datasetAPI.preview(id, 10)
      setPreviewData(res.data.data)
      setPreviewModal(true)
    } catch {
      message.error('预览失败')
    }
  }

  const columns = [
    { title: '名称', dataIndex: 'name' },
    { title: '数据源ID', dataIndex: 'datasourceId' },
    { title: '表名', dataIndex: 'tableName' },
    { title: '类型', dataIndex: 'type' },
    { title: '状态', dataIndex: 'status', render: (status: number) => (status === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>) },
    {
      title: '操作',
      render: (_: unknown, record: Dataset) => (
        <Space>
          <Button type="link" onClick={() => handlePreview(record.id)}>预览</Button>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => setModalVisible(true)}>新建数据集</Button>
      </div>
      <Table rowKey="id" columns={columns} dataSource={list} loading={loading} />
      <Modal title="新建数据集" open={modalVisible} onCancel={() => setModalVisible(false)} footer={null}>
        <Form form={form} onFinish={handleCreate} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="datasourceId" label="数据源" rules={[{ required: true }]}>
            <Select options={datasources.map(d => ({ value: d.id, label: d.name }))} />
          </Form.Item>
          <Form.Item name="databaseName" label="数据库名">
            <Input />
          </Form.Item>
          <Form.Item name="tableName" label="表名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select options={[{ value: 'db', label: '数据库表' }]} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">创建</Button>
          </Form.Item>
        </Form>
      </Modal>
      <Modal title="数据预览" open={previewModal} onCancel={() => setPreviewModal(false)} footer={null} width={800}>
        {previewData && (
          <Table
            size="small"
            columns={previewData.fields.map(f => ({ title: f.name, dataIndex: f.name, key: f.name }))}
            dataSource={previewData.data.map((row, idx) => ({ ...row, key: idx }))}
            pagination={false}
          />
        )}
      </Modal>
    </div>
  )
}
