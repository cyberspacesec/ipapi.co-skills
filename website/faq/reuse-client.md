# ❓ 能复用 Client 吗

## 问题

我在一个长运行的服务里要查很多次 IP，需要每次查询都 `NewClient()` 新建一个 `Client` 吗？同一个 `Client` 能不能多次调用、跨 goroutine 复用？

## 简答

能，而且**应该**复用——`Client` 是线程安全的，底层会复用连接池，长期保持同一个实例性能最好。

## 详解

::: tip 🎨 一图抵千言
下图对比「复用同一个 Client」与「每次新建 Client」两条路径，直观展示连接池复用带来的差异。
:::

```mermaid
flowchart TD
    subgraph REUSE["✅ 推荐：复用 Client"]
        R1["进程启动时 NewClient 一次"] --> R2["共享 HTTPClient + Transport"]
        R2 --> R3["连接池持续生效<br/>空闲连接按 host 缓存"]
        R3 --> R4["每次 GetIPInfo<br/>直接拿已有连接"]
        R4 --> R5["省去 DNS / TCP / TLS 握手<br/>延迟最低"]
    end

    subgraph NEW["❌ 反例：每次新建"]
        N1["每次查询 NewClient"] --> N2["新建 HTTPClient + Transport"]
        N2 --> N3["连接池形同虚设"]
        N3 --> N4["每次都要 DNS + TCP + TLS 握手"]
        N4 --> N5["高频场景又慢又浪费资源"]
    end
```

### ✅ 为什么能复用

`Client` 在每次发起请求时都会用 [`http.NewRequestWithContext`](https://pkg.go.dev/net/http#NewRequestWithContext) **新建一个独立的 `*http.Request`**，方法之间不共享任何可变状态：

```go
// api.go 内部，每次调用都新建 request
req, err := newGetRequest(ctx, c.BaseURL, ip, format)
c.applyAuth(req)   // 只改本次 request 的 header / query
c.setHeaders(req)  // 只改本次 request 的 User-Agent
resp, err := c.doRequest(req)
```

字段层面也分得很清楚：

| 字段 | 可变性 | 说明 |
|------|--------|------|
| `HTTPClient` | 创建后只读 | 底层 `*http.Client` + `Transport`，天然并发安全 |
| `BaseURL` / `APIKey` / `UserAgent` | 创建后只读 | 配置项，运行期不改动 |
| `Retries` / `RateLimiter` | 读取即用 | `RateLimiter` 是 `<-chan time.Time`，多 goroutine 自动排队 |
| `*http.Request` | 每次新建 | 不在 `Client` 上存留 |

因此多个 goroutine 同时调用 `client.GetIPInfo(...)` 完全安全，**无需额外加锁**。

### 🚀 复用 = 复用连接池

`NewClient` 默认创建的 `*http.Client` 底层是 Go 标准库的 `http.Transport`，它自带**连接池**：

- 🤝 复用 TCP / TLS 连接，省掉重复握手开销
- ♻️ 空闲连接按 host 缓存，下次请求直接拿
- ⏱ 默认 `IdleConnTimeout` 自动回收闲置连接

如果你**每次都 `NewClient()`**，等于每次都建一个新的 `http.Client` + 新的 `Transport`，连接池形同虚设，每次查询都要重新 DNS、TCP、TLS 握手——在高频场景下既慢又浪费资源。

```go
// ❌ 反例：每次请求都新建，连接池用不上
func lookup(ip string) (*ipapi.IPInfo, error) {
    client := ipapi.NewClient() // 别这么干
    return client.GetIPInfo(context.Background(), ip, "json")
}
```

```go
// ✅ 正例：复用同一个实例，连接池持续生效
var client = ipapi.NewClient() // 进程级单例

func lookup(ip string) (*ipapi.IPInfo, error) {
    return client.GetIPInfo(context.Background(), ip, "json")
}
```

### 🧪 在 HTTP 服务里复用

最常见的场景：把 `Client` 做成包级变量或注入到依赖里，handler 里并发调用：

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "os"

    "github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
)

// 进程启动时建一次，后续所有请求复用
var client = ipapi.NewClient(
    ipapi.WithAPIKey(os.Getenv("IPAPI_KEY")),
)

