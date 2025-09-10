import { useState } from 'react'
import { Card, Form, Input, Button, message, Typography, Divider } from 'antd'
import { UserOutlined, MailOutlined, LockOutlined } from '@ant-design/icons'
import { useAuthStore } from '../stores/authStore'
import api from '../utils/api'

const { Title } = Typography

const Profile = () => {
  const { user } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [passwordForm] = Form.useForm()

  const handlePasswordChange = async (values: { 
    currentPassword: string
    newPassword: string
    confirmPassword: string 
  }) => {
    setLoading(true)
    try {
      await api.put('/api/profile/password', {
        current_password: values.currentPassword,
        new_password: values.newPassword,
      })
      message.success('密码修改成功')
      passwordForm.resetFields()
    } catch (error: any) {
      message.error(error.response?.data?.error || '密码修改失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ maxWidth: 800, margin: '0 auto' }}>
      <Title level={2} style={{ marginBottom: 32 }}>
        个人资料
      </Title>

      <Card title="基本信息" style={{ marginBottom: 24 }}>
        <div style={{ padding: '16px 0' }}>
          <div style={{ 
            display: 'flex', 
            alignItems: 'center', 
            marginBottom: 16 
          }}>
            <UserOutlined style={{ marginRight: 8, color: '#1890ff' }} />
            <strong>用户ID：</strong>
            <span style={{ marginLeft: 8 }}>{user?.id}</span>
          </div>
          
          <div style={{ 
            display: 'flex', 
            alignItems: 'center', 
            marginBottom: 16 
          }}>
            <MailOutlined style={{ marginRight: 8, color: '#1890ff' }} />
            <strong>邮箱地址：</strong>
            <span style={{ marginLeft: 8 }}>{user?.email}</span>
          </div>

          <div style={{ 
            display: 'flex', 
            alignItems: 'center', 
            marginBottom: 16 
          }}>
            <strong>账户状态：</strong>
            <span style={{ 
              marginLeft: 8,
              color: user?.is_active ? '#52c41a' : '#ff4d4f'
            }}>
              {user?.is_active ? '已激活' : '未激活'}
            </span>
          </div>

          <div style={{ 
            display: 'flex', 
            alignItems: 'center', 
            marginBottom: 16 
          }}>
            <strong>用户类型：</strong>
            <span style={{ 
              marginLeft: 8,
              color: user?.is_admin ? '#722ed1' : '#666'
            }}>
              {user?.is_admin ? '管理员' : '普通用户'}
            </span>
          </div>

          <div style={{ 
            display: 'flex', 
            alignItems: 'center' 
          }}>
            <strong>注册时间：</strong>
            <span style={{ marginLeft: 8 }}>
              {user?.created_at ? new Date(user.created_at).toLocaleString('zh-CN') : '-'}
            </span>
          </div>
        </div>
      </Card>

      <Card title="修改密码">
        <Form
          form={passwordForm}
          layout="vertical"
          onFinish={handlePasswordChange}
          style={{ maxWidth: 400 }}
        >
          <Form.Item
            name="currentPassword"
            label="当前密码"
            rules={[{ required: true, message: '请输入当前密码' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="请输入当前密码"
            />
          </Form.Item>

          <Form.Item
            name="newPassword"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码至少6位' }
            ]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="请输入新密码"
            />
          </Form.Item>

          <Form.Item
            name="confirmPassword"
            label="确认新密码"
            dependencies={['newPassword']}
            rules={[
              { required: true, message: '请确认新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('newPassword') === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'))
                },
              }),
            ]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="请确认新密码"
            />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
            >
              修改密码
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}

export default Profile
