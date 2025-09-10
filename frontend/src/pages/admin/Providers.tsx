import { useEffect, useState } from 'react'
import { 
  Table, 
  Button, 
  Modal, 
  Form, 
  Input, 
  Select,
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
const { Option } = Select
const { TextArea } = Input

interface DNSProvider {
  id: number
  name: string
  type: string
  config: string
  is_active: boolean
  created_at: string
}

const AdminProviders = () => {
  const [providers, setProviders] = useState<DNSProvider[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingProvider, setEditingProvider] = useState<DNSProvider | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchProviders()
  }, [])

  const fetchProviders = async () => {
    setLoading(true)
    try {
      const response = await api.get('/api/admin/dns-providers')
      setProviders(response.data.providers || [])
    } catch (error) {
      message.error('获取DNS服务商列表失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = () => {
    setEditingProvider(null)
    form.resetFields()
    form.setFieldsValue({ 
      is_active: true,
      type: 'dnspod'
    })
    setModalVisible(true)
  }

  const handleEdit = (provider: DNSProvider) => {
    setEditingProvider(provider)
    form.setFieldsValue({
      name: provider.name,
      type: provider.type,
      config: provider.config,
      is_active: provider.is_active,
    })
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await api.delete(`/api/admin/dns-providers/${id}`)
      message.success('DNS服务商删除成功')
      fetchProviders()
    } catch (error: any) {
      message.error(error.response?.data?.error || '删除失败')
    }
  }

  const handleSubmit = async (values: any) => {
    try {
      if (editingProvider) {
        await api.put(`/api/admin/dns-providers/${editingProvider.id}`, values)
        message.success('DNS服务商更新成功')
      } else {
        await api.post('/api/admin/dns-providers', values)
        message.success('DNS服务商创建成功')
      }
      setModalVisible(false)
      fetchProviders()
    } catch (error: any) {
      message.error(error.response?.data?.error || '操作失败')
    }
  }

  const getConfigTemplate = (type: string) => {
    switch (type) {
      case 'dnspod':
        return JSON.stringify({
          token: "your_dnspod_token_here"
        }, null, 2)
      case 'cloudflare':
        return JSON.stringify({
          api_token: "your_cloudflare_api_token_here",
          zone_id: "your_zone_id_here"
        }, null, 2)
      default:
        return '{}'
    }
  }

  const handleTypeChange = (type: string) => {
    const template = getConfigTemplate(type)
    form.setFieldsValue({ config: template })
  }

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <strong>{name}</strong>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => (
        <Tag color={
          type === 'dnspod' ? 'blue' :
          type === 'cloudflare' ? 'orange' : 'default'
        }>
          {type.toUpperCase()}
        </Tag>
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
      render: (_: any, record: DNSProvider) => (
        <Space>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个DNS服务商吗？"
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
          DNS服务商管理
        </Title>
        <Button 
          type="primary" 
          icon={<PlusOutlined />}
          onClick={handleCreate}
        >
          添加服务商
        </Button>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={providers}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个DNS服务商`,
          }}
        />
      </Card>

      <Modal
        title={editingProvider ? '编辑DNS服务商' : '添加DNS服务商'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        okText="确定"
        cancelText="取消"
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            name="name"
            label="服务商名称"
            rules={[{ required: true, message: '请输入服务商名称' }]}
          >
            <Input placeholder="例如: DNSPod主服务商" />
          </Form.Item>

          <Form.Item
            name="type"
            label="服务商类型"
            rules={[{ required: true, message: '请选择服务商类型' }]}
          >
            <Select 
              placeholder="选择服务商类型"
              onChange={handleTypeChange}
            >
              <Option value="dnspod">DNSPod</Option>
              <Option value="cloudflare">Cloudflare (暂未支持)</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="config"
            label="配置信息 (JSON格式)"
            rules={[
              { required: true, message: '请输入配置信息' },
              {
                validator: (_, value) => {
                  try {
                    JSON.parse(value)
                    return Promise.resolve()
                  } catch (error) {
                    return Promise.reject(new Error('请输入有效的JSON格式'))
                  }
                }
              }
            ]}
          >
            <TextArea 
              rows={6}
              placeholder="请输入JSON格式的配置信息"
            />
          </Form.Item>

          <Form.Item
            name="is_active"
            label="状态"
            valuePropName="checked"
          >
            <Switch checkedChildren="启用" unCheckedChildren="禁用" />
          </Form.Item>

          <div style={{ 
            background: '#f6f8fa', 
            padding: '12px', 
            borderRadius: '6px',
            marginTop: '16px'
          }}>
            <h4 style={{ margin: '0 0 8px 0', fontSize: '14px' }}>配置说明:</h4>
            <ul style={{ margin: 0, paddingLeft: '20px', fontSize: '12px', color: '#666' }}>
              <li>DNSPod: 需要提供token字段</li>
              <li>Cloudflare: 需要提供api_token和zone_id字段</li>
              <li>请确保API凭证的有效性和权限设置</li>
            </ul>
          </div>
        </Form>
      </Modal>
    </div>
  )
}

export default AdminProviders
