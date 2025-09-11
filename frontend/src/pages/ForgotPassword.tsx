import { Form, Input, Button, Card, message } from 'antd'
import { MailOutlined, ArrowLeftOutlined } from '@ant-design/icons'
import { Link } from 'react-router-dom'
import { useState } from 'react'
import api from '../utils/api'

const ForgotPassword = () => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [sent, setSent] = useState(false)

  const onFinish = async (values: { email: string }) => {
    setLoading(true)
    try {
      await api.post('/api/forgot-password', { email: values.email })
      setSent(true)
      message.success('如果该邮箱存在，重置链接已发送')
    } catch (error: any) {
      message.error(error.response?.data?.error || '发送失败')
    } finally {
      setLoading(false)
    }
  }

  if (sent) {
    return (
      <div style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        padding: '0 16px'
      }}>
        <Card style={{
          width: '100%',
          maxWidth: 400,
          boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
          textAlign: 'center'
        }}>
          <div style={{ padding: '20px 0' }}>
            <div style={{ fontSize: '48px', marginBottom: '20px' }}>📧</div>
            <h2 style={{ color: '#1890ff', marginBottom: '16px' }}>邮件已发送</h2>
            <p style={{ color: '#666', marginBottom: '24px' }}>
              如果该邮箱存在于我们的系统中，您将收到一封包含密码重置链接的邮件。
            </p>
            <p style={{ color: '#999', fontSize: '14px', marginBottom: '24px' }}>
              没有收到邮件？请检查垃圾邮件文件夹，或等待几分钟后重试。
            </p>
            <Button 
              type="primary" 
              icon={<ArrowLeftOutlined />}
              onClick={() => setSent(false)}
            >
              返回重新发送
            </Button>
            <div style={{ marginTop: '16px' }}>
              <Link to="/login">返回登录</Link>
            </div>
          </div>
        </Card>
      </div>
    )
  }

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      padding: '0 16px'
    }}>
      <Card style={{
        width: '100%',
        maxWidth: 400,
        boxShadow: '0 4px 12px rgba(0,0,0,0.15)'
      }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <h1 style={{ fontSize: 24, margin: 0, color: '#1890ff' }}>
            🔑 找回密码
          </h1>
          <p style={{ color: '#666', marginTop: 8 }}>
            输入您的邮箱地址，我们将发送重置链接
          </p>
        </div>

        <Form
          form={form}
          name="forgot-password"
          onFinish={onFinish}
          layout="vertical"
          size="large"
        >
          <Form.Item
            name="email"
            label="邮箱地址"
            rules={[
              { required: true, message: '请输入邮箱地址' },
              { type: 'email', message: '请输入有效的邮箱地址' }
            ]}
          >
            <Input
              prefix={<MailOutlined />}
              placeholder="请输入您的注册邮箱"
            />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              style={{ width: '100%' }}
            >
              发送重置链接
            </Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: 'center' }}>
          <Link to="/login">
            <ArrowLeftOutlined /> 返回登录
          </Link>
        </div>
      </Card>
    </div>
  )
}

export default ForgotPassword
