import { Typography, Row, Col, Card, Statistic } from 'antd'
import { UserOutlined, DnsOutlined, GlobalOutlined } from '@ant-design/icons'

const { Title } = Typography

export default function Dashboard() {
  return (
    <div>
      <Title level={2}>仪表板</Title>
      <Row gutter={16}>
        <Col span={8}>
          <Card>
            <Statistic
              title="域名总数"
              value={12}
              prefix={<GlobalOutlined />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="DNS 记录"
              value={48}
              prefix={<DnsOutlined />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="用户数"
              value={8}
              prefix={<UserOutlined />}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}