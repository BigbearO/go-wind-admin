# API Key 认证方案

## 1. 方案概述

本方案通过 API Key + API Secret 实现微服务间的身份认证，适用于定时任务、消息队列消费者等无法携带用户 Token 的场景。

### 核心特点

- **复用现有表结构**：服务账号存储在 `sys_users` 表，凭证存储在 `sys_user_credentials` 表
- **双重验证**：API Key（标识）+ API Secret（密钥）
- **最小化改动**：扩展现有认证中间件，无需引入复杂 OAuth 流程

---

## 2. 数据模型

### 2.1 服务账号（复用 `sys_users` 表）

| 字段 | 说明 | 示例值 |
|------|------|--------|
| `id` | 服务账号 ID（预留范围 90001-99999） | 90001 |
| `username` | 服务账号标识（`svc_` 前缀） | `svc_scheduler` |
| `nickname` | 服务名称 | `定时任务服务` |
| `user_type` | 用户类型（新增枚举） | `service_account` |
| `status` | 账号状态 | `active` |

### 2.2 API 凭证（复用 `sys_user_credentials` 表）

| 字段 | 说明 | 存储方式 |
|------|------|----------|
| `user_id` | 关联服务账号 ID | 90001 |
| `credential_type` | 凭证类型 | `API_KEY` / `API_SECRET` |
| `credential_value` | 凭证值 | API Key 明文 / API Secret bcrypt 哈希 |
| `expires_at` | 过期时间（可选） | `2025-12-31 23:59:59` |
| `status` | 凭证状态 | `active` / `revoked` |

---

## 3. 认证流程

### 3.1 请求格式

```http
GET /api/v1/users HTTP/1.1
Host: api.example.com
X-Api-Key: ak_scheduler_20240101_abc123
X-Api-Secret: sk_live_REDACTED
X-Api-Timestamp: 1704067200
X-Api-Signature: a1b2c3d4e5f6...
```

### 3.2 认证流程图

```
┌──────────────┐       ┌──────────────┐       ┌──────────────┐
│   调用服务    │       │  认证中间件   │       │    数据库    │
└──────┬───────┘       └──────┬───────┘       └──────┬───────┘
       │                      │                      │
       │ 1. 携带 API Key/Secret                     │
       │─────────────────────>│                      │
       │                      │                      │
       │                      │ 2. 查询凭证信息       │
       │                      │─────────────────────>│
       │                      │                      │
       │                      │ 3. 返回凭证数据       │
       │                      │<─────────────────────│
       │                      │                      │
       │                      │ 4. 验证 API Secret    │
       │                      │   (bcrypt 对比)      │
       │                      │                      │
       │                      │ 5. 注入用户上下文     │
       │                      │                      │
       │ 6. 继续处理请求       │                      │
       │<─────────────────────│                      │
       │                      │                      │
```

### 3.3 验证步骤

1. **提取请求头**：从请求中获取 `X-Api-Key`、`X-Api-Secret` 等
2. **查询凭证**：根据 API Key 查询对应的用户 ID 和凭证信息
3. **验证用户类型**：确认用户为服务账号类型
4. **验证 API Key**：对比存储的 API Key 是否匹配
5. **验证 API Secret**：使用 bcrypt 验证 Secret 哈希
6. **检查凭证状态**：验证是否过期或已撤销
7. **注入上下文**：将服务账号信息注入请求上下文

---

## 4. 签名机制（可选增强）

### 4.1 签名生成算法

```go
// 签名计算公式
signature = HmacSHA256(
    apiKey + apiSecret + timestamp + requestPath,
    apiSecret
)
```

### 4.2 Go 实现

```go
func generateSignature(apiKey, apiSecret, timestamp, requestPath string) string {
    data := apiKey + apiSecret + timestamp + requestPath
    h := hmac.New(sha256.New, []byte(apiSecret))
    h.Write([]byte(data))
    return hex.EncodeToString(h.Sum(nil))
}
```

### 4.3 时间戳验证

```go
const timestampTTL = 5 * time.Minute

func validateTimestamp(timestamp string) error {
    ts, err := strconv.ParseInt(timestamp, 10, 64)
    if err != nil {
        return ErrTimestampExpired
    }
    
    if time.Since(time.Unix(ts, 0)) > timestampTTL {
        return ErrTimestampExpired
    }
    return nil
}
```

---

## 5. 代码实现

### 5.1 常量定义

`pkg/middleware/auth/constants.go`

