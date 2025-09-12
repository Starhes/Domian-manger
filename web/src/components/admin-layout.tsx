import { Layout as AntLayout, Menu, Avatar, Dropdown, Button } from 'antd'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { 
  DashboardOutlined, 
  UserOutlined, 
  DnsOutlined,
  SettingOutlined,
  MailOutlined,
  LogoutOutlined 
} from '@ant-design/icons'
import { useAuthStore } from '@/stores/auth-store'

const { Header, Sider, Content } = AntLayout

export default function AdminLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuthStore()

  const menuItems = [
    {
      key: '/admin',
      icon: <DashboardOutlined />,
      label: '管理仪表板',
    },
    {
      key: '/admin/users',
      icon: <UserOutlined />,
      label: '用户管理',
    },
    {
      key: '/admin/domains',
      icon: <DnsOutlined />,
      label: '域名管理',
    },
    {
      key: '/admin/providers',
      icon: <SettingOutlined />,
      label: 'DNS 提供商',
    },
    {
      key: '/admin/smtp-configs',
      icon: <MailOutlined />,
      label: 'SMTP 配置',
    },
  ]

  const userMenuItems = [
    {
      key: 'back-to-user',
      label: '返回用户界面',
      onClick: () => navigate('/'),
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: () => {
        logout()
        navigate('/login')
      },
    },
  ]

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Sider
        breakpoint="lg"
        collapsedWidth="0"
        onBreakpoint={(broken) => {
          console.log(broken)
        }}
        onCollapse={(collapsed, type) => {
          console.log(collapsed, type)
        }}
      >
        <div style={{ 
          height: 64, 
          margin: 16, 
          background: 'rgba(255, 255, 255, 0.2)',
          borderRadius: 6,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: 'white',
          fontSize: '18px',
          fontWeight: 'bold'
        }}>
          Admin Panel
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <AntLayout>
        <Header style={{ 
          padding: '0 24px', 
          background: '#fff',
          display: 'flex',
          justifyContent: 'flex-end',
          alignItems: 'center'
        }}>
          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <Button type="text" style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <Avatar icon={<UserOutlined />} />
              <span>{user?.name} (管理员)</span>
            </Button>
          </Dropdown>
        </Header>
        <Content style={{ margin: '24px 16px 0' }}>
          <div style={{ 
            padding: 24, 
            minHeight: 360, 
            background: '#fff',
            borderRadius: 8
          }}>
            <Outlet />
          </div>
        </Content>
      </AntLayout>
    </AntLayout>
  )
}