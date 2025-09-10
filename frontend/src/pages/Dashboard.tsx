import { useEffect, useState } from 'react'
import { Card, Row, Col, Statistic, Typography, Spin } from 'antd'
import { 
  UserOutlined, 
  GlobalOutlined, 
  ClockCircleOutlined,
  CheckCircleOutlined 
} from '@ant-design/icons'
import api from '../utils/api'

const { Title } = Typography

interface Stats {
  totalRecords: number
  activeRecords: number
  availableDomains: number
  recentActivity: number
}

const Dashboard = () => {
  const [stats, setStats] = useState<Stats | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchStats()
  }, [])

  const fetchStats = async () => {
    try {
      // 获取用户的DNS记录
      const recordsResponse = await api.get('/api/dns-records')
      const records = recordsResponse.data.records || []

      // 获取可用域名
      const domainsResponse = await api.get('/api/domains')
      const domains = domainsResponse.data.domains || []

      const stats: Stats = {
        totalRecords: records.length,
        activeRecords: records.filter((r: any) => r.status === 'active').length,
        availableDomains: domains.length,
        recentActivity: records.filter((r: any) => {
          const createdAt = new Date(r.created_at)
          const weekAgo = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000)
          return createdAt > weekAgo
        }).length
      }

      setStats(stats)
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
        仪表盘
      </Title>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总DNS记录"
              value={stats?.totalRecords || 0}
              prefix={<GlobalOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="活跃记录"
              value={stats?.activeRecords || 0}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="可用域名"
              value={stats?.availableDomains || 0}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="本周活动"
              value={stats?.recentActivity || 0}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={12}>
          <Card title="快速操作" style={{ height: 300 }}>
            <div style={{ 
              display: 'flex', 
              flexDirection: 'column', 
              gap: '16px',
              padding: '20px 0'
            }}>
              <div>
                <h4>管理DNS记录</h4>
                <p style={{ color: '#666' }}>
                  添加、编辑或删除您的DNS记录
                </p>
              </div>
              <div>
                <h4>查看域名</h4>
                <p style={{ color: '#666' }}>
                  浏览所有可用的域名选项
                </p>
              </div>
              <div>
                <h4>更新资料</h4>
                <p style={{ color: '#666' }}>
                  修改您的个人账户信息
                </p>
              </div>
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title="系统状态" style={{ height: 300 }}>
            <div style={{ 
              display: 'flex', 
              flexDirection: 'column', 
              gap: '16px',
              padding: '20px 0'
            }}>
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between',
                alignItems: 'center'
              }}>
                <span>DNS服务</span>
                <span style={{ color: '#52c41a' }}>正常</span>
              </div>
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between',
                alignItems: 'center'
              }}>
                <span>API服务</span>
                <span style={{ color: '#52c41a' }}>正常</span>
              </div>
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between',
                alignItems: 'center'
              }}>
                <span>数据库</span>
                <span style={{ color: '#52c41a' }}>正常</span>
              </div>
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between',
                alignItems: 'center'
              }}>
                <span>邮件服务</span>
                <span style={{ color: '#fa8c16' }}>维护中</span>
              </div>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default Dashboard
