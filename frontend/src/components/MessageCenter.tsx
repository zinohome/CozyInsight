import { useState, useEffect, useCallback } from 'react'
import { Badge, Dropdown, List, Button, Empty } from 'antd'
import { BellOutlined, CheckOutlined, DeleteOutlined } from '@ant-design/icons'
import { messageAPI } from '@/api/message'
import type { Message } from '@/types/message'

export default function MessageCenter() {
  const [messages, setMessages] = useState<Message[]>([])
  const [unreadCount, setUnreadCount] = useState(0)
  const [loading, setLoading] = useState(false)
  const [open, setOpen] = useState(false)

  const fetchMessages = useCallback(async () => {
    setLoading(true)
    try {
      const [msgs, count] = await Promise.all([
        messageAPI.list(),
        messageAPI.countUnread(),
      ])
      setMessages(msgs)
      setUnreadCount(count)
    } catch {
      // ignore
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (open) {
      fetchMessages()
    }
  }, [open, fetchMessages])

  useEffect(() => {
    const timer = setInterval(() => {
      messageAPI.countUnread().then(setUnreadCount).catch(() => {})
    }, 30000)
    return () => clearInterval(timer)
  }, [])

  const handleMarkAsRead = async (id: number) => {
    try {
      await messageAPI.markAsRead(id)
      setMessages((prev) =>
        prev.map((m) => (m.id === id ? { ...m, isRead: 1 } : m))
      )
      setUnreadCount((c) => Math.max(0, c - 1))
    } catch {
      // ignore
    }
  }

  const handleMarkAllAsRead = async () => {
    try {
      await messageAPI.markAllAsRead()
      setMessages((prev) => prev.map((m) => ({ ...m, isRead: 1 })))
      setUnreadCount(0)
    } catch {
      // ignore
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await messageAPI.remove(id)
      setMessages((prev) => prev.filter((m) => m.id !== id))
    } catch {
      // ignore
    }
  }

  const dropdownContent = (
    <div style={{ width: 360, maxHeight: 400, overflow: 'auto', background: '#fff' }}>
      <div style={{ padding: '8px 16px', borderBottom: '1px solid #f0f0f0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span style={{ fontWeight: 500 }}>消息中心</span>
        <Button type="link" size="small" onClick={handleMarkAllAsRead}>
          <CheckOutlined /> 全部已读
        </Button>
      </div>
      <List
        size="small"
        loading={loading}
        dataSource={messages}
        locale={{ emptyText: <Empty description="暂无消息" /> }}
        renderItem={(msg) => (
          <List.Item
            style={{ padding: '8px 16px', cursor: 'pointer', background: msg.isRead ? '#fff' : '#f0f7ff' }}
            actions={[
              <Button
                key="read"
                type="text"
                size="small"
                icon={<CheckOutlined />}
                onClick={() => handleMarkAsRead(msg.id)}
                disabled={msg.isRead === 1}
              />,
              <Button
                key="delete"
                type="text"
                size="small"
                danger
                icon={<DeleteOutlined />}
                onClick={() => handleDelete(msg.id)}
              />,
            ]}
          >
            <List.Item.Meta
              title={<span style={{ fontWeight: msg.isRead ? 400 : 600 }}>{msg.title}</span>}
              description={
                <div>
                  <div style={{ color: '#666', fontSize: 12 }}>{msg.content}</div>
                  <div style={{ color: '#999', fontSize: 11, marginTop: 4 }}>{msg.createdAt}</div>
                </div>
              }
            />
          </List.Item>
        )}
      />
    </div>
  )

  return (
    <Dropdown
      dropdownRender={() => dropdownContent}
      trigger={['click']}
      open={open}
      onOpenChange={setOpen}
    >
      <Badge count={unreadCount} size="small">
        <BellOutlined style={{ fontSize: 20, cursor: 'pointer', color: '#666' }} />
      </Badge>
    </Dropdown>
  )
}
