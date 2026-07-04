// examples/advanced_usage/main.go
package main

import (
	"context"
	"fmt"
	"github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
	"log"
	"net/http"
	"time"
)

func main() {
	// 创建定制化客户端
	client := ipapi.NewClient(
		ipapi.WithAPIKey("your_api_key_here"),
		ipapi.WithCustomHTTPClient(&http.Client{
			Timeout:   15 * time.Second,
			Transport: &http.Transport{MaxIdleConns: 10},
		}),
		ipapi.WithErrorHandler(customErrorHandler),
	)

	ctx := context.Background()

	// 获取特定IP的详细信息
	printIPInfo(client, ctx, "8.8.8.8")
	printIPInfo(client, ctx, "2001:4860:4860::8888")

	// 获取特定字段
	if org, err := client.GetField(ctx, "8.8.8.8", "org"); err == nil {
		fmt.Printf("\nDNS服务器所属组织: %s\n", org)
	}
}

func printIPInfo(client *ipapi.Client, ctx context.Context, ip string) {
	info, err := client.GetIPInfo(ctx, ip, "json")
	if err != nil {
		log.Printf("查询IP %s 失败: %v", ip, err)
		return
	}

	fmt.Printf("\n%s 的详细信息:\n", ip)
	fmt.Printf("地理位置: %s, %s\n网络信息: %s (%s)\n时区: %s (UTC%s)\n",
		info.City, info.CountryName,
		info.ASN, info.Org,
		info.Timezone, info.UTCOffset)
}

func customErrorHandler(err error) error {
	fmt.Printf("\n自定义错误处理: %v\n", err)
	return err
}
