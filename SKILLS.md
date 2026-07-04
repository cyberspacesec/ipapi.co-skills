# ipapi CLI — AI Agent 接入说明

> 本文件面向 **AI Agent / 自动化脚本**。如果你是 Agent，读完这一页就能用 `ipapi` CLI 完成 IP 地理位置查询，无需翻文档。

## 这是什么

`ipapi` 是一个命令行 IP 地理位置查询工具，封装 [ipapi.co](https://ipapi.co) API。**默认输出 JSON**，错误用稳定退出码区分，专为程序化调用设计。

## 安装

```bash
go install github.com/cyberspacesec/ipapi.co-skills/cmd/ipapi@latest
```

或从 [releases](https://github.com/cyberspacesec/ipapi.co-skills/releases) 下载二进制。

## 命令一览

| 命令 | 作用 | 需要网络 |
|---|---|---|
| `ipapi info <ip>` | 查询指定 IP 完整信息 | 是 |
| `ipapi me` | 查询本机公网 IP 完整信息 | 是 |
| `ipapi field <ip> <field>` | 查询指定 IP 的单个字段 | 是 |
| `ipapi me-field <field>` | 查询本机 IP 的单个字段 | 是 |
| `ipapi raw <ip> -f <fmt>` | 查询指定 IP 的原始格式（xml/csv/yaml/jsonp） | 是 |
| `ipapi me-raw -f <fmt>` | 查询本机 IP 的原始格式 | 是 |
| `ipapi fields` | 列出全部可查字段（分组） | **否** |
| `ipapi version` | 版本信息 | 否 |
| `ipapi completion <shell>` | 生成 shell 补全 | 否 |

## 输出协议（重要）

**默认 JSON 信封**，成功写到 stdout：

```json
{
  "ok": true,
  "command": "info",
  "args": {"ip": "8.8.8.8", "format": "json"},
  "data": { ...IPInfo 28 字段... },
  "meta": {"format": "json", "durationMs": 312, "retrievedAt": "2026-07-04T10:01:22Z"}
}
```

**错误信封**写到 **stderr**（stdout 保持纯净，可安全 pipe）：

```json
{
  "ok": false,
  "command": "info",
  "args": {"ip": "999.1.1.1"},
  "error": {
    "code": "INVALID_IP",
    "message": "invalid IP address",
    "sentinel": "ErrInvalidIP",
    "retryable": false
  }
}
```

- 判成功：`ok == true`（或退出码 == 0）
- 判错误类型：读 `error.code`（稳定字符串，见下表），**不要解析 message**
- 判是否值得重试：`error.retryable == true`

## 退出码

| 码 | code | 含义 | retryable |
|---|---|---|---|
| 0 | — | 成功 | — |
| 2 | USAGE | 参数/旗标错误 | 否 |
| 3 | INVALID_IP | 无效 IP 地址 | 否 |
| 4 | INVALID_FIELD | 无效字段名 | 否 |
| 5 | INVALID_FORMAT | 无效响应格式 | 否 |
| 6 | RATE_LIMITED | API 限流 | **是** |
| 7 | RESERVED_IP | 保留 IP | 否 |
| 8 | NOT_FOUND | 资源未找到 | **是** |
| 9 | SERVER_ERROR | 服务端错误 | **是** |
| 10 | METHOD_NOT_ALLOWED | 方法不允许 | 否 |
| 11 | INVALID_KEY | 无效 API key | 否 |
| 12 | UNEXPECTED_DATA | 意外响应 | 否 |
| 70 | INTERNAL | 其他内部错误 | 否 |

Agent 可用 `$?` / exit code 快速分支：`[ $? -eq 0 ]` 成功；`[ $? -ge 64 ]` 非业务错误；6/8/9 值得退避重试。

## 可查字段（28 个）

运行 `ipapi fields` 或 `ipapi fields --json` 获取完整列表。常用：

- **地理**：`city` `region` `country` `country_name` `country_code` `postal` `latitude` `longitude` `latlong`
- **网络**：`asn` `org` `network` `version` `hostname`
- **时区**：`timezone` `utc_offset`
- **经济**：`currency` `currency_name` `country_calling_code`
- **统计**：`country_area` `country_population`

## 配置优先级

旗标 > 环境变量 > `~/.ipapi.json` > 默认值

常用环境变量：`IPAPI_API_KEY`、`IPAPI_FORMAT`、`IPAPI_BASE_URL`、`IPAPI_RETRIES`、`IPAPI_TIMEOUT`。

## Agent 调用模板

**查一个 IP 的完整信息：**
```bash
ipapi info 8.8.8.8 | jq '.data.country_name'
```

**只取一个字段（最省配额）：**
```bash
ipapi field 8.8.8.8 country --human   # human 模式直接输出纯值 "US"
```

**查本机公网 IP：**
```bash
ipapi me | jq '.data.ip'
```

**先探查能查什么再查：**
```bash
ipapi fields --json | jq '.all[]'      # 拿到全部字段名
ipapi field 8.8.8.8 asn                # 再查具体字段
```

**错误处理（bash）：**
```bash
if ! ipapi info "$IP" > /tmp/info.json 2>/tmp/err.json; then
  code=$(jq -r '.error.code' /tmp/err.json)
  case "$code" in
    RATE_LIMITED) sleep 60 && retry;;
    INVALID_IP) echo "bad ip";;
    *) cat /tmp/err.json;;
  esac
fi
```

**批量查询（带限速）：**
```bash
for ip in 8.8.8.8 1.1.1.1 9.9.9.9; do
  ipapi field "$ip" country --human
  sleep 1
done
```

## 何时用 CLI vs Go SDK

- **CLI**：脚本、自动化、CI、Agent 接入、临时查询 — 90% 场景
- **Go SDK**（`pkg/ipapi`）：需要在 Go 程序内嵌入、要复用 Client 连接池、要自定义错误处理 — 见 [库开发指南](https://cyberspacesec.github.io/ipapi.co-skills/guide/intro)

## 限制

- `info`/`me` 仅支持 `--format json`（结构化解析）；xml/csv/yaml/jsonp 用 `raw`/`me-raw`
- 匿名调用有配额限制（ipapi.co 限制），生产环境用 `IPAPI_API_KEY` 配置 API key
- `field`/`me-field` 的 `--human` 输出纯值一行（便于 pipe）；不加 `--human` 输出 JSON 信封

## 完整文档

- 🖥 [CLI 文档首页](https://cyberspacesec.github.io/ipapi.co-skills/cli/)
- 📋 [命令速查](https://cyberspacesec.github.io/ipapi.co-skills/cli/commands)
- 🚪 [退出码详解](https://cyberspacesec.github.io/ipapi.co-skills/cli/exit-codes)
- 🤖 [Agent 接入指南](https://cyberspacesec.github.io/ipapi.co-skills/cli/agent)
