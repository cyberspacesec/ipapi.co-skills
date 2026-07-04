// Package main 是 ipapi 命令行工具的入口。
//
// ipapi CLI 是 ipapi.co-skills Go SDK 的终端形态，专为 AI Agent 接入设计：
// 默认输出 JSON 信封，--human/-H 切人类可读，错误用稳定退出码区分。
package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {
	err := Execute()
	switch {
	case err == nil:
		os.Exit(exitOK)
	default:
		// exitError：业务错误，信封已输出到 stderr，直接用其携带的退出码。
		var ee *exitError
		if errors.As(err, &ee) {
			os.Exit(ee.code)
		}
		// 其它错误（用法错误、PreRunE 校验失败、子命令未识别的错误）。
		// 打印简短信息到 stderr，返回 usage 码。
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitUsage)
	}
}
