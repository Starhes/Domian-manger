import { useEffect, useState } from 'react'
import { Card, Row, Col, Statistic, Typography, Spin, Table } from 'antd'
import { 
  UserOutlined, 
  GlobalOutlined, 
  SettingOutlined,
  DatabaseOutlined 
} from '@ant-design/icons'
import api from '../../utils/api'

const { Title } = Typography

interface AdminStats {
  users: {
    total: number
    active: number
  }
  domains: {
    total: number
    active: number
  }
  records: number
  providers: number
}

const AdminDashboard = () => {
  const [stats, setStats] = useState<AdminStats | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchStats()
  }, [])

  const fetchStats = async () => {
    try {
      const response = await api.get('/api/admin/stats')
      setStats(response.data.stats)
    } catch (error) {
      console.error('获取统计数据失败:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div>
      <Title level={2} style={{ marginBottom: 32 }}>
        管理仪表盘
      </Title>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总用户数"
              value={stats?.users.total || 0}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
            <div style={{ marginTop: 8, fontSize: 12, color: '#666' }}>
              活跃用户: {stats?.users.active || 0}
            </div>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总域名数"
              value={stats?.domains.total || 0}
              prefix={<GlobalOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
            <div style={{ marginTop: 8, fontSize: 12, color: '#666' }}>
              活跃域名: {stats?.domains.active || 0}
            </div>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="DNS记录数"
              value={stats?.records || 0}
              prefix={<DatabaseOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="DNS服务商"
              value={stats?.providers || 0}
              prefix={<SettingOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={12}>
          <Card title="系统概览" style={{ height: 400 }}>
            <div style={{ 
              display: 'flex', 
              flexDirection: 'column', 
              gap: '20px',
              padding: '20px 0'
            }}>
              <div>
                <h4>用户管理</h4>
                <p style={{ color: '#666' }}>
                  管理系统中的所有用户账户，包括激活、禁用和权限设置
                </p>
              </div>
              <div>
                <h4>域名管理</h4>
                <p style={{ color: '#666' }}>
                  添加和管理可供用户使用的域名资源
                </p>
              </div>
              <div>
                <h4>DNS服务商</h4>
                <p style={{ color: '#666' }}>
                  配置和管理DNS服务商的API凭证和设置
                </p>
              </div>
              <div>
                <h4>系统监控</h4>
                <p style={{ color: '#666' }}>
                  监控系统运行状态和性能指标
                </p>
              </div>
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title="最近活动" style={{ height: 400 }}>
            <div style={{ 
              display: 'flex', 
              flexDirection: 'column', 
              gap: '16px',
              padding: '20px 0'
            }}>
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '8px 0',
                borderBottom: '1px solid #f0f0f0'
              }}>
                <span>新用户注册</span>
                <span style={{ color: '#666' }}>2小时前</span>
              </div>
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '8px 0',
                borderBottom: '1px solid #f0f0f0'
              }}>
                <span>DNS记录更新</span>
                <span style={{ color: '#666' }}>5小时前</span>
              </div>
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '8px 0',
                borderBottom: '1px solid #f0f0f0'
              }}>
                <span>域名添加</span>
                <span style={{ color: '#666' }}>1天前</span>
              </div>
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '8px 0',
                borderBottom: '1px solid #f0f0f0'
              }}>
                <span>系统配置更新</span>
                <span style={{ color: '#666' }}>2天前</span>
              </div>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default AdminDashboard
