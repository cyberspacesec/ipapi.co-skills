// examples/error_handling/main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
	"log"
)

func main() {
	client := ipapi.NewClient()

	// 测试各种错误场景
	testCases := []struct {
		ip     string
		field  string
		format string
	}{
		{"invalid.ip", "city", "json"},     // 无效IP
		{"8.8.8.8", "invalid_field", ""},   // 无效字段
		{"10.0.0.1", "city", "json"},       // 保留IP
		{"999.999.999.999", "country", ""}, // 非法格式
	}

	for _, tc := range testCases {
		fmt.Printf("\n测试用例: IP=%s, Field=%s\n", tc.ip, tc.field)

		if _, err := client.GetIPInfo(context.Background(), tc.ip, tc.format); err != nil {
			handleError(err)
		}

		if _, err := client.GetField(context.Background(), tc.ip, tc.field); err != nil {
			handleError(err)
		}
	}
}

func handleError(err error) {
	switch {
	case errors.Is(err, ipapi.ErrInvalidIP):
		fmt.Println("→ 无效IP地址错误")
	case errors.Is(err, ipapi.ErrInvalidField):
		fmt.Println("→ 请求字段不存在")
	case errors.Is(err, ipapi.ErrReservedIP):
		fmt.Println("→ 保留IP地址错误")
	case errors.Is(err, ipapi.ErrRateLimited):
		fmt.Println("→ 触发速率限制，建议稍后重试")
	default:
		log.Printf("未处理的错误类型: %v", err)
	}
}
