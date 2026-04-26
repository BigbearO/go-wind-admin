# Game 服务调用 Admin 服务的 RPC Client 方案

## 一、现状分析

### 1.1 项目结构

```
app/game/service/
├── internal/
│   ├── data/
│   │   ├── admin_client.go        # 当前使用硬编码地址
│   │   └── providers/
│   │       └── wire_set.go
│   └── service/
│       ├── test_service.go
│       └── providers/
├── configs/
│   ├── client.yaml               # 已配置 discovery，但未使用
│   ├── config.yaml
│   ├── data.yaml
│   └── ...
└── cmd/server/
    ├── main.go
    └── wire.go
```

### 1.2 已有配置 (client.yaml)

```yaml
client:
  discovery:
    type: nacos
    nacos:
      addr: "127.0.0.1:8848"
      namespace: "public"
      group: "DEFAULT_GROUP"
```

### 1.3 现有问题

- `data/admin_client.go` 使用硬编码地址 `http://127.0.0.1:7788`
- 未利用已有的 discovery 配置
- 无法动态发现服务地址

---

## 二、方案设计

### 2.1 配置文件优化

```yaml
# configs/client.yaml
client:
  discovery:
    type: nacos          # 支持: nacos, consul, etcd, eureka, kubernetes, polaris
    nacos:
      addr: "127.0.0.1:8848"
      namespace: "public"
      group: "DEFAULT_GROUP"

  # 新增：HTTP Client 配置
  http_clients:
    admin:
      service_name: go-wind-admin    # 注册中心的服务名
      timeout: 3s
      enabled: true
```

### 2.2 目录结构设计

```
app/game/service/
├── internal/
│   ├── data/
│   │   ├── admin_client.go        # 改造：使用工厂创建客户端
│   │   ├── providers/
│   │   │   └── wire_set.go
│   │   └── client/                 # 新增：客户端工厂
│   │       ├── client_factory.go   # HTTP Client 工厂
│   │       ├── registry.go        # 注册中心创建
│   │       └── wire.go             # Wire 注入
│   └── service/
└── configs/
```

---

## 三、核心代码设计

### 3.1 注册中心创建 (client/registry.go)

```go
package client

import (
    "fmt"

    "github.com/go-kratos/kratos/v2/registry"
    "github.com/tx7do/kratos-bootstrap/registry/nacos"
    "github.com/tx7do/kratos-bootstrap/registry/consul"
    "github.com/tx7do/kratos-bootstrap/registry/etcd"
)

type RegistryConfig struct {
    Type   string
    Nacos  *NacosConfig
    Consul *ConsulConfig
    Etcd   *EtcdConfig
}

type NacosConfig struct {
    Addr      string
    Namespace  string
    Group     string
}

type ConsulConfig struct {
    Addr string
}

type EtcdConfig struct {
    Addr string
}

func NewDiscovery(cfg *RegistryConfig) (registry.Discovery, error) {
    switch cfg.Type {
    case "nacos":
        return nacos.NewRegistry([]string{cfg.Nacos.Addr})
    case "consul":
        return consul.NewRegistry([]string{cfg.Consul.Addr})
    case "etcd":
        return etcd.NewRegistry([]string{cfg.Etcd.Addr})
    default:
        return nil, fmt.Errorf("unsupported registry: %s", cfg.Type)
    }
}
```

### 3.2 HTTP Client 工厂 (client/client_factory.go)

```go
package client

import (
    "context"
    "time"

    "github.com/go-kratos/kratos/v2/registry"
    "github.com/go-kratos/kratos/v2/transport/http"

    adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
    identityV1 "go-wind-admin/api/gen/go/identity/service/v1"
)

type HTTPClientFactory struct {
    registry registry.Discovery
    clients  map[string]*http.Client
}

type ClientConfig struct {
    ServiceName string
    Timeout     time.Duration
    Enabled     bool
}

func NewHTTPClientFactory(reg registry.Discovery) *HTTPClientFactory {
    return &HTTPClientFactory{
        registry: reg,
        clients:  make(map[string]*http.Client),
    }
}

func (f *HTTPClientFactory) getClient(serviceName string, timeout time.Duration) (*http.Client, error) {
    if c, ok := f.clients[serviceName]; ok {
        return c, nil
    }

    conn, err := http.NewClient(
        context.Background(),
        http.WithEndpoint("discovery:///" + serviceName),  // 从注册中心发现
        http.WithDiscovery(f.registry),
        http.WithTimeout(timeout),
    )
    if err != nil {
        return nil, err
    }

    f.clients[serviceName] = conn
    return conn, nil
}

// 为各服务生成便捷方法
func (f *HTTPClientFactory) NewUserServiceClient() (adminV1.UserServiceHTTPClient, error) {
    conn, err := f.getClient("go-wind-admin", 3*time.Second)
    if err != nil {
        return nil, err
    }
    return adminV1.NewUserServiceHTTPClient(conn), nil
}

func (f *HTTPClientFactory) NewIdentityUserClient() (identityV1.UserServiceHTTPClient, error) {
    conn, err := f.getClient("go-wind-admin", 3*time.Second)
    if err != nil {
        return nil, err
    }
    return identityV1.NewUserServiceHTTPClient(conn), nil
}

func (f *HTTPClientFactory) NewRoleServiceClient() (adminV1.RoleServiceHTTPClient, error) {
    conn, err := f.getClient("go-wind-admin", 3*time.Second)
    if err != nil {
        return nil, err
    }
    return adminV1.NewRoleServiceHTTPClient(conn), nil
}
```

