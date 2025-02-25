# go-ipapi.co

[![Go Reference](https://pkg.go.dev/badge/github.com/cyberspacesec/go-ipapi.co.svg)](https://pkg.go.dev/github.com/cyberspacesec/go-ipapi.co)
[![GitHub License](https://img.shields.io/github/license/cyberspacesec/go-ipapi.co)](https://github.com/cyberspacesec/go-ipapi.co/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/cyberspacesec/go-ipapi.co)](https://goreportcard.com/report/github.com/cyberspacesec/go-ipapi.co)

Go客户端库用于[ipapi.co](https://ipapi.co/api/#introduction)的IP地理位置查询API，支持IPv4/IPv6查询、字段级数据获取和高级错误处理。

## 功能特性

- ✅ 完整IP信息查询（JSON/XML/CSV/YAML）
- 🎯 单个字段查询（国家/城市/时区等）
- 🌍 客户端IP自动检测
- 🔒 支持HTTPS和API密钥认证
- 🔄 自动重试和速率限制处理
- 🛠 可定制HTTP客户端和上下文支持
- 📡 支持IPv4和IPv6地址
- 🧪 完整单元测试覆盖

## 安装

```bash
go get github.com/cyberspacesec/go-ipapi.co
```

## 快速开始

```go
package main

import (
	"context"
	"fmt"
	"github.com/cyberspacesec/go-ipapi.co/pkg/ipapi"
	"time"
)

func main() {
	client := ipapi.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 获取当前IP信息
	info, _ := client.GetClientIPInfo(ctx, "json")
	fmt.Printf("Location: %s, %s\nTimezone: %s\n",
		info.City, info.CountryName, info.Timezone)

	// 查询特定IP的ASN
	asn, _ := client.GetField(ctx, "8.8.8.8", "asn")
	fmt.Println("Google DNS ASN:", asn)
}
```

## 高级用法

### 自定义配置客户端

```go
client := ipapi.NewClient(
    ipapi.WithAPIKey("your_api_key"),
    ipapi.WithCustomHTTPClient(&http.Client{
        Timeout: 30 * time.Second,
    }),
    ipapi.WithErrorHandler(func(err error) error {
        log.Printf("API Error: %v", err)
        return err
    }),
)
```

### 批量查询处理

```go
func batchLookup(ips []string) {
    for _, ip := range ips {
        info, err := client.GetIPInfo(context.Background(), ip, "json")
        if err != nil {
            continue
        }
        fmt.Printf("%s | %s | %s\n", 
            info.IP, info.CountryCode, info.ASN)
    }
}
```

## 错误处理

### 错误类型检查

```go
info, err := client.GetIPInfo(ctx, "invalid.ip", "json")
if err != nil {
    switch {
    case errors.Is(err, ipapi.ErrInvalidIP):
        fmt.Println("无效IP地址")
    case errors.Is(err, ipapi.ErrRateLimited):
        time.Sleep(1 * time.Minute) // 等待重试
    default:
        log.Fatal(err)
    }
}
```

### 自定义错误处理

```go
client := ipapi.NewClient(
    ipapi.WithErrorHandler(func(err error) error {
        if apiErr, ok := err.(*ipapi.APIError); ok {
            metrics.LogError(apiErr.Reason)
        }
        return err
    }),
)
```

## 贡献指南

欢迎通过以下方式参与贡献：
1. 提交Issue报告问题
2. Fork仓库并提交Pull Request
3. 完善测试用例和文档
4. 分享使用案例

请确保：
- 代码符合Go语言规范
- 新增功能包含测试用例
- 文档保持同步更新

## 许可证

本项目基于 MIT 许可证发布，详见 [LICENSE](LICENSE) 文件。

---

> 📌 注意：使用API需遵守ipapi.co的服务条款，生产环境建议申请API密钥。



