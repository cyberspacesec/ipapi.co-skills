# ipapi.co-skills

[![CLI Release](https://img.shields.io/github/v/release/cyberspacesec/ipapi.co-skills)](https://github.com/cyberspacesec/ipapi.co-skills/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/cyberspacesec/ipapi.co-skills.svg)](https://pkg.go.dev/github.com/cyberspacesec/ipapi.co-skills)
[![CI](https://github.com/cyberspacesec/ipapi.co-skills/actions/workflows/ci.yml/badge.svg)](https://github.com/cyberspacesec/ipapi.co-skills/actions/workflows/ci.yml)
[![GitHub License](https://img.shields.io/github/license/cyberspacesec/ipapi.co-skills)](https://github.com/cyberspacesec/ipapi.co-skills/blob/main/LICENSE)
[![Docs](https://img.shields.io/badge/docs-vitepress-3c8c4a)](https://cyberspacesec.github.io/ipapi.co-skills/)

**命令行 IP 地理位置查询工具**，封装 [ipapi.co](https://ipapi.co) API。默认输出 JSON、稳定退出码、AI Agent 友好；背后是一个零依赖的 Go SDK。

📖 **完整文档**：[https://cyberspacesec.github.io/ipapi.co-skills/](https://cyberspacesec.github.io/ipapi.co-skills/) · 📄 **English**: [README.md](./README.md)

## ✨ 为什么用它

- 🤖 **Agent 友好**：默认 JSON 信封 `{ok, command, data, meta}`，错误用稳定 `code` + 退出码区分，Agent 无需读文档即可调用
- 🚪 **稳定退出码**：`0` 成功，`3-12` 对应 10 类业务错误，`6/8/9` 标记可重试
- 🎯 **字段级查询**：`ipapi field 8.8.8.8 country` 一行取一个字段，`--human` 直出纯值便于 pipe
- 📡 **多格式**：JSON / JSONP / XML / CSV / YAML，`raw` 命令直出原始字节
- 🧭 **自描述**：`ipapi fields` 列出全部 28 个可查字段并分组
- 🧩 **也是 Go SDK**：CLI 只是薄壳，`pkg/ipapi` 是零运行时依赖的 Go 库，可直接嵌入程序

## 📥 安装

```bash
go install github.com/cyberspacesec/ipapi.co-skills/cmd/ipapi@latest
```

或从 [Releases](https://github.com/cyberspacesec/ipapi.co-skills/releases) 下载预编译二进制。

## ⚡ 30 秒上手

```bash
ipapi 8.8.8.8                    # 查询指定 IP（默认 JSON 信封）
ipapi 8.8.8.8 -H                 # 人类可读表格
ipapi field 8.8.8.8 country      # 只取一个字段
ipapi field 8.8.8.8 country -H   # 纯值输出，便于 shell pipe
ipapi me                          # 查本机公网 IP
ipapi fields                      # 列出全部可查字段
ipapi raw 8.8.8.8 -f yaml         # 原始 YAML 响应
```

JSON 输出示例：

```json
{
  "ok": true,
  "command": "info",
  "args": {"ip": "8.8.8.8", "format": "json"},
  "data": {
    "ip": "8.8.8.8",
    "city": "Mountain View",
    "country_name": "United States",
    "country_code": "US",
    "latitude": 37.42301,
    "longitude": -122.083352,
    "timezone": "America/Los_Angeles",
    "asn": "AS15169",
    "org": "Google LLC"
  },
  "meta": {"format": "json", "durationMs": 312, "retrievedAt": "2026-07-04T10:01:22Z"}
}
```

错误时退出码区分类型，便于 Agent 分支：

```bash
ipapi info 999.1.1.1; echo "exit=$?"   # exit=3 (INVALID_IP)
```

## 🤖 AI Agent 接入

Agent 调用模式：先 `ipapi fields` 探查可查字段，再用 `ipapi info/field` 查询，靠退出码与 `error.code` 判断结果。详见 [SKILLS.md](./SKILLS.md) 与 [Agent 接入指南](https://cyberspacesec.github.io/ipapi.co-skills/cli/agent)。

```bash
# 提取国家名
ipapi info 8.8.8.8 | jq -r '.data.country_name'

# 错误处理
if ! ipapi info "$IP" > /tmp/info.json 2>/tmp/err.json; then
  code=$(jq -r '.error.code' /tmp/err.json)
  case "$code" in
    RATE_LIMITED) sleep 60 && retry ;;
    INVALID_IP)   echo "bad ip" ;;
  esac
fi
```

## 📋 命令一览

| 命令 | 作用 |
|---|---|
| `ipapi info <ip>` | 查询指定 IP 完整信息 |
| `ipapi me` | 查询本机公网 IP 完整信息 |
| `ipapi field <ip> <field>` | 查询指定 IP 的单个字段 |
| `ipapi me-field <field>` | 查询本机 IP 的单个字段 |
| `ipapi raw <ip> -f <fmt>` | 查询指定 IP 的原始格式（xml/csv/yaml/jsonp） |
| `ipapi me-raw -f <fmt>` | 查询本机 IP 的原始格式 |
| `ipapi fields` | 列出全部可查字段（本地，无网络） |
| `ipapi version` | 版本信息 |
| `ipapi completion <shell>` | 生成 shell 补全 |

全局旗标：`--api-key`、`--api-key-mode`、`-f/--format`、`--base-url`、`--retries`、`--timeout`、`-H/--human`、`--config`、`--callback`。配置优先级：旗标 > 环境变量 > `~/.ipapi.json` > 默认值。

## 🧩 作为 Go SDK 使用

CLI 是 `pkg/ipapi` 的薄壳。若需在 Go 程序中嵌入：

```bash
go get github.com/cyberspacesec/ipapi.co-skills
```

```go
package main

import (
	"context"
	"fmt"
	"github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
	"time"
)

func main() {
	client := ipapi.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, _ := client.GetIPInfo(ctx, "8.8.8.8", string(ipapi.FormatJSON))
	fmt.Printf("Location: %s, %s\n", info.City, info.CountryName)
}
```

SDK 特性：10 个函数式选项（`WithAPIKey`/`WithCustomHTTPClient`/`WithErrorHandler`/`WithCallback`/`WithBaseURL`/`WithUserAgent`/`WithRetries`/`WithTimeout`/`WithRateLimiter` 等）、5 种格式、10 种哨兵错误（`errors.Is` 匹配）、自动重试与限流、IPv4/IPv6 全支持、100% 测试覆盖。

- 📖 [库开发指南](https://cyberspacesec.github.io/ipapi.co-skills/guide/intro)
- 📚 [API 参考](https://cyberspacesec.github.io/ipapi.co-skills/api/methods)
- 🐙 [Go 文档](https://pkg.go.dev/github.com/cyberspacesec/ipapi.co-skills)

## 🏗 构建

```bash
# CLI
go build ./cmd/ipapi/

# 测试
go test ./...
```

发布跨平台二进制由 GitHub Actions + GoReleaser 自动完成（打 `v*` tag 触发）。

## 🤝 贡献

欢迎提交 Issue 与 PR。请确保：
- 代码符合 Go 规范，`go vet` 通过
- 新增功能包含测试
- 文档保持同步

## 📄 许可证

MIT，详见 [LICENSE](LICENSE)。

---

> 📌 使用 API 需遵守 ipapi.co 服务条款，生产环境建议申请 API Key。
