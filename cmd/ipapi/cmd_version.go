// Package main: cmd_version.go 实现 `ipapi version`。
// version/commit/date 由 goreleaser ldflags 注入；未注入时显示 dev。
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// 这些变量由 goreleaser 的 ldflags 注入（见 .goreleaser.yaml）。
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func versionString() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "打印版本信息",
	Long:  `打印 ipapi CLI 的版本、commit、构建时间与 Go 运行时版本。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		asJSON, _ := cmd.Flags().GetBool("json")
		if asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(map[string]string{
				"version":   version,
				"commit":    commit,
				"date":      date,
				"goVersion": runtime.Version(),
			})
		}
		fmt.Printf("ipapi %s\nGo %s\n", versionString(), runtime.Version())
		return nil
	},
}

func init() {
	versionCmd.Flags().Bool("json", false, "JSON 输出")
}

// buildCompletionCmd 构造 cobra 内置的 completion 子命令树。
func buildCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [shell]",
		Short: "生成 shell 自动补全脚本",
		Long: `生成 bash/zsh/fish/powershell 的 shell 补全脚本。

  ipapi completion bash > /etc/bash_completion.d/ipapi
  ipapi completion zsh > "${fpath[1]}/_ipapi"
  ipapi completion fish | source`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]
			switch shell {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s (bash|zsh|fish|powershell)", shell)
			}
		},
	}
}
