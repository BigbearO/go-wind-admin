# 服务账号与 API Key 认证方案

## 1. 方案概述

本方案旨在解决微服务架构中服务间调用的认证问题，特别是定时任务、消息队列消费者等场景下无法携带用户 Token 的问题。采用 **Service Account（服务账号）** 模式，结合 **API Key + API Secret** 双重认证机制。

### 核心设计原则

- **复用现有表结构**：服务账号存储在 `sys_users` 表，凭证存储在 `sys_user_credentials` 表
- **双认证方式支持**：
  - 直接 API Key 认证：适用于服务间简单调用
  - OAuth 2.0 client_credentials 流程：适用于需要标准 OAuth 令牌的场景
- **最小化代码改动**：复用现有认证中间件和用户体系

---

## 2. 数据模型设计

### 2.1 服务账号（复用 `sys_users` 表）

服务账号作为特殊用户存储在现有 `sys_users` 表中：

| 字段 | 说明 | 示例值 |
|------|------|--------|
| `id` | 服务账号 ID，预留范围 90001-99999 | 90001 |
| `username` | 服务账号标识，`svc_` 前缀 | `svc_scheduler` |
| `nickname` | 服务名称 | `定时任务服务` |
| `status` | 账号状态 | `active` |
| `user_type` | 用户类型（新增枚举值） | `service_account` |
| `email` | 可选，用于告警通知 | `scheduler@example.com` |

#### 用户类型扩展

```go
// 用户类型枚举
const (
    UserTypeRegularUser     = "user"           // 普通用户
    UserTypeServiceAccount  = "service_account" // 服务账号
)
```

### 2.2 API 凭证（复用 `sys_user_credentials` 表）

凭证存储在现有 `sys_user_credentials` 表中：

| 字段 | 说明 | 示例值 |
|------|------|--------|
| `user_id` | 关联服务账号 ID | 90001 |
| `credential_type` | 凭证类型 | `API_KEY` / `API_SECRET` |
| `credential_value` | 凭证值（API Secret 存储 bcrypt 哈希） | 哈希值或明文 Key |
| `expires_at` | 过期时间（可选） | `2025-12-31 23:59:59` |
| `status` | 凭证状态 | `active` / `revoked` |

#### 凭证类型说明

```go
// 凭证类型枚举（现有定义中已包含）
const (
    CredentialTypePassword  = "PASSWORD"
    CredentialTypeAPIKey    = "API_KEY"      // API Key（明文存储）
    CredentialTypeAPISecret = "API_SECRET"   // API Secret（bcrypt 哈希存储）
)
```

---

## 3. 认证方式

### 3.1 方式一：直接 API Key 认证

#### 请求头格式

```http
X-Api-Key: ak_scheduler_20240101_abc123
X-Api-Secret: sk_live_xyzxxxxxxx
X-Api-Timestamp: 1704067200
X-Api-Signature: a1b2c3d4e5f6...  // 可选，签名增强安全性
```

#### 认证流程

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│  调用服务    │         │  API Gateway │         │  目标服务    │
└──────┬──────┘         └──────┬──────┘         └──────┬──────┘
       │                       │                       │
       │ 1. 携带 API Key/Secret│                       │
       │──────────────────────>│                       │
       │                       │                       │
       │                       │ 2. 验证凭证            │
       │                       │──────────────────────>│
       │                       │                       │
       │                       │ 3. 返回服务账号信息    │
       │                       │<──────────────────────│
       │                       │                       │
       │ 4. 转发请求            │                       │
       │<──────────────────────│                       │
       │                       │                       │
```

#### 签名算法（可选增强）

```go
// 签名生成公式
signature = HmacSHA256(
    apiKey + apiSecret + timestamp + requestPath,
    apiSecret
)
```

### 3.2 方式二：OAuth 2.0 client_credentials 流程

#### 请求格式

```http
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
&client_id=ak_scheduler_20240101_abc123
&client_secret=sk_live_xyzxxxxxxx
&scope=api:read api:write
```

#### 响应格式

```json
{
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "scope": "api:read api:write"
}
```

#### 认证流程

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│  调用服务    │         │   Auth 服务  │         │  目标服务    │
└──────┬──────┘         └──────┬──────┘         └──────┬──────┘
       │                       │                       │
       │ 1. client_credentials │                       │
       │──────────────────────>│                       │
       │                       │                       │
       │ 2. 返回 Access Token  │                       │
       │<──────────────────────│                       │
       │                       │                       │
       │ 3. 携带 Bearer Token  │                       │
       │───────────────────────────────────────────────>│
       │                       │                       │
       │ 4. 返回响应           │                       │
       │<───────────────────────────────────────────────│
       │                       │                       │
```

---

## 4. 实现细节

### 4.1 文件结构

```
backend/
├── pkg/
│   └── middleware/
│       └── auth/
│           ├── auth.go              # 认证中间件（扩展 API Key 支持）
│           ├── api_key_validator.go # API Key 验证器
│           ├── constants.go         # 常量定义
│           ├── errors.go            # 错误定义
│           └── options.go           # 配置选项
└── app/admin/service/
    ├── internal/
    │   ├── data/
    │   │   └── user_credential.go   # 凭证数据访问
    │   └── service/
    │       └── authentication_service.go  # 认证服务（添加 client_credentials）
    └── internal/server/
        └── rest_server.go           # 中间件配置
```

### 4.2 核心代码实现

#### 4.2.1 常量定义（`pkg/middleware/auth/constants.go`）

```go
package auth

const (
    // 请求头
    HeaderAPIKey      = "X-Api-Key"
    HeaderAPISecret   = "X-Api-Secret"
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

#### 4.2.2 错误定义（`pkg/middleware/auth/errors.go`）

```go
package auth

import "errors"

var (
    ErrInvalidAPIKey       = errors.New("invalid API key")
    ErrInvalidAPISecret    = errors.New("invalid API secret")
    ErrExpiredCredential   = errors.New("credential has expired")
    ErrRevokedCredential   = errors.New("credential has been revoked")
    ErrTimestampExpired    = errors.New("request timestamp expired")
    ErrInvalidSignature    = errors.New("invalid request signature")
    ErrNotServiceAccount   = errors.New("user is not a service account")
)
```

#### 4.2.3 API Key 验证器（`pkg/middleware/auth/api_key_validator.go`）

```go
package auth

import (
    "context"
    "time"
    
    "github.com/go-kratos/kratos/v2/log"
    "golang.org/x/crypto/bcrypt"
)

type APIKeyValidator struct {
    credentialRepo CredentialRepository
    logger         *log.Helper
    allowSignature bool
    timestampTTL   time.Duration
}

type CredentialRepository interface {
    GetCredentialByType(ctx context.Context, userID int64, credType string) (*UserCredential, error)
    GetUserByID(ctx context.Context, userID int64) (*User, error)
}

type UserCredential struct {
    ID             int64
    UserID         int64
    CredentialType string
    CredentialValue string
    ExpiresAt      *time.Time
    Status         string
}

type User struct {
    ID       int64
    Username string
    UserType string
    Status   string
}

func NewAPIKeyValidator(repo CredentialRepository, logger log.Logger) *APIKeyValidator {
    return &APIKeyValidator{
        credentialRepo: repo,
        logger:         log.NewHelper(logger),
        allowSignature: true,
        timestampTTL:   5 * time.Minute,
    }
}

// ValidateAPIKey 验证 API Key 和 Secret
func (v *APIKeyValidator) ValidateAPIKey(ctx context.Context, apiKey, apiSecret string) (*User, error) {
    // 1. 根据 API Key 查找用户 ID（API Key 格式: ak_{service}_{date}_{random}）
    userID, err := v.parseAPIKeyToUserID(ctx, apiKey)
    if err != nil {
        return nil, ErrInvalidAPIKey
    }
    
    // 2. 获取用户信息，验证是否为服务账号
    user, err := v.credentialRepo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    if user.UserType != UserTypeServiceAccount {
        return nil, ErrNotServiceAccount
    }
    
    // 3. 验证 API Key
    storedAPIKey, err := v.credentialRepo.GetCredentialByType(ctx, userID, CredentialTypeAPIKey)
    if err != nil || storedAPIKey.CredentialValue != apiKey {
        return nil, ErrInvalidAPIKey
    }
    
    if err := v.checkCredentialStatus(storedAPIKey); err != nil {
        return nil, err
    }
    
    // 4. 验证 API Secret（bcrypt 对比）
    storedAPISecret, err := v.credentialRepo.GetCredentialByType(ctx, userID, CredentialTypeAPISecret)
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
    // 1. 验证时间戳（防止重放攻击）
    ts, err := strconv.ParseInt(timestamp, 10, 64)
    if err != nil {
        return nil, ErrTimestampExpired
    }
    
    if time.Since(time.Unix(ts, 0)) > v.timestampTTL {
        return nil, ErrTimestampExpired
    }
    
    // 2. 验证签名
    expectedSig := v.generateSignature(apiKey, apiSecret, timestamp, requestPath)
    if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
        return nil, ErrInvalidSignature
    }
    
    // 3. 验证 API Key 和 Secret
    return v.ValidateAPIKey(ctx, apiKey, apiSecret)
}

func (v *APIKeyValidator) parseAPIKeyToUserID(ctx context.Context, apiKey string) (int64, error) {
    // 从缓存或数据库查询 API Key 对应的用户 ID
    // 实现依赖具体的存储方案
    return 0, nil
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

func (v *APIKeyValidator) generateSignature(apiKey, apiSecret, timestamp, path string) string {
    data := apiKey + apiSecret + timestamp + path
    h := hmac.New(sha256.New, []byte(apiSecret))
    h.Write([]byte(data))
    return hex.EncodeToString(h.Sum(nil))
}
```

#### 4.2.4 认证中间件扩展（`pkg/middleware/auth/auth.go`）

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
    }
    for _, opt := range opts {
        opt(o)
    }
    
    return func(handler middleware.Handler) middleware.Handler {
        return func(ctx context.Context, req interface{}) (interface{}, error) {
            // 1. 从请求头获取 Token
            if header, ok := transport.FromServerContext(ctx); ok {
                // 优先检查 Bearer Token
                token := extractToken(header)
                if token != "" {
                    // JWT 验证逻辑（现有）
                    return handleJWTAuth(ctx, req, handler, token)
                }
                
                // 检查 API Key 认证
                if o.apiKeyEnabled {
                    apiKey := header.RequestHeader().Get(HeaderAPIKey)
                    apiSecret := header.RequestHeader().Get(HeaderAPISecret)
                    
                    if apiKey != "" && apiSecret != "" {
                        user, err := apiKeyValidator.ValidateAPIKey(ctx, apiKey, apiSecret)
                        if err != nil {
                            return nil, err
                        }
                        
                        // 将服务账号信息注入上下文
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
}

func WithAPIKeyAuth(enabled bool) Option {
    return func(o *options) {
        o.apiKeyEnabled = enabled
    }
}
```

#### 4.2.5 client_credentials 流程（`app/admin/service/internal/service/authentication_service.go`）

```go
package service

import (
    "context"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
)

// OAuthTokenRequest OAuth 令牌请求
type OAuthTokenRequest struct {
    GrantType    string `json:"grant_type"`
    ClientID     string `json:"client_id"`
    ClientSecret string `json:"client_secret"`
    Scope        string `json:"scope"`
}

// OAuthTokenResponse OAuth 令牌响应
type OAuthTokenResponse struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
    ExpiresIn   int64  `json:"expires_in"`
    Scope       string `json:"scope"`
}

// OAuthToken 处理 OAuth 令牌请求
func (s *AuthenticationService) OAuthToken(ctx context.Context, req *OAuthTokenRequest) (*OAuthTokenResponse, error) {
    if req.GrantType != "client_credentials" {
        return nil, errors.New("unsupported grant type")
    }
    
    // 1. 验证 client_id 和 client_secret（即 API Key 和 Secret）
    user, err := s.apiKeyValidator.ValidateAPIKey(ctx, req.ClientID, req.ClientSecret)
    if err != nil {
        return nil, err
    }
    
    // 2. 生成 JWT Token
    expiresAt := time.Now().Add(1 * time.Hour)
    claims := &jwt.MapClaims{
        "sub":       user.ID,
        "username":  user.Username,
        "user_type": "service_account",
        "scope":     req.Scope,
        "exp":       expiresAt.Unix(),
        "iat":       time.Now().Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(s.jwtSecret)
    if err != nil {
        return nil, err
    }
    
    return &OAuthTokenResponse{
        AccessToken: tokenString,
        TokenType:   "Bearer",
        ExpiresIn:   3600,
        Scope:       req.Scope,
    }, nil
}
```

### 4.3 中间件配置

在 `rest_server.go` 中配置中间件，支持 API Key 认证：

```go
// 创建 API Key 验证器
apiKeyValidator := auth.NewAPIKeyValidator(data.NewCredentialRepo(dataClient, rdb), log.DefaultLogger)

// 配置认证中间件
ms := []middleware.Middleware{
    selector.Server(
        auth.Server(apiKeyValidator, auth.WithAPIKeyAuth(true)),
        authz.Server(),
    ).Match(rpc.NewRestWhiteListMatcher()).Build(),
}
```

---

## 5. 服务账号管理

### 5.1 创建服务账号

```sql
-- 1. 创建服务账号
INSERT INTO sys_users (id, username, nickname, user_type, status, created_at)
VALUES (90001, 'svc_scheduler', '定时任务服务', 'service_account', 'active', NOW());

-- 2. 生成 API Key（可读格式）
-- API Key: ak_{service}_{date}_{random}
INSERT INTO sys_user_credentials (user_id, credential_type, credential_value, status, created_at)
VALUES (90001, 'API_KEY', 'ak_scheduler_20240101_abc123', 'active', NOW());

-- 3. 生成 API Secret（bcrypt 哈希）
-- 原始 Secret: sk_live_xyzxxxxxxx
INSERT INTO sys_user_credentials (user_id, credential_type, credential_value, status, created_at)
VALUES (90001, 'API_SECRET', '$2a$10$...', 'active', NOW());
```

### 5.2 服务账号权限控制

服务账号的权限通过现有的 RBAC 体系管理：

```sql
-- 为服务账号分配角色
INSERT INTO sys_user_roles (user_id, role_id)
VALUES (90001, 100);  -- 假设角色 100 是定时任务专用角色

-- 角色权限配置（现有表）
-- sys_role_menus 等表配置角色的具体权限
```

---

## 6. 使用示例

### 6.1 直接 API Key 方式

#### Go 客户端示例

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "time"
)

