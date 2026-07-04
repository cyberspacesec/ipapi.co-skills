# 🌐 IPv4 查询

> 查询 IPv4 地址的地理位置，最常见用法。

## 什么是 IPv4

IPv4 是 32 位地址，形如 `8.8.8.8`，是互联网最普及的地址格式。本库完全支持。

## 基本查询

```go
info, err := client.GetIPInfo(ctx, "8.8.8.8", "json")
if err != nil {
	log.Fatal(err)
}
fmt.Printf("%s 位于 %s, %s\n", info.IP, info.City, info.CountryName)
```

::: tip 🎨 一图抵千言
下方流程图展示了 `GetIPInfo` 从输入到结果的完整调用链路，含校验、鉴权、请求与错误处理。
:::

```mermaid
flowchart TD
    A[🌍 输入 IPv4 地址] --> B[🔍 ValidateIP 合法性校验]
    B -->|非法| C[🚫 返回 ErrInvalidIP]
    B -->|合法| D[📐 ValidateFormat 格式校验]
    D -->|非法| E[🚫 返回 ErrInvalidFormat]
    D -->|合法| F[🏗️ newGetRequest 构造请求]
    F --> G[🔑 applyAuth 注入鉴权]
    G --> H[📨 setHeaders 设置请求头]
    H --> I[🌐 doRequest 发送请求]
    I --> J{响应状态}
    J -->|5xx 或网络错误| K[🔁 限流后重试]
    K --> I
    J -->|4xx 或 APIError| L[⚠️ handleError 转换错误]
    J -->|2xx 成功| M[📦 解码 JSON 为 IPInfo]
    M --> N[🕒 标记 RetrievedAt 时间]
    N --> O[✅ 返回 IPInfo 结果]
```

## 校验

[`ValidateIP`](/api/validate-ip) 用 `net.ParseIP` 判断合法性：

```go
if err := ipapi.ValidateIP("8.8.8.8"); err != nil {
	// 合法
}
if err := ipapi.ValidateIP("999.999.999.999"); err != nil {
	// err == ErrInvalidIP
}
```

`GetIPInfo` 内部已调用，非法 IP 直接返回 [`ErrInvalidIP`](/api/errors)，不发请求。

## 单字段

```go
city, _ := client.GetField(ctx, "8.8.8.8", "city")
asn, _ := client.GetField(ctx, "8.8.8.8", "asn")
```

## 常见 IPv4 示例

| IP | 说明 |
|----|------|
| `8.8.8.8` | Google Public DNS |
| `1.1.1.1` | Cloudflare DNS |
| `208.67.222.222` | OpenDNS |

## 错误场景

- 🚫 `999.999.999.999` → `ErrInvalidIP`
- 🚫 `10.0.0.1` / `192.168.x.x` → `ErrReservedIP`（保留/私有地址）
- 🚫 `0.0.0.0` → `ErrReservedIP`

私有地址无地理意义，详见 [保留 IP](./reserved-ip)。

下面的时序图从客户端与 ipapi.co 服务端交互的视角，对比了合法 IPv4、非法 IPv4、保留 IP 三种典型场景下各自的调用时序与返回点：

::: tip 🎨 一图抵千言
:::

```mermaid
sequenceDiagram
    participant C as 🧑‍💻 调用方
    participant S as 📦 ipapi SDK
    participant A as 🌐 ipapi.co API
    Note over C,A: 场景一：合法公网 IPv4 (8.8.8.8)
    C->>S: GetIPInfo(ctx, "8.8.8.8", "json")
    S->>S: ValidateIP / ValidateFormat 校验
    S->>A: GET /8.8.8.8 (注入鉴权)
    A-->>S: 200 OK + JSON
    S-->>C: IPInfo{City,CountryName,ASN...}
    Note over C,A: 场景二：非法 IPv4 (999.999.999.999)
    C->>S: GetIPInfo(ctx, "999.999.999.999", "json")
    S->>S: ValidateIP 失败
    S-->>C: 🚫 ErrInvalidIP（未发请求）
    Note over C,A: 场景三：保留/私有 IP (10.0.0.1)
    C->>S: GetIPInfo(ctx, "10.0.0.1", "json")
    S->>S: ValidateIP 通过
    S->>A: GET /10.0.0.1
    A-->>S: 200 OK（保留地址标记）
    S-->>C: 🚫 ErrReservedIP
```

## 下一步

- 🌍 学 [IPv6 查询](./ipv6)
- 🧪 看 [查询指定 IP 示例](/examples/lookup-specific-ip)
