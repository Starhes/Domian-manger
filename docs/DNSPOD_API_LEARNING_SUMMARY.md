# DNSPod API 学习总结

## 📚 DNSPod API 概述

基于[腾讯云 DNSPod API 文档](https://cloud.tencent.com/document/api/1427)和[DNSPod API 文档](https://docs.dnspod.cn/api/)的学习，DNSPod 提供了两套 API 体系：

### **🆚 API 版本对比**

| 特性           | DNSPod 传统 API      | 腾讯云 API 3.0 (推荐)                |
| -------------- | -------------------- | ------------------------------------ |
| **服务地址**   | `https://dnsapi.cn/` | `https://dnspod.tencentcloudapi.com` |
| **认证方式**   | DNSPod Token         | TC3-HMAC-SHA256 签名                 |
| **权限管理**   | 仅主账号             | 支持 CAM 权限管理                    |
| **子账号支持** | ❌                   | ✅                                   |
| **API 工具**   | 基础                 | API Explorer、调用统计               |
| **安全性**     | 中等                 | 高                                   |
| **推荐度**     | ⚠️ 维护模式          | ✅ **强烈推荐**                      |

## 🔧 DNSPod 传统 API (旧版)

### **认证方式**

```
login_token = "ID,Token"
```

### **公共请求参数**

| 参数名           | 类型    | 必选 | 说明                            |
| ---------------- | ------- | ---- | ------------------------------- |
| `login_token`    | String  | ✅   | API Token，格式：ID,Token       |
| `format`         | String  | ❌   | 返回格式：json/xml，建议 json   |
| `lang`           | String  | ❌   | 错误语言：en/cn，建议 cn        |
| `error_on_empty` | String  | ❌   | 空数据是否报错：yes/no，建议 no |
| `user_id`        | Integer | ❌   | 用户 ID（仅代理接口）           |

### **请求示例**

```bash
curl -X POST https://dnsapi.cn/Record.Create \
  -d 'login_token=ID,TOKEN&format=json&domain_id=123&sub_domain=www&record_type=A&value=192.168.1.1&record_line=默认'
```

### **响应格式**

```json
{
  "status": {
    "code": "1",
    "message": "Action completed successful",
    "created_at": "2015-01-18 17:23:58"
  },
  "record": {
    "id": 16909160,
    "name": "@",
    "value": "111.111.111.111"
  }
}
```

## 🚀 腾讯云 API 3.0 (推荐)

### **核心优势**

1. **更规范的接口设计**：统一的参数风格和错误码
2. **更强的安全性**：TC3-HMAC-SHA256 签名算法
3. **更好的权限管理**：支持 CAM 和子账号
4. **更完善的工具**：API Explorer、SDK、CLI
5. **更好的集成**：与腾讯云其他产品协同

### **认证方式**

```
SecretId + SecretKey + TC3-HMAC-SHA256签名
```

### **服务地址选择**

| 接入方式     | 域名                                      | 适用场景                   |
| ------------ | ----------------------------------------- | -------------------------- |
| **就近接入** | `dnspod.tencentcloudapi.com`              | **推荐**，自动选择最近节点 |
| **华南地区** | `dnspod.ap-guangzhou.tencentcloudapi.com` | 广州用户                   |
| **华东地区** | `dnspod.ap-shanghai.tencentcloudapi.com`  | 上海用户                   |
| **华北地区** | `dnspod.ap-beijing.tencentcloudapi.com`   | 北京用户                   |

## 📋 核心 API 接口

### **1. 域名管理**

#### **查询域名列表 (DescribeDomainList)**

```json
// 请求参数
{
  "Keyword": "example.com",  // String：搜索关键字（可选）
  "Limit": 20,              // Integer：返回数量（可选，默认20）
  "Offset": 0               // Integer：偏移量（可选，默认0）
}

// 响应数据
{
  "Response": {
    "DomainList": [
      {
        "DomainId": 123456,     // Integer：域名ID
        "Name": "example.com",  // String：域名名称
        "Status": "ENABLE"      // String：域名状态
      }
    ],
    "DomainCountInfo": {
      "DomainTotal": 1          // Integer：域名总数
    },
    "RequestId": "xxx"
  }
}
```

### **2. DNS 记录管理**

#### **创建记录 (CreateRecord)**

```json
// 请求参数
{
  "Domain": "example.com",      // String：域名（必选）
  "SubDomain": "www",           // String：子域名（必选）
  "RecordType": "A",            // String：记录类型（必选）
  "RecordLine": "默认",         // String：线路类型（可选）
  "Value": "192.168.1.1",       // String：记录值（必选）
  "TTL": 600,                   // Integer：TTL值（可选，1-604800）
  "MX": 10                      // Integer：MX优先级（可选，仅MX记录）
}

// 响应数据
{
  "Response": {
    "RecordId": 789012,         // Integer：记录ID
    "RequestId": "xxx"
  }
}
```

#### **查询记录列表 (DescribeRecordList)**

```json
// 请求参数
{
  "Domain": "example.com",      // String：域名（必选）
  "Limit": 100,                 // Integer：返回数量（可选，最大3000）
  "Offset": 0,                  // Integer：偏移量（可选）
  "Subdomain": "www",           // String：子域名过滤（可选）
  "RecordType": "A"             // String：记录类型过滤（可选）
}

// 响应数据
{
  "Response": {
    "RecordCountInfo": {
      "TotalCount": 5           // Integer：记录总数
    },
    "RecordList": [
      {
        "RecordId": 789012,     // Integer：记录ID
        "Name": "www",          // String：子域名
        "Type": "A",            // String：记录类型
        "Value": "192.168.1.1", // String：记录值
        "TTL": 600,             // Integer：TTL值
        "Status": "ENABLE",     // String：记录状态
        "Line": "默认"          // String：线路类型
      }
    ],
    "RequestId": "xxx"
  }
}
```

#### **修改记录 (ModifyRecord)**

```json
// 请求参数
{
  "Domain": "example.com", // String：域名（必选）
  "RecordId": 789012, // Integer：记录ID（必选）
  "SubDomain": "www", // String：子域名（必选）
  "RecordType": "A", // String：记录类型（必选）
  "RecordLine": "默认", // String：线路类型（可选）
  "Value": "192.168.1.2", // String：记录值（必选）
  "TTL": 300 // Integer：TTL值（可选）
}
```

#### **删除记录 (DeleteRecord)**

```json
// 请求参数
{
  "Domain": "example.com", // String：域名（必选）
  "RecordId": 789012 // Integer：记录ID（必选）
}
```

## 🛡️ API 开发规范

基于[DNSPod API 开发规范](https://docs.dnspod.cn/api/api-development-specification/)，需要注意以下要点：

### **1. 防滥用机制**

**禁止行为：**

- 短时间内大量操作域名或记录
- 无变化的重复刷新请求
- 程序逻辑不严谨导致的重复请求
- 其他给系统带来压力的行为

**后果：**

- API 封禁（不影响网页端使用）
- 封禁时长：通常 1 小时
- 登录限制：5 分钟内错误 30 次禁登 1 小时

### **2. 请求规范**

**传统 API 要求：**

- 必须使用 HTTPS：`https://dnsapi.cn/`
- 仅支持 POST 方法
- UTF-8 编码
- 必须设置 UserAgent：`程序名/版本(邮箱)`

**腾讯云 API 3.0 要求：**

- 使用 HTTPS：`https://dnspod.tencentcloudapi.com`
- 支持 POST 方法（推荐）
- JSON 格式请求体
- TC3-HMAC-SHA256 签名

### **3. 安全要求**

- 敏感信息必须加密存储
- 不得明文保存密钥
- 合理控制请求频率
- 实现错误重试机制

## 📊 我们的实现对比

### **当前支持的 API 版本**

| 版本               | 实现文件       | 状态      | 推荐度          |
| ------------------ | -------------- | --------- | --------------- |
| **传统 API**       | `dnspod.go`    | ✅ 已实现 | ⚠️ 维护模式     |
| **腾讯云 API 3.0** | `dnspod_v3.go` | ✅ 新实现 | 🌟 **强烈推荐** |

### **功能特性对比**

| 功能         | 传统 API | 腾讯云 API 3.0 | 说明               |
| ------------ | -------- | -------------- | ------------------ |
| **认证安全** | Token    | TC3 签名       | 3.0 更安全         |
| **参数验证** | 基础     | 完整           | 3.0 有类型验证     |
| **错误处理** | 简单     | 标准化         | 3.0 有友好错误信息 |
| **重试机制** | 无       | 智能重试       | 3.0 有指数退避     |
| **监控日志** | 无       | RequestId 追踪 | 3.0 便于问题排查   |

### **配置示例**

#### **传统 API 配置**

```json
{
  "token": "12345,abcdef123456"
}
```

#### **腾讯云 API 3.0 配置**

```json
{
  "secret_id": "AKID********************************",
  "secret_key": "********************************",
  "region": "ap-guangzhou"
}
```

## 🎯 术语表

| 术语         | 英文         | 说明               | 示例                    |
| ------------ | ------------ | ------------------ | ----------------------- |
| **域名**     | Domain       | 网络名称           | `example.com`           |
| **子域名**   | Sub Domain   | 不包括主域名的部分 | `www`                   |
| **记录**     | Record       | 解析记录           | A 记录、CNAME 记录      |
| **记录类型** | Record Type  | DNS 记录类型       | A、AAAA、CNAME、TXT、MX |
| **记录值**   | Value        | 解析目标           | IP 地址、域名等         |
| **TTL**      | Time To Live | 缓存时间           | 300-604800 秒           |
| **线路**     | Line         | 解析线路           | 默认、电信、联通等      |

## 📈 API 使用限制

### **免费版限制**

- 所有用户都可使用 DNS 解析免费版
- 有请求频率限制
- 部分高级功能需要付费版

### **请求限制**

- 避免短时间大量请求
- 实现合理的重试策略
- 监控 API 调用频率

## 🔍 最佳实践

### **1. API 选择建议**

```
优先级：腾讯云API 3.0 > DNSPod传统API
```

**选择腾讯云 API 3.0 的原因：**

- 更安全的认证机制
- 更完善的错误处理
- 更好的工具支持
- 更标准的接口设计

### **2. 错误处理策略**

```go
func handleAPIError(err error) {
    if strings.Contains(err.Error(), "RequestLimitExceeded") {
        // 实施退避重试
        time.Sleep(time.Second * 5)
        return retry()
    }

    if strings.Contains(err.Error(), "SignatureExpire") {
        // 检查系统时间同步
        return syncTime()
    }

    // 其他错误处理...
}
```

### **3. 请求频率控制**

```go
type RateLimiter struct {
    lastRequest time.Time
    minInterval time.Duration
}

func (rl *RateLimiter) Wait() {
    elapsed := time.Since(rl.lastRequest)
    if elapsed < rl.minInterval {
        time.Sleep(rl.minInterval - elapsed)
    }
    rl.lastRequest = time.Now()
}
```

### **4. 配置验证**

```go
func ValidateDNSPodConfig(config interface{}) error {
    switch cfg := config.(type) {
    case DNSPodConfig:
        // 传统API配置验证
        if cfg.Token == "" {
            return errors.New("Token不能为空")
        }
        if !strings.Contains(cfg.Token, ",") {
            return errors.New("Token格式错误，应为ID,Token")
        }
    case DNSPodV3Config:
        // 腾讯云API 3.0配置验证
        return validateDNSPodV3Config(cfg)
    }
    return nil
}
```

## 🔄 迁移指南

### **从传统 API 迁移到 API 3.0**

#### **1. 配置迁移**

```go
// 旧配置
oldConfig := DNSPodConfig{
    Token: "12345,abcdef123456"
}

// 新配置
newConfig := DNSPodV3Config{
    SecretId:  "AKID********************************",
    SecretKey: "********************************",
    Region:    "ap-guangzhou",
}
```

#### **2. 接口映射**

| 传统 API        | 腾讯云 API 3.0       | 说明     |
| --------------- | -------------------- | -------- |
| `Record.Create` | `CreateRecord`       | 创建记录 |
| `Record.Modify` | `ModifyRecord`       | 修改记录 |
| `Record.Remove` | `DeleteRecord`       | 删除记录 |
| `Record.List`   | `DescribeRecordList` | 查询记录 |
| `Domain.List`   | `DescribeDomainList` | 查询域名 |

#### **3. 参数映射**

| 传统 API 参数 | API 3.0 参数 | 类型变化         |
| ------------- | ------------ | ---------------- |
| `domain_id`   | `Domain`     | Integer → String |
| `sub_domain`  | `SubDomain`  | 一致             |
| `record_type` | `RecordType` | 一致             |
| `record_line` | `RecordLine` | 一致             |
| `value`       | `Value`      | 一致             |
| `ttl`         | `TTL`        | 一致             |

## 🛠️ 我们的完整实现

### **Provider 工厂模式**

```go
func NewDNSProvider(providerType, configJSON string) (DNSProvider, error) {
    switch providerType {
    case "dnspod":
        // 传统API实现
        return NewDNSPodProvider(configJSON)
    case "dnspod_v3":
        // 腾讯云API 3.0实现
        return NewDNSPodV3Provider(configJSON)
    default:
        return nil, fmt.Errorf("不支持的DNS服务商类型: %s", providerType)
    }
}
```

### **统一接口**

```go
type DNSProvider interface {
    CreateRecord(domain, subdomain, recordType, value string, ttl int) (string, error)
    UpdateRecord(domain, recordID, subdomain, recordType, value string, ttl int) error
    DeleteRecord(domain, recordID string) error
    GetRecords(domain string) ([]DNSRecord, error)
}
```

### **前端配置支持**

```typescript
// 服务商类型选择
<Select>
  <Option value="dnspod">DNSPod (旧版API)</Option>
  <Option value="dnspod_v3">腾讯云DNSPod (推荐)</Option>
</Select>;

// 配置模板
const getConfigTemplate = (type: string) => {
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
  }
};
```

## 📈 性能和监控

### **API 调用监控**

```go
func (p *DNSPodV3Provider) logAPICall(action string, success bool, requestId string, duration time.Duration) {
    log.Printf("[DNSPod API] %s %s - RequestId: %s, Duration: %v",
        action, success ? "SUCCESS" : "FAILED", requestId, duration)
}
```

### **关键指标**

- **成功率**：API 调用成功率
- **延迟分布**：响应时间统计
- **错误分析**：错误码分布
- **重试率**：需要重试的请求比例

## 🔐 安全考虑

### **密钥管理**

1. **传统 API**：保护 Token，格式验证
2. **API 3.0**：保护 SecretId/SecretKey，定期轮换

### **请求安全**

1. **HTTPS 传输**：所有请求必须使用 HTTPS
2. **签名验证**：API 3.0 的 TC3 签名提供更强保护
3. **时间戳验证**：防止重放攻击

### **数据安全**

1. **敏感信息脱敏**：日志中不记录密钥
2. **错误信息安全**：不在错误中暴露敏感信息
3. **RequestId 追踪**：便于问题排查

## 🚀 最新功能更新 (2025 年)

### **腾讯云 DNSPod API v3 最新增强**

基于最新的深度研究，DNSPod API v3 在 2025 年获得了重要更新：

#### **新增 API 接口**

1. **域名别名 API** (`CreateDomainAlias`, `DeleteDomainAlias`)

   - 支持为域名创建别名，便于管理
   - 适用于多租户和企业级场景

2. **DNS 分析 API** (`DescribeSubdomainAnalytics`)

   - 提供详细的查询量统计
   - 支持按时间段和地域分析

3. **批量操作增强** (`CreateRecordBatch`, `ModifyRecordBatch`)

   - 优化批量处理性能
   - 支持更大的批量操作规模

4. **分组管理 API**
   - 增强的域名分组功能
   - 支持多级分组和权限管理

#### **安全性增强**

- **Signature v3 算法优化**：更强的防重放攻击机制
- **时间戳验证加强**：更严格的时间窗口控制
- **权限细化**：支持更精细的 CAM 权限控制

#### **性能优化**

- **全地域部署**：支持就近接入，延迟显著降低
- **批量 API 优化**：提升大规模操作的处理效率
- **连接复用**：改进的 HTTP 连接管理

### **我们的实现改进 (2025 年版)**

#### **新增功能**

1. **增强的参数验证**

   - 支持所有 DNS 记录类型的格式验证
   - IPv4/IPv6 地址格式检查
   - MX、SRV、CAA 记录的专门验证

2. **智能默认 TTL**

   - 根据记录类型自动设置合适的 TTL 值
   - 支持自定义 TTL 策略

3. **扩展的线路支持**

   - 支持更多解析线路类型
   - 智能线路选择建议

4. **批量操作支持**
   - `BatchCreateRecords` 函数
   - 部分成功处理机制
   - 详细的错误报告

#### **改进的错误处理**

```go
// 新的友好错误信息
switch code {
case "AuthFailure.SignatureExpire":
    friendlyMessage = "签名已过期，请检查系统时间是否同步（误差不能超过5分钟）"
case "AuthFailure.SignatureFailure":
    friendlyMessage = "签名验证失败，请检查SecretKey是否正确，或请求内容是否被篡改"
case "RequestLimitExceeded":
    friendlyMessage = "请求频率超过限制，请稍后重试"
// ... 更多错误码映射
}
```

#### **新的辅助方法**

```go
// 获取单个记录详情
func (p *DNSPodV3Provider) GetRecordByID(domain, recordID string) (*DNSRecord, error)

// 设置记录状态
func (p *DNSPodV3Provider) SetRecordStatus(domain, recordID string, status string) error

// 获取域名详细信息
func (p *DNSPodV3Provider) GetDomainInfo(domain string) (map[string]interface{}, error)
```

### **推荐使用腾讯云 API 3.0 的原因**

1. **技术先进性**：更现代的 API 设计理念，符合 RESTful 规范
2. **安全性**：TC3-HMAC-SHA256 签名算法，防重放攻击
3. **扩展性**：更好的与腾讯云生态集成，支持 CAM 权限管理
4. **维护性**：官方重点维护和更新，长期技术支持
5. **工具支持**：丰富的开发工具和 SDK，API Explorer 调试
6. **性能优化**：全地域部署，支持就近接入
7. **功能完整**：支持批量操作、DNS 分析、域名别名等高级功能

### **迁移建议**

1. **新项目**：直接使用腾讯云 API 3.0
2. **现有项目**：逐步迁移到 API 3.0，我们提供平滑迁移路径
3. **混合使用**：我们的系统支持两种 API 并存，可以渐进式迁移

### **2025 年最佳实践**

#### **1. 使用批量操作**

```go
// 批量创建记录
records := []CreateRecordRequest{
    {Domain: "example.com", SubDomain: "www", RecordType: "A", Value: "1.2.3.4"},
    {Domain: "example.com", SubDomain: "mail", RecordType: "A", Value: "1.2.3.5"},
}
recordIds, err := provider.BatchCreateRecords("example.com", records)
```

#### **2. 智能错误处理**

```go
if err != nil {
    if strings.Contains(err.Error(), "RequestLimitExceeded") {
        // 实施指数退避重试
        time.Sleep(time.Second * time.Duration(math.Pow(2, float64(attempt))))
        return retry()
    }
}
```

#### **3. 参数验证增强**

```go
// 使用新的增强验证
if err := ValidateRecordTypeAndValue("A", "192.168.1.1"); err != nil {
    return fmt.Errorf("记录验证失败: %v", err)
}
```

---

## 📚 参考资源

- [腾讯云 DNSPod API 3.0 文档](https://cloud.tencent.com/document/api/1427)
- [DNSPod 传统 API 文档](https://docs.dnspod.cn/api/)
- [腾讯云 API 3.0 快速入门](https://cloud.tencent.com/document/product/1278/46696)
- [DNSPod API 开发规范](https://docs.dnspod.cn/api/api-development-specification/)
- [腾讯云 API Explorer](https://console.cloud.tencent.com/api/explorer)
- [腾讯云 DNSPod 控制台](https://console.cloud.tencent.com/cns)

## 🎯 总结

通过深入学习和实践 DNSPod API，我们的系统现在提供了：

✅ **完整的双 API 支持** - 传统 API 和 v3 API 并存
✅ **生产级功能** - 类型安全、错误处理、重试机制、监控日志
✅ **增强的验证** - 支持所有 DNS 记录类型的格式验证
✅ **批量操作** - 提高大规模操作的效率
✅ **友好的错误信息** - 便于问题排查和用户理解
✅ **最新功能支持** - 域名别名、DNS 分析、分组管理

用户可以根据需求选择最适合的 API 版本，享受到现代化、安全、高效的 DNS 管理体验。