```go
package auth

const (
    // 请求头名称
    HeaderAPIKey       = "X-Api-Key"
    HeaderAPISecret    = "X-Api-Secret"
    HeaderAPITimestamp = "X-Api-Timestamp"
    HeaderAPISignature = "X-Api-Signature"
    
    // 服务账号 ID 范围
    ServiceAccountIDMin = 90001
    ServiceAccountIDMax = 99999
    
    // 用户类型
    UserTypeServiceAccount = "service_account"
    
    // 凭证类型
    CredentialTypeAPIKey    = "API_KEY"
    CredentialTypeAPISecret = "API_SECRET"
)
```

### 5.2 错误定义

`pkg/middleware/auth/errors.go`

```go
package auth

import "errors"

var (
    ErrInvalidAPIKey     = errors.New("invalid API key")
    ErrInvalidAPISecret  = errors.New("invalid API secret")
    ErrExpiredCredential = errors.New("credential has expired")
    ErrRevokedCredential = errors.New("credential has been revoked")
    ErrTimestampExpired  = errors.New("request timestamp expired")
    ErrInvalidSignature  = errors.New("invalid request signature")
    ErrNotServiceAccount = errors.New("user is not a service account")
)
```

### 5.3 API Key 验证器

`pkg/middleware/auth/api_key_validator.go`

```go
package auth

import (
    "context"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "strconv"
    "time"
    
    "github.com/go-kratos/kratos/v2/log"
    "golang.org/x/crypto/bcrypt"
)

// CredentialRepository 凭证数据访问接口
type CredentialRepository interface {
    GetCredentialByType(ctx context.Context, userID int64, credType string) (*UserCredential, error)
    GetUserByID(ctx context.Context, userID int64) (*User, error)
    GetUserIDByAPIKey(ctx context.Context, apiKey string) (int64, error)
}

// UserCredential 凭证信息
type UserCredential struct {
    ID              int64
    UserID          int64
    CredentialType  string
    CredentialValue string
    ExpiresAt       *time.Time
    Status          string
}

// User 用户信息
type User struct {
    ID       int64
    Username string
    UserType string
    Status   string
}

// APIKeyValidator API Key 验证器
type APIKeyValidator struct {
    repo          CredentialRepository
    logger        *log.Helper
    timestampTTL  time.Duration
    signEnabled   bool
}

// NewAPIKeyValidator 创建验证器
func NewAPIKeyValidator(repo CredentialRepository, logger log.Logger) *APIKeyValidator {
    return &APIKeyValidator{
        repo:         repo,
        logger:       log.NewHelper(logger),
        timestampTTL: 5 * time.Minute,
        signEnabled:  true,
    }
}

// Validate 验证 API Key 和 Secret
func (v *APIKeyValidator) Validate(ctx context.Context, apiKey, apiSecret string) (*User, error) {
    // 1. 根据 API Key 查询用户 ID
    userID, err := v.repo.GetUserIDByAPIKey(ctx, apiKey)
    if err != nil {
        v.logger.WithContext(ctx).Errorf("failed to get user by api key: %v", err)
        return nil, ErrInvalidAPIKey
    }
    
    // 2. 获取用户信息
    user, err := v.repo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // 3. 验证是否为服务账号
    if user.UserType != UserTypeServiceAccount {
        return nil, ErrNotServiceAccount
    }
    
    // 4. 验证 API Key
    storedAPIKey, err := v.repo.GetCredentialByType(ctx, userID, CredentialTypeAPIKey)
    if err != nil || storedAPIKey.CredentialValue != apiKey {
        return nil, ErrInvalidAPIKey
    }
    if err := v.checkCredentialStatus(storedAPIKey); err != nil {
        return nil, err
    }
    
    // 5. 验证 API Secret（bcrypt 对比）
    storedAPISecret, err := v.repo.GetCredentialByType(ctx, userID, CredentialTypeAPISecret)
    if err != nil {
        return nil, ErrInvalidAPISecret
    }
    if err := v.checkCredentialStatus(storedAPISecret); err != nil {
        return nil, err
    }
    if err := bcrypt.CompareHashAndPassword(
        []byte(storedAPISecret.CredentialValue),
        []byte(apiSecret),
    ); err != nil {
        return nil, ErrInvalidAPISecret
    }
    
    return user, nil
}

// ValidateWithSignature 验证带签名的请求
func (v *APIKeyValidator) ValidateWithSignature(
    ctx context.Context,
    apiKey, apiSecret, timestamp, signature, requestPath string,
) (*User, error) {
    // 1. 验证时间戳
    if err := v.validateTimestamp(timestamp); err != nil {
        return nil, err
    }
    
    // 2. 验证签名
    expectedSig := v.generateSignature(apiKey, apiSecret, timestamp, requestPath)
    if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
        return nil, ErrInvalidSignature
    }
    
    // 3. 验证凭证
    return v.Validate(ctx, apiKey, apiSecret)
}

func (v *APIKeyValidator) checkCredentialStatus(cred *UserCredential) error {
    if cred.Status == "revoked" {
        return ErrRevokedCredential
    }
    if cred.ExpiresAt != nil && time.Now().After(*cred.ExpiresAt) {
        return ErrExpiredCredential
    }
    return nil
}

func (v *APIKeyValidator) validateTimestamp(timestamp string) error {
    ts, err := strconv.ParseInt(timestamp, 10, 64)
    if err != nil {
        return ErrTimestampExpired
    }
    if time.Since(time.Unix(ts, 0)) > v.timestampTTL {
        return ErrTimestampExpired
    }
    return nil
}

func (v *APIKeyValidator) generateSignature(apiKey, apiSecret, timestamp, path string) string {
    data := apiKey + apiSecret + timestamp + path
    h := hmac.New(sha256.New, []byte(apiSecret))
    h.Write([]byte(data))
    return hex.EncodeToString(h.Sum(nil))
}
```

### 5.4 认证中间件

`pkg/middleware/auth/auth.go`

```go
package auth

import (
    "context"
    
    "github.com/go-kratos/kratos/v2/middleware"
    "github.com/go-kratos/kratos/v2/transport"
)

// Server 认证中间件
func Server(apiKeyValidator *APIKeyValidator, opts ...Option) middleware.Middleware {
    o := &options{
        apiKeyEnabled: true,
        signEnabled:   true,
    }
    for _, opt := range opts {
        opt(o)
    }
    
    return func(handler middleware.Handler) middleware.Handler {
        return func(ctx context.Context, req interface{}) (interface{}, error) {
            // 从传输层获取请求头
            if header, ok := transport.FromServerContext(ctx); ok {
                // 优先检查 Bearer Token
                token := extractToken(header)
                if token != "" {
                    // JWT 验证（现有逻辑）
                    return handleJWTAuth(ctx, req, handler, token)
                }
                
                // 检查 API Key 认证
                if o.apiKeyEnabled {
                    apiKey := header.RequestHeader().Get(HeaderAPIKey)
                    apiSecret := header.RequestHeader().Get(HeaderAPISecret)
                    
                    if apiKey != "" && apiSecret != "" {
                        var user *User
                        var err error
                        
                        // 可选：验证签名
                        if o.signEnabled {
                            timestamp := header.RequestHeader().Get(HeaderAPITimestamp)
                            signature := header.RequestHeader().Get(HeaderAPISignature)
                            requestPath := header.Operation()
                            
                            user, err = apiKeyValidator.ValidateWithSignature(
                                ctx, apiKey, apiSecret, timestamp, signature, requestPath,
                            )
                        } else {
                            user, err = apiKeyValidator.Validate(ctx, apiKey, apiSecret)
                        }
                        
                        if err != nil {
                            return nil, err
                        }
                        
                        // 注入上下文
                        ctx = context.WithValue(ctx, "user_id", user.ID)
                        ctx = context.WithValue(ctx, "username", user.Username)
                        ctx = context.WithValue(ctx, "auth_type", "api_key")
                        
                        return handler(ctx, req)
                    }
                }
            }
            
            return handler(ctx, req)
        }
    }
}

// Option 配置选项
type Option func(*options)

type options struct {
    apiKeyEnabled bool
    signEnabled   bool
}

// WithAPIKeyAuth 启用/禁用 API Key 认证
func WithAPIKeyAuth(enabled bool) Option {
    return func(o *options) {
        o.apiKeyEnabled = enabled
    }
}

// WithSignature 启用/禁用签名验证
func WithSignature(enabled bool) Option {
    return func(o *options) {
        o.signEnabled = enabled
    }
}

// extractToken 提取 Bearer Token
func extractToken(header transport.Transporter) string {
    auth := header.RequestHeader().Get("Authorization")
    if len(auth) > 7 && auth[:7] == "Bearer " {
        return auth[7:]
    }
    return ""
}

// handleJWTAuth JWT 认证处理（现有逻辑）
func handleJWTAuth(ctx context.Context, req interface{}, handler middleware.Handler, token string) (interface{}, error) {
    // 现有 JWT 验证逻辑
    return handler(ctx, req)
}
```

