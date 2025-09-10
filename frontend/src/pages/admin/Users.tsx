import { useEffect, useState } from 'react'
import { 
  Table, 
  Button, 
  Modal, 
  Form, 
  Input, 
  Switch, 
  Space, 
  Popconfirm, 
  message,
  Typography,
  Tag,
  Card
} from 'antd'
import { EditOutlined, DeleteOutlined, SearchOutlined } from '@ant-design/icons'
import api from '../../utils/api'

const { Title } = Typography
const { Search } = Input

interface User {
  id: number
  email: string
  is_active: boolean
  is_admin: boolean
  created_at: string
}

const AdminUsers = () => {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  })
  const [searchText, setSearchText] = useState('')
  const [form] = Form.useForm()

  useEffect(() => {
    fetchUsers()
  }, [pagination.current, pagination.pageSize, searchText])

  const fetchUsers = async () => {
    setLoading(true)
    try {
      const response = await api.get('/api/admin/users', {
        params: {
          page: pagination.current,
          pageSize: pagination.pageSize,
          search: searchText,
        }
      })
      
      const { users, pagination: paginationData } = response.data
      setUsers(users || [])
      setPagination(prev => ({
        ...prev,
        total: paginationData.total,
      }))
    } catch (error) {
      message.error('获取用户列表失败')
    } finally {
      setLoading(false)
    }
  }

  const handleEdit = (user: User) => {
    setEditingUser(user)
    form.setFieldsValue({
      email: user.email,
      is_active: user.is_active,
      is_admin: user.is_admin,
    })
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await api.delete(`/api/admin/users/${id}`)
      message.success('用户删除成功')
      fetchUsers()
    } catch (error: any) {
      message.error(error.response?.data?.error || '删除失败')
    }
  }

  const handleSubmit = async (values: any) => {
    if (!editingUser) return

    try {
      await api.put(`/api/admin/users/${editingUser.id}`, values)
      message.success('用户更新成功')
      setModalVisible(false)
      fetchUsers()
    } catch (error: any) {
      message.error(error.response?.data?.error || '更新失败')
    }
  }

  const handleSearch = (value: string) => {
    setSearchText(value)
    setPagination(prev => ({ ...prev, current: 1 }))
  }

  const handleTableChange = (paginationData: any) => {
    setPagination(prev => ({
      ...prev,
      current: paginationData.current,
      pageSize: paginationData.pageSize,
    }))
  }

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'green' : 'red'}>
          {isActive ? '已激活' : '未激活'}
        </Tag>
      ),
    },
    {
      title: '用户类型',
      dataIndex: 'is_admin',
      key: 'is_admin',
      render: (isAdmin: boolean) => (
        <Tag color={isAdmin ? 'purple' : 'default'}>
          {isAdmin ? '管理员' : '普通用户'}
        </Tag>
      ),
    },
    {
      title: '注册时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleString('zh-CN'),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: User) => (
        <Space>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个用户吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center',
        marginBottom: 24 
      }}>
        <Title level={2} style={{ margin: 0 }}>
          用户管理
        </Title>
        <Search
          placeholder="搜索用户邮箱"
          allowClear
          enterButton={<SearchOutlined />}
          style={{ width: 300 }}
          onSearch={handleSearch}
        />
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={users}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条记录`,
          }}
          onChange={handleTableChange}
        />
      </Card>

      <Modal
        title="编辑用户"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        okText="确定"
        cancelText="取消"
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            name="email"
            label="邮箱地址"
          >
            <Input disabled />
          </Form.Item>

          <Form.Item
            name="is_active"
            label="账户状态"
            valuePropName="checked"
          >
            <Switch checkedChildren="已激活" unCheckedChildren="未激活" />
          </Form.Item>

          <Form.Item
            name="is_admin"
            label="管理员权限"
            valuePropName="checked"
          >
            <Switch checkedChildren="管理员" unCheckedChildren="普通用户" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default AdminUsers
