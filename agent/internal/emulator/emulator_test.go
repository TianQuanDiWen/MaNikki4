package emulator

import (
	"encoding/json"
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