### 5.5 中间件配置

`app/admin/service/internal/server/rest_server.go`

```go
// 创建 API Key 验证器
apiKeyValidator := auth.NewAPIKeyValidator(
    data.NewCredentialRepo(dataClient, rdb),
    log.DefaultLogger,
)

// 配置认证中间件
ms := []middleware.Middleware{
    selector.Server(
        auth.Server(
            apiKeyValidator,
            auth.WithAPIKeyAuth(true),
            auth.WithSignature(false), // 可选：签名验证
        ),
        authz.Server(),
    ).Match(rpc.NewRestWhiteListMatcher()).Build(),
}
```

---

## 6. 服务账号管理

### 6.1 创建服务账号

```sql
-- 1. 创建服务账号
INSERT INTO sys_users (id, username, nickname, user_type, status, created_at)
VALUES (90001, 'svc_scheduler', '定时任务服务', 'service_account', 'active', NOW());

-- 2. 生成 API Key（明文存储）
INSERT INTO sys_user_credentials (user_id, credential_type, credential_value, status, created_at)
VALUES (90001, 'API_KEY', 'ak_scheduler_20240101_abc123', 'active', NOW());

-- 3. 生成 API Secret（bcrypt 哈希存储）
-- 原始 Secret: sk_live_REDACTED
-- bcrypt 哈希: $2a$10$N9qo8uLOickgx2ZMRZoMy...
INSERT INTO sys_user_credentials (user_id, credential_type, credential_value, status, created_at)
VALUES (90001, 'API_SECRET', '$2a$10$N9qo8uLOickgx2ZMRZoMy...', 'active', NOW());

-- 4. 分配权限
INSERT INTO sys_user_roles (user_id, role_id)
VALUES (90001, 100);
```

### 6.2 生成凭证

```go
package main

import (
    "fmt"
    "crypto/rand"
    "encoding/hex"
    "golang.org/x/crypto/bcrypt"
)

func generateCredentials(serviceName string) (string, string) {
    // 生成 API Key
    randomKey := make([]byte, 8)
    rand.Read(randomKey)
    apiKey := fmt.Sprintf("ak_%s_%s_%s", 
        serviceName,
        time.Now().Format("20060102"),
        hex.EncodeToString(randomKey)[:12],
    )
    
    // 生成 API Secret
    randomSecret := make([]byte, 16)
    rand.Read(randomSecret)
    apiSecret := "sk_live_REDACTED" + hex.EncodeToString(randomSecret)
    
    // 生成 bcrypt 哈希
    hashedSecret, _ := bcrypt.GenerateFromPassword([]byte(apiSecret), bcrypt.DefaultCost)
    
    fmt.Printf("API Key: %s\n", apiKey)
    fmt.Printf("API Secret (原始): %s\n", apiSecret)
    fmt.Printf("API Secret (哈希): %s\n", string(hashedSecret))
    
    return apiKey, string(hashedSecret)
}

func main() {
    generateCredentials("scheduler")
}
```

### 6.3 凭证管理接口

```go
// CreateServiceAccount 创建服务账号
func (s *Service) CreateServiceAccount(ctx context.Context, req *CreateServiceAccountRequest) (*ServiceAccount, error) {
    // 1. 创建用户记录
    user := &ent.User{
        Username: req.Username,
        Nickname: req.Nickname,
        UserType: "service_account",
        Status:   "active",
    }
    
    // 2. 生成凭证
    apiKey, hashedSecret := generateCredentials(req.ServiceName)
    
    // 3. 保存到数据库
    // ...
    
    return &ServiceAccount{
        ID:       user.ID,
        Username: user.Username,
        APIKey:   apiKey,
        Secret:   hashedSecret,
    }, nil
}

// RevokeCredential 撤销凭证
func (s *Service) RevokeCredential(ctx context.Context, userID int64, credType string) error {
    return s.repo.UpdateCredentialStatus(ctx, userID, credType, "revoked")
}

// RotateCredential 轮换凭证
func (s *Service) RotateCredential(ctx context.Context, userID int64) (*Credential, error) {
    // 1. 创建新凭证
    // 2. 更新数据库
    // 3. 返回新凭证（原凭证仍有效一段时间用于过渡）
}
```

