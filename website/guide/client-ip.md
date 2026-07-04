# 📡 客户端 IP 检测

> 不知道要查哪个 IP？查询「发起请求的这台机器」的公网 IP。

## 客户端 IP 端点

ipapi.co 提供 `GET /{format}/` 端点（路径里没有 IP），自动返回**调用方出口 IP** 的信息。本库封装为 [`GetClientIPInfo`](/api/get-client-ip-info)：

```go
info, err := client.GetClientIPInfo(ctx, "json")
fmt.Printf("我的公网 IP: %s\n位置: %s\n", info.IP, info.City)
```

::: tip 🎨 一图抵千言
下面这张时序图展示了 `GetClientIPInfo` 内部的完整调用链：从格式校验、构造请求、鉴权，到 ipapi.co 识别调用方出口 IP 并返回结果。
:::

```mermaid
sequenceDiagram
    autonumber
    participant SDK as 🧰 SDK 客户端
    participant Auth as 🔑 applyAuth 鉴权
    participant HTTP as 🌐 doRequest 网络层
    participant API as 🛰️ ipapi.co 服务

    SDK->>SDK: 📋 ValidateFormat 校验格式
    SDK->>SDK: 🏗️ newGetRequest 构造 GET /{format}/
    Note over SDK: 路径不带 IP 由服务端识别出口地址
    SDK->>Auth: 📨 注入凭证
    Auth-->>SDK: ✅ Bearer 头或 ?key= 就绪
    SDK->>SDK: 🏷️ setHeaders 设置 User-Agent
    SDK->>HTTP: 📤 发起请求

    HTTP->>HTTP: ⏳ 限流 tick（若开启）
    HTTP->>API: 🔎 GET /json/
    API->>API: 📍 识别调用方出口 IP
    API-->>HTTP: 📦 返回该 IP 的地理信息

    alt 状态码 ≥ 400
        HTTP->>HTTP: ⚠️ 解析 APIError 或映射状态码
        HTTP-->>SDK: ❌ 返回错误（经 handleError）
    else 5xx 或网络错误
        HTTP->>HTTP: 🔁 按 Retries 重试
    else 成功 2xx
        HTTP-->>SDK: 📄 响应体
        SDK->>SDK: 🧩 json.Decode 反序列化 IPInfo
        SDK->>SDK: 🕐 设置 RetrievedAt 时间戳
        SDK-->>SDK: 🎉 返回调用方出口 IP 信息
    end
```

## 系列方法

| 方法 | 返回 | 用途 |
|------|------|------|
| `GetClientIPInfo` | `*IPInfo` | 拿自己完整信息 |
| `GetClientIPInfoRaw` | `[]byte` | 拿自己原始响应（XML/CSV/YAML） |
| `GetClientField` | `string` | 拿自己单个字段 |

```go
// 只想知道自己公网 IP
myIP, _ := client.GetClientField(ctx, "ip")
fmt.Println("我的 IP:", myIP)
```

::: tip 🎨 一图抵千言
三种系列方法对应不同的「取结果」状态：完整结构化对象、原始字节、单个字段。下面用状态图展示一次客户端 IP 检测从发起到拿到所需形态的流转。
:::

```mermaid
stateDiagram-v2
    [*] --> 已发起 : client.GetClient* 调用
    已发起 --> 校验中 : ValidateFormat
    校验中 --> 已发起 : 格式非法 → 返回错误
    校验中 --> 已鉴权 : applyAuth 注入凭证
    已鉴权 --> 已请求 : doRequest GET /{format}/
    已请求 --> 重试中 : 5xx 或网络错误
    重试中 --> 已请求 : Retries 内重试
    重试中 --> 失败 : 重试耗尽
    已请求 --> 已响应 : 2xx 返回响应体
    已响应 --> 完整对象 : GetClientIPInfo → json.Decode
    已响应 --> 原始字节 : GetClientIPInfoRaw → []byte
    已响应 --> 单字段 : GetClientField → 解析某字段
    完整对象 --> [*] : *IPInfo
    原始字节 --> [*] : XML/CSV/YAML
    单字段 --> [*] : string
    失败 --> [*] : 返回 error
```

## 典型场景

### 🌐 「我的 IP 是什么」

```go
info, _ := client.GetClientIPInfo(ctx, "json")
fmt.Println(info.IP)
```

### 📍 服务部署位置自检

服务器调用 `GetClientIPInfo`，看返回的城市是否是预期机房位置。

### 🔄 出口 IP 探测

多出口/CDN 环境下，探测实际出口 IP 用于调试。

## ⚠ 注意事项

::: warning 🚧 这是「调用方的出口 IP」
`GetClientIPInfo` 返回的是 **SDK 所在机器** 看到的出口 IP，**不是**终端用户的 IP。

如果你的服务在网关/代理后，要查终端用户 IP，应从 `X-Forwarded-For` 等头取出真实 IP，再用 `GetIPInfo` 查询：
:::

```go
// 从请求头取真实用户 IP
userIP := r.Header.Get("X-Forwarded-For")
info, _ := client.GetIPInfo(ctx, userIP, "json")
```

## 下一步

- 📖 看 [`GetClientIPInfo` API](/api/get-client-ip-info)
- 🧪 看 [获取客户端 IP 示例](/examples/lookup-client-ip)
- 🌐 学 [IPv4](./ipv4) / [IPv6](./ipv6)