### 3.3 Wire 注入 (client/wire.go)

```go
//go:build wireinject
// +build wireinject

package client

import (
    "github.com/google/wire"
)

var ProviderSet = wire.NewSet(
    NewDiscovery,
    NewHTTPClientFactory,
)
```

### 3.4 改造现有 AdminClient (data/admin_client.go)

```go
package data

import (
    "go-wind-admin/app/game/service/internal/data/client"

    adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
    identityV1 "go-wind-admin/api/gen/go/identity/service/v1"
)

type AdminClient struct {
    UserService     adminV1.UserServiceHTTPClient
    IdentityUserSvc identityV1.UserServiceHTTPClient
}

func NewAdminClient(
    factory *client.HTTPClientFactory,
) (*AdminClient, error) {
    userSvc, err := factory.NewUserServiceClient()
    if err != nil {
        return nil, err
    }

    identityUserSvc, err := factory.NewIdentityUserClient()
    if err != nil {
        return nil, err
    }

    return &AdminClient{
        UserService:     userSvc,
        IdentityUserSvc: identityUserSvc,
    }, nil
}
```

### 3.5 更新 Wire Set (data/providers/wire_set.go)

```go
var ProviderSet = wire.NewSet(
    data.NewEntClient,
    data.NewGameAccountRepo,
    data.NewAdminClient,
    client.ProviderSet,  // 新增
)
```

---

## 四、Kratos 原生方式（参考）

如果不使用工厂模式，也可以直接使用 Kratos 原生方式：

```go
// 1. 创建注册中心
r := nacos.NewRegistry([]string{"127.0.0.1:8848"})

// 2. 创建带服务发现的HTTP Client
conn, err := http.NewClient(
    context.Background(),
    http.WithEndpoint("discovery:///go-wind-admin"),  // 格式: discovery:///<serviceName>
    http.WithDiscovery(r),
    http.WithTimeout(3 * time.Second),
)

// 3. 使用生成的HTTP Client
client := adminV1.NewUserServiceHTTPClient(conn)
```

---

## 五、方案优势

| 特性 | 说明 |
|------|------|
| 多注册中心 | 通过配置切换 Nacos/Consul/Etcd/K8s 等 |
| Wire 集成 | 与现有项目保持一致 |
| 复用配置 | 利用现有的 client.yaml discovery 配置 |
| 扩展性 | 新增服务只需配置 + 添加方法 |
| 服务发现 | 动态从注册中心获取地址 |

---

## 六、实施步骤

1. **修改配置文件** - `configs/client.yaml`
2. **创建 client 包** - `internal/data/client/`
3. **更新 wire_set** - 添加新 provider
4. **改造 admin_client.go** - 使用工厂创建客户端

---

## 七、依赖说明

需要引入以下 kratos-bootstrap 注册中心模块（根据使用的注册中心选择）：

```go
import (
    // Nacos
    _ "github.com/tx7do/kratos-bootstrap/registry/nacos"
    // Consul
    _ "github.com/tx7do/kratos-bootstrap/registry/consul"
    // Etcd
    _ "github.com/tx7do/kratos-bootstrap/registry/etcd"
    // Eureka
    _ "github.com/tx7do/kratos-bootstrap/registry/eureka"
    // Kubernetes
    _ "github.com/tx7do/kratos-bootstrap/registry/kubernetes"
    // Polaris
    _ "github.com/tx7do/kratos-bootstrap/registry/polaris"
)
```
