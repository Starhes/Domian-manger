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
  Statistic,
  Progress
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
    form.setFieldsValue({ ttl: 600, allow_private_ip: false })
    setSelectedRecordType('')
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
      priority: record.priority,
      weight: record.weight,
      port: record.port,
      comment: record.comment,
    })
    setSelectedRecordType(record.type)
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await api.delete(`/api/dns-records/${id}`)
      message.success('DNS记录删除成功')
      fetchRecords()
      fetchStats()
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
      fetchStats()
    } catch (error: any) {
      message.error(error.response?.data?.error || '操作失败')
    }
  }

  const handleBatchCreate = () => {
    batchForm.resetFields()
    setBatchModalVisible(true)
  }

  const handleBatchSubmit = async (values: any) => {
    try {
      const records = values.records.map((record: any) => ({
        ...record,
        ttl: record.ttl || 600,
      }))
      
      const response = await api.post('/api/dns-records/batch', { records })
      
      if (response.data.success_count > 0) {
        message.success(`成功创建 ${response.data.success_count} 条记录`)
      }
      
      if (response.data.error_count > 0) {
        message.warning(`${response.data.error_count} 条记录创建失败`)
        console.error('批量创建错误:', response.data.errors)
      }
      
      setBatchModalVisible(false)
      fetchRecords()
      fetchStats()
    } catch (error: any) {
      message.error(error.response?.data?.error || '批量创建失败')
    }
  }

  const handleExport = async () => {
    try {
      const response = await api.get('/api/dns-records/export?format=file', {
        responseType: 'blob'
      })
      
      const blob = new Blob([response.data], { type: 'application/json' })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `dns_records_export_${new Date().toISOString().split('T')[0]}.json`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
      
      message.success('DNS记录导出成功')
    } catch (error) {
      message.error('导出失败')
    }
  }

  const handleImport = () => {
    importForm.resetFields()
    setImportModalVisible(true)
  }

  const handleImportSubmit = async (values: any) => {
    try {
      let records = []
      
      if (values.import_type === 'json') {
        const jsonData = JSON.parse(values.json_content)
        records = jsonData.records || jsonData
      } else {
        // 处理文本格式导入
        const lines = values.text_content.split('\n')
        records = lines.filter((line: string) => line.trim() && !line.startsWith('#'))
          .map((line: string) => {
            const parts = line.trim().split(/\s+/)
            if (parts.length < 3) return null
            
            const [fullDomain, type, value, ...rest] = parts
            const domainParts = fullDomain.split('.')
            const subdomain = domainParts[0]
            const domain = domainParts.slice(1).join('.')
            
            return {
              subdomain,
              domain,
              type: type.toUpperCase(),
              value,
              ttl: rest[0] ? parseInt(rest[0]) || 600 : 600,
              priority: rest[1] ? parseInt(rest[1]) || 0 : 0,
              weight: rest[2] ? parseInt(rest[2]) || 0 : 0,
              port: rest[3] ? parseInt(rest[3]) || 0 : 0,
              comment: rest.slice(4).join(' ') || ''
            }
          }).filter(Boolean)
      }
      
      const response = await api.post('/api/dns-records/import', { records })
      
      if (response.data.success_count > 0) {
        message.success(`成功导入 ${response.data.success_count} 条记录`)
      }
      
      if (response.data.error_count > 0) {
        message.warning(`${response.data.error_count} 条记录导入失败`)
        console.error('导入错误:', response.data.errors)
      }
      
      setImportModalVisible(false)
      fetchRecords()
      fetchStats()
    } catch (error: any) {
      message.error(error.response?.data?.error || '导入失败')
    }
  }

  const getSelectedRecordType = () => {
    return recordTypes.find(type => type.value === selectedRecordType)
  }

  const needsPriority = () => {
    const type = getSelectedRecordType()
    return type?.fields.includes('priority')
  }

  const needsWeight = () => {
    const type = getSelectedRecordType()
    return type?.fields.includes('weight')
  }

  const needsPort = () => {
    const type = getSelectedRecordType()
    return type?.fields.includes('port')
  }

  const columns = [
    {
      title: '子域名',
      dataIndex: 'subdomain',
      key: 'subdomain',
      render: (subdomain: string, record: DNSRecord) => (
        <div>
          <strong>{subdomain}.{record.domain.name}</strong>
          {record.comment && (
            <div style={{ fontSize: '12px', color: '#666', marginTop: '2px' }}>
              {record.comment}
            </div>
          )}
        </div>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => {
        const colors: { [key: string]: string } = {
          'A': 'blue',
          'AAAA': 'cyan',
          'CNAME': 'green',
          'MX': 'orange',
          'TXT': 'purple',
          'NS': 'red',
          'PTR': 'magenta',
          'SRV': 'gold',
          'CAA': 'lime'
        }
        return <Tag color={colors[type] || 'default'}>{type}</Tag>
      },
    },
    {
      title: '记录值',
      dataIndex: 'value',
      key: 'value',
      ellipsis: true,
      render: (value: string, record: DNSRecord) => (
        <div>
          <div>{value}</div>
          {(record.priority > 0 || record.weight > 0 || record.port > 0) && (
            <div style={{ fontSize: '12px', color: '#666' }}>
              {record.priority > 0 && `优先级: ${record.priority} `}
              {record.weight > 0 && `权重: ${record.weight} `}
              {record.port > 0 && `端口: ${record.port}`}
            </div>
          )}
        </div>
      ),
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
      {/* 统计信息 */}
      {stats && (
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col span={6}>
            <Card>
              <Statistic
                title="总记录数"
                value={stats.total_records}
                suffix={`/ ${stats.quota}`}
              />
              <Progress
                percent={stats.quota_used}
                size="small"
                status={stats.quota_used > 80 ? 'exception' : 'normal'}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="域名数量"
                value={stats.domain_stats.length}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="记录类型"
                value={stats.type_stats.length}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="配额使用率"
                value={stats.quota_used}
                precision={1}
                suffix="%"
              />
            </Card>
          </Col>
        </Row>
      )}

      <div style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center',
        marginBottom: 24 
      }}>
        <Title level={2} style={{ margin: 0 }}>
          DNS记录管理
        </Title>
        <Space>
          <Button 
            icon={<DownloadOutlined />}
            onClick={handleExport}
          >
            导出记录
          </Button>
          <Button 
            icon={<UploadOutlined />}
            onClick={handleImport}
          >
            导入记录
          </Button>
          <Button 
            icon={<AppstoreAddOutlined />}
            onClick={handleBatchCreate}
          >
            批量创建
          </Button>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            添加记录
          </Button>
        </Space>
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

      {/* 创建/编辑记录模态框 */}
      <Modal
        title={editingRecord ? '编辑DNS记录' : '添加DNS记录'}
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
            label={
              <span>
                子域名
                <Tooltip title="支持通配符(*)和下划线(_)，如 www, mail, *, _dmarc">
                  <InfoCircleOutlined style={{ marginLeft: 4 }} />
                </Tooltip>
              </span>
            }
            rules={[{ required: true, message: '请输入子域名' }]}
          >
            <Input placeholder="例如: www, mail, *, _dmarc" />
          </Form.Item>

          <Form.Item
            name="type"
            label="记录类型"
            rules={[{ required: true, message: '请选择记录类型' }]}
          >
            <Select 
              placeholder="选择记录类型"
              onChange={setSelectedRecordType}
            >
              {recordTypes.map(type => (
                <Option key={type.value} value={type.value}>
                  <div>
                    <strong>{type.label}</strong>
                    <div style={{ fontSize: '12px', color: '#666' }}>
                      {type.description}
                    </div>
                  </div>
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="value"
            label="记录值"
            rules={[{ required: true, message: '请输入记录值' }]}
          >
            <Input placeholder="例如: 192.168.1.1 或 example.com" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="ttl"
                label="TTL (秒)"
                rules={[{ required: true, message: '请选择TTL值' }]}
              >
                <Select placeholder="选择TTL">
                  {ttlOptions.map(option => (
                    <Option key={option.value} value={option.value}>
                      <div>
                        <strong>{option.label}</strong>
                        <div style={{ fontSize: '12px', color: '#666' }}>
                          {option.description}
                        </div>
                      </div>
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            {needsPriority() && (
              <Col span={12}>
                <Form.Item
                  name="priority"
                  label="优先级"
                  rules={[{ required: true, message: '请输入优先级' }]}
                >
                  <InputNumber min={0} max={65535} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
            )}
          </Row>

          {needsWeight() && (
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="weight"
                  label="权重"
                  rules={[{ required: true, message: '请输入权重' }]}
                >
                  <InputNumber min={0} max={65535} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
              {needsPort() && (
                <Col span={12}>
                  <Form.Item
                    name="port"
                    label="端口"
                    rules={[{ required: true, message: '请输入端口' }]}
                  >
                    <InputNumber min={1} max={65535} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              )}
            </Row>
          )}

          <Form.Item
            name="comment"
            label="备注"
          >
            <Input.TextArea rows={2} placeholder="可选的记录备注信息" />
          </Form.Item>

          {selectedRecordType === 'A' && (
            <Form.Item
              name="allow_private_ip"
              valuePropName="checked"
            >
              <Checkbox>
                允许使用私有IP地址 (10.x.x.x, 172.16-31.x.x, 192.168.x.x)
              </Checkbox>
            </Form.Item>
          )}
        </Form>
      </Modal>

      {/* 批量创建模态框 */}
      <Modal
        title="批量创建DNS记录"
        open={batchModalVisible}
        onCancel={() => setBatchModalVisible(false)}
        onOk={() => batchForm.submit()}
        okText="批量创建"
        cancelText="取消"
        width={800}
      >
        <Form
          form={batchForm}
          onFinish={handleBatchSubmit}
        >
          <Form.List name="records">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Card key={key} size="small" style={{ marginBottom: 16 }}>
                    <Row gutter={16}>
                      <Col span={6}>
                        <Form.Item
                          {...restField}
                          name={[name, 'domain_id']}
                          rules={[{ required: true, message: '请选择域名' }]}
                        >
                          <Select placeholder="域名">
                            {domains.map(domain => (
                              <Option key={domain.id} value={domain.id}>
                                {domain.name}
                              </Option>
                            ))}
                          </Select>
                        </Form.Item>
                      </Col>
                      <Col span={4}>
                        <Form.Item
                          {...restField}
                          name={[name, 'subdomain']}
                          rules={[{ required: true, message: '子域名' }]}
                        >
                          <Input placeholder="子域名" />
                        </Form.Item>
                      </Col>
                      <Col span={4}>
                        <Form.Item
                          {...restField}
                          name={[name, 'type']}
                          rules={[{ required: true, message: '类型' }]}
                        >
                          <Select placeholder="类型">
                            {recordTypes.map(type => (
                              <Option key={type.value} value={type.value}>
                                {type.value}
                              </Option>
                            ))}
                          </Select>
                        </Form.Item>
                      </Col>
                      <Col span={6}>
                        <Form.Item
                          {...restField}
                          name={[name, 'value']}
                          rules={[{ required: true, message: '记录值' }]}
                        >
                          <Input placeholder="记录值" />
                        </Form.Item>
                      </Col>
                      <Col span={3}>
                        <Form.Item
                          {...restField}
                          name={[name, 'ttl']}
                        >
                          <InputNumber placeholder="TTL" min={60} max={604800} />
                        </Form.Item>
                      </Col>
                      <Col span={1}>
                        <Button 
                          type="link" 
                          danger 
                          onClick={() => remove(name)}
                          icon={<DeleteOutlined />}
                        />
                      </Col>
                    </Row>
                  </Card>
                ))}
                <Form.Item>
                  <Button 
                    type="dashed" 
                    onClick={() => add()} 
                    block 
                    icon={<PlusOutlined />}
                  >
                    添加记录
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>
        </Form>
      </Modal>

      {/* 导入记录模态框 */}
      <Modal
        title="导入DNS记录"
        open={importModalVisible}
        onCancel={() => setImportModalVisible(false)}
        onOk={() => importForm.submit()}
        okText="导入"
        cancelText="取消"
        width={600}
      >
        <Form
          form={importForm}
          layout="vertical"
          onFinish={handleImportSubmit}
        >
          <Form.Item
            name="import_type"
            label="导入格式"
            initialValue="json"
          >
            <Select>
              <Option value="json">JSON格式</Option>
              <Option value="text">文本格式</Option>
            </Select>
          </Form.Item>

          <Form.Item
            noStyle
            shouldUpdate={(prevValues, currentValues) => 
              prevValues.import_type !== currentValues.import_type
            }
          >
            {({ getFieldValue }) => {
              const importType = getFieldValue('import_type')
              
              if (importType === 'json') {
                return (
                  <Form.Item
                    name="json_content"
                    label="JSON内容"
                    rules={[{ required: true, message: '请输入JSON内容' }]}
                  >
                    <TextArea
                      rows={10}
                      placeholder='{"records": [{"subdomain": "www", "domain": "example.com", "type": "A", "value": "1.2.3.4", "ttl": 600}]}'
                    />
                  </Form.Item>
                )
              } else {
                return (
                  <Form.Item
                    name="text_content"
                    label="文本内容"
                    rules={[{ required: true, message: '请输入文本内容' }]}
                    extra="格式: 子域名.域名 类型 值 [TTL] [优先级] [权重] [端口] [备注]"
                  >
                    <TextArea
                      rows={10}
                      placeholder={`www.example.com A 1.2.3.4 600
mail.example.com MX mail.example.com 600 10
_sip._tcp.example.com SRV sip.example.com 600 10 5 5060`}
                    />
                  </Form.Item>
                )
              }
            }}
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default DNSRecords