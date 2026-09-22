package matcher

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
)

// MatchRowButtonParam 是 MatchRowButton 自定义识别器的配置参数。
type MatchRowButtonParam struct {
	Keyword    string `json:"keyword"`     // 匹配左侧文本的关键字或正则表达式，支持使用 | 分隔多个候选词
	Button     string `json:"button"`      // 匹配右侧按钮的文本或正则表达式，如 "前往"、"购买" 或 "前往|购买"
	YTolerance int    `json:"y_tolerance"` // 判定为同一行的垂直中心高度最大允许距离（像素），默认 60
}

// DefaultYTolerance 默认同一行判定高度容差。
const DefaultYTolerance = 60

// NewMatchRowButtonRunner 创建通用的“行按钮绑定”自定义识别器。
// 职责：在给定 ROI 区域内通过 OCR 检索指定 keyword，并在同一水平行（垂直高度容差内）寻找对应的 button，返回按钮真实坐标。
func NewMatchRowButtonRunner() maa.CustomRecognitionRunner {
	return maa.CustomRecognitionFunc(func(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
		if arg == nil || arg.Img == nil {
			return nil, false
		}

		// 解析传入的自定义参数
		param := MatchRowButtonParam{
			YTolerance: DefaultYTolerance,
		}
		if arg.CustomRecognitionParam != "" && arg.CustomRecognitionParam != "null" {
			raw := strings.TrimSpace(arg.CustomRecognitionParam)
			if strings.HasPrefix(raw, "\"") && strings.HasSuffix(raw, "\"") {
				var unquoted string
				if err := json.Unmarshal([]byte(raw), &unquoted); err == nil {
					raw = strings.TrimSpace(unquoted)
				}
			}
			if err := json.Unmarshal([]byte(raw), &param); err != nil {
				fmt.Printf("[MatchRowButton] 解析参数失败: %v, param=%s\n", err, arg.CustomRecognitionParam)
				return nil, false
			}
		}

		if param.Keyword == "__DISABLED__" {
			return nil, false
		}
		if param.Keyword == "" || param.Button == "" {
			fmt.Printf("[MatchRowButton] 错误: keyword 或 button 为空 (keyword=%q, button=%q)\n", param.Keyword, param.Button)
			return nil, false
		}
		if param.YTolerance <= 0 {
			param.YTolerance = DefaultYTolerance
		}

		kwRegex, err := regexp.Compile(param.Keyword)
		if err != nil {
			fmt.Printf("[MatchRowButton] 编译 keyword 正则失败: %v\n", err)
			return nil, false
		}

		btnRegex, err := regexp.Compile(param.Button)
		if err != nil {
			fmt.Printf("[MatchRowButton] 编译 button 正则失败: %v\n", err)
			return nil, false
		}

		// 确定 OCR 搜索区域（优先使用节点传入的 ROI）
		roi := arg.Roi
		if roi[2] <= 0 || roi[3] <= 0 {
			b := arg.Img.Bounds()
			roi = maa.Rect{0, 0, b.Dx(), b.Dy()}
		}

		// 调用 MaaFramework 原生 OCR 识别
		reco, err := ctx.RunRecognitionDirect(
			maa.RecognitionTypeOCR,
			&maa.OCRParam{ROI: maa.NewTargetRect(roi)},
			arg.Img,
		)
		if err != nil || reco == nil || !reco.Hit || reco.Results == nil || len(reco.Results.All) == 0 {
			return nil, false
		}

		var kwItems []TextItem
		var btnItems []TextItem

		for _, res := range reco.Results.All {
			ocrRes, ok := res.AsOCR()
			if !ok || ocrRes == nil || ocrRes.Text == "" {
				continue
			}

			cleanText := strings.ReplaceAll(ocrRes.Text, " ", "")
			centerY := ocrRes.Box[1] + ocrRes.Box[3]/2
			item := TextItem{
				Text:    ocrRes.Text,
				Box:     ocrRes.Box,
				CenterY: centerY,
			}

			if kwRegex.MatchString(ocrRes.Text) || kwRegex.MatchString(cleanText) {
				kwItems = append(kwItems, item)
			}
			if btnRegex.MatchString(ocrRes.Text) || btnRegex.MatchString(cleanText) {
				btnItems = append(btnItems, item)
			}
		}

		if len(kwItems) == 0 || len(btnItems) == 0 {
			return nil, false
		}

		box, kwText, btnText, yDiff, ok := FindRowButtonMatch(kwItems, btnItems, param.YTolerance)
		if ok {
			fmt.Printf("[MatchRowButton] 成功命中同行按钮 -> keyword: %q, button: %q (box=%v, yDiff=%d)\n",
				kwText, btnText, box, yDiff)
			return &maa.CustomRecognitionResult{
				Box:    box,
				Detail: fmt.Sprintf(`{"keyword":%q,"button":%q,"y_diff":%d}`, kwText, btnText, yDiff),
			}, true
		}

		return nil, false
	})
}

// TextItem 表示一个 OCR 识别出的文本项及其坐标信息。
type TextItem struct {
	Text    string
	Box     maa.Rect
	CenterY int
}

// FindRowButtonMatch 在关键字候选集与按钮候选集中寻找同行的最佳按钮：
// 1. 按钮位于关键字右侧（X_btn > X_kw）；
// 2. 垂直中心距离在 yTolerance 范围内。
func FindRowButtonMatch(kwItems, btnItems []TextItem, yTolerance int) (matchedBox maa.Rect, kwText, btnText string, yDiff int, ok bool) {
	for _, kw := range kwItems {
		for _, btn := range btnItems {
			if btn.Box[0] <= kw.Box[0] {
				continue
			}

			diff := btn.CenterY - kw.CenterY
			if diff < 0 {
				diff = -diff
			}
			if diff <= yTolerance {
				return btn.Box, kw.Text, btn.Text, diff, true
			}
		}
	}
	return maa.Rect{}, "", "", 0, false
}

