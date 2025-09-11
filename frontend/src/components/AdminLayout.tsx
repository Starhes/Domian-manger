import { Layout as AntLayout, Menu, Avatar, Dropdown, Space, Button } from 'antd'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
	DashboardOutlined,
	UserOutlined,
	GlobalOutlined,
	SettingOutlined,
	LogoutOutlined,
	HomeOutlined,
	MenuOutlined,
	MailOutlined
} from '@ant-design/icons'
import { useAuthStore } from '../stores/authStore'
import { useState } from 'react'

const { Header, Sider, Content } = AntLayout

const AdminLayout = () => {
	const navigate = useNavigate()
	const location = useLocation()
	const { user, logout } = useAuthStore()
	const [collapsed, setCollapsed] = useState(false)
	const [isMobile, setIsMobile] = useState(false)

	const menuItems = [
		{
			key: '/admin',
			icon: <DashboardOutlined />,
			label: '管理仪表盘',
		},
		{
			key: '/admin/users',
			icon: <UserOutlined />,
			label: '用户管理',
		},
		{
			key: '/admin/domains',
			icon: <GlobalOutlined />,
			label: '域名管理',
		},
		{
			key: '/admin/providers',
			icon: <SettingOutlined />,
			label: 'DNS服务商',
		},
		{
			key: '/admin/smtp-configs',
			icon: <MailOutlined />,
			label: 'SMTP配置',
		},
	]

	const userMenuItems = [
		{
			key: 'home',
			icon: <HomeOutlined />,
			label: '返回首页',
			onClick: () => navigate('/'),
		},
		{
			key: 'profile',
			icon: <UserOutlined />,
			label: '个人资料',
			onClick: () => navigate('/profile'),
		},
		{
			type: 'divider' as const,
		},
		{
			key: 'logout',
			icon: <LogoutOutlined />,
			label: '退出登录',
			onClick: logout,
		},
	]

	return (
		<AntLayout style={{ minHeight: '100vh' }}>
			<Sider
				collapsible
				collapsed={collapsed}
				onCollapse={setCollapsed}
				breakpoint="lg"
				onBreakpoint={setIsMobile}
				collapsedWidth="0"
				trigger={null}
				width={250}
				style={{
					overflow: 'auto',
					height: '100vh',
					position: 'fixed',
					left: 0,
					top: 0,
					bottom: 0,
					zIndex: 10,
				}}
			>
				<div style={{
					height: 32,
					margin: 16,
					background: 'rgba(255, 255, 255, 0.3)',
					borderRadius: 6,
					display: 'flex',
					alignItems: 'center',
					justifyContent: 'center',
					color: 'white',
					fontSize: 16,
					fontWeight: 'bold'
				}}>
					{isMobile || !collapsed ? '管理后台' : '管理'}
				</div>

				<Menu
					theme="dark"
					mode="inline"
					selectedKeys={[location.pathname]}
					items={menuItems}
					onClick={({ key }) => navigate(key)}
				/>
			</Sider>

			<AntLayout style={{ marginLeft: isMobile ? 0 : (collapsed ? 80 : 250), transition: 'margin-left 0.2s' }}>
				<Header style={{
					padding: '0 24px',
					background: '#fff',
					display: 'flex',
					justifyContent: 'space-between',
					alignItems: 'center'
				}}>
					<Button
						type="text"
						icon={<MenuOutlined />}
						onClick={() => setCollapsed(!collapsed)}
						style={{
							fontSize: '16px',
						}}
					/>
					<div style={{ flexGrow: 1 }} />
					<Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
						<Space style={{ cursor: 'pointer' }}>
							<Avatar icon={<UserOutlined />} />
							<span>管理员: {user?.email}</span>
						</Space>
					</Dropdown>
				</Header>

				<Content style={{ margin: '24px 16px 0', overflow: 'initial' }}>
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

export default AdminLayout
