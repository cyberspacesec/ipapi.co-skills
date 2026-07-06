// Package main: root.go 定义 CLI 的根命令与全局旗标。
//
// rootCmd 不实现 Run（只打印 help），所有逻辑在子命令。
// 全局旗标用 PersistentFlags，子命令本地旗标用 Flags。
// PersistentPreRunE 负责合并配置并构造 *ipapi.Client，注入到 rootState 供子命令取用。
package main

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
)

// rootState 是 PreRunE 构造好、供子命令共享的运行时状态。
type rootState struct {
	cfg    *Config
	client *ipapi.Client
}

var state = &rootState{}

// rootCmd 是 CLI 入口。
var rootCmd = &cobra.Command{
	Use:   "ipapi",
	Short: "IP 地理位置查询命令行工具",
	Long: `ipapi 是基于 ipapi.co 的命令行 IP 地理位置查询工具。

默认输出 JSON 信封（适合脚本与 AI Agent 解析），加 --human/-H 切换为人类可读。

快速开始:
  ipapi 8.8.8.8                    # 查询指定 IP（JSON）
  ipapi 8.8.8.8 -H                 # 人类可读
  ipapi field 8.8.8.8 country      # 只取一个字段
  ipapi me                          # 查本机公网 IP
  ipapi fields                      # 列出全部可查字段

环境变量: IPAPI_API_KEY, IPAPI_FORMAT, IPAPI_BASE_URL, ...
配置文件: ~/.ipapi.json

完整文档: https://cyberspacesec.github.io/ipapi.co-skills/cli/`,
	Version: versionString(),
	// 不设 Run，让 cobra 默认打印 help
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 跳过 help/completion/version 等不需要 client 的命令
		if skipPreRun(cmd) {
			return nil
		}
		cfg := gatherConfig(cmd)
		if err := cfg.validate(); err != nil {
			return err
		}
		state.cfg = &cfg
		state.client = newClient(&cfg)
		return nil
	},
	SilenceUsage:  true, // 出错时不重复打印 usage（错误信封已足够）
	SilenceErrors: true, // 错误由我们自己渲染到 stderr（JSON 信封），不要 cobra 再打印 "Error: ..."
}

// skipPreRun 判断是否跳过 PreRun 的配置加载（help/completion/version 不需要 client）。
func skipPreRun(cmd *cobra.Command) bool {
	switch cmd.Name() {
	case "help", "completion", "version", "fields":
		return true
	}
	return false
}

// gatherConfig 从 cobra 旗标 + env + 文件 + 默认值 合并出最终 Config。
func gatherConfig(cmd *cobra.Command) Config {
	cfg := defaultConfig()

	// 1. 配置文件（最低优先级，先于 env 与旗标）
	cfg.ConfigPath = resolveConfigPath(getString(cmd, "config"))
	_ = loadConfigFile(&cfg, cfg.ConfigPath)

	// 2. 环境变量
	applyEnv(&cfg)

	// 3. 命令行旗标（最高优先级，仅当用户显式设置时覆盖）
	applyFlags(cmd, &cfg)

	// human 旗标单独处理
	cfg.Human, _ = cmd.Flags().GetBool("human")
	return cfg
}

// applyFlags 把显式设置的旗标覆盖到 cfg。
// 用 Changed 判断，避免默认值覆盖 env。
func applyFlags(cmd *cobra.Command, cfg *Config) {
	if cmd.Flags().Changed("api-key") {
		cfg.APIKey = getString(cmd, "api-key")
	}
	if cmd.Flags().Changed("api-key-mode") {
		cfg.APIKeyMode = getString(cmd, "api-key-mode")
	}
	if cmd.Flags().Changed("format") {
		cfg.Format = getString(cmd, "format")
	}
	if cmd.Flags().Changed("base-url") {
		cfg.BaseURL = getString(cmd, "base-url")
	}
	if cmd.Flags().Changed("user-agent") {
		cfg.UserAgent = getString(cmd, "user-agent")
	}
	if cmd.Flags().Changed("retries") {
		cfg.Retries = getInt(cmd, "retries")
	}
	if cmd.Flags().Changed("timeout") {
		cfg.Timeout = getDuration(cmd, "timeout")
	}
	if cmd.Flags().Changed("callback") {
		cfg.Callback = getString(cmd, "callback")
	}
}

// getString/getInt/getDuration 是 cobra flag 读取的薄包装，出错时返回零值。
func getString(cmd *cobra.Command, name string) string {
	v, _ := cmd.Flags().GetString(name)
	return v
}
func getInt(cmd *cobra.Command, name string) int {
	v, _ := cmd.Flags().GetInt(name)
	return v
}
func getDuration(cmd *cobra.Command, name string) time.Duration {
	v, _ := cmd.Flags().GetDuration(name)
	return v
}

// init 注册全局旗标与子命令。
func init() {
	p := rootCmd.PersistentFlags()
	p.String("api-key", "", "API 密钥（env IPAPI_API_KEY）")
	p.String("api-key-mode", "header", "认证方式：header（Bearer，默认）或 query（?key=）")
	p.StringP("format", "f", "json", "响应格式：json|jsonp|xml|csv|yaml（info/me 仅支持 json）")
	p.String("base-url", "https://ipapi.co/", "API 基础 URL（env IPAPI_BASE_URL）")
	p.String("user-agent", "ipapi-cli/0.1.0", "自定义 User-Agent")
	p.Int("retries", 2, "网络错误/5xx 重试次数（总请求次数 = retries+1）")
	p.Duration("timeout", 10*time.Second, "单次请求超时")
	p.BoolP("human", "H", false, "人类可读输出（默认 JSON 信封）")
	p.String("config", "", "配置文件路径（默认 ~/.ipapi.json）")
	p.String("callback", "", "JSONP 回调函数名（仅 --format jsonp + raw/me-raw 生效）")

	rootCmd.AddCommand(infoCmd, meCmd, fieldCmd, meFieldCmd, rawCmd, meRawCmd, fieldsCmd, quotaCmd, versionCmd)
	completionCmd := buildCompletionCmd()
	rootCmd.AddCommand(completionCmd)
}

// Execute 是 main 调用的入口。
func Execute() error {
	return rootCmd.Execute()
}
