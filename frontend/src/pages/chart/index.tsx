import { useState, useEffect } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, Select, message } from 'antd'
import { chartAPI } from '@/api/chart'
import { datasetAPI } from '@/api/dataset'
import type { Chart } from '@/types/chart'
import type { Dataset } from '@/types/dataset'

export default function ChartPage() {
  const [list, setList] = useState<Chart[]>([])
  const [datasets, setDatasets] = useState<Dataset[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [form] = Form.useForm()

  const fetchList = async () => {
    setLoading(true)
    try {
      const [dsRes, listRes] = await Promise.all([datasetAPI.list(), chartAPI.list()])
      setDatasets(dsRes.data.data)
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

  const handleCreate = async (values: { title: string; type: string; datasetId: number; config: string }) => {
    try {
      await chartAPI.create(values)
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
      await chartAPI.remove(id)
      message.success('删除成功')
      fetchList()
    } catch {
      message.error('删除失败')
    }
  }

  const columns = [
    { title: '标题', dataIndex: 'title' },
    { title: '类型', dataIndex: 'type', render: (type: string) => <Tag>{type}</Tag> },
    { title: '数据集ID', dataIndex: 'datasetId' },
    { title: '状态', dataIndex: 'status', render: (status: number) => (status === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>) },
    {
      title: '操作',
      render: (_: unknown, record: Chart) => (
        <Space>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => setModalVisible(true)}>新建图表</Button>
      </div>
      <Table rowKey="id" columns={columns} dataSource={list} loading={loading} />
      <Modal title="新建图表" open={modalVisible} onCancel={() => setModalVisible(false)} footer={null}>
        <Form form={form} onFinish={handleCreate} layout="vertical">
          <Form.Item name="title" label="标题" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select options={[{ value: 'bar', label: '柱状图' }, { value: 'line', label: '折线图' }, { value: 'pie', label: '饼图' }, { value: 'table', label: '表格' }]} />
          </Form.Item>
          <Form.Item name="datasetId" label="数据集" rules={[{ required: true }]}>
            <Select options={datasets.map(d => ({ value: d.id, label: d.name }))} />
          </Form.Item>
          <Form.Item name="config" label="配置 (JSON)">
            <Input.TextArea rows={4} placeholder='{}' />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">创建</Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