---

## 7. 客户端使用示例

### 7.1 Go 客户端

```go
package api_client

import (
    "bytes"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "net/http"
    "strconv"
    "time"
)

type Client struct {
    BaseURL   string
    APIKey    string
    APISecret string
    SignEnabled bool
}

func NewClient(baseURL, apiKey, apiSecret string) *Client {
    return &Client{
        BaseURL:     baseURL,
        APIKey:      apiKey,
        APISecret:   apiSecret,
        SignEnabled: true,
    }
}

func (c *Client) Do(method, path string, body interface{}) (*http.Response, error) {
    jsonData, _ := json.Marshal(body)
    req, _ := http.NewRequest(method, c.BaseURL+path, bytes.NewBuffer(jsonData))
    
    // 设置认证头
    timestamp := strconv.FormatInt(time.Now().Unix(), 10)
    req.Header.Set(HeaderAPIKey, c.APIKey)
    req.Header.Set(HeaderAPISecret, c.APISecret)
    req.Header.Set(HeaderAPITimestamp, timestamp)
    req.Header.Set("Content-Type", "application/json")
    
    // 可选：签名
    if c.SignEnabled {
        signature := c.signRequest(path, timestamp)
        req.Header.Set(HeaderAPISignature, signature)
    }
    
    return http.DefaultClient.Do(req)
}

func (c *Client) signRequest(path, timestamp string) string {
    data := c.APIKey + c.APISecret + timestamp + path
    h := hmac.New(sha256.New, []byte(c.APISecret))
    h.Write([]byte(data))
    return hex.EncodeToString(h.Sum(nil))
}

// 使用示例
func main() {
    client := NewClient(
        "http://localhost:8000",
        "ak_scheduler_20240101_abc123",
        "sk_live_REDACTED",
    )
    
    resp, err := client.Do("GET", "/api/v1/users", nil)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    // 处理响应...
}
```

### 7.2 Python 客户端

```python
import hmac
import hashlib
import time
import requests

class APIClient:
    def __init__(self, base_url, api_key, api_secret, sign_enabled=True):
        self.base_url = base_url
        self.api_key = api_key
        self.api_secret = api_secret
        self.sign_enabled = sign_enabled
    
    def request(self, method, path, data=None):
        timestamp = str(int(time.time()))
        headers = {
            'X-Api-Key': self.api_key,
            'X-Api-Secret': self.api_secret,
            'X-Api-Timestamp': timestamp,
            'Content-Type': 'application/json',
        }
        
        if self.sign_enabled:
            signature = self._sign_request(path, timestamp)
            headers['X-Api-Signature'] = signature
        
        response = requests.request(
            method,
            f"{self.base_url}{path}",
            json=data,
            headers=headers
        )
        return response
    
    def _sign_request(self, path, timestamp):
        data = f"{self.api_key}{self.api_secret}{timestamp}{path}"
        signature = hmac.new(
            self.api_secret.encode(),
            data.encode(),
            hashlib.sha256
        ).hexdigest()
        return signature


# 使用示例
if __name__ == '__main__':
    client = APIClient(
        'http://localhost:8000',
        'ak_scheduler_20240101_abc123',
        'sk_live_REDACTED'
    )
    
    response = client.request('GET', '/api/v1/users')
    print(response.json())
```

### 7.3 cURL 示例

```bash
# 基本认证
curl -X GET "http://localhost:8000/api/v1/users" \
  -H "X-Api-Key: ak_scheduler_20240101_abc123" \
  -H "X-Api-Secret: sk_live_REDACTED"

# 带签名认证
TIMESTAMP=$(date +%s)
PATH="/api/v1/users"
API_KEY="ak_scheduler_20240101_abc123"
API_SECRET="sk_live_REDACTED"

SIGNATURE=$(echo -n "${API_KEY}${API_SECRET}${TIMESTAMP}${PATH}" | \
  openssl dgst -sha256 -hmac "${API_SECRET}" | \
  awk '{print $2}')

curl -X GET "http://localhost:8000${PATH}" \
  -H "X-Api-Key: ${API_KEY}" \
  -H "X-Api-Secret: ${API_SECRET}" \
  -H "X-Api-Timestamp: ${TIMESTAMP}" \
  -H "X-Api-Signature: ${SIGNATURE}"
```

---

## 8. 凭证命名规范

### 8.1 API Key 格式

```
ak_{service}_{date}_{random}
```

