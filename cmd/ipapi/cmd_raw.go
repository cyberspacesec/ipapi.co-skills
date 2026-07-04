// Package main: cmd_raw.go 实现 `ipapi raw <ip>` 与 `ipapi me-raw`。
// 调用 GetIPInfoRaw/GetClientIPInfoRaw，返回原始字节（xml/csv/yaml/jsonp）。
// 成功时直接把原始字节写到 stdout，不包信封；出错时仍走错误信封到 stderr。
package main

import (
	"context"
	"time"

	"github.com/spf13/cobra"
)

var rawCmd = &cobra.Command{
	Use:   "raw <ip>",
	Short: "查询指定 IP 的原始格式响应",
	Long: `查询指定 IP，返回 --format 指定格式的原始响应字节（不解析）。

支持 json/jsonp/xml/csv/yaml。成功时直接输出原始字节到 stdout，
适合管道喂给其它工具（如 jq、csv 处理器）。错误信封输出到 stderr。

示例:
  ipapi raw 8.8.8.8 -f yaml
  ipapi raw 8.8.8.8 -f csv | column -t -s,
  ipapi raw 8.8.8.8 -f jsonp --callback myCb`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ip := args[0]
		argsMap := map[string]string{"ip": ip, "format": state.cfg.Format}
		return runRaw("raw", argsMap, func(ctx context.Context, format string) ([]byte, error) {
			return state.client.GetIPInfoRaw(ctx, ip, format)
		})
	},
}

var meRawCmd = &cobra.Command{
	Use:   "me-raw",
	Short: "查询本机 IP 的原始格式响应",
	Long: `查询本机公网 IP，返回 --format 指定格式的原始响应字节。

示例:
  ipapi me-raw -f csv
  ipapi me-raw -f yaml`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		argsMap := map[string]string{"format": state.cfg.Format}
		return runRaw("me-raw", argsMap, func(ctx context.Context, format string) ([]byte, error) {
			return state.client.GetClientIPInfoRaw(ctx, format)
		})
	},
}

// runRaw 是 raw/me-raw 共享逻辑：只调一次 fn，成功则原始字节直出 stdout，
// 失败则错误信封到 stderr，保持 stdout 纯净。
func runRaw(command string, argsMap map[string]string, fn func(ctx context.Context, format string) ([]byte, error)) error {
	ctx, cancel := context.WithTimeout(context.Background(), state.cfg.Timeout+time.Second)
	defer cancel()
	b, err := fn(ctx, state.cfg.Format)
	if err != nil {
		return renderError(command, argsMap, err)
	}
	printRaw(b)
	return nil
}

// 让 time 包被使用（context 超时用 time.Second）。
var _ = time.Second
