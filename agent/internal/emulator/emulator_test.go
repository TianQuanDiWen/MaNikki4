package emulator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestStripJSONComments(t *testing.T) {
	input := `{
		// 这是整行单行注释
		"url": "https://example.com/api/v1//test",
		"path": "C:\\Program Files//test\\app.exe", // 行尾注释
		"normal": 123
	}`

	stripped := stripJSONComments([]byte(input))

	var data struct {
		URL    string `json:"url"`
		Path   string `json:"path"`
		Normal int    `json:"normal"`
	}

	if err := json.Unmarshal(stripped, &data); err != nil {
		t.Fatalf("json.Unmarshal failed: %v\nStripped content:\n%s", err, string(stripped))
	}

	if data.URL != "https://example.com/api/v1//test" {
		t.Errorf("expected URL to be preserved, got %q", data.URL)
	}
	if data.Path != "C:\\Program Files//test\\app.exe" {
		t.Errorf("expected Path to be preserved, got %q", data.Path)
	}
	if data.Normal != 123 {
		t.Errorf("expected Normal to be 123, got %d", data.Normal)
	}
}

func TestParseOtherRunningInstances(t *testing.T) {
	// 单实例 JSON（仅 0 号机在运行）
	singleRunning := []byte(`{
		"index": "0",
		"is_process_started": true,
		"pid": 1234
	}`)
	if parseOtherRunningInstances(singleRunning, 0) {
		t.Errorf("expected false for single running instance when querying same index")
	}
	if !parseOtherRunningInstances(singleRunning, 1) {
		t.Errorf("expected true when querying different index from running single instance")
	}

	// 数组形式多实例（0 号和 1 号，其中 1 号在运行）
	multiRunning := []byte(`[
		{"index": "0", "is_process_started": false, "pid": 0},
		{"index": "1", "is_process_started": true, "pid": 5678}
	]`)
	if !parseOtherRunningInstances(multiRunning, 0) {
		t.Errorf("expected true for index 0 when index 1 is running")
	}
	if parseOtherRunningInstances(multiRunning, 1) {
		t.Errorf("expected false for index 1 when only index 1 is running")
	}

	// 数组形式多实例（所有实例均未运行）
	noneRunning := []byte(`[
		{"index": "0", "is_process_started": false, "pid": 0},
		{"index": "1", "is_process_started": false, "pid": 0}
	]`)
	if parseOtherRunningInstances(noneRunning, 0) {
		t.Errorf("expected false when no instances are running")
	}

	// 数值类型 index 兼容性测试（"index": 0 为 number 而非 string）
	numericRunning := []byte(`[
		{"index": 0, "is_process_started": false, "pid": 0},
		{"index": 1, "is_process_started": true, "pid": 5678}
	]`)
	if !parseOtherRunningInstances(numericRunning, 0) {
		t.Errorf("expected true for numeric index 0 when numeric index 1 is running")
	}
	if parseOtherRunningInstances(numericRunning, 1) {
		t.Errorf("expected false for numeric index 1 when only numeric index 1 is running")
	}

	// 异常数据 Fail-Safe 保守测试（损坏 JSON 应默认返回 true 保护多开）
	corrupted := []byte(`invalid json data`)
	if !parseOtherRunningInstances(corrupted, 0) {
		t.Errorf("expected true (Fail-Safe) when parsing corrupted JSON")
	}
}

func TestDetectServer(t *testing.T) {
	// 1. 命令行入参测试
	if res := detectServer(`{"ServerOption":"TW"}`, "官服"); res != "TW" {
		t.Errorf("expected TW from CLI rawJSON, got %s", res)
	}
	if res := detectServer(`{"ServerOption_TW":"Yes"}`, "官服"); res != "TW" {
		t.Errorf("expected TW from ServerOption_TW, got %s", res)
	}

	// 2. 环境变量 PI_RESOURCE 测试
	t.Setenv("PI_RESOURCE", "台服")
	if res := detectServer("", "官服"); res != "TW" {
		t.Errorf("expected TW from PI_RESOURCE env var, got %s", res)
	}
	t.Setenv("PI_RESOURCE", "")

	// 3. 配置 resource 测试
	if res := detectServer("", "台服"); res != "TW" {
		t.Errorf("expected TW from cfgResource, got %s", res)
	}
	if res := detectServer("", "官服"); res != "CN" {
		t.Errorf("expected CN from cfgResource, got %s", res)
	}
}

func TestResolveMXUResource(t *testing.T) {
	tempDir := t.TempDir()
	cfgDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	mxuContent := `{
		"instances": [
			{
				"id": "inst_1",
				"resourceName": "台服"
			}
		],
		"lastActiveInstanceId": "inst_1"
	}`
	if err := os.WriteFile(filepath.Join(cfgDir, "mxu-MaNikki4.json"), []byte(mxuContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if res := resolveMXUResource(tempDir); res != "台服" {
		t.Errorf("expected 台服 from mxu-MaNikki4.json, got %s", res)
	}
}
