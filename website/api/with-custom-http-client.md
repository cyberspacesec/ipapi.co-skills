# WithCustomHTTPClient

> 替换底层 `*http.Client`，定制超时、代理、连接池等。

## 签名

```go
func WithCustomHTTPClient(client *http.Client) ClientOption
```

## 作用

设置 `Client.HTTPClient` 字段，覆盖默认的 10s 超时 + 3 跳转限制客户端。

::: tip 🎨 一图抵千言
`WithCustomHTTPClient` 如何替换默认客户端，影响后续所有请求的传输层。
:::

```mermaid
flowchart LR
    subgraph Default["默认 HTTPClient 🟦"]
        D1["Timeout 10s"]
        D2["CheckRedirect 3 跳"]
    end
    subgraph Custom["自定义 HTTPClient 🟩"]
        C1["Timeout 调节"]
        C2["Transport 连接池"]
        C3["Proxy 代理"]
    end
    NewClient["NewClient"] -->|"未调用本选项"| Default
    NewClient -->|"调用 WithCustomHTTPClient"| Custom
    Default --> doReq["doRequest"]
    Custom --> doReq
    doReq --> Net["Go net/http 网络 🌐"]
```

下面这张类图展示**注入视角**：`WithCustomHTTPClient` 作为 `ClientOption` 如何改写 `*Client` 的 `HTTPClient` 字段，进而决定 `Timeout` 与 `Transport` 的归属。

```mermaid
classDiagram
    class ClientOption {
        <<func>>
        +Apply(c *Client)
    }
    class WithCustomHTTPClient {
        +client *http.Client
    }
    class Client {
        +BaseURL string
        +APIKey string
        +HTTPClient *http.Client
        +Retries int
        +Timeout time.Duration
        +doRequest(ctx, req) []byte
    }
    class http_Client {
        +Timeout time.Duration
        +Transport http.RoundTripper
        +CheckRedirect func()
    }
    class http_Transport {
        +MaxIdleConns int
        +Proxy func(*http.Request) (*url.URL, error)
    }
    WithCustomHTTPClient --|> ClientOption : 实现
    WithCustomHTTPClient ..> Client : "Apply 写入 HTTPClient"
    Client o-- http_Client : 持有
    http_Client o-- http_Transport : 持有
    note for Client "默认 HTTPClient 由 SDK 构造\n调用本选项后整体替换"
    note for http_Client "Timeout / Transport\n均由调用方负责"
```

## 默认 vs 自定义对照

| 维度 | 默认 Client | 自定义 Client |
|------|------------|--------------|
| 超时 | 固定 10s | 任意 |
| 跳转限制 | 3 次 | 需自行设 `CheckRedirect` |
| 连接池 | 默认 | 可调 `MaxIdleConns` |
| 代理 | 无 | 可配 `Transport.Proxy` |

## 示例

### 调超时

```go
client := ipapi.NewClient(
	ipapi.WithCustomHTTPClient(&http.Client{
		Timeout: 30 * time.Second,
	}),
)
```

### 连接池

```go
client := ipapi.NewClient(
	ipapi.WithCustomHTTPClient(&http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},
	}),
)
```

### 代理

```go
proxyURL, _ := url.Parse("http://proxy:8080")
client := ipapi.NewClient(
	ipapi.WithCustomHTTPClient(&http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
	}),
)
```

下面这张决策树展示**选择视角**：何时需要本选项、何时保持默认即可，避免无谓的覆盖。

```mermaid
flowchart TD
    Start["要改传输层吗？"] --> Q1{"仅改超时？"}
    Q1 -->|"是，且 ≤ 默认策略"| Keep["保持默认 Client\n（10s + 3 跳转）"]
    Q1 -->|"否"| Q2{"需要代理 / 连接池 / 自定义跳转？"}
    Q2 -->|"否，只改超时"| Simple["WithCustomHTTPClient\n仅设 Timeout"]
    Q2 -->|"是"| Q3{"多 Client 共享 Transport？"}
    Q3 -->|"否"| Solo["每个 Client 独立 Transport\n避免相互污染"]
    Q3 -->|"是"| Shared["共享 Transport\n注意任一修改影响全局"]
    Keep --> Done["doRequest"]
    Simple --> Done
    Solo --> Done
    Shared --> Done

    classDef decide fill:#fff3cd,stroke:#b8860b
    classDef ok fill:#d4edda,stroke:#28a745
    class Q1,Q2,Q3 decide
    class Keep,Simple,Solo,Shared,Done ok
```

::: details 🔍 决策依据
默认客户端已内置 10s 超时与 3 次跳转限制，覆盖绝大多数查询场景。只有当需要更长超时、代理、连接池调优或自定义 `CheckRedirect` 时才动用本选项。可参考 [`Client`](./) 字段定义与 [`WithAPIKey`](./with-api-key) 等其他选项的组合方式。
:::

## 内部

```go
func WithCustomHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = client
	}
}
```

## 默认客户端（被替换的）

```go
&http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects { // 3
			return fmt.Errorf("stopped after %d redirects", maxRedirects)
		}
		return nil
	},
}
```

## 注意

- 替换后**丢失**默认的跳转限制，需自行设 `CheckRedirect`
- 多个 `Client` 共享同一 `Transport` 可复用连接池

::: warning ⚠️ 替换即重置
`WithCustomHTTPClient` 是**整体替换**，不是合并。你传入的 `*http.Client` 会完全覆盖默认实例，包括超时、跳转策略、Transport。若只想要更长超时而保留跳转限制，需在自定义 Client 里一并设 `CheckRedirect`。
:::

::: danger 🚨 共享 Transport 的坑
多个 `Client` 共享同一个 `*http.Transport` 时连接池会被复用，但**任一 Client 的 Transport 修改都会影响所有**。生产环境建议为不同用途（不同超时/代理）的 Client 配置独立 Transport，避免相互污染。
:::

::: details 📖 常用 Transport 参数速查
```go
&http.Transport{
    MaxIdleConns:        100,   // 全局最大空闲连接
    MaxIdleConnsPerHost: 20,    // 单 host 最大空闲连接
    IdleConnTimeout:     90 * time.Second, // 空闲超时
    TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时
    ExpectContinueTimeout: 1 * time.Second,
    DisableCompression:  false,
}
```
高并发查询建议 `MaxIdleConnsPerHost` 至少等于并发数，避免频繁建连。
:::

## 下一步

- 🛠 学 [自定义 HTTP 客户端](/guide/custom-http)
- ⏱ 学 [Context 超时](/guide/context)
- 🧪 看 [高级示例](/examples/advanced-usage)
