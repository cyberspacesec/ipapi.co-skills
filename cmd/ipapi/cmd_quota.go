// Package main: cmd_quota.go 实现 `ipapi quota` 子命令。
//
// 调用 SDK 的 GetQuota 查询当前 API key 的剩余 IP 查询配额。
// 该能力对应 ipapi.co 的 GET /quota/ endpoint（官方 api-doc 未文档化，但稳定可用）。
//
// 用法:
//
//	ipapi quota                   # JSON 信封
//	ipapi quota -H                # 人类可读
//	IPAPI_API_KEY=xxx ipapi quota # 用 key 查询真实剩余配额
package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
)

var quotaCmd = &cobra.Command{
	Use:   "quota",
	Short: "查询当前 API key 的剩余配额",
	Long: `查询当前 API key 的剩余 IP 查询配额（对应 ipapi.co 的 GET /quota/）。

无需指定参数，认证复用全局 --api-key / IPAPI_API_KEY。
- 有效 key：返回剩余查询次数
- 未配置 key：返回 "API key needed"
- 无效 key：错误信封，退出码 11 (INVALID_KEY)

示例:
  ipapi quota
  ipapi quota -H
  IPAPI_API_KEY=xxx ipapi quota`,
	Args: cobra.NoArgs,
	RunE: runQuota,
}

// runQuota 调用 GetQuota 并渲染信封。
func runQuota(cmd *cobra.Command, args []string) error {
	command := "quota"
	argsMap := map[string]string{}
	ctx, cancel := context.WithTimeout(context.Background(), state.cfg.Timeout+time.Second)
	defer cancel()
	start := time.Now()

	q, err := state.client.GetQuota(ctx)
	if err != nil {
		return renderError(command, argsMap, err)
	}

	m := &meta{
		Format:      state.cfg.Format,
		DurationMs:  msSince(start),
		RetrievedAt: nowUTC(),
	}
	renderOK(command, argsMap, q, m, state.cfg.Human)
	return nil
}

// printQuotaHuman 渲染 quota 的人类可读输出。
// 由 output.go 的 printHuman 按 command 名分派调用。
func printQuotaHuman(w io.Writer, q *ipapi.Quota) {
	fmt.Fprintln(w, "\n  📊 IP 查询配额")
	fmt.Fprintln(w, "  ─────────────────────────────────────────")
	if n, ok := q.AvailableInt(); ok {
		fmt.Fprintf(w, "  %-16s %d 次\n", "剩余配额", n)
	} else {
		fmt.Fprintf(w, "  %-16s %s\n", "状态", q.Available)
	}
	fmt.Fprintln(w, "  ─────────────────────────────────────────")
	fmt.Fprintln(w)
}
