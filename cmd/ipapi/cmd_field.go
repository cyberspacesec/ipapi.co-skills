// Package main: cmd_field.go 实现 `ipapi field <ip> <field>` 与 `ipapi me-field <field>`。
// 调用 GetField/GetClientField，返回单字段字符串。JSON 信封里 data 为 {field, value}。
package main

import (
	"context"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
)

var fieldCmd = &cobra.Command{
	Use:   "field <ip> <field>",
	Short: "查询指定 IP 的单个字段",
	Long: `查询指定 IP 的某个字段值（如 country、city、asn、org、timezone 等）。

字段必须是 ipapi.co 支持的有效字段名，运行 ` + "`ipapi fields`" + ` 查看全部。
返回的是字段原始字符串值。

示例:
  ipapi field 8.8.8.8 country
  ipapi field 8.8.8.8 asn
  ipapi field 8.8.8.8 org --human    # human 模式只打印纯值，便于 shell pipe`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ip, field := args[0], args[1]
		argsMap := map[string]string{"ip": ip, "field": field}
		return runField(cmd, "field", argsMap, func(ctx context.Context) (string, error) {
			return state.client.GetField(ctx, ip, field)
		}, field)
	},
}

var meFieldCmd = &cobra.Command{
	Use:   "me-field <field>",
	Short: "查询本机 IP 的单个字段",
	Long: `查询本机公网 IP 的某个字段值。

示例:
  ipapi me-field asn
  ipapi me-field country --human`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		field := args[0]
		argsMap := map[string]string{"field": field}
		return runField(cmd, "me-field", argsMap, func(ctx context.Context) (string, error) {
			return state.client.GetClientField(ctx, field)
		}, field)
	},
}

// runField 是 field/me-field 共享逻辑：调 SDK → trim → 渲染信封。
// GetField/GetClientField 返回原始响应体字符串，可能含尾换行，需 TrimSpace。
func runField(cmd *cobra.Command, command string, argsMap map[string]string, fn func(ctx context.Context) (string, error), field string) error {
	ctx, cancel := context.WithTimeout(context.Background(), state.cfg.Timeout+time.Second)
	defer cancel()
	start := time.Now()
	val, err := fn(ctx)
	if err != nil {
		return renderError(command, argsMap, err)
	}
	val = strings.TrimSpace(val)
	m := &meta{
		Format:     state.cfg.Format,
		DurationMs: msSince(start),
		// field 端点不返回 IPInfo，RetrievedAt 用本地时间
		RetrievedAt: nowUTC(),
	}
	data := &fieldValue{Field: field, Value: val}
	renderOK(command, argsMap, data, m, state.cfg.Human)
	return nil
}

// 编译期确认 ipapi 包被使用（field 命令本身不直接调 ipapi 符号，但保持导入一致）。
var _ = ipapi.FormatJSON
