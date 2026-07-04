// Package main: client.go 把合并后的 Config 映射成一个 *ipapi.Client。
package main

import (
	"strings"

	"github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
)

// newClient 把 Config 映射到 SDK 的 *ipapi.Client，全部走函数式选项。
//
// 选项按传入顺序应用：WithTimeout 在末尾，落在 NewClient 内建（或用户
// 传入的）*http.Client 上，保留 CheckRedirect 策略，仅覆盖 Timeout。
// Config 加载阶段已校验 Timeout>0，故 WithTimeout 总会生效。
func newClient(cfg *Config) *ipapi.Client {
	opts := []ipapi.ClientOption{
		ipapi.WithBaseURL(cfg.BaseURL),
		ipapi.WithUserAgent(cfg.UserAgent),
		ipapi.WithRetries(cfg.Retries),
		ipapi.WithTimeout(cfg.Timeout),
	}
	if cfg.APIKey != "" {
		opts = append(opts, ipapi.WithAPIKey(cfg.APIKey))
		if strings.EqualFold(cfg.APIKeyMode, "query") {
			opts = append(opts, ipapi.WithAPIKeyQuery())
		}
	}
	if cfg.Callback != "" {
		opts = append(opts, ipapi.WithCallback(cfg.Callback))
	}
	return ipapi.NewClient(opts...)
}