type APIClient struct {
    BaseURL   string
    APIKey    string
    APISecret string
}

func (c *APIClient) DoRequest(method, path string, body interface{}) (*http.Response, error) {
    jsonData, _ := json.Marshal(body)
    req, _ := http.NewRequest(method, c.BaseURL+path, bytes.NewBuffer(jsonData))
    
    // 设置 API Key 认证头
    timestamp := strconv.FormatInt(time.Now().Unix(), 10)
    signature := c.generateSignature(path, timestamp)
    
    req.Header.Set("X-Api-Key", c.APIKey)
    req.Header.Set("X-Api-Secret", c.APISecret)
    req.Header.Set("X-Api-Timestamp", timestamp)
    req.Header.Set("X-Api-Signature", signature)
    req.Header.Set("Content-Type", "application/json")
    
    return http.DefaultClient.Do(req)
}

func main() {
    client := &APIClient{
        BaseURL:   "http://localhost:8000",
        APIKey:    "ak_scheduler_20240101_abc123",
        APISecret: "sk_live_xyzxxxxxxx",
    }
    
    resp, err := client.DoRequest("GET", "/api/v1/users", nil)
    // ...
}
```

### 6.2 client_credentials 方式

#### Go 客户端示例

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/url"
    "strings"
)

type OAuthClient struct {
    TokenURL    string
    ClientID    string
    ClientSecret string
    Scope       string
    accessToken string
}

// GetAccessToken 获取访问令牌
func (c *OAuthClient) GetAccessToken() error {
    data := url.Values{}
    data.Set("grant_type", "client_credentials")
    data.Set("client_id", c.ClientID)
    data.Set("client_secret", c.ClientSecret)
    data.Set("scope", c.Scope)
    
    resp, err := http.Post(c.TokenURL, 
        "application/x-www-form-urlencoded",
        strings.NewReader(data.Encode()))
    if err != nil {
        return err
    }
    
    var result struct {
        AccessToken string `json:"access_token"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    
    c.accessToken = result.AccessToken
    return nil
}

func main() {
    client := &OAuthClient{
        TokenURL:     "http://localhost:8000/oauth/token",
        ClientID:     "ak_scheduler_20240101_abc123",
        ClientSecret: "sk_live_xyzxxxxxxx",
        Scope:        "api:read",
    }
    
    // 获取 Token
    client.GetAccessToken()
    
    // 使用 Token 调用 API
    req, _ := http.NewRequest("GET", "http://localhost:8000/api/v1/users", nil)
    req.Header.Set("Authorization", "Bearer "+client.accessToken)
    
    resp, _ := http.DefaultClient.Do(req)
    // ...
}
```

---

## 7. 安全考虑

### 7.1 凭证安全

- **API Secret 使用 bcrypt 哈希存储**，即使数据库泄露也无法逆向
- **API Key 可读存储**，但仅作为标识符，不作为敏感凭证
- **支持凭证过期时间**，定期轮换凭证

### 7.2 传输安全

- **强制 HTTPS**：所有 API 请求必须通过 HTTPS
- **签名机制**：可选的请求签名防止篡改和重放攻击
- **时间戳验证**：请求需携带时间戳，过期请求拒绝（默认 5 分钟）

### 7.3 权限控制

- **最小权限原则**：服务账号仅分配必要的权限
- **审计日志**：记录所有服务账号的操作日志
- **定期审计**：定期检查服务账号的权限和凭证状态

### 7.4 凭证管理

- **凭证轮换**：定期更换 API Secret（建议 90 天）
- **紧急撤销**：支持立即撤销凭证
- **多凭证管理**：支持同一服务账号创建多个凭证对（用于轮换过渡期）

---

## 8. 最佳实践

### 8.1 凭证命名规范

```
API Key: ak_{service}_{date}_{random}
  - ak: API Key 前缀
  - service: 服务名称（如 scheduler、mq-consumer）
  - date: 创建日期（YYYYMMDD）
  - random: 随机字符串（8-16 位）

