import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { shareAPI } from '@/api/share'
import type { Dashboard } from '@/types/dashboard'

export default function ShareView() {
  const { token } = useParams<{ token: string }>()
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!token) return
    shareAPI.get(token)
      .then(res => {
        if (res.code === 200) {
          setDashboard(res.data)
        } else {
          setError(res.error || '分享链接无效或已过期')
        }
      })
      .catch(() => setError('加载失败'))
      .finally(() => setLoading(false))
  }, [token])

  if (loading) return <div style={{ padding: 40, textAlign: 'center' }}>加载中...</div>
  if (error) return <div style={{ padding: 40, textAlign: 'center', color: '#999' }}>{error}</div>
  if (!dashboard) return null

  return (
    <div style={{ padding: 24, maxWidth: 1200, margin: '0 auto' }}>
      <h2>{dashboard.title}</h2>
      <p style={{ color: '#666' }}>通过分享链接查看</p>
    </div>
  )
}
