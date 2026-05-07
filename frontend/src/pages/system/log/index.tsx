import { useState, useEffect } from 'react'
import { Table, Tag, message } from 'antd'
import { logAPI } from '@/api/log'
import type { OperationLog } from '@/types/log'

export default function LogPage() {
  const [list, setList] = useState<OperationLog[]>([])
  const [loading, setLoading] = useState(false)

  const fetchList = async () => {
    setLoading(true)
    try {
      const data = await logAPI.list(100)
      setList(data)
    } catch {
      message.error('获取日志失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const columns = [
    { title: '用户', dataIndex: 'username' },
    { title: '方法', dataIndex: 'method', render: (v: string) => <Tag>{v}</Tag> },
    { title: '路径', dataIndex: 'path' },
    { title: '状态码', dataIndex: 'statusCode' },
    { title: '耗时(ms)', dataIndex: 'duration' },
    { title: 'IP', dataIndex: 'ip' },
    { title: '时间', dataIndex: 'createdAt' },
  ]

  return (
    <div style={{ padding: 24 }}>
      <Table rowKey="id" columns={columns} dataSource={list} loading={loading} pagination={{ pageSize: 20 }} />
    </div>
  )
}