示例: ak_scheduler_20240101_abc123

API Secret: sk_{env}_{random}
  - sk: Secret Key 前缀
  - env: 环境标识（live、test、dev）
  - random: 随机字符串（16-32 位）

示例: sk_live_xyzxxxxxxxghi012
```

### 8.2 凭证存储

服务端存储：
- API Key：明文存储（用于查询）
- API Secret：bcrypt 哈希存储（cost=10）

客户端存储：
- 使用环境变量或密钥管理服务（如 Vault）
- 禁止硬编码在代码中
- 禁止提交到版本控制系统

### 8.3 权限规划

```yaml
# 服务账号权限规划示例
service_accounts:
  - name: svc_scheduler
    role: scheduler_role
    permissions:
      - api:users:read
      - api:reports:write
    
  - name: svc_mq_consumer
    role: mq_consumer_role
    permissions:
      - api:orders:read
      - api:notifications:write
```

---

## 9. 监控与告警

### 9.1 监控指标

- 认证失败率
- 凭证过期预警
- 服务账号活跃度
- API 调用量统计

### 9.2 告警规则

```yaml
alerts:
  - name: api_key_auth_failure_high
    expr: rate(auth_failure_total{auth_type="api_key"}[5m]) > 0.1
    severity: warning
    message: "API Key 认证失败率过高"
  
  - name: credential_expiring_soon
    expr: credential_expires_in_hours < 168  # 7 天
    severity: warning
    message: "服务账号凭证即将过期"
```

---

## 10. 迁移计划

### 10.1 阶段一：基础设施准备

1. 扩展用户类型枚举
2. 实现 API Key 验证器
3. 扩展认证中间件
4. 实现 client_credentials 流程

### 10.2 阶段二：创建服务账号

1. 创建服务账号（定时任务、MQ 消费者等）
2. 生成 API Key/Secret 对
3. 配置权限
4. 分发凭证

### 10.3 阶段三：应用集成

1. 调用方集成 API Key 或 OAuth 客户端
2. 测试验证
3. 灰度上线
4. 监控告警配置

---

## 11. 常见问题

### Q1: API Key 和 JWT Token 的区别？

| 特性 | API Key | JWT Token |
|------|---------|-----------|
| 适用场景 | 服务间调用 | 用户会话 |
| 过期机制 | 可配置过期时间 | 短期有效（如 1 小时） |
| 携带方式 | 请求头 | Authorization Bearer |
| 权限粒度 | 粗粒度（服务级别） | 细粒度（用户级别） |
| 管理复杂度 | 低 | 中 |

### Q2: 为什么复用现有用户表？

- 避免数据冗余和表结构膨胀
- 复用现有的 RBAC 权限体系
- 简化代码改动，降低维护成本
- 统一的用户管理界面

### Q3: 如何处理凭证泄露？

1. 立即撤销泄露的凭证
2. 生成新的凭证对
3. 更新所有使用该凭证的服务
4. 审计泄露期间的访问日志
5. 加强凭证管理流程

### Q4: 多实例部署如何处理？

- 所有实例共享同一套凭证
- 使用分布式缓存存储 API Key -> UserID 映射
- 凭证验证无状态，支持水平扩展

---

## 12. 参考资料

- [OAuth 2.0 Client Credentials Grant](https://oauth.net/2/grant-types/client-credentials/)
- [API Key Best Practices](https://www.googleapis.com/auth/apikeys)
- [Kratos Middleware](https://go-kratos.dev/docs/component/middleware)
- [bcrypt Hashing](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
