import { useEffect, useState } from 'react'
import { 
  Table, 
  Button, 
  Modal, 
  Form, 
  Input, 
  Select, 
  Space, 
  Popconfirm, 
  message,
  Typography,
  Tag,
  Card,
  Upload,
  Checkbox,
  InputNumber,
  Tooltip,
  Divider,
  Row,
  Col,
  Statistic
} from 'antd'
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  DownloadOutlined, 
  UploadOutlined,
  InfoCircleOutlined,
  AppstoreAddOutlined
} from '@ant-design/icons'
import api from '../utils/api'

const { Title, Text } = Typography
const { Option } = Select
const { TextArea } = Input

interface DNSRecord {
  id: number
  subdomain: string
  type: string
  value: string
  ttl: number
  priority: number
  weight: number
  port: number
  comment: string
  created_at: string
  domain: {
    id: number
    name: string
  }
}

interface Domain {
  id: number
  name: string
  is_active: boolean
}

interface DNSRecordType {
  value: string
  label: string
  description: string
  fields: string[]
}

interface TTLOption {
  value: number
  label: string
  description: string
}

interface RecordStats {
  total_records: number
  quota: number
  quota_used: number
  type_stats: Array<{ type: string; count: number }>
  domain_stats: Array<{ domain_name: string; count: number }>
}

const DNSRecords = () => {
  const [records, setRecords] = useState<DNSRecord[]>([])
  const [domains, setDomains] = useState<Domain[]>([])
  const [recordTypes, setRecordTypes] = useState<DNSRecordType[]>([])
  const [ttlOptions, setTTLOptions] = useState<TTLOption[]>([])
  const [stats, setStats] = useState<RecordStats | null>(null)
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [batchModalVisible, setBatchModalVisible] = useState(false)
  const [importModalVisible, setImportModalVisible] = useState(false)
  const [editingRecord, setEditingRecord] = useState<DNSRecord | null>(null)
  const [selectedRecordType, setSelectedRecordType] = useState<string>('')
  const [form] = Form.useForm()
  const [batchForm] = Form.useForm()
  const [importForm] = Form.useForm()

  useEffect(() => {
    fetchRecords()
    fetchDomains()
    fetchRecordTypes()
    fetchTTLOptions()
    fetchStats()
  }, [])

  const fetchRecords = async () => {
    setLoading(true)
    try {
      const response = await api.get('/api/dns-records')
      setRecords(response.data.records || [])
    } catch (error) {
      message.error('获取DNS记录失败')
    } finally {
      setLoading(false)
    }
  }

  const fetchDomains = async () => {
    try {
      const response = await api.get('/api/domains')
      setDomains(response.data.domains || [])
    } catch (error) {
      message.error('获取域名列表失败')
    }
  }

  const fetchRecordTypes = async () => {
    try {
      const response = await api.get('/api/dns-records/types')
      setRecordTypes(response.data.types || [])
    } catch (error) {
      console.error('获取记录类型失败:', error)
    }
  }

  const fetchTTLOptions = async () => {
    try {
      const response = await api.get('/api/dns-records/ttl-options')
      setTTLOptions(response.data.options || [])
    } catch (error) {
      console.error('获取TTL选项失败:', error)
    }
  }

  const fetchStats = async () => {
    try {
      const response = await api.get('/api/dns-records/stats')
      setStats(response.data)
    } catch (error) {
      console.error('获取统计信息失败:', error)
    }
  }

  const handleCreate = () => {
    setEditingRecord(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: DNSRecord) => {
    setEditingRecord(record)
    form.setFieldsValue({
      domain_id: record.domain.id,
      subdomain: record.subdomain,
      type: record.type,
      value: record.value,
      ttl: record.ttl,
    })
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await api.delete(`/api/dns-records/${id}`)
      message.success('DNS记录删除成功')
      fetchRecords()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleSubmit = async (values: any) => {
    try {
      if (editingRecord) {
        await api.put(`/api/dns-records/${editingRecord.id}`, values)
        message.success('DNS记录更新成功')
      } else {
        await api.post('/api/dns-records', values)
        message.success('DNS记录创建成功')
      }
      setModalVisible(false)
      fetchRecords()
    } catch (error: any) {
      message.error(error.response?.data?.error || '操作失败')
    }
  }

  const columns = [
    {
      title: '子域名',
      dataIndex: 'subdomain',
      key: 'subdomain',
      render: (subdomain: string, record: DNSRecord) => (
        <strong>{subdomain}.{record.domain.name}</strong>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => (
        <Tag color={
          type === 'A' ? 'blue' :
          type === 'CNAME' ? 'green' :
          type === 'TXT' ? 'orange' : 'default'
        }>
          {type}
        </Tag>
      ),
    },
    {
      title: '记录值',
      dataIndex: 'value',
      key: 'value',
      ellipsis: true,
    },
    {
      title: 'TTL',
      dataIndex: 'ttl',
      key: 'ttl',
      render: (ttl: number) => `${ttl}s`,
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
      render: (_: any, record: DNSRecord) => (
        <Space>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这条DNS记录吗？"
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
          DNS记录管理
        </Title>
        <Button 
          type="primary" 
          icon={<PlusOutlined />}
          onClick={handleCreate}
        >
          添加记录
        </Button>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={records}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
          scroll={{ x: 'max-content' }}
        />
      </Card>

      <Modal
        title={editingRecord ? '编辑DNS记录' : '添加DNS记录'}
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
            name="domain_id"
            label="选择域名"
            rules={[{ required: true, message: '请选择域名' }]}
          >
            <Select placeholder="选择一个域名">
              {domains.map(domain => (
                <Option key={domain.id} value={domain.id}>
                  {domain.name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="subdomain"
            label="子域名"
            rules={[{ required: true, message: '请输入子域名' }]}
          >
            <Input placeholder="例如: www, mail, ftp" />
          </Form.Item>

          <Form.Item
            name="type"
            label="记录类型"
            rules={[{ required: true, message: '请选择记录类型' }]}
          >
            <Select placeholder="选择记录类型">
              <Option value="A">A</Option>
              <Option value="CNAME">CNAME</Option>
              <Option value="TXT">TXT</Option>
              <Option value="MX">MX</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="value"
            label="记录值"
            rules={[{ required: true, message: '请输入记录值' }]}
          >
            <Input placeholder="例如: 192.168.1.1 或 example.com" />
          </Form.Item>

          <Form.Item
            name="ttl"
            label="TTL (秒)"
            initialValue={600}
            rules={[{ required: true, message: '请输入TTL值' }]}
          >
            <Select>
              <Option value={300}>5分钟 (300)</Option>
              <Option value={600}>10分钟 (600)</Option>
              <Option value={1800}>30分钟 (1800)</Option>
              <Option value={3600}>1小时 (3600)</Option>
              <Option value={43200}>12小时 (43200)</Option>
              <Option value={86400}>1天 (86400)</Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default DNSRecords
