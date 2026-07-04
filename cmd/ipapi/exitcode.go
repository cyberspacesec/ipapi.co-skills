// Package main: exitcode.go 把 SDK 的哨兵错误映射成稳定的进程退出码。
//
// 设计原则：
//   - 0  = 成功
//   - 2  = 用法错误（参数/旗标不合法）
//   - 3..12 = SDK 哨兵错误，每个错误一个码，便于 AI Agent 用 $? 分支
//   - 70 = 其他未分类错误（EX_UNAVAILABLE，参考 sysexits.h）
//
// 错误码与 error.code 字符串的对应关系是公开契约，文档站 exit-codes.md 同步。
package main

import (
	"errors"

	"github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
)

// 错误码常量。命名与 error.code 字符串保持一致（去掉 Err 前缀、转大写）。
const (
	exitOK               = 0
	exitUsage            = 2
	exitInvalidIP        = 3
	exitInvalidField     = 4
	exitInvalidFormat    = 5
	exitRateLimited      = 6
	exitReservedIP       = 7
	exitNotFound         = 8
	exitServerError      = 9
	exitMethodNotAllowed = 10
	exitInvalidKey       = 11
	exitUnexpectedData   = 12
	exitInternal         = 70
)

// errCodePair 把哨兵错误与退出码、code 字符串、sentinel 名绑定。
type errCodePair struct {
	sentinel error
	code     int
	name     string // 大写 code，如 "INVALID_IP"
	label    string // SDK 变量名，如 "ErrInvalidIP"
}

var errCodeTable = []errCodePair{
	{ipapi.ErrInvalidIP, exitInvalidIP, "INVALID_IP", "ErrInvalidIP"},
	{ipapi.ErrInvalidField, exitInvalidField, "INVALID_FIELD", "ErrInvalidField"},
	{ipapi.ErrInvalidFormat, exitInvalidFormat, "INVALID_FORMAT", "ErrInvalidFormat"},
	{ipapi.ErrRateLimited, exitRateLimited, "RATE_LIMITED", "ErrRateLimited"},
	{ipapi.ErrReservedIP, exitReservedIP, "RESERVED_IP", "ErrReservedIP"},
	{ipapi.ErrNotFound, exitNotFound, "NOT_FOUND", "ErrNotFound"},
	{ipapi.ErrServerError, exitServerError, "SERVER_ERROR", "ErrServerError"},
	{ipapi.ErrMethodNotAllowed, exitMethodNotAllowed, "METHOD_NOT_ALLOWED", "ErrMethodNotAllowed"},
	{ipapi.ErrInvalidKey, exitInvalidKey, "INVALID_KEY", "ErrInvalidKey"},
	{ipapi.ErrUnexpectedData, exitUnexpectedData, "UNEXPECTED_DATA", "ErrUnexpectedData"},
}

// classifyError 返回 err 对应的 (code, codeName, sentinelName, retryable)。
// 未匹配时返回 INTERNAL。
func classifyError(err error) (code int, codeName, sentinelName string, retryable bool) {
	if err == nil {
		return exitOK, "", "", false
	}
	for _, p := range errCodeTable {
		if errors.Is(err, p.sentinel) {
			return p.code, p.name, p.label, ipapi.IsRetryableError(p.sentinel)
		}
	}
	return exitInternal, "INTERNAL", "", false
}

// errToExitCode 返回 err 对应的退出码。识别 exitError（携带预计算码）。
func errToExitCode(err error) int {
	if err == nil {
		return exitOK
	}
	if ee, ok := err.(*exitError); ok {
		return ee.code
	}
	code, _, _, _ := classifyError(err)
	return code
}
