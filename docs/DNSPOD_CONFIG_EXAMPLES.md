# DNSPod 配置示例

本文档提供了 DNSPod 各种配置的详细示例，帮助用户快速配置和使用 DNS 服务。

## 🔧 基础配置

### 腾讯云 DNSPod API v3 配置 (推荐)

```json
{
  "secret_id": "AKID********************************",
  "secret_key": "********************************",
  "region": "ap-guangzhou"
}
```

**配置说明：**

- `secret_id`: 腾讯云 API 密钥 ID，以 `AKID` 开头，长度为 36 位
- `secret_key`: 腾讯云 API 密钥，长度为 32 位
- `region`: 可选，指定地域，默认为空（就近接入）

**支持的地域：**

```
ap-guangzhou    # 广州
ap-shanghai     # 上海
ap-nanjing      # 南京
ap-beijing      # 北京
ap-chengdu      # 成都
ap-chongqing    # 重庆
ap-hongkong     # 香港
ap-singapore    # 新加坡
```

### DNSPod 传统 API 配置

```json
{
  "token": "12345,abcdef123456789abcdef123456789abc"
}
```

**配置说明：**

- `token`: DNSPod Token，格式为 `ID,Token`

## 🚀 前端配置示例

### 服务商选择器

```typescript
// 服务商类型定义
type ProviderType = "dnspod" | "dnspod_v3";

// 配置模板
const getConfigTemplate = (type: ProviderType): string => {
  switch (type) {
    case "dnspod":
      return JSON.stringify(
        {
          token: "ID,TOKEN",
        },
        null,
        2
      );

    case "dnspod_v3":
      return JSON.stringify(
        {
          secret_id: "AKID********************************",
          secret_key: "********************************",
          region: "ap-guangzhou",
        },
        null,
        2
      );

    default:
      return "{}";
  }
};

// React 组件示例
const ProviderConfig: React.FC = () => {
  const [providerType, setProviderType] = useState<ProviderType>("dnspod_v3");
  const [config, setConfig] = useState("");

  useEffect(() => {
    setConfig(getConfigTemplate(providerType));
  }, [providerType]);

  return (
    <div>
      <Select
        value={providerType}
        onChange={setProviderType}
        placeholder="选择 DNS 服务商"
      >
        <Option value="dnspod_v3">腾讯云 DNSPod (推荐)</Option>
        <Option value="dnspod">DNSPod 传统 API</Option>
      </Select>

      <TextArea
        value={config}
        onChange={(e) => setConfig(e.target.value)}
        placeholder="请输入配置 JSON"
        rows={6}
      />
    </div>
  );
};
```

## 📝 API 使用示例

### 创建 DNS 记录

```go
// 使用腾讯云 API v3
provider, err := NewDNSProvider("dnspod_v3", configJSON)
if err != nil {
    log.Fatal(err)
}

// 创建 A 记录
recordID, err := provider.CreateRecord(
    "example.com",  // 域名
    "www",          // 子域名
    "A",            // 记录类型
    "192.168.1.1",  // 记录值
    600,            // TTL (秒)
)
if err != nil {
    log.Printf("创建记录失败: %v", err)
} else {
    log.Printf("记录创建成功，ID: %s", recordID)
}
```

### 批量创建记录

```go
// 准备批量记录
records := []CreateRecordRequest{
    {
        Domain:     "example.com",
        SubDomain:  "www",
        RecordType: "A",
        Value:      "192.168.1.1",
        TTL:        &[]uint64{600}[0],
    },
    {
        Domain:     "example.com",
        SubDomain:  "mail",
        RecordType: "A",
        Value:      "192.168.1.2",
        TTL:        &[]uint64{600}[0],
    },
    {
        Domain:     "example.com",
        SubDomain:  "ftp",
        RecordType: "CNAME",
        Value:      "www.example.com",
        TTL:        &[]uint64{600}[0],
    },
}

// 批量创建
if v3Provider, ok := provider.(*DNSPodV3Provider); ok {
    recordIDs, err := v3Provider.BatchCreateRecords("example.com", records)
    if err != nil {
        log.Printf("批量创建失败: %v", err)
    } else {
        log.Printf("批量创建成功，记录IDs: %v", recordIDs)
    }
}
```

### 查询和管理记录

```go
// 获取所有记录
records, err := provider.GetRecords("example.com")
if err != nil {
    log.Printf("查询记录失败: %v", err)
} else {
    for _, record := range records {
        fmt.Printf("记录: %s.%s %s %s (TTL: %d)\n",
            record.Subdomain, "example.com",
            record.Type, record.Value, record.TTL)
    }
}

// 更新记录
err = provider.UpdateRecord(
    "example.com",   // 域名
    recordID,        // 记录ID
    "www",           // 子域名
    "A",             // 记录类型
    "192.168.1.100", // 新的记录值
    300,             // 新的TTL
)
if err != nil {
    log.Printf("更新记录失败: %v", err)
}

// 删除记录
err = provider.DeleteRecord("example.com", recordID)
if err != nil {
    log.Printf("删除记录失败: %v", err)
}
```

## 🛡️ 安全最佳实践

### 1. 密钥管理

```go
// 从环境变量读取敏感信息
config := DNSPodV3Config{
    SecretId:  os.Getenv("TENCENTCLOUD_SECRET_ID"),
    SecretKey: os.Getenv("TENCENTCLOUD_SECRET_KEY"),
    Region:    os.Getenv("TENCENTCLOUD_REGION"),
}

// 验证配置
if err := validateDNSPodV3Config(config); err != nil {
    log.Fatal("配置验证失败:", err)
}
```

### 2. 错误处理

```go
func handleDNSError(err error) {
    if strings.Contains(err.Error(), "AuthFailure") {
        log.Error("认证失败，请检查密钥配置")
        // 发送告警通知
        sendAlert("DNS认证失败", err.Error())
    } else if strings.Contains(err.Error(), "RequestLimitExceeded") {
        log.Warn("请求频率超限，等待重试")
        // 实施退避策略
        time.Sleep(time.Second * 5)
    } else {
        log.Error("DNS操作失败:", err)
    }
}
```

### 3. 请求频率控制

```go
type RateLimiter struct {
    lastRequest time.Time
    minInterval time.Duration
    mu          sync.Mutex
}

func (rl *RateLimiter) Wait() {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    elapsed := time.Since(rl.lastRequest)
    if elapsed < rl.minInterval {
        time.Sleep(rl.minInterval - elapsed)
    }
    rl.lastRequest = time.Now()
}

// 使用示例
rateLimiter := &RateLimiter{
    minInterval: time.Second * 2, // 每2秒最多一个请求
}

for _, domain := range domains {
    rateLimiter.Wait()
    records, err := provider.GetRecords(domain)
    // 处理结果...
}
```

## 🔍 调试和监控

### 1. 日志配置

```go
// 启用详细日志
func enableDebugLogging() {
    // 可以通过环境变量控制
    if os.Getenv("DNS_DEBUG") == "true" {
        log.SetLevel(log.DebugLevel)
    }
}

// 记录API调用
func logAPICall(action, domain string, duration time.Duration, err error) {
    status := "SUCCESS"
    if err != nil {
        status = "FAILED"
    }

    log.WithFields(log.Fields{
        "action":   action,
        "domain":   domain,
        "duration": duration,
        "status":   status,
    }).Info("DNS API调用")

    if err != nil {
        log.WithError(err).Error("DNS API调用失败")
    }
}
```

### 2. 健康检查

```go
func healthCheck(provider DNSProvider) error {
    // 尝试获取域名列表
    domains, err := provider.GetDomains()
    if err != nil {
        return fmt.Errorf("健康检查失败: %v", err)
    }

    if len(domains) == 0 {
        return fmt.Errorf("没有可用的域名")
    }

    log.Printf("健康检查通过，发现 %d 个域名", len(domains))
    return nil
}
```

## 🚀 高级用法

### 1. 动态 DNS (DDNS)

```go
func updateDynamicDNS(provider DNSProvider, domain, subdomain string) error {
    // 获取当前公网IP
    currentIP, err := getCurrentPublicIP()
    if err != nil {
        return err
    }

    // 查找现有记录
    records, err := provider.GetRecords(domain)
    if err != nil {
        return err
    }

    var targetRecord *DNSRecord
    for _, record := range records {
        if record.Subdomain == subdomain && record.Type == "A" {
            targetRecord = &record
            break
        }
    }

    if targetRecord == nil {
        // 创建新记录
        _, err = provider.CreateRecord(domain, subdomain, "A", currentIP, 300)
    } else if targetRecord.Value != currentIP {
        // 更新现有记录
        err = provider.UpdateRecord(domain, targetRecord.ID, subdomain, "A", currentIP, 300)
    }

    return err
}

func getCurrentPublicIP() (string, error) {
    resp, err := http.Get("https://api.ipify.org")
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(body), nil
}
```

### 2. 记录同步

```go
func syncRecords(srcProvider, dstProvider DNSProvider, domain string) error {
    // 获取源记录
    srcRecords, err := srcProvider.GetRecords(domain)
    if err != nil {
        return fmt.Errorf("获取源记录失败: %v", err)
    }

    // 获取目标记录
    dstRecords, err := dstProvider.GetRecords(domain)
    if err != nil {
        return fmt.Errorf("获取目标记录失败: %v", err)
    }

    // 构建目标记录映射
    dstMap := make(map[string]DNSRecord)
    for _, record := range dstRecords {
        key := fmt.Sprintf("%s-%s", record.Subdomain, record.Type)
        dstMap[key] = record
    }

    // 同步记录
    for _, srcRecord := range srcRecords {
        key := fmt.Sprintf("%s-%s", srcRecord.Subdomain, srcRecord.Type)

        if dstRecord, exists := dstMap[key]; exists {
            // 更新现有记录
            if dstRecord.Value != srcRecord.Value || dstRecord.TTL != srcRecord.TTL {
                err := dstProvider.UpdateRecord(domain, dstRecord.ID,
                    srcRecord.Subdomain, srcRecord.Type, srcRecord.Value, srcRecord.TTL)
                if err != nil {
                    log.Printf("更新记录失败 %s: %v", key, err)
                }
            }
        } else {
            // 创建新记录
            _, err := dstProvider.CreateRecord(domain, srcRecord.Subdomain,
                srcRecord.Type, srcRecord.Value, srcRecord.TTL)
            if err != nil {
                log.Printf("创建记录失败 %s: %v", key, err)
            }
        }
    }

    return nil
}
```

## 📊 性能优化

### 1. 连接池

```go
var httpClient = &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

### 2. 并发控制

```go
func processDomainsConcurrently(provider DNSProvider, domains []string) {
    // 限制并发数
    semaphore := make(chan struct{}, 5)
    var wg sync.WaitGroup

    for _, domain := range domains {
        wg.Add(1)
        go func(d string) {
            defer wg.Done()
            semaphore <- struct{}{} // 获取信号量
            defer func() { <-semaphore }() // 释放信号量

            records, err := provider.GetRecords(d)
            if err != nil {
                log.Printf("处理域名 %s 失败: %v", d, err)
                return
            }

            log.Printf("域名 %s 有 %d 条记录", d, len(records))
        }(domain)
    }

    wg.Wait()
}
```

## 🎯 故障排查

### 常见错误及解决方案

| 错误信息                       | 可能原因        | 解决方案               |
| ------------------------------ | --------------- | ---------------------- |
| `AuthFailure.SignatureExpire`  | 签名过期        | 检查系统时间同步       |
| `AuthFailure.SecretIdNotFound` | SecretId 不存在 | 检查控制台中的密钥状态 |
| `InvalidParameter`             | 参数错误        | 检查参数格式和取值范围 |
| `RequestLimitExceeded`         | 请求频率超限    | 实施退避重试策略       |
| `ResourceNotFound`             | 资源不存在      | 确认域名或记录 ID 正确 |

### 调试工具

```bash
# 测试DNS解析
nslookup example.com

# 查看DNS传播状态
dig @8.8.8.8 example.com

# 检查TTL值
dig example.com | grep -E "^example.com"
```

---

通过这些配置示例和最佳实践，您可以快速上手并高效使用 DNSPod API 进行 DNS 管理。
