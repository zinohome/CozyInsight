import { useState, useEffect } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, Select, message } from 'antd'
import type { FormInstance } from 'antd'
import { EditOutlined } from '@ant-design/icons'
import { datasetAPI } from '@/api/dataset'
import { datasourceAPI } from '@/api/datasource'
import type { Dataset, CreateDatasetRequest, PreviewDataResponse } from '@/types/dataset'
import type { Datasource } from '@/types/datasource'

export default function DatasetPage() {
  const [list, setList] = useState<Dataset[]>([])
  const [datasources, setDatasources] = useState<Datasource[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [editRecord, setEditRecord] = useState<Dataset | null>(null)
  const [previewModal, setPreviewModal] = useState(false)
  const [previewData, setPreviewData] = useState<PreviewDataResponse | null>(null)
  const [searchText, setSearchText] = useState('')
  const [datasetType, setDatasetType] = useState<'db' | 'sql'>('db')
  const [editDatasetType, setEditDatasetType] = useState<'db' | 'sql'>('db')
  const [form] = Form.useForm()
  const [editForm] = Form.useForm()

  const fetchList = async () => {
    setLoading(true)
    try {
      const [dsRes, listRes] = await Promise.all([datasourceAPI.list(), datasetAPI.list()])
      setDatasources(dsRes)
      setList(listRes)
    } catch {
      message.error('获取数据失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const handleCreate = async (values: CreateDatasetRequest & { sql?: string }) => {
    try {
      await datasetAPI.create(values)
      message.success('创建成功')
      setModalVisible(false)
      form.resetFields()
      setDatasetType('db')
      fetchList()
    } catch {
      message.error('创建失败')
    }
  }

  const handleUpdate = async (values: CreateDatasetRequest & { sql?: string }) => {
    if (!editRecord) return
    try {
      await datasetAPI.update(editRecord.id, values)
      message.success('更新成功')
      setEditModalVisible(false)
      setEditRecord(null)
      editForm.resetFields()
      setEditDatasetType('db')
      fetchList()
    } catch {
      message.error('更新失败')
    }
  }

  const openEdit = (record: Dataset) => {
    setEditRecord(record)
    setEditDatasetType(record.type as 'db' | 'sql')
    editForm.setFieldsValue({
      name: record.name,
      datasourceId: record.datasourceId,
      databaseName: record.databaseName,
      tableName: record.tableName,
      sql: record.sql || '',
      type: record.type,
    })
    setEditModalVisible(true)
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
      setPreviewData(res)
      setPreviewModal(true)
    } catch {
      message.error('预览失败')
    }
  }

  const filteredList = list.filter(d => !searchText || d.name.toLowerCase().includes(searchText.toLowerCase()))

  const columns = [
    { title: '名称', dataIndex: 'name' },
    { title: '数据源ID', dataIndex: 'datasourceId' },
    { title: '表名', dataIndex: 'tableName' },
    { title: '类型', dataIndex: 'type', render: (type: string) => type === 'sql' ? <Tag color="blue">SQL</Tag> : <Tag color="green">数据库表</Tag> },
    { title: '状态', dataIndex: 'status', render: (status: number) => (status === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>) },
    {
      title: '操作',
      render: (_: unknown, record: Dataset) => (
        <Space>
          <Button type="link" onClick={() => handlePreview(record.id)}>预览</Button>
          <Button type="link" icon={<EditOutlined />} onClick={() => openEdit(record)}>编辑</Button>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]

  const renderDatasetForm = (formInstance: FormInstance, type: 'db' | 'sql') => {
    return (
      <>
        <Form.Item name="name" label="名称" rules={[{ required: true }]}>
          <Input />
        </Form.Item>
        <Form.Item name="datasourceId" label="数据源" rules={[{ required: true }]}>
          <Select options={datasources.map(d => ({ value: d.id, label: d.name }))} />
        </Form.Item>
        <Form.Item name="type" label="类型" rules={[{ required: true }]}>
          <Select
            options={[
              { value: 'db', label: '数据库表' },
              { value: 'sql', label: 'SQL' },
            ]}
            onChange={(v) => {
              if (formInstance === form) {
                setDatasetType(v)
              } else {
                setEditDatasetType(v)
              }
            }}
          />
        </Form.Item>
        {type === 'db' ? (
          <>
            <Form.Item name="databaseName" label="数据库名">
              <Input />
            </Form.Item>
            <Form.Item name="tableName" label="表名" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
          </>
        ) : (
          <>
            <Form.Item name="sql" label="SQL语句" rules={[{ required: true }]}>
              <Input.TextArea rows={6} placeholder="SELECT ..." />
            </Form.Item>
            <Form.Item name="tableName" label="标识名" rules={[{ required: true }]}>
              <Input placeholder="用于标识该SQL数据集" />
            </Form.Item>
          </>
        )}
      </>
    )
  }

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Button type="primary" onClick={() => setModalVisible(true)}>新建数据集</Button>
        <Input.Search
          placeholder="搜索名称"
          allowClear
          style={{ width: 250 }}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
        />
      </div>
      <Table rowKey="id" columns={columns} dataSource={filteredList} loading={loading} pagination={{ pageSize: 10, showSizeChanger: true }} />

      <Modal title="新建数据集" open={modalVisible} onCancel={() => { setModalVisible(false); setDatasetType('db'); form.resetFields() }} footer={null}>
        <Form form={form} onFinish={handleCreate} layout="vertical">
          {renderDatasetForm(form, datasetType)}
          <Form.Item>
            <Button type="primary" htmlType="submit">创建</Button>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="编辑数据集"
        open={editModalVisible}
        onCancel={() => { setEditModalVisible(false); setEditDatasetType('db'); setEditRecord(null); editForm.resetFields() }}
        footer={null}
      >
        <Form form={editForm} onFinish={handleUpdate} layout="vertical">
          {renderDatasetForm(editForm, editDatasetType)}
          <Form.Item>
            <Button type="primary" htmlType="submit">保存</Button>
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
