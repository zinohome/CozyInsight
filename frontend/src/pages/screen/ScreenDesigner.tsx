import { useParams } from 'react-router-dom'
import { message } from 'antd'
import { useState, useEffect, useCallback } from 'react'
import { dashboardAPI } from '@/api/dashboard'

export default function ScreenDesigner() {
  const { id } = useParams<{ id: string }>()
  const [dashboard, setDashboard] = useState<{ title: string } | null>(null)

  const fetchDashboard = useCallback(async () => {
    if (!id) return
    try {
      const d = await dashboardAPI.get(Number(id))
      if (d.type !== 'screen') {
        message.error('该资源不是数据大屏')
        return
      }
      setDashboard(d)
    } catch {
      message.error('加载数据大屏失败')
    }
  }, [id])

  useEffect(() => {
    fetchDashboard()
  }, [fetchDashboard])

  return (
    <div style={{ height: 'calc(100vh - 64px)' }}>
      <h3>{dashboard?.title || '数据大屏设计器'}</h3>
      <p>设计中...</p>
    </div>
  )
}