func ipHandler(w http.ResponseWriter, r *http.Request) {
    // 多个请求并发进来，安全复用同一个 client
    info, err := client.GetClientIPInfo(r.Context(), "json")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    fmt.Fprintf(w, "你的 IP %s 来自 %s\n", info.IP, info.CountryName)
}

func main() {
    http.HandleFunc("/ip", ipHandler)
    http.ListenAndServe(":8080", nil)
}
```

### 🧵 批量并发查询

复用同一个 `Client` 配合 goroutine 做批量查询，再用 `sync.WaitGroup` 收敛：

```go
func lookupMany(ips []string) {
    var wg sync.WaitGroup
    client := ipapi.NewClient() // 建一次

    for _, ip := range ips {
        wg.Add(1)
        go func(ip string) {
            defer wg.Done()
            // 并发复用 client，安全
            info, err := client.GetIPInfo(context.Background(), ip, "json")
            if err != nil {
                fmt.Printf("%s 查询失败: %v\n", ip, err)
                return
            }
            fmt.Printf("%s → %s\n", ip, info.CountryName)
        }(ip)
    }
    wg.Wait()
}
```

::: tip 💡 高并发记得配限流
并发复用 `Client` 时，配合 `RateLimiter` 通道控制 QPS，避免把免费额度或服务端打爆。
:::

### 🔄 什么时候该换一个 Client

复用是默认选择，下面这些情况才需要新建或替换：

| 场景 | 做法 |
|------|------|
| 切换 API Key / 认证方式 | 新建带不同 `WithAPIKey` / `WithAPIKeyQuery` 的实例 |
| 不同超时档位 | 新建带 `WithCustomHTTPClient` 的实例，或共享 `Transport` 只换 `Timeout` |
| 不同基地址（自建镜像） | 新建后改 `BaseURL` |
| 测试隔离 | 每个用例独立 `Client`，避免相互污染 |

::: warning ⚠️ 运行期不要改字段
`APIKey`、`BaseURL`、`UserAgent` 这类字段设计为**创建后只读**。运行期并发改写会造成数据竞争，要换配置就**新建一个 `Client`**，而不是热改现有实例。
:::

### 📦 多个 Client 共享 Transport

如果确实需要多个 `Client`（比如不同超时），可以共享同一个 `Transport` 来复用底层连接池：

::: details 🔬 为什么共享 Transport 也能复用连接池？
Go 标准库的 `http.Transport` 持有连接池（`idleConn` map），连接是按 `host` 维度缓存的，与上层的 `*http.Client` 无关。多个 `*http.Client` 只要指向同一个 `Transport`，就会共用同一池空闲连接。

本 SDK 的 `applyAuth` 只改本次 `*http.Request` 的 header/query，不触碰 `Transport`，所以不同 API Key 的 Client 共享同一个 `Transport` 也不会串号——每个请求的认证信息随 request 独立发送，连接复用仅作用于传输层。

注意：`Transport` 的 `MaxIdleConnsPerHost` 默认只有 2，高并发场景务必显式调大，否则连接池会频繁「建-拆」而不是复用。
:::

```go
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 20,
    IdleConnTimeout:     90 * time.Second,
}

// 两个 Client，不同超时，但共享连接池
fast := ipapi.NewClient(ipapi.WithCustomHTTPClient(
    &http.Client{Transport: transport, Timeout: 5 * time.Second},
))
slow := ipapi.NewClient(ipapi.WithCustomHTTPClient(
    &http.Client{Transport: transport, Timeout: 30 * time.Second},
))
```

## 相关

- 🧭 [Client 客户端](../guide/client-concept) — `Client` 结构、线程安全说明与默认值
- 🛠 [自定义 HTTP 客户端](../guide/custom-http) — 超时、代理、连接池调优与 `Transport` 共享
- 🔄 [重试与限流](../guide/retry-concept) — `RateLimiter` 通道配合复用 Client 控制并发
- 📖 [`Client` API 参考](../api/client) — `Client` 结构体字段定义
- 🏗 [`NewClient`](../api/new-client) — 创建客户端的入口与默认配置
- 🔧 [`WithCustomHTTPClient`](../api/with-custom-http-client) — 替换底层 `*http.Client`
- 📚 [`GetIPInfo` 等方法](../api/get-ip-info) — 六个查询方法的线程安全调用
