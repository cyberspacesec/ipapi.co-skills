# 🔒 认证机制

> ipapi.co 的 API Key 认证，本库支持两种方式。

## 为什么需要认证

ipapi.co 提供免费额度，但**未认证请求配额低、且共享 IP 池**。生产环境强烈建议申请 API Key：

- 📈 更高配额
- 🧮 独立计费
- 🛡 稳定可用性

在 [ipapi.co](https://ipapi.co) 注册后在控制台获取 Key。

## 两种认证方式

### 方式 1：Bearer Header（默认）

Key 放在 `Authorization` 头：

```go
client := ipapi.NewClient(
	ipapi.WithAPIKey("your_api_key"),
)
// 内部发送: Authorization: Bearer your_api_key
```

✅ 推荐。Key 不出现在 URL/日志中，更安全。

### 方式 2：Query Parameter

Key 放在 `?key=` 查询参数：

```go
client := ipapi.NewClient(
	ipapi.WithAPIKey("your_api_key"),
	ipapi.WithAPIKeyQuery(),  // 切换为 query 模式
)
// 内部发送: https://ipapi.co/8.8.8.8/json/?key=your_api_key
```

⚠️ Key 会出现在 URL、访问日志、Referer 中。仅在 Header 不可用时用（如某些受限代理、CDN）。

## 内部实现

[`applyAuth`](/api/with-api-key) 根据字段 `APIKeyMode` 分支：

```go
switch c.APIKeyMode {
case APIKeyQuery:
	q.Set("key", c.APIKey)        // 查询参数
default:
	req.Header.Set("Authorization", "Bearer "+c.APIKey) // 头
}
```

`APIKeyMode` 是 `int` 枚举：

```go
const (
	APIKeyHeader APIKeyMode = iota // 0，默认
	APIKeyQuery                    // 1
)
```

::: tip 🎨 一图抵千言
下面的流程图展示了 [`applyAuth`](/api/with-api-key) 在每次请求时如何根据 `APIKey` 与 `APIKeyMode` 选择认证方式。
:::

```mermaid
flowchart TD
    A["🚀 发起请求 GetIPInfo/GetField"] --> B["⚙️ applyAuth 处理请求"]
    B --> C{"🔑 APIKey 是否为空?"}
    C -->|"是, 未配置 WithAPIKey"| D["🙈 匿名放行, 享受免费额度"]
    C -->|"否, 已设置 Key"| E{"🔀 APIKeyMode 取值?"}
    E -->|"APIKeyQuery 模式1"| F["📦 拼到 URL 查询参数 ?key="]
    E -->|"APIKeyHeader 默认0"| G["🛡 写入 Authorization Bearer 头"]
    F --> H["⚠️ Key 暴露在 URL 日志 Referer"]
    G --> I["✅ Key 不出现在 URL, 更安全"]
    H --> J["🌐 发出请求 doRequest"]
    I --> J
    D --> J
    J --> K{"📥 响应状态码?"}
    K -->|"403"| L["🚫 ErrInvalidKey, 检查 Key"]
    K -->|"2xx"| M["🎉 返回 IP 信息结果"]
```

## 何时用哪种

| 场景 | 推荐 |
|------|------|
| 一般后端服务 | Header ✅ |
| 需要在 CDN/网关层按 URL 鉴权 | Query |
| 前端 JSONP（无法设 Header） | Query |
| 安全要求高 | Header（避免日志泄露） |

::: tip 🎨 选型决策树
上面的表格给出了场景与推荐的对应关系。下面这张决策图从**使用者创建 [`Client`](./getting-started) 时的选型视角**出发，引导你一步步判断该用 Header 还是 Query（与上一节运行时 [`applyAuth`](/api/with-api-key) 分支互补）。
:::

```mermaid
flowchart LR
    Start(["🧭 创建 Client 选认证模式"]) --> Q1{"🔑 需要更高配额 / 独立计费?"}
    Q1 -->|"否"| Anon["🙈 省略 WithAPIKey<br/>匿名走免费额度"]
    Q1 -->|"是"| Q2{"🌐 运行环境能否自定义请求头?"}
    Q2 -->|"能, 后端 / 服务端"| Hdr["✅ Header 模式<br/>仅 WithAPIKey(key)"]
    Q2 -->|"否, JSONP / 受限 CDN / 受限代理"| Q3{"🛡 能否接受 Key 出现在 URL 与日志?"}
    Q3 -->|"能, 网关需按 URL 鉴权"| Qry["⚠️ Query 模式<br/>WithAPIKey + WithAPIKeyQuery"]
    Q3 -->|"否, 安全要求高"| Q4{"🔧 能换环境 / 换代理以支持 Header?"}
    Q4 -->|"能"| Hdr
    Q4 -->|"否"| Qry
    Anon --> Done(["📦 Client 就绪"])
    Hdr --> Done
    Qry --> Done

    classDef safe fill:#d4edda,stroke:#28a745,color:#155724
    classDef warn fill:#fff3cd,stroke:#ffc107,color:#856404
    classDef neutral fill:#e2e3e5,stroke:#6c757d,color:#383d41
    class Hdr safe
    class Qry warn
    class Anon neutral
```

## 安全最佳实践

::: danger 🔐 切勿硬编码 Key
**不要**把 Key 写进源码或提交进 Git。用环境变量：

```go
client := ipapi.NewClient(
	ipapi.WithAPIKey(os.Getenv("IPAPI_KEY")),
)
```

`.env` / `.gitignore` 中排除敏感文件。
:::

- 🗝 Key 当作密码保管，定期轮换
- 📋 服务端 403 时检查 Key（[`ErrInvalidKey`](/api/errors)）
- 🚫 不要把带 Key 的 URL 打到日志

::: warning 🎨 安全权衡一览
两种认证模式在 Key 的**暴露面**上有本质差异。下图对比 Key 从发出到落地各环节的可观测性，帮助理解为何默认推荐 Header。
:::

```mermaid
flowchart TD
    subgraph Header["🛡 Header 模式 (默认, APIKeyHeader)"]
        direction LR
        H1["📤 Authorization: Bearer xxx<br/>仅存在于请求头"] --> H2{"🔍 谁能看到 Key?"}
        H2 -->|"客户端进程内存"| H3["✅ 受控"]
        H2 -->|"HTTPS 加密通道"| H4["✅ 中间人不可见"]
        H2 -->|"ipapi 服务端"| H5["✅ 预期接收方"]
        H2 -->|"访问日志 / Referer"| H6["✅ 不出现"]
    end

    subgraph Query["⚠️ Query 模式 (APIKeyQuery)"]
        direction LR
        Q1["📤 ?key=xxx 拼入 URL"] --> Q2{"🔍 谁能看到 Key?"}
        Q2 -->|"客户端进程内存"| Q3["✅ 受控"]
        Q2 -->|"HTTPS 加密通道"| Q4["✅ 中间人不可见<br/>但 URL 仍可能被记录"]
        Q2 -->|"网关 / 代理 / CDN 访问日志"| Q5["❌ 明文记录"]
        Q2 -->|"浏览器 Referer / 历史记录"| Q6["❌ 明文泄露"]
        Q2 -->|"错误页 / 崩溃堆栈含 URL"| Q7["❌ 明文泄露"]
    end

    Header -.->|"同等配额 / 同等鉴权效果<br/>安全面更小, 故默认"| Query
    Query -.->|"仅当 Header 不可用时<br/>作为退路使用"| Header

    classDef ok fill:#d4edda,stroke:#28a745,color:#155724
    classDef bad fill:#f8d7da,stroke:#dc3545,color:#721c24
    class H3,H4,H5,H6 ok
    class Q5,Q6,Q7 bad
```

## 不认证也能用

省略 `WithAPIKey` 即可，享受免费额度：

```go
client := ipapi.NewClient() // 匿名
```

## 下一步

- 📖 看 [`WithAPIKey`](/api/with-api-key)、[`WithAPIKeyQuery`](/api/with-api-key-query)
- 🧪 看 [带 API Key 示例](/examples/with-api-key)
- 🛡 学 [错误处理](./error-concept)
