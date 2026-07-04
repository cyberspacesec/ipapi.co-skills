// examples/basic_usage/main.go
package main

import (
	"context"
	"fmt"
	"github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
	"log"
	"time"
)

func main() {
	// 创建默认客户端
	client := ipapi.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 获取当前IP的完整信息
	info, err := client.GetClientIPInfo(ctx, "json")
	if err != nil {
		log.Fatalf("获取IP信息失败: %v", err)
	}

	fmt.Printf("IP基本信息:\n%s\n国家: %s\n城市: %s\n时区: %s\n经纬度: %s\n",
		info.IP, info.CountryName, info.City, info.Timezone, info.LatLong)

	// 解析经纬度坐标
	lat, lon, err := info.ParseLatLong()
	if err == nil {
		fmt.Printf("解析后的坐标: 纬度=%.4f, 经度=%.4f\n", lat, lon)
	}
}
