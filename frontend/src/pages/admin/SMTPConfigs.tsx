import { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Switch,
  message,
  Popconfirm,
  Tag,
  Tooltip,
  InputNumber,
  Row,
  Col
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  CheckOutlined,
  StarOutlined,
  ExperimentOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons'
import api from '../../utils/api'

interface SMTPConfig {
  id: number
  name: string
  host: string
  port: number
  username: string
  from_email: string
  from_name: string
  is_active: boolean
  is_default: boolean
  use_tls: boolean
  description: string
  last_test_at?: string
  test_result?: string
  created_at: string
  updated_at: string
}

const SMTPConfigs = () => {
  const [configs, setConfigs] = useState<SMTPConfig[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [testModalVisible, setTestModalVisible] = useState(false)
  const [editingConfig, setEditingConfig] = useState<SMTPConfig | null>(null)
  const [testingConfig, setTestingConfig] = useState<SMTPConfig | null>(null)
  const [form] = Form.useForm()
  const [testForm] = Form.useForm()

  const fetchConfigs = async () => {
    setLoading(true)
    try {
      const response = await api.get('/api/admin/smtp-configs')
      setConfigs(response.data.configs)
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取SMTP配置列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchConfigs()
  }, [])

  const handleCreate = () => {
    setEditingConfig(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (config: SMTPConfig) => {
    setEditingConfig(config)
    form.setFieldsValue({
      name: config.name,
      host: config.host,
      port: config.port,
      username: config.username,
      from_email: config.from_email,
      from_name: config.from_name,
      use_tls: config.use_tls,
      description: config.description
    })
    setModalVisible(true)
  }

  const handleSubmit = async (values: any) => {
    try {
      if (editingConfig) {
        await api.put(`/api/admin/smtp-configs/${editingConfig.id}`, values)
        message.success('SMTP配置更新成功')
      } else {
        await api.post('/api/admin/smtp-configs', values)
        message.success('SMTP配置创建成功')
      }
      setModalVisible(false)
      fetchConfigs()
    } catch (error: any) {
      message.error(error.response?.data?.error || '保存失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await api.delete(`/api/admin/smtp-configs/${id}`)
      message.success('SMTP配置删除成功')
      fetchConfigs()
    } catch (error: any) {
      message.error(error.response?.data?.error || '删除失败')
    }
  }

  const handleActivate = async (id: number) => {
    try {
      await api.post(`/api/admin/smtp-configs/${id}/activate`)
      message.success('SMTP配置激活成功')
      fetchConfigs()
    } catch (error: any) {
      message.error(error.response?.data?.error || '激活失败')
    }
  }

  const handleSetDefault = async (id: number) => {
    try {
      await api.post(`/api/admin/smtp-configs/${id}/set-default`)
      message.success('默认SMTP配置设置成功')
      fetchConfigs()
    } catch (error: any) {
      message.error(error.response?.data?.error || '设置失败')
    }
  }

  const handleTest = (config: SMTPConfig) => {
    setTestingConfig(config)
    testForm.resetFields()
    setTestModalVisible(true)
  }

  const handleTestSubmit = async (values: { to_email: string }) => {
    if (!testingConfig) return
    
    try {
      await api.post(`/api/admin/smtp-configs/${testingConfig.id}/test`, values)
      message.success('测试邮件发送成功')
      setTestModalVisible(false)
      fetchConfigs() // 刷新列表以显示测试结果
    } catch (error: any) {
      message.error(error.response?.data?.error || '测试失败')
    }
  }

  const columns = [
    {
      title: '配置名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: SMTPConfig) => (
        <div>
          <span>{text}</span>
          <div style={{ fontSize: '12px', color: '#666' }}>
            {record.from_email}
          </div>
        </div>
      ),
    },
    {
      title: 'SMTP服务器',
      key: 'server',
      render: (record: SMTPConfig) => (
        <div>
          <div>{record.host}:{record.port}</div>
          <div style={{ fontSize: '12px', color: '#666' }}>
            TLS: {record.use_tls ? '启用' : '禁用'}
          </div>
        </div>
      ),
    },
    {
      title: '状态',
      key: 'status',
      render: (record: SMTPConfig) => (
        <Space direction="vertical" size="small">
          <Tag color={record.is_active ? 'green' : 'default'}>
            {record.is_active ? '激活' : '未激活'}
          </Tag>
          {record.is_default && (
            <Tag color="blue" icon={<StarOutlined />}>
              默认
            </Tag>
          )}
        </Space>
      ),
    },
    {
      title: '最后测试',
      key: 'test_result',
      render: (record: SMTPConfig) => {
        if (!record.last_test_at) {
          return <span style={{ color: '#999' }}>未测试</span>
        }
        
        const isSuccess = record.test_result?.includes('成功')
        return (
          <div>
            <Tag color={isSuccess ? 'green' : 'red'}>
              {isSuccess ? '成功' : '失败'}
            </Tag>
            <div style={{ fontSize: '12px', color: '#666' }}>
              {new Date(record.last_test_at).toLocaleString()}
            </div>
            {!isSuccess && record.test_result && (
              <Tooltip title={record.test_result}>
                <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />
              </Tooltip>
            )}
          </div>
        )
      },
    },
    {
      title: '操作',
      key: 'actions',
      render: (record: SMTPConfig) => (
        <Space>
          <Button
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          
          <Button
            size="small"
            icon={<ExperimentOutlined />}
            onClick={() => handleTest(record)}
          >
            测试
          </Button>
          
          {!record.is_active && (
            <Button
              size="small"
              type="primary"
              icon={<CheckOutlined />}
              onClick={() => handleActivate(record.id)}
            >
              激活
            </Button>
          )}
          
          {!record.is_default && (
            <Button
              size="small"
              icon={<StarOutlined />}
              onClick={() => handleSetDefault(record.id)}
            >
              设为默认
            </Button>
          )}
          
          <Popconfirm
            title="确定要删除这个SMTP配置吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              size="small"
              danger
              icon={<DeleteOutlined />}
              disabled={record.is_default}
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
      <Card
        title="SMTP配置管理"
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            添加配置
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={configs}
          rowKey="id"
          loading={loading}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>

      {/* 创建/编辑模态框 */}
      <Modal
        title={editingConfig ? '编辑SMTP配置' : '创建SMTP配置'}
        visible={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        width={600}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="配置名称"
                name="name"
                rules={[
                  { required: true, message: '请输入配置名称' },
                  { max: 100, message: '名称长度不能超过100位' }
                ]}
              >
                <Input placeholder="如：Gmail SMTP" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="SMTP服务器"
                name="host"
                rules={[
                  { required: true, message: '请输入SMTP服务器地址' },
                  { max: 255, message: '地址长度不能超过255位' }
                ]}
              >
                <Input placeholder="如：smtp.gmail.com" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                label="端口"
                name="port"
                rules={[
                  { required: true, message: '请输入端口号' },
                  { type: 'number', min: 1, max: 65535, message: '端口号范围1-65535' }
                ]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  placeholder="587"
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="使用TLS"
                name="use_tls"
                valuePropName="checked"
                initialValue={true}
              >
                <Switch />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="用户名"
                name="username"
                rules={[
                  { required: true, message: '请输入用户名' },
                  { max: 255, message: '用户名长度不能超过255位' }
                ]}
              >
                <Input placeholder="SMTP认证用户名" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label={editingConfig ? '密码（留空则不修改）' : '密码'}
                name="password"
                rules={editingConfig ? [] : [
                  { required: true, message: '请输入密码' },
                  { max: 255, message: '密码长度不能超过255位' }
                ]}
              >
                <Input.Password placeholder="SMTP认证密码" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="发件人邮箱"
                name="from_email"
                rules={[
                  { required: true, message: '请输入发件人邮箱' },
                  { type: 'email', message: '请输入有效的邮箱地址' },
                  { max: 255, message: '邮箱长度不能超过255位' }
                ]}
              >
                <Input placeholder="发送邮件时的发件人地址" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="发件人名称"
                name="from_name"
                rules={[
                  { max: 100, message: '名称长度不能超过100位' }
                ]}
              >
                <Input placeholder="发送邮件时的发件人姓名" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="描述"
            name="description"
            rules={[
              { max: 500, message: '描述长度不能超过500位' }
            ]}
          >
            <Input.TextArea
              rows={3}
              placeholder="配置描述（可选）"
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* 测试模态框 */}
      <Modal
        title={`测试SMTP配置: ${testingConfig?.name}`}
        visible={testModalVisible}
        onCancel={() => setTestModalVisible(false)}
        onOk={() => testForm.submit()}
        destroyOnClose
      >
        <Form
          form={testForm}
          layout="vertical"
          onFinish={handleTestSubmit}
        >
          <Form.Item
            label="测试邮箱地址"
            name="to_email"
            rules={[
              { required: true, message: '请输入测试邮箱地址' },
              { type: 'email', message: '请输入有效的邮箱地址' }
            ]}
          >
            <Input placeholder="输入要接收测试邮件的邮箱地址" />
          </Form.Item>
          
          <div style={{ color: '#666', fontSize: '14px' }}>
            点击确定后，系统将向指定邮箱发送一封测试邮件，用于验证SMTP配置是否正常工作。
          </div>
        </Form>
      </Modal>
    </div>
  )
}

export default SMTPConfigs
