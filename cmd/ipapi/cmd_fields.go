// Package main: cmd_fields.go 实现 `ipapi fields` —— 列出全部可查字段。
// 这是 AI Agent 自发现的关键命令：先 fields 拿到字段清单，再决定查什么。
// 本地枚举，无网络请求。
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
)

// fieldGroup 把字段按语义分组，方便人类阅读与 Agent 理解字段含义。
type fieldGroup struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields"`
}

// fieldsIndex 定义字段分组（仅用于展示，不影响 SDK 校验）。
var fieldsIndex = []fieldGroup{
	{"identity", []string{"ip", "network", "version", "hostname"}},
	{"geo", []string{"city", "region", "region_code", "country", "country_name", "country_code", "country_code_iso3", "country_capital", "country_tld", "continent_code", "in_eu", "postal", "latitude", "longitude", "latlong"}},
	{"time", []string{"timezone", "utc_offset"}},
	{"network", []string{"asn", "org"}},
	{"culture", []string{"languages", "country_calling_code"}},
	{"economy", []string{"currency", "currency_name"}},
	{"stats", []string{"country_area", "country_population"}},
}

var fieldsCmd = &cobra.Command{
	Use:   "fields",
	Short: "列出全部可查字段（本地，无网络）",
	Long: `列出 ipapi field / ipapi me-field 支持的全部字段名，按语义分组。

Agent 可用此命令自发现能查什么，无需读文档。

示例:
  ipapi fields                    # 人类可读分组列表
  ipapi fields --json             # JSON 数组
  ipapi fields --group geo        # 只看 geo 分组`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		groupFilter, _ := cmd.Flags().GetString("group")
		asJSON, _ := cmd.Flags().GetBool("json")

		// 校验 group 过滤值（若有）
		if groupFilter != "" && groupFilter != "all" {
			valid := false
			for _, g := range fieldsIndex {
				if g.Name == groupFilter {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("--group: 未知分组 %q，可选: %s", groupFilter, allGroupNames())
			}
		}

		groups := fieldsIndex
		if groupFilter != "" && groupFilter != "all" {
			groups = filterGroups(fieldsIndex, groupFilter)
		}

		if asJSON {
			return printFieldsJSON(groups)
		}
		printFieldsHuman(groups)
		return nil
	},
}

func init() {
	fieldsCmd.Flags().String("group", "all", "只显示指定分组（identity/geo/time/network/culture/economy/stats）")
	fieldsCmd.Flags().Bool("json", false, "JSON 输出（即便未加也默认 JSON 信封，此旗标用于 fields 不走信封的纯数组形式）")
}

func printFieldsJSON(groups []fieldGroup) error {
	out := map[string]interface{}{
		"groups": groups,
		"all":    ipapi.ValidFields(),
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func printFieldsHuman(groups []fieldGroup) {
	fmt.Println("\n📡 ipapi 支持的可查字段（共 28 个）")
	fmt.Println("──────────────────────────────────────────")
	for _, g := range groups {
		fmt.Printf("\n  ▸ %s\n", g.Name)
		sort.Strings(g.Fields)
		for _, f := range g.Fields {
			fmt.Printf("      %s\n", f)
		}
	}
	fmt.Println("\n──────────────────────────────────────────")
	fmt.Println("用法: ipapi field <ip> <field>   |   ipapi me-field <field>")
	fmt.Println()
}

func filterGroups(all []fieldGroup, name string) []fieldGroup {
	for _, g := range all {
		if g.Name == name {
			return []fieldGroup{g}
		}
	}
	return nil
}

func allGroupNames() string {
	names := make([]string, 0, len(fieldsIndex))
	for _, g := range fieldsIndex {
		names = append(names, g.Name)
	}
	return joinStrings(names, "/")
}

// joinStrings 是 strings.Join 的本地薄包装（避免在 fields 文件再 import strings）。
func joinStrings(ss []string, sep string) string {
	if len(ss) == 0 {
		return ""
	}
	out := ss[0]
	for _, s := range ss[1:] {
		out += sep + s
	}
	return out
}