- `ak`: API Key 前缀（固定）
- `service`: 服务名称（如 `scheduler`、`mq-consumer`、`data-sync`）
- `date`: 创建日期，格式 YYYYMMDD
- `random`: 随机字符串，12 位

**示例**：
- `ak_scheduler_20240101_abc123def456`
- `ak_mq_consumer_20240115_xyz789ghi012`

### 8.2 API Secret 格式

```
sk_{env}_{random}
```

- `sk`: Secret Key 前缀（固定）
- `env`: 环境标识（`live`、`test`、`dev`）
- `random`: 随机字符串，32 位

**示例**：
 

---

## 9. 安全最佳实践

### 9.1 凭证存储

| 环境 | 存储方式 |
|------|----------|
| 生产环境 | 密钥管理服务（Vault、AWS Secrets Manager） |
| 开发环境 | 环境变量 |
| CI/CD | 加密的环境变量或 Secrets |

**禁止**：
- ❌ 硬编码在代码中
- ❌ 提交到版本控制系统
- ❌ 通过日志、聊天工具传输

### 9.2 凭证轮换

建议周期：**90 天**

轮换流程：
1. 生成新凭证对
2. 更新客户端配置
3. 验证新凭证可用
4. 撤销旧凭证
5. 记录审计日志

### 9.3 权限控制

遵循**最小权限原则**：

```yaml
service_accounts:
  - name: svc_scheduler
    role: scheduler
    permissions:
      - api:users:read
      - api:reports:write
      # 只分配必要的权限
```

### 9.4 监控告警

关键指标：
- 认证失败率
- 凭证过期预警
- 异常访问模式

```yaml
alerts:
  - name: api_key_auth_failure_high
    expr: rate(auth_failure_total{auth_type="api_key"}[5m]) > 0.1
    message: "API Key 认证失败率过高"
    
  - name: credential_expiring_soon
    expr: credential_expires_in_days < 7
    message: "服务账号凭证即将过期"
```

---

## 10. 故障排查

### 10.1 常见错误

| 错误 | 原因 | 解决方案 |
|------|------|----------|
| `invalid API key` | API Key 不存在或错误 | 检查 API Key 是否正确 |
| `invalid API secret` | Secret 错误或哈希不匹配 | 检查 Secret 是否正确 |
| `credential has expired` | 凭证已过期 | 轮换凭证 |
| `credential has been revoked` | 凭证已被撤销 | 生成新凭证 |
| `request timestamp expired` | 时间戳超过 5 分钟 | 同步服务器时间 |
| `invalid request signature` | 签名验证失败 | 检查签名算法 |

### 10.2 调试技巧

```go
// 在验证器中添加调试日志
func (v *APIKeyValidator) Validate(ctx context.Context, apiKey, apiSecret string) (*User, error) {
    v.logger.Infof("validating API Key: %s", apiKey[:20]+"...")
    
    // 验证过程...
    
    if err != nil {
        v.logger.Warnf("validation failed: %v", err)
    }
    
    return user, err
}
```

---

## 11. 实施清单

### 11.1 准备阶段

- [ ] 扩展用户类型枚举（添加 `service_account`）
- [ ] 确认 `sys_user_credentials` 表包含 `API_KEY` 和 `API_SECRET` 类型
- [ ] 实现凭证数据访问接口

### 11.2 开发阶段

- [ ] 创建常量定义文件
- [ ] 创建错误定义文件
- [ ] 实现 API Key 验证器
- [ ] 扩展认证中间件
- [ ] 添加凭证管理接口

### 11.3 部署阶段

- [ ] 创建服务账号
- [ ] 生成并分发凭证
- [ ] 配置权限
- [ ] 配置监控告警
- [ ] 编写运维文档

### 11.4 运维阶段

- [ ] 定期轮换凭证
- [ ] 审计访问日志
- [ ] 监控凭证状态
- [ ] 权限定期审计

---

## 12. 总结

本方案通过 **API Key + API Secret** 实现简洁高效的服务间认证：

**优点**：
- 实现简单，无需复杂的 OAuth 流程
- 复用现有表结构，改动最小
- 支持可选签名机制增强安全性
- 易于管理和轮换

**适用场景**：
- 定时任务服务调用 API
- 消息队列消费者访问服务
- 微服务间内部调用
- 第三方系统集成

**不适用场景**：
- 用户身份认证（使用 JWT）
- 需要细粒度授权的场景（考虑 OAuth）
- 公开 API（考虑 API Gateway）
