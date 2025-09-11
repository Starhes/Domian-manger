import { Form, Input, Button, Card, message, Tooltip } from 'antd'
import { UserOutlined, LockOutlined, MailOutlined, InfoCircleOutlined } from '@ant-design/icons'
import { Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'

const Register = () => {
  const [form] = Form.useForm()
  const navigate = useNavigate()
  const { register, isLoading } = useAuthStore()

  const onFinish = async (values: { 
    email: string; 
    password: string; 
    confirmPassword: string;
    nickname: string;
  }) => {
    try {
      await register(values.email, values.password, values.confirmPassword, values.nickname)
      message.success('注册成功，请检查邮箱激活账户')
      navigate('/login')
    } catch (error: any) {
      message.error(error.response?.data?.error || '注册失败')
    }
  }

  // 密码强度验证
  const validatePassword = (_: any, value: string) => {
    if (!value) {
      return Promise.reject(new Error('请输入密码'))
    }
    
    if (value.length < 8) {
      return Promise.reject(new Error('密码长度至少8位'))
    }
    
    if (value.length > 100) {
      return Promise.reject(new Error('密码长度不能超过100位'))
    }
    
    if (!/[a-zA-Z]/.test(value)) {
      return Promise.reject(new Error('密码必须包含字母'))
    }
    
    if (!/[0-9]/.test(value)) {
      return Promise.reject(new Error('密码必须包含数字'))
    }
    
    if (value.length < 12 && !/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]/.test(value)) {
      return Promise.reject(new Error('密码长度小于12位时必须包含特殊字符'))
    }
    
    return Promise.resolve()
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
            域名管理系统
          </h1>
          <p style={{ color: '#666', marginTop: 8 }}>
            创建您的账户
          </p>
        </div>

        <Form
          form={form}
          name="register"
          onFinish={onFinish}
          layout="vertical"
          size="large"
        >
          <Form.Item
            label="邮箱地址"
            name="email"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' },
              { max: 255, message: '邮箱长度不能超过255位' }
            ]}
          >
            <Input
              prefix={<MailOutlined />}
              placeholder="请输入邮箱地址"
            />
          </Form.Item>

          <Form.Item
            label="昵称（可选）"
            name="nickname"
            rules={[
              { max: 100, message: '昵称长度不能超过100位' }
            ]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="请输入昵称（可选）"
            />
          </Form.Item>

          <Form.Item
            label={
              <span>
                密码&nbsp;
                <Tooltip title={
                  <div>
                    <div>• 长度至少8位，最多100位</div>
                    <div>• 必须包含字母和数字</div>
                    <div>• 长度小于12位时必须包含特殊字符</div>
                  </div>
                }>
                  <InfoCircleOutlined style={{ color: '#1890ff' }} />
                </Tooltip>
              </span>
            }
            name="password"
            rules={[{ validator: validatePassword }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="请输入密码"
            />
          </Form.Item>

          <Form.Item
            label="确认密码"
            name="confirmPassword"
            dependencies={['password']}
            rules={[
              { required: true, message: '请确认密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'))
                },
              }),
            ]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="请再次输入密码"
            />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={isLoading}
              style={{ width: '100%' }}
            >
              注册
            </Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: 'center' }}>
          <span>已有账户？ </span>
          <Link to="/login">立即登录</Link>
        </div>
      </Card>
    </div>
  )
}

export default Register
