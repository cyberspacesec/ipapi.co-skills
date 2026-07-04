# WithAPIKeyQuery

> 切换 API Key 认证为 `?key=` 查询参数方式。

## 签名

```go
func WithAPIKeyQuery() ClientOption
```

## 作用

设置 `Client.APIKeyMode = APIKeyQuery`。配合 [`WithAPIKey`](./with-api-key)，请求 URL 会带：

```
https://ipapi.co/8.8.8.8/json/?key=<key>
```

::: tip 🎨 一图抵千言
下图对比两种认证模式下 `applyAuth` 的分流路径。
:::

```mermaid
flowchart TD
    A["调用 doRequest"] --> B{"APIKeyMode"}
    B -->|"APIKeyHeader 默认"| C["写入 Authorization 头"]
    B -->|"APIKeyQuery 本选项"| D["URL 追加 ?key= 参数"]
    C --> E["发出请求 📤"]
    D --> E
    E --> F["响应处理 ✅"]
    C -.->|"JSONP 不可用 ❌"| G["script 无法设头"]
    D -.->|"JSONP 可用 ✅"| H["script 可带 URL"]
```

::: tip 🎨 一图抵千言
下图展示 **配置到出网的全链路视角**：从 [`WithAPIKeyQuery`](./with-api-key-query) 配置 `APIKeyMode`，到 [`newGetRequest`](./with-api-key-query) 拼 `?key=`，再到对比 Header 模式的最终 URL 形态。
:::

```mermaid
flowchart LR
    subgraph Q["Query 模式（本选项）"]
        Q1["WithAPIKeyQuery()"] --> Q2["APIKeyMode = APIKeyQuery"]
        Q2 --> Q3["newGetRequest 拼 URL"]
        Q3 --> Q4["baseURL + ?key=<key>"]
        Q4 --> Q5["GET https://ipapi.co/8.8.8.8/json/?key=***"]
    end
    subgraph H["Header 模式（默认）"]
        H1["未设置本选项"] --> H2["APIKeyMode = APIKeyHeader"]
        H2 --> H3["setHeaders 写头"]
        H3 --> H4["Authorization: Bearer ***"]
        H4 --> H5["GET https://ipapi.co/8.8.8.8/json/"]
    end
    Q5 --> R["统一 doRequest 出网 📤"]
    H5 --> R
```

## 示例

```go
client := ipapi.NewClient(
	ipapi.WithAPIKey("your_api_key"),
	ipapi.WithAPIKeyQuery(),
)
```

## 与默认 Header 模式对比

| 维度 | Header（默认） | Query |
|------|---------------|-------|
| 位置 | `Authorization` 头 | `?key=` URL 参数 |
| 日志泄露 | 不易 | 易出现在访问日志 |
| 前端可见 | 否 | 是 |
| JSONP 可用 | ❌ | ✅ |

## 何时用

- 📞 JSONP 场景（`<script>` 无法设头）
- 🚪 某些代理/CDN 只支持 URL 鉴权
- 🔧 调试时方便直接看 URL

::: tip 🎨 一图抵千言
下图展示 **选型决策树视角**：按运行环境与日志敏感度反推应选哪种模式。
:::

```mermaid
flowchart TD
    Start(["需要带 API Key 调用"]) --> Q1{"运行在浏览器/JSONP？"}
    Q1 -->|"是"| Q2["必须用 Query（script 无法设头）"]
    Q1 -->|"否"| Q2b{"代理/CDN 只认 URL 鉴权？"}
    Q2b -->|"是"| Q2
    Q2b -->|"否"| Q3{"日志/Referer 敏感？"}
    Q3 -->|"是，需隐蔽 Key"| H1["选 Header（默认）"]
    Q3 -->|"否，后端服务"| H1
    Q2 --> End(["WithAPIKeyQuery()"])
    H1 --> End2(["不设本选项"])
```

::: info 📊 两种模式速查
**Header**：默认、隐蔽、JSONP 不可用、适合后端服务。
**Query**：URL 暴露、JSONP 可用、适合前端/受限代理。
:::

::: warning ⚠️ 安全权衡
Query 方式 Key 会出现在 URL、日志、Referer。仅在必要时用，且用受限/临时 Key。
:::

::: details 🔍 Key 泄露路径分析
- **访问日志**：反向代理/Nginx 默认记录完整 URL
- **Referer 头**：用户点外链时 Key 会被带向第三方站点
- **浏览器历史**：本地留痕，共享设备可见
- **CDN 缓存**：部分 CDN 可能缓存带 Key 的 URL
建议：用 Query 时务必配合短期/受限 Key，并定期轮换。
:::

## 内部

```go
func WithAPIKeyQuery() ClientOption {
	return func(c *Client) {
		c.APIKeyMode = APIKeyQuery
	}
}
```

`applyAuth` 据 `APIKeyMode` 分支。

## 下一步

- 🔒 学 [认证机制](/guide/auth-concept)
- 📖 看 [`WithAPIKey`](./with-api-key)
- 📞 学 [JSONP 回调](/guide/jsonp)
