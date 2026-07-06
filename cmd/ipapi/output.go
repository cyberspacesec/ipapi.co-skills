// Package main: output.go 负责所有输出渲染。
//
// 两种模式：
//   - JSON 信封（默认）：{ok, command, args, data|error, meta}，stdout 输出成功，
//     stderr 输出错误。便于 AI Agent 程序化解析。
//   - --human/-H：人类可读的对齐表格或纯值，仅 stdout。
//
// raw/me-raw 命令特殊：成功时直接把原始字节写到 stdout，不包信封；
// 但出错时仍走错误信封到 stderr，保持 stdout 纯净。
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
)

// envelope 是统一的 JSON 输出信封。
type envelope struct {
	OK      bool        `json:"ok"`
	Command string      `json:"command"`
	Args    interface{} `json:"args,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *errDetail  `json:"error,omitempty"`
	Meta    *meta       `json:"meta,omitempty"`
}

// errDetail 是错误信封里的 error 子对象。
type errDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Sentinel  string `json:"sentinel,omitempty"`
	Retryable bool   `json:"retryable"`
}

// meta 是成功信封里的元信息。
type meta struct {
	Format      string    `json:"format"`
	DurationMs  int64     `json:"durationMs"`
	RetrievedAt time.Time `json:"retrievedAt"`
}

// renderOK 输出成功信封到 stdout。
func renderOK(cmd string, args, data interface{}, m *meta, human bool) {
	if human {
		printHuman(os.Stdout, cmd, data, m)
		return
	}
	env := envelope{OK: true, Command: cmd, Args: args, Data: data, Meta: m}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(env)
}

// exitError 携带退出码的错误。renderError 返回它，子命令直接 return，
// main 用它取出退出码。其 Error() 返回空串，避免 cobra 再打印 usage。
type exitError struct {
	code int
}

func (e *exitError) Error() string { return "" }

// renderError 输出错误信封到 stderr，并返回一个 exitError 供子命令 return。
// 无论 human 与否，错误永远走 JSON 到 stderr（机器可读）。
func renderError(cmd string, args interface{}, err error) error {
	code, codeName, sentinelName, retryable := classifyError(err)
	env := envelope{
		OK:      false,
		Command: cmd,
		Args:    args,
		Error: &errDetail{
			Code:      codeName,
			Message:   err.Error(),
			Sentinel:  sentinelName,
			Retryable: retryable,
		},
	}
	enc := json.NewEncoder(os.Stderr)
	enc.SetIndent("", "  ")
	_ = enc.Encode(env)
	return &exitError{code: code}
}

// printRaw 把原始字节直接写到 stdout（raw/me-raw 命令成功路径）。
func printRaw(b []byte) {
	os.Stdout.Write(b)
}

// printHuman 渲染人类可读输出。根据 command 分派不同格式。
func printHuman(w io.Writer, cmd string, data interface{}, m *meta) {
	switch cmd {
	case "info", "me":
		printIPInfoHuman(w, data, m)
	case "quota":
		if q, ok := data.(*ipapi.Quota); ok {
			printQuotaHuman(w, q)
			return
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(data)
	case "field", "me-field":
		// data 形如 {field, value}，human 模式只打印纯值一行，便于 shell pipe
		if fv, ok := data.(*fieldValue); ok {
			fmt.Fprintln(w, fv.Value)
			return
		}
		// 兜底：JSON 打印
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(data)
	default:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(data)
	}
}

// fieldValue 是 field/me-field 命令的 data 结构。
type fieldValue struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

// printIPInfoHuman 把 *ipapi.IPInfo 渲染成对齐的 key:value 表格。
func printIPInfoHuman(w io.Writer, data interface{}, m *meta) {
	info, ok := data.(*ipapi.IPInfo)
	if !ok {
		// 非 IPInfo（理论上不会到这里），兜底 JSON
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(data)
		return
	}
	if info.IP != "" {
		fmt.Fprintf(w, "\n  🌐 %s\n", info.IP)
	}
	fmt.Fprintln(w, "  ─────────────────────────────────────────")
	row(w, "Network", info.Network)
	row(w, "Version", info.Version)
	row(w, "City", info.City)
	row(w, "Region", info.Region)
	row(w, "Region Code", info.RegionCode)
	if info.CountryName != "" || info.CountryCode != "" {
		row(w, "Country", fmt.Sprintf("%s (%s)", info.CountryName, info.CountryCode))
	}
	row(w, "Continent", info.ContinentCode)
	row(w, "In EU", boolStr(info.InEU))
	row(w, "Postal", info.GetPostal())
	row(w, "Latitude", fmt.Sprintf("%g", info.Latitude))
	row(w, "Longitude", fmt.Sprintf("%g", info.Longitude))
	row(w, "LatLong", info.LatLong)
	row(w, "Timezone", info.Timezone)
	row(w, "UTC Offset", info.UTCOffset)
	row(w, "Calling Code", info.CountryCallingCode)
	row(w, "Currency", fmt.Sprintf("%s (%s)", info.CurrencyName, info.Currency))
	row(w, "Languages", info.Languages)
	if info.CountryArea > 0 {
		row(w, "Country Area", fmt.Sprintf("%g km²", info.CountryArea))
	}
	if info.CountryPopulation > 0 {
		row(w, "Population", fmt.Sprintf("%d", info.CountryPopulation))
	}
	row(w, "ASN", info.ASN)
	row(w, "Org", info.Org)
	row(w, "Hostname", info.Hostname)
	fmt.Fprintln(w, "  ─────────────────────────────────────────")
	if m != nil {
		row(w, "Retrieved at", m.RetrievedAt.Local().Format("2006-01-02 15:04:05"))
		row(w, "Elapsed", fmt.Sprintf("%d ms", m.DurationMs))
	}
	fmt.Fprintln(w)
}

func row(w io.Writer, key, val string) {
	if val == "" {
		return
	}
	fmt.Fprintf(w, "  %-16s %s\n", key, val)
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// nowUTC 返回当前 UTC 时间，封装以便测试替换。
var nowUTC = func() time.Time { return time.Now().UTC() }

// msSince 返回自 start 以来的毫秒数。
func msSince(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}

// joinArgs 把任意 args 渲染成简短字符串（用于 human 模式的命令行回显，可选）。
func joinArgs(args interface{}) string {
	b, _ := json.Marshal(args)
	return strings.TrimSpace(string(b))
}
