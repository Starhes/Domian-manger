import { Typography, Card } from 'antd'

const { Title, Paragraph } = Typography

export default function Register() {
  return (
    <div style={{ 
      display: 'flex', 
      justifyContent: 'center', 
      alignItems: 'center', 
      minHeight: '100vh',
      background: '#f0f2f5'
    }}>
      <Card style={{ width: 400 }}>
        <Title level={3}>注册</Title>
        <Paragraph>注册功能开发中...</Paragraph>
      </Card>
    </div>
  )
}