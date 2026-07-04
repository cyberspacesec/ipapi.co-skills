// Package main implements the ipapi command-line tool.
//
// config.go 负责把"命令行旗标 > 环境变量 > 配置文件 > 默认值"四层来源
// 合并成一份最终配置，供 newClient 使用。配置文件用 JSON（stdlib 即可解析，
// 不引入 yaml/viper 等额外依赖）。
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
)

// Config 是 CLI 运行时所需的全部可配置项。每个字段都对应一个全局旗标、
// 一个环境变量或一个配置文件键。
type Config struct {
	APIKey      string        `json:"api_key"`
	APIKeyMode  string        `json:"api_key_mode"` // "header" | "query"
	Format      string        `json:"format"`
	BaseURL     string        `json:"base_url"`
	UserAgent   string        `json:"user_agent"`
	Retries     int           `json:"retries"`
	Timeout     time.Duration `json:"timeout"`
	Human       bool          `json:"-"` // 仅旗标，不持久化
	ConfigPath  string        `json:"-"` // 仅旗标，不持久化
	Callback    string        `json:"callback,omitempty"`
}

// defaultConfig 返回内置默认值。所有后续来源都以此为基础覆盖。
func defaultConfig() Config {
	return Config{
		APIKey:     "",
		APIKeyMode: "header",
		Format:     string(ipapi.FormatJSON),
		BaseURL:    "https://ipapi.co/",
		UserAgent:  "ipapi-cli/0.1.0",
		Retries:    2,
		Timeout:    10 * time.Second,
	}
}

// envKey 是每个配置字段对应的环境变量名。
var envKeys = map[string]string{
	"APIKey":     "IPAPI_API_KEY",
	"APIKeyMode": "IPAPI_API_KEY_MODE",
	"Format":     "IPAPI_FORMAT",
	"BaseURL":    "IPAPI_BASE_URL",
	"UserAgent":  "IPAPI_USER_AGENT",
	"Retries":    "IPAPI_RETRIES",
	"Timeout":    "IPAPI_TIMEOUT",
}

// loadConfigFile 从 path 读取 JSON 配置，覆盖到 cfg 上。
// 文件不存在视为"无配置文件"，静默跳过（不报错）。
// 文件存在但解析失败则返回错误（用户主动配置错了应该提示）。
func loadConfigFile(cfg *Config, path string) error {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 配置文件可选
		}
		return fmt.Errorf("read config %s: %w", path, err)
	}
	// 用临时结构解析 duration 字段：JSON 里 timeout 是字符串（如 "10s"）。
	var raw struct {
		APIKey     string `json:"api_key"`
		APIKeyMode string `json:"api_key_mode"`
		Format     string `json:"format"`
		BaseURL    string `json:"base_url"`
		UserAgent  string `json:"user_agent"`
		Retries    int    `json:"retries"`
		Timeout    string `json:"timeout"`
		Callback   string `json:"callback"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse config %s: %w", path, err)
	}
	if raw.APIKey != "" {
		cfg.APIKey = raw.APIKey
	}
	if raw.APIKeyMode != "" {
		cfg.APIKeyMode = raw.APIKeyMode
	}
	if raw.Format != "" {
		cfg.Format = raw.Format
	}
	if raw.BaseURL != "" {
		cfg.BaseURL = raw.BaseURL
	}
	if raw.UserAgent != "" {
		cfg.UserAgent = raw.UserAgent
	}
	if raw.Retries > 0 {
		cfg.Retries = raw.Retries
	}
	if raw.Timeout != "" {
		d, err := time.ParseDuration(raw.Timeout)
		if err != nil {
			return fmt.Errorf("parse timeout in %s: %w", path, err)
		}
		cfg.Timeout = d
	}
	if raw.Callback != "" {
		cfg.Callback = raw.Callback
	}
	return nil
}

// applyEnv 把环境变量覆盖到 cfg 上（仅当对应 env 存在时）。
func applyEnv(cfg *Config) {
	if v, ok := os.LookupEnv(envKeys["APIKey"]); ok && v != "" {
		cfg.APIKey = v
	}
	if v, ok := os.LookupEnv(envKeys["APIKeyMode"]); ok && v != "" {
		cfg.APIKeyMode = v
	}
	if v, ok := os.LookupEnv(envKeys["Format"]); ok && v != "" {
		cfg.Format = v
	}
	if v, ok := os.LookupEnv(envKeys["BaseURL"]); ok && v != "" {
		cfg.BaseURL = v
	}
	if v, ok := os.LookupEnv(envKeys["UserAgent"]); ok && v != "" {
		cfg.UserAgent = v
	}
	if v, ok := os.LookupEnv(envKeys["Retries"]); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			cfg.Retries = n
		}
	}
	if v, ok := os.LookupEnv(envKeys["Timeout"]); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Timeout = d
		}
	}
}

// validate 校验合并后的配置合法性。返回 sentinel-friendly 错误。
func (cfg *Config) validate() error {
	if err := ipapi.ValidateFormat(cfg.Format); err != nil {
		return fmt.Errorf("--format: %w", err)
	}
	switch strings.ToLower(cfg.APIKeyMode) {
	case "header", "query":
	default:
		return fmt.Errorf("--api-key-mode: must be 'header' or 'query', got %q", cfg.APIKeyMode)
	}
	if cfg.Retries < 0 {
		return fmt.Errorf("--retries: must be >= 0, got %d", cfg.Retries)
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("--timeout: must be > 0, got %s", cfg.Timeout)
	}
	return nil
}

// resolveConfigPath 决定配置文件路径：--config 显式 > $IPAPI_CONFIG > ./.ipapi.json > ~/.ipapi.json。
// 返回第一个存在的路径（用于读取），或空串。
func resolveConfigPath(explicit string) string {
	candidates := []string{}
	if explicit != "" {
		candidates = append(candidates, explicit)
	}
	if v := os.Getenv("IPAPI_CONFIG"); v != "" {
		candidates = append(candidates, v)
	}
	candidates = append(candidates, ".ipapi.json")
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, home+"/.ipapi.json")
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
