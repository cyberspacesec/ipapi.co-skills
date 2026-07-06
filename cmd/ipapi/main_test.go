// Package main 的单元测试。覆盖：输出信封、退出码映射、配置合并、fields 枚举。
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
)

// ---- exitcode 测试 ----

func TestClassifyError(t *testing.T) {
	cases := []struct {
		err      error
		wantCode int
		wantName string
		wantRetry bool
	}{
		{ipapi.ErrInvalidIP, exitInvalidIP, "INVALID_IP", false},
		{ipapi.ErrInvalidField, exitInvalidField, "INVALID_FIELD", false},
		{ipapi.ErrInvalidFormat, exitInvalidFormat, "INVALID_FORMAT", false},
		{ipapi.ErrRateLimited, exitRateLimited, "RATE_LIMITED", true},
		{ipapi.ErrReservedIP, exitReservedIP, "RESERVED_IP", false},
		{ipapi.ErrNotFound, exitNotFound, "NOT_FOUND", true},
		{ipapi.ErrServerError, exitServerError, "SERVER_ERROR", true},
		{ipapi.ErrMethodNotAllowed, exitMethodNotAllowed, "METHOD_NOT_ALLOWED", false},
		{ipapi.ErrInvalidKey, exitInvalidKey, "INVALID_KEY", false},
		{ipapi.ErrUnexpectedData, exitUnexpectedData, "UNEXPECTED_DATA", false},
		{errors.New("something else"), exitInternal, "INTERNAL", false},
		{nil, exitOK, "", false},
	}
	for _, c := range cases {
		code, name, _, retry := classifyError(c.err)
		if code != c.wantCode {
			t.Errorf("err=%v: code=%d want %d", c.err, code, c.wantCode)
		}
		if name != c.wantName {
			t.Errorf("err=%v: name=%q want %q", c.err, name, c.wantName)
		}
		if retry != c.wantRetry {
			t.Errorf("err=%v: retry=%v want %v", c.err, retry, c.wantRetry)
		}
	}
}

func TestErrToExitCodeExitError(t *testing.T) {
	ee := &exitError{code: exitInvalidIP}
	if got := errToExitCode(ee); got != exitInvalidIP {
		t.Errorf("exitError code=%d want %d", got, exitInvalidIP)
	}
}

// ---- output 信封测试 ----

func TestRenderOKJSON(t *testing.T) {
	// 捕获 stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()

	info := &ipapi.IPInfo{IP: "8.8.8.8", City: "Mountain View", CountryName: "United States"}
	m := &meta{Format: "json", DurationMs: 100, RetrievedAt: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)}
	renderOK("info", map[string]string{"ip": "8.8.8.8"}, info, m, false)
	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	var env envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if !env.OK {
		t.Error("OK should be true")
	}
	if env.Command != "info" {
		t.Errorf("Command=%q want info", env.Command)
	}
	if env.Meta.DurationMs != 100 {
		t.Errorf("DurationMs=%d want 100", env.Meta.DurationMs)
	}
}

func TestRenderErrorToStderr(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = old }()

	err := renderError("info", map[string]string{"ip": "999.1.1.1"}, ipapi.ErrInvalidIP)
	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	var env envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if env.OK {
		t.Error("OK should be false")
	}
	if env.Error.Code != "INVALID_IP" {
		t.Errorf("Error.Code=%q want INVALID_IP", env.Error.Code)
	}
	if env.Error.Sentinel != "ErrInvalidIP" {
		t.Errorf("Error.Sentinel=%q want ErrInvalidIP", env.Error.Sentinel)
	}
	// renderError 应返回 exitError 携带正确退出码
	ee, ok := err.(*exitError)
	if !ok {
		t.Fatalf("renderError should return *exitError, got %T", err)
	}
	if ee.code != exitInvalidIP {
		t.Errorf("exitError.code=%d want %d", ee.code, exitInvalidIP)
	}
}

func TestPrintIPInfoHuman(t *testing.T) {
	var buf bytes.Buffer
	info := &ipapi.IPInfo{IP: "8.8.8.8", City: "X", CountryName: "Y", CountryCode: "US"}
	m := &meta{Format: "json", DurationMs: 5, RetrievedAt: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)}
	printIPInfoHuman(&buf, info, m)
	out := buf.String()
	if !strings.Contains(out, "8.8.8.8") {
		t.Errorf("human output missing IP: %s", out)
	}
	if !strings.Contains(out, "City") {
		t.Errorf("human output missing City: %s", out)
	}
}

