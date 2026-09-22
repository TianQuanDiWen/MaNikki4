package matcher

import (
	"regexp"
	"strings"
	"testing"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
)

func TestFindRowButtonMatch_Basic(t *testing.T) {
	// 模拟 screenshot:
	// Row 1: "时空回廊《雪鹰少校》" at x=230, y=360, center=370; Button "前往" at x=520, y=340, center=370
	// Row 2: "主线：6-10" at x=230, y=570, center=580; Button "前往" at x=520, y=550, center=580
	kwItems := []TextItem{
		{Text: "时空回廊《雪鹰少校》获取", Box: maa.Rect{230, 360, 200, 30}, CenterY: 375},
	}
	btnItems := []TextItem{
		{Text: "前往", Box: maa.Rect{520, 340, 130, 60}, CenterY: 370},
		{Text: "前往", Box: maa.Rect{520, 550, 130, 60}, CenterY: 580},
	}

	box, kw, btn, diff, ok := FindRowButtonMatch(kwItems, btnItems, 60)
	if !ok {
		t.Fatalf("expected to match row button, but got false")
	}
	if box[1] != 340 {
		t.Errorf("expected box y=340 (row 1), got %d", box[1])
	}
	if kw != "时空回廊《雪鹰少校》获取" || btn != "前往" {
		t.Errorf("unexpected matched texts: kw=%q, btn=%q", kw, btn)
	}
	if diff != 5 {
		t.Errorf("expected diff=5, got %d", diff)
	}
}

func TestFindRowButtonMatch_Tolerance(t *testing.T) {
	kwItems := []TextItem{
		{Text: "时空回廊", Box: maa.Rect{230, 360, 100, 30}, CenterY: 375},
	}
	// 只有第 2 行的按钮（y=580），垂直距离差 205 像素
	btnItems := []TextItem{
		{Text: "前往", Box: maa.Rect{520, 550, 130, 60}, CenterY: 580},
	}

	_, _, _, _, ok := FindRowButtonMatch(kwItems, btnItems, 60)
	if ok {
		t.Errorf("expected false when y difference exceeds tolerance, got true")
	}
}

func TestFindRowButtonMatch_ButtonLeftOfKeyword(t *testing.T) {
	// 如果按钮在关键字左侧，应忽略
	kwItems := []TextItem{
		{Text: "时空回廊", Box: maa.Rect{520, 360, 100, 30}, CenterY: 375},
	}
	btnItems := []TextItem{
		{Text: "前往", Box: maa.Rect{230, 340, 130, 60}, CenterY: 370},
	}

	_, _, _, _, ok := FindRowButtonMatch(kwItems, btnItems, 60)
	if ok {
		t.Errorf("expected false when button is to the left of keyword, got true")
	}
}

func TestRegexMultiKeyword_PipeSeparator(t *testing.T) {
	pattern := "时空回廊|雪鹰少校"
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("failed to compile regex: %v", err)
	}

	tests := []struct {
		input string
		want  bool
	}{
		{"时空回廊", true},
		{"雪鹰少校", true},
		{"时空回廊《雪鹰少校》获取", true},
		{"主线：6-10", false},
		{"服装店：购买", false},
	}

	for _, tt := range tests {
		if got := re.MatchString(tt.input); got != tt.want {
			t.Errorf("MatchString(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestRegexSpaceInsensitive(t *testing.T) {
	pattern := "时空回廊"
	re := regexp.MustCompile(pattern)

	ocrText := "时 空 回 廊"
	cleanText := strings.ReplaceAll(ocrText, " ", "")

	if !re.MatchString(cleanText) {
		t.Errorf("expected cleanText %q to match %q", cleanText, pattern)
	}
}
