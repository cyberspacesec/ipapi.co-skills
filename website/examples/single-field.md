# 💡 单字段查询

> 只需要某个字段，用单字段端点省流量。

## 场景

只需国家代码、ASN 等单个字段，无需拉全量 JSON。

::: tip 🎨 一图抵千言
单字段查询的调用结构：[`GetField`](/api/get-field) 按 `field` 名取单值，返回值经 `TrimSpace` 去除首尾空白后交给调用方；客户端会在发请求前用 [`IsValidField`](/api/validate-format) 校验字段名，非法字段直接返回 [`ErrInvalidField`](/api/errors)。

```mermaid
flowchart TD
    A["main 调用 client.GetField(ctx, ip, field)"] --> B{"IsValidField(field)?"}
    B -->|合法| C["newGetRequest 构造单字段 URL<br/>/{ip}/{field}/"]
    B -->|非法| X["返回 ErrInvalidField"]
    C --> D["applyAuth 注入鉴权<br/>doRequest 发送 HTTP"]
    D --> E{"状态码"}
    E -->|2xx| F["读取响应体原始字节"]
    E -->|4xx/5xx| G["mapStatusCodeToError"]
    F --> H["strings.TrimSpace(string(body))"]
    H --> I["返回单值字符串"]
    G --> Y["返回对应哨兵错误"]
```
:::

## 代码

```go
func main() {
	client := ipapi.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ip := "8.8.8.8"

	country, _ := client.GetField(ctx, ip, "country_code")
	city, _ := client.GetField(ctx, ip, "city")
	asn, _ := client.GetField(ctx, ip, "asn")
	org, _ := client.GetField(ctx, ip, "org")

	fmt.Printf("%s: %s, %s, %s (%s)\n", ip, country, city, asn, org)
}
```

## 输出

```
8.8.8.8: US, Mountain View, AS15169 (Google LLC)
```

## 客户端 IP 单字段

```go
myCity, _ := client.GetClientField(ctx, "city")
myTZ, _ := client.GetClientField(ctx, "timezone")
```

## 合法字段

28 个合法字段，详见 [字段总览](/api/fields)。非法字段返回 `ErrInvalidField`（客户端校验）。

## ⚠ 配额权衡

每次 `GetField` 都消耗 **1 次配额**。需要 ≥3 个字段时，用 [`GetIPInfo`](/api/get-ip-info) 一次拿全更划算：

```go
// 不推荐：3 次单字段 = 3 次配额
a, _ := client.GetField(ctx, ip, "city")
b, _ := client.GetField(ctx, ip, "country")
c, _ := client.GetField(ctx, ip, "asn")

// 推荐：1 次完整 = 1 次配额
info, _ := client.GetIPInfo(ctx, ip, "json")
// info.City / info.Country / info.ASN
```

详见 [字段概念](/guide/field-concept)。

::: tip 🎨 一图抵千言
下面的时序图对比两种调用的网络往返与配额消耗：左侧逐个单字段查询触发多次 HTTP 往返、累计多次配额；右侧一次完整查询只走一次往返、只计一次配额。

```mermaid
sequenceDiagram
    autonumber
    participant Main as ["main 调用方"]
    participant SDK as ["ipapi.Client"]
    participant API as ["ipapi.co 端点"]

    Note over Main,API: ❌ 逐个单字段（3 次往返 = 3 次配额）
    Main->>SDK: GetField(ctx, ip, "city")
    SDK->>API: GET /{ip}/city/
    API-->>SDK: 200 "Mountain View"
    SDK-->>Main: "Mountain View"（计 1 次配额）
    Main->>SDK: GetField(ctx, ip, "country")
    SDK->>API: GET /{ip}/country/
    API-->>SDK: 200 "US"
    SDK-->>Main: "US"（再计 1 次配额）
    Main->>SDK: GetField(ctx, ip, "asn")
    SDK->>API: GET /{ip}/asn/
    API-->>SDK: 200 "AS15169"
    SDK-->>Main: "AS15169"（再计 1 次配额）

    Note over Main,API: ✅ 一次完整查询（1 次往返 = 1 次配额）
    Main->>SDK: GetIPInfo(ctx, ip, "json")
    SDK->>API: GET /{ip}/json/
    API-->>SDK: 200 完整 JSON
    SDK-->>Main: *ipapi.IPInfo（City/Country/ASN 同体）
    Note over Main,SDK: 共 1 次配额，字段内联可达
```
:::

::: details 运行预期输出与常见问题
**预期输出**

```txt
8.8.8.8: US, Mountain View, AS15169 (Google LLC)
```

**常见问题**

- **返回值带换行/空格？** SDK 已对原始响应体执行 `strings.TrimSpace`，正常无需再处理；若你拼接后仍有空白，检查是否自己又包了一层字符串。
- **拿到空字符串？** 多半是该 IP 此字段确实为空（如未分配 ASN 的内网段），而非失败。需要区分"空值"与"错误"时，务必检查 `err` 而非判断空串。
- **字段名写错？** 单字段端点在客户端用 [`IsValidField`](/api/validate-format) 预校验，非法字段返回 [`ErrInvalidField`](/api/errors)，**不会**消耗请求。
- **频繁 429？** 每次 [`GetField`](/api/get-field) 都计 1 次配额；连续多字段请改用 [`GetIPInfo`](/api/get-ip-info) 一次拿全。
:::

## 下一步

- 📖 看 [`GetField`](/api/get-field) / [`GetClientField`](/api/get-client-field)
- 📋 看 [字段总览](/api/fields)
- 🧭 学 [字段查询概念](/guide/field-concept)
