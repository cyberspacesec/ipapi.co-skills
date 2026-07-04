# 🛠 自定义 HTTP 客户端

> 用 `WithCustomHTTPClient` 替换底层 `*http.Client`，实现超时、代理、连接池等定制。

::: tip 🎨 一图抵千言
下图展示自定义 HTTP 客户端在 SDK 请求链路中的位置：`WithCustomHTTPClient` 注入的 `*http.Client` 会被 `doRequest` 用于真正的网络发包，前置的鉴权与重试、后置的错误处理则由 SDK 内部统一完成。
:::

```mermaid
flowchart TD
    Start["🚀 NewClient 初始化"] --> Opt["⚙️ 应用 WithCustomHTTPClient 选项"]
    Opt --> Replace["🔄 替换 c.HTTPClient 为自定义实例"]
    Replace --> Call["📞 调用 GetIPInfo / GetField 等方法"]
    Call --> Validate["✅ 校验 IP 与 format / field"]
    Validate --> BuildReq["🧱 newGetRequest 构造 GET 请求"]
    BuildReq --> Auth["🔑 applyAuth 设置鉴权"]
    Auth --> Header["📝 setHeaders 设置 User-Agent"]
    Header --> DoReq["🌐 doRequest 发起请求"]
    DoReq --> RateLimit["⏳ 命中 RateLimiter 则等待令牌"]
    RateLimit --> Send["📡 c.HTTPClient.Do 真正发包"]
    Send --> CheckErr{"❓ 网络错误或 5xx?"}
    CheckErr -- 是 --> Retry{"🔁 已达 Retries 上限?"}
    Retry -- 否 --> Wait["💤 defaultRetryDelay 后重试"] --> Send
    Retry -- 是 --> HandleFail["⚠️ 返回请求失败错误"]
    CheckErr -- 否 --> CheckCode{"❓ 状态码 ≥ 400?"}
    CheckCode -- 是 --> ParseErr["🔎 解析 APIError / mapStatusCodeToError"]
    ParseErr --> HandleFail
    CheckCode -- 否 --> Decode["📦 解析 JSON / 读取原始字节"]
    Decode --> Done["✅ 返回结果"]
    HandleFail --> HandleErr["🧯 handleError 走自定义或默认错误处理"]
    HandleErr --> EndErr["❌ 返回错误"]
    Done --> EndOK["🎉 完成"]
```

## 默认 HTTP 客户端

[`NewClient`](/api/new-client) 默认创建：

```go
&http.Client{
	Timeout: 10 * time.Second, // 默认超时
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects { // 最多 3 次跳转
			return fmt.Errorf("stopped after %d redirects", maxRedirects)
		}
		return nil
	},
}
```

## 替换它

[`WithCustomHTTPClient`](/api/with-custom-http-client) 接受任意 `*http.Client`：

```go
client := ipapi.NewClient(
	ipapi.WithCustomHTTPClient(&http.Client{
		Timeout:   15 * time.Second,
		Transport: &http.Transport{MaxIdleConns: 10},
	}),
)
```

## 常见定制场景

> 下面的 classDiagram 从结构视角拆解 `*http.Client` 可定制的组件：`Timeout` 直接挂在 Client 上，而代理、连接池、TLS 等都通过 `Transport` 配置，多个 Client 还可共享同一个 Transport 复用连接。

```mermaid
classDiagram
    class HTTPClient {
        +Duration Timeout
        +Transport Transport
        +CheckRedirect CheckRedirect
    }
    class HTTPTransport {
        +Proxy Proxy
        +int MaxIdleConns
        +int MaxIdleConnsPerHost
        +Duration IdleConnTimeout
        +TLSClientConfig TLSClientConfig
    }
    class TLSConfig {
        +bool InsecureSkipVerify
    }
    class ProxyFunc {
        <<function>>
        +ProxyURL(url) Proxy
    }
    class RedirectPolicy {
        <<function>>
        maxRedirects = 3
    }
    HTTPClient --> HTTPTransport : 通过 Transport 配置网络层
    HTTPClient --> RedirectPolicy : 限制最多 3 次跳转
    HTTPTransport --> ProxyFunc : 配置代理出口
    HTTPTransport --> TLSConfig : 定制 TLS 握手
    note for HTTPClient "Timeout 10s 默认 / WithCustomHTTPClient 替换"
    note for HTTPTransport "连接池调优 + 代理 + TLS 的统一入口"
```

### ⏱ 调整超时

```go
&http.Client{Timeout: 30 * time.Second}
```

### 🌐 配置代理

```go
&http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyURL(must(url.Parse("http://proxy:8080"))),
	},
}
```

### 🔗 连接池调优

```go
&http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	},
}
```

### 🔒 自定义 TLS

```go
&http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false, // 生产别跳过校验
		},
	},
}
```

### 📦 复用 Transport

```go
transport := &http.Transport{MaxIdleConns: 100}
// 多个 Client 共享 transport，复用连接
c1 := ipapi.NewClient(ipapi.WithCustomHTTPClient(&http.Client{Transport: transport, Timeout: 10 * time.Second}))
c2 := ipapi.NewClient(ipapi.WithCustomHTTPClient(&http.Client{Transport: transport, Timeout: 30 * time.Second}))
```

## 完整示例

来自 [`examples/advanced_usage`](/examples/advanced-usage)：

```go
client := ipapi.NewClient(
	ipapi.WithAPIKey("your_api_key"),
	ipapi.WithCustomHTTPClient(&http.Client{
		Timeout:   15 * time.Second,
		Transport: &http.Transport{MaxIdleConns: 10},
	}),
	ipapi.WithErrorHandler(customErrorHandler),
)
```

## 下一步

- 📖 看 [`WithCustomHTTPClient`](/api/with-custom-http-client)
- 🧪 看 [高级示例](/examples/advanced-usage)
- ⏱ 学 [Context 超时](./context)