// ---- config 测试 ----

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		wantErr bool
	}{
		{"valid", defaultConfig(), false},
		{"bad format", func() Config { c := defaultConfig(); c.Format = "xml"; return c }(), false},
		{"invalid format", func() Config { c := defaultConfig(); c.Format = "toml"; return c }(), true},
		{"bad keymode", func() Config { c := defaultConfig(); c.APIKeyMode = "weird"; return c }(), true},
		{"query mode", func() Config { c := defaultConfig(); c.APIKeyMode = "query"; return c }(), false},
		{"neg retries", func() Config { c := defaultConfig(); c.Retries = -1; return c }(), true},
		{"zero timeout", func() Config { c := defaultConfig(); c.Timeout = 0; return c }(), true},
	}
	for _, c := range cases {
		err := c.cfg.validate()
		if (err != nil) != c.wantErr {
			t.Errorf("%s: err=%v wantErr=%v", c.name, err, c.wantErr)
		}
	}
}

func TestLoadConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ipapi.json")
	content := `{"api_key":"test123","format":"xml","retries":5,"timeout":"30s","api_key_mode":"query"}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := defaultConfig()
	if err := loadConfigFile(&cfg, path); err != nil {
		t.Fatalf("loadConfigFile: %v", err)
	}
	if cfg.APIKey != "test123" {
		t.Errorf("APIKey=%q want test123", cfg.APIKey)
	}
	if cfg.Format != "xml" {
		t.Errorf("Format=%q want xml", cfg.Format)
	}
	if cfg.Retries != 5 {
		t.Errorf("Retries=%d want 5", cfg.Retries)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout=%s want 30s", cfg.Timeout)
	}
	if cfg.APIKeyMode != "query" {
		t.Errorf("APIKeyMode=%q want query", cfg.APIKeyMode)
	}
}

func TestLoadConfigFileNotExist(t *testing.T) {
	cfg := defaultConfig()
	before := cfg
	err := loadConfigFile(&cfg, "/nonexistent/path/.ipapi.json")
	if err != nil {
		t.Errorf("missing file should not error: %v", err)
	}
	if cfg != before {
		t.Error("missing file should not modify cfg")
	}
}

func TestApplyEnv(t *testing.T) {
	t.Setenv("IPAPI_API_KEY", "envkey")
	t.Setenv("IPAPI_FORMAT", "csv")
	t.Setenv("IPAPI_RETRIES", "7")
	t.Setenv("IPAPI_TIMEOUT", "15s")
	t.Setenv("IPAPI_API_KEY_MODE", "query")

	cfg := defaultConfig()
	applyEnv(&cfg)
	if cfg.APIKey != "envkey" {
		t.Errorf("APIKey=%q want envkey", cfg.APIKey)
	}
	if cfg.Format != "csv" {
		t.Errorf("Format=%q want csv", cfg.Format)
	}
	if cfg.Retries != 7 {
		t.Errorf("Retries=%d want 7", cfg.Retries)
	}
	if cfg.Timeout != 15*time.Second {
		t.Errorf("Timeout=%s want 15s", cfg.Timeout)
	}
	if cfg.APIKeyMode != "query" {
		t.Errorf("APIKeyMode=%q want query", cfg.APIKeyMode)
	}
}

func TestNewClientMapsFields(t *testing.T) {
	cfg := Config{
		APIKey:     "k",
		APIKeyMode: "query",
		BaseURL:    "http://example.com/",
		UserAgent:  "test-agent",
		Retries:    3,
		Timeout:    5 * time.Second,
		Format:     "json",
		Callback:   "cb",
	}
	c := newClient(&cfg)
	if c.APIKey != "k" {
		t.Errorf("APIKey=%q", c.APIKey)
	}
	if c.APIKeyMode != ipapi.APIKeyQuery {
		t.Errorf("APIKeyMode=%v want query", c.APIKeyMode)
	}
	if c.BaseURL != "http://example.com/" {
		t.Errorf("BaseURL=%q", c.BaseURL)
	}
	if c.UserAgent != "test-agent" {
		t.Errorf("UserAgent=%q", c.UserAgent)
	}
	if c.Retries != 3 {
		t.Errorf("Retries=%d", c.Retries)
	}
	if c.Callback != "cb" {
		t.Errorf("Callback=%q", c.Callback)
	}
	if c.HTTPClient.Timeout != 5*time.Second {
		t.Errorf("HTTPClient.Timeout=%s", c.HTTPClient.Timeout)
	}
}

// ---- fields 测试 ----

func TestFieldsIndexHasAllFields(t *testing.T) {
	// fieldsIndex 里列出的字段总数应等于 SDK ValidFields 的数量
	collected := map[string]struct{}{}
	for _, g := range fieldsIndex {
		for _, f := range g.Fields {
			collected[f] = struct{}{}
		}
	}
	valid := ipapi.ValidFields()
	if len(collected) != len(valid) {
		t.Errorf("fieldsIndex covers %d unique fields, SDK has %d — 漏字段", len(collected), len(valid))
	}
	// 确认 fieldsIndex 的字段都在 SDK validFields 里
	for f := range collected {
		if !ipapi.IsValidField(f) {
			t.Errorf("fieldsIndex 含非法字段 %q", f)
		}
	}
}

// ---- 端到端 mock 测试 ----

// newMockServer 启动一个模拟 ipapi.co 的 httptest server。
// 返回的 JSON 用 ipapi.co 的真实字段命名。
func newMockServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func TestInfoEndToEndMock(t *testing.T) {
	body := `{"ip":"8.8.8.8","city":"Mountain View","country_name":"United States","asn":"AS15169"}`
	srv := newMockServer(t, 200, body)
	defer srv.Close()

	cfg := defaultConfig()
	cfg.BaseURL = srv.URL + "/"
	cfg.Format = "json"
	client := newClient(&cfg)
	state.cfg = &cfg
	state.client = client

	// 捕获 stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	argsMap := map[string]string{"ip": "8.8.8.8", "format": "json"}
	err := runIPInfo(nil, "info", argsMap, func(ctx context.Context, format string) (*ipapi.IPInfo, error) {
		return state.client.GetIPInfo(ctx, "8.8.8.8", format)
	})
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("runIPInfo: %v", err)
	}
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	var env envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if !env.OK {
		t.Error("OK should be true")
	}
}

func TestQuotaEndToEndMock(t *testing.T) {
	srv := newMockServer(t, 200, `{"available":"12345"}`)
	defer srv.Close()

	cfg := defaultConfig()
	cfg.BaseURL = srv.URL + "/"
	cfg.Format = "json"
	cfg.APIKey = "test-key"
	client := newClient(&cfg)
	state.cfg = &cfg
	state.client = client

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runQuota(nil, nil)
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("runQuota: %v", err)
	}
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	var env envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if !env.OK {
		t.Fatal("OK should be true")
	}
	if env.Command != "quota" {
		t.Errorf("expected command=quota, got %q", env.Command)
	}
	// data should be the Quota struct marshalled: {"available":"12345"}
	raw, _ := json.Marshal(env.Data)
	if !bytes.Contains(raw, []byte("12345")) {
		t.Errorf("expected data to contain 12345, got %s", raw)
	}
}

func TestQuotaEndToEndInvalidKey(t *testing.T) {
	srv := newMockServer(t, 200, `{"error":true,"reason":"Invalid Key","message":"Invalid key. SignUp @ https://ipapi.co/pricing/ "}`)
	defer srv.Close()

	cfg := defaultConfig()
	cfg.BaseURL = srv.URL + "/"
	cfg.Format = "json"
	cfg.APIKey = "bad"
	client := newClient(&cfg)
	state.cfg = &cfg
	state.client = client

	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	err := runQuota(nil, nil)
	w.Close()
	os.Stderr = old

	if err == nil {
		t.Fatal("expected error for invalid key")
	}
	ee, ok := err.(*exitError)
	if !ok {
		t.Fatalf("expected *exitError, got %T", err)
	}
	if ee.code != exitInvalidKey {
		t.Errorf("expected exit code %d (INVALID_KEY), got %d", exitInvalidKey, ee.code)
	}
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	var env envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("invalid JSON on stderr: %v\n%s", err, buf.String())
	}
	if env.OK {
		t.Fatal("OK should be false for invalid key")
	}
	if env.Error == nil || env.Error.Code != "INVALID_KEY" {
		t.Errorf("expected error.code=INVALID_KEY, got %+v", env.Error)
	}
}

func TestPrintQuotaHuman(t *testing.T) {
	var buf bytes.Buffer
	printQuotaHuman(&buf, &ipapi.Quota{Available: "42"})
	out := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("42")) {
		t.Errorf("expected human output to contain 42, got %q", out)
	}
	if !bytes.Contains(buf.Bytes(), []byte("配额")) {
		t.Errorf("expected human output to contain 配额, got %q", out)
	}
}
