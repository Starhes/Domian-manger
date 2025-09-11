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
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import api from '../../utils/api'

const { Title } = Typography

interface Domain {
  id: number
  name: string
  is_active: boolean
  created_at: string
}

const AdminDomains = () => {
  const [domains, setDomains] = useState<Domain[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingDomain, setEditingDomain] = useState<Domain | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchDomains()
  }, [])

  const fetchDomains = async () => {
    setLoading(true)
    try {
      const response = await api.get('/api/admin/domains')
      setDomains(response.data.domains || [])
    } catch (error) {
      message.error('获取域名列表失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = () => {
    setEditingDomain(null)
    form.resetFields()
    form.setFieldsValue({ is_active: true })
    setModalVisible(true)
  }

  const handleEdit = (domain: Domain) => {
    setEditingDomain(domain)
    form.setFieldsValue({
      name: domain.name,
      is_active: domain.is_active,
    })
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await api.delete(`/api/admin/domains/${id}`)
      message.success('域名删除成功')
      fetchDomains()
    } catch (error: any) {
      message.error(error.response?.data?.error || '删除失败')
    }
  }

  const handleSubmit = async (values: any) => {
    try {
      if (editingDomain) {
        await api.put(`/api/admin/domains/${editingDomain.id}`, values)
        message.success('域名更新成功')
      } else {
        await api.post('/api/admin/domains', values)
        message.success('域名创建成功')
      }
      setModalVisible(false)
      fetchDomains()
    } catch (error: any) {
      message.error(error.response?.data?.error || '操作失败')
    }
  }

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '域名',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <strong>{name}</strong>
      ),
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'green' : 'red'}>
          {isActive ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleString('zh-CN'),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: Domain) => (
        <Space>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个域名吗？"
            description="删除域名前请确保没有关联的DNS记录"
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
          域名管理
        </Title>
        <Button 
          type="primary" 
          icon={<PlusOutlined />}
          onClick={handleCreate}
        >
          添加域名
        </Button>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={domains}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个域名`,
          }}
          scroll={{ x: 'max-content' }}
        />
      </Card>

      <Modal
        title={editingDomain ? '编辑域名' : '添加域名'}
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
            name="name"
            label="域名"
            rules={[
              { required: true, message: '请输入域名' },
              { 
                pattern: /^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9]\.[a-zA-Z]{2,}$/,
                message: '请输入有效的域名格式'
              }
            ]}
          >
            <Input 
              placeholder="例如: example.com" 
              disabled={!!editingDomain}
            />
          </Form.Item>

          <Form.Item
            name="is_active"
            label="状态"
            valuePropName="checked"
          >
            <Switch checkedChildren="启用" unCheckedChildren="禁用" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default AdminDomains
