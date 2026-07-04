# GetClientIPInfoRaw 📡🧱

> 查询客户端出口 IP，返回原始响应字节。

::: tip 🎯 跨域利器
`GetClientIPInfoRaw` 最经典的用法是后端把 JSONP 原始字节直接返回给浏览器，实现跨域"查我自己 IP"。也可拿客户端 IP 的 XML/CSV/YAML 透传下游。
:::

## 签名

```go
func (c *Client) GetClientIPInfoRaw(ctx context.Context, format string) ([]byte, error)
```

## 端点

```
GET https://ipapi.co/{format}/
```

与 [`GetClientIPInfo`](./get-client-ip-info) 同端点，返回**原始字节**。

## 客户端 IP 查询对照

| 方法 | 返回 | 解码 | 格式 | 典型场景 |
|------|------|------|------|----------|
| [`GetClientIPInfo`](./get-client-ip-info) | `*IPInfo` | `json.Decode` | 仅 `json` | 业务消费 |
| `GetClientIPInfoRaw` | `[]byte` | `io.ReadAll` | 全 5 种 | 透传/JSONP/非 JSON |

## 参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 超时/取消 |
| `format` | `string` | `json`/`jsonp`/`xml`/`csv`/`yaml` |

## 返回

- `[]byte`
- `error`

## 示例

```go
csvData, _ := client.GetClientIPInfoRaw(ctx, string(ipapi.FormatCSV))
fmt.Println(string(csvData))
```

JSONP：

```go
client := ipapi.NewClient(ipapi.WithCallback("cb"))
data, _ := client.GetClientIPInfoRaw(ctx, string(ipapi.FormatJSONP))
// cb({"ip":"...","city":"..."})
```

## 内部流程

::: tip 🎨 一图抵千言
路径无 IP，省去 `ValidateIP`；与 `GetClientIPInfo` 的分叉点在最后——这里用 `io.ReadAll` 读原始字节，**不** `json.Decode`。
:::

```mermaid
flowchart TD
    Start([🎯 调用 GetClientIPInfoRaw format]) --> VFmt{ValidateFormat}
    VFmt -->|非法| E2[❌ ErrInvalidFormat]
    VFmt -->|合法| Build[🔗 newGetRequest URL: /format/]
    Build --> Auth[🔐 applyAuth]
    Auth --> UA[📝 setHeaders]
    UA --> Do[🔄 doRequest 限流+重试]
    Do --> Read[📥 io.ReadAll resp.Body]
    Read --> Out([✅ 返回 []byte])
    E2 --> HE[⚠️ handleError]
    Do -.错误.-> HE
    HE --> ErrOut([❌ 返回 error])

    classDef raw fill:#fef3c7,stroke:#d97706
    class Read,Out raw
```

```mermaid
flowchart LR
    A[GetClientIPInfoRaw] -->|同端点 同链路| B{返回形态分叉}
    B -->|io.ReadAll| C[✅ []byte 原始]
    B -.对比.-> D[GetClientIPInfo<br/>json.Decode → *IPInfo]
    style C fill:#fef3c7,stroke:#d97706
    style D fill:#dcfce7,stroke:#16a34a
```

## 何时用

- 需要客户端 IP 的 XML/CSV/YAML/JSONP 格式
- 后端把 JSONP 直接返回给浏览器跨域调用

::: warning ⚠️ 安全提示
JSONP 跨域 handler 若把 `callback` 参数直接拼进响应而不校验，存在 XSS 风险。建议对 `callback` 做白名单或字符校验（仅允许 `[A-Za-z0-9_.]+`）后再透传。
:::

```go
func handler(w http.ResponseWriter, r *http.Request) {
	cb := r.URL.Query().Get("callback")
	c := ipapi.NewClient(ipapi.WithCallback(cb))
	data, _ := c.GetClientIPInfoRaw(r.Context(), string(ipapi.FormatJSONP))
	w.Header().Set("Content-Type", "application/javascript")
	w.Write(data)
}
```

## 下一步

- 📖 看 [`GetClientIPInfo`](./get-client-ip-info)
- 📞 学 [JSONP 回调](/guide/jsonp)
- 🧪 看 [JSONP 示例](/examples/jsonp)
