// Package main: cmd_info.go 实现 `ipapi info <ip>` 与 `ipapi me` 子命令。
// 二者都调用返回 *IPInfo 的 SDK 方法，输出统一走 JSON 信封或 human 表格。
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
)

var infoCmd = &cobra.Command{
	Use:   "info <ip>",
	Short: "查询指定 IP 的完整信息",
	Long: `查询指定 IPv4/IPv6 地址的完整地理位置信息，返回结构化 IPInfo。

默认输出 JSON 信封，加 --human/-H 输出对齐表格。仅支持 --format json
（其它格式请用 ipapi raw <ip> -f <format>）。

示例:
  ipapi info 8.8.8.8
  ipapi info 8.8.8.8 -H
  ipapi info 2001:4860:4860::8888
  IPAPI_API_KEY=xxx ipapi info 8.8.8.8`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ip := args[0]
		return runIPInfo(cmd, "info", map[string]string{"ip": ip, "format": state.cfg.Format}, func(ctx context.Context, format string) (*ipapi.IPInfo, error) {
			return state.client.GetIPInfo(ctx, ip, format)
		})
	},
}

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "查询本机公网 IP 的完整信息",
	Long: `查询调用方（本机出口）公网 IP 的完整地理位置信息。

无需指定 IP 地址，ipapi.co 会自动识别来源 IP。

示例:
  ipapi me
  ipapi me -H`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runIPInfo(cmd, "me", map[string]string{"format": state.cfg.Format}, func(ctx context.Context, format string) (*ipapi.IPInfo, error) {
			return state.client.GetClientIPInfo(ctx, format)
		})
	},
}

// runIPInfo 是 info/me 共享的执行逻辑：限流校验 → 调 SDK → 渲染信封。
func runIPInfo(cmd *cobra.Command, command string, argsMap map[string]string, fn func(ctx context.Context, format string) (*ipapi.IPInfo, error)) error {
	if err := requireJSONFormat(state.cfg.Format, command); err != nil {
		return renderError(command, argsMap, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), state.cfg.Timeout+time.Second)
	defer cancel()
	start := time.Now()
	info, err := fn(ctx, state.cfg.Format)
	if err != nil {
		return renderError(command, argsMap, err)
	}
	m := &meta{
		Format:      state.cfg.Format,
		DurationMs:  msSince(start),
		RetrievedAt: nowUTC(),
	}
	if info != nil && info.RetrievedAt.IsZero() {
		info.RetrievedAt = m.RetrievedAt
	}
	renderOK(command, argsMap, info, m, state.cfg.Human)
	return nil
}

// requireJSONFormat 限制 info/me 只接受 json（GetIPInfo/GetClientIPInfo 只解 JSON）。
func requireJSONFormat(format, command string) error {
	if format != string(ipapi.FormatJSON) {
		return fmt.Errorf("--format %q 不被 %s 支持：info/me 仅支持 json；其它格式请用 raw/me-raw", format, command)
	}
	return nil
}
