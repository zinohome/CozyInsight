import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Form, Input, Button, message, Avatar, Divider, Descriptions } from 'antd'
import { UserOutlined, LockOutlined, ArrowLeftOutlined } from '@ant-design/icons'
import { userAPI } from '@/api/user'
import { useAuthStore } from '@/store/auth'
import type { UserInfo } from '@/types/auth'

export default function ProfilePage() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const [form] = Form.useForm()
  const [pwdForm] = Form.useForm()
  const [userInfo, setUserInfo] = useState<UserInfo | null>(null)
  const [loading, setLoading] = useState(false)
  const [pwdLoading, setPwdLoading] = useState(false)

  useEffect(() => {
    if (user) {
      setUserInfo(user)
      form.setFieldsValue({
        username: user.username,
        email: user.email,
        nickName: user.nickName,
      })
    }
  }, [user, form])

  const handleUpdate = async (values: { email: string; nickName: string }) => {
    if (!userInfo) return
    setLoading(true)
    try {
      await userAPI.update(userInfo.id, values)
      message.success('更新成功')
      // Refresh user info
      const fresh = await userAPI.profile()
      useAuthStore.getState().setUser(fresh)
      setUserInfo(fresh)
    } catch {
      message.error('更新失败')
    } finally {
      setLoading(false)
    }
  }

  const handleChangePassword = async (values: {
    oldPassword: string
    newPassword: string
    confirmPassword: string
  }) => {
    if (values.newPassword !== values.confirmPassword) {
      message.error('两次输入的新密码不一致')
      return
    }
    setPwdLoading(true)
    try {
      await userAPI.changePassword({
        oldPassword: values.oldPassword,
        newPassword: values.newPassword,
      })
      message.success('密码修改成功')
      pwdForm.resetFields()
    } catch {
      message.error('密码修改失败')
    } finally {
      setPwdLoading(false)
    }
  }

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/')}>
          返回工作台
        </Button>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '300px 1fr', gap: 24 }}>
        <Card>
          <div style={{ textAlign: 'center', padding: '24px 0' }}>
            <Avatar size={80} icon={<UserOutlined />} />
            <h3 style={{ marginTop: 16, marginBottom: 4 }}>
              {userInfo?.nickName || userInfo?.username || '用户'}
            </h3>
            <p style={{ color: '#999', margin: 0 }}>{userInfo?.email || '-'}</p>
          </div>
          <Divider />
          <Descriptions column={1} size="small">
            <Descriptions.Item label="用户名">{userInfo?.username || '-'}</Descriptions.Item>
            <Descriptions.Item label="角色">{userInfo?.isAdmin ? '管理员' : '普通用户'}</Descriptions.Item>
          </Descriptions>
        </Card>

        <div>
          <Card title="基本信息" style={{ marginBottom: 24 }}>
            <Form form={form} onFinish={handleUpdate} layout="vertical">
              <Form.Item name="username" label="用户名">
                <Input disabled />
              </Form.Item>
              <Form.Item
                name="email"
                label="邮箱"
                rules={[{ type: 'email', message: '请输入有效邮箱' }]}
              >
                <Input />
              </Form.Item>
              <Form.Item name="nickName" label="昵称">
                <Input />
              </Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading}>
                  保存
                </Button>
              </Form.Item>
            </Form>
          </Card>

          <Card title="修改密码">
            <Form form={pwdForm} onFinish={handleChangePassword} layout="vertical">
              <Form.Item
                name="oldPassword"
                label="当前密码"
                rules={[{ required: true, message: '请输入当前密码' }]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder="当前密码" />
              </Form.Item>
              <Form.Item
                name="newPassword"
                label="新密码"
                rules={[
                  { required: true, message: '请输入新密码' },
                  { min: 6, message: '密码至少6位' },
                ]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder="新密码" />
              </Form.Item>
              <Form.Item
                name="confirmPassword"
                label="确认新密码"
                rules={[{ required: true, message: '请确认新密码' }]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder="确认新密码" />
              </Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit" loading={pwdLoading}>
                  修改密码
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </div>
      </div>
    </div>
  )
}
