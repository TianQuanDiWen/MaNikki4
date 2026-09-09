package arena

import (
	"fmt"
	"image"
	"regexp"
	"strconv"
	"strings"
	"sync"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
)

var digitRegex = regexp.MustCompile(`\d+`)

// DefaultSelfPowerROI 默认自身战力区域（720x1280）
var DefaultSelfPowerROI = maa.Rect{230, 670, 220, 35}

// PowerChecker 负责暂存自身战力并执行与对手的战力比对。
// 自身战力和对手战力的区域均由各自节点的 ROI 动态配置传入，Go 侧不硬编码。
type PowerChecker struct {
	mu        sync.RWMutex
	selfPower int
}

func NewPowerChecker() *PowerChecker {
	return &PowerChecker{}
}

// RecordSelfPower 从传入的自身战力 ROI 中识别数字并暂存。
func (c *PowerChecker) RecordSelfPower(ctx *maa.Context, img image.Image, roi maa.Rect) (int, bool) {
	if roi[2] <= 0 || roi[3] <= 0 {
		roi = DefaultSelfPowerROI
	}

	power := c.ocrNumber(ctx, img, roi)
	if power <= 0 {
		fmt.Printf("[Arena] 识别自身战力失败或战力为0 (ROI: %v)\n", roi)
		return 0, false
	}

	c.mu.Lock()
	c.selfPower = power
	c.mu.Unlock()

	fmt.Printf("[Arena] 成功记录自身战力 -> %d (ROI: %v)\n", power, roi)
	return power, true
}

// GetSelfPower 获取已暂存的自身战力。
func (c *PowerChecker) GetSelfPower() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.selfPower
}

// CheckCanBeat 检查已暂存的自身战力是否高于指定 ROI 区域内的对手战力。
func (c *PowerChecker) CheckCanBeat(ctx *maa.Context, img image.Image, oppROI maa.Rect) bool {
	if oppROI[2] <= 0 || oppROI[3] <= 0 {
		fmt.Printf("[Arena] 警告: 传入的对手战力 ROI 无效: %v\n", oppROI)
		return false
	}

	self := c.GetSelfPower()
	if self <= 0 {
		fmt.Println("[Arena] 警告: 自身战力尚未记录(为0)，无法进行对战判定")
		return false
	}

	// 动态 OCR 当前对手节点指定的战力区域
	oppPower := c.ocrNumber(ctx, img, oppROI)
	canBeat := self > oppPower && oppPower > 0

	fmt.Printf("[Arena] 节点对比 -> 对手战力: %d (ROI: %v) | 暂存自身战力: %d | 可挑战: %v\n",
		oppPower, oppROI, self, canBeat)
	return canBeat
}

// ocrNumber 在指定 ROI 运行 OCR 并提取连续纯数字。
func (c *PowerChecker) ocrNumber(ctx *maa.Context, img image.Image, roi maa.Rect) int {
	reco, err := ctx.RunRecognitionDirect(
		maa.RecognitionTypeOCR,
		&maa.OCRParam{ROI: maa.NewTargetRect(roi)},
		img,
	)
	if err != nil || reco == nil || !reco.Hit || reco.Results == nil || reco.Results.Best == nil {
		return 0
	}

	ocrRes, ok := reco.Results.Best.AsOCR()
	if !ok || ocrRes == nil {
		return 0
	}

	cleanText := strings.ReplaceAll(ocrRes.Text, " ", "")
	nums := digitRegex.FindString(cleanText)
	if nums == "" {
		return 0
	}

	val, _ := strconv.Atoi(nums)
	return val
}

// ResolveROI 获取节点配置传入的识别区域。
func ResolveROI(arg *maa.CustomRecognitionArg) maa.Rect {
	return arg.Roi
}

// NewArenaRecordSelfPowerRunner 创建记录自身战力的识别器。
func NewArenaRecordSelfPowerRunner(c *PowerChecker) maa.CustomRecognitionRunner {
	return maa.CustomRecognitionFunc(func(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
		roi := ResolveROI(arg)
		if power, ok := c.RecordSelfPower(ctx, arg.Img, roi); ok {
			return &maa.CustomRecognitionResult{
				Box:    roi,
				Detail: fmt.Sprintf(`{"self_power":%d}`, power),
			}, true
		}
		return nil, false
	})
}

// NewArenaCanBeatRunner 创建通用的对手比对识别器。
func NewArenaCanBeatRunner(c *PowerChecker) maa.CustomRecognitionRunner {
	return maa.CustomRecognitionFunc(func(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
		oppROI := ResolveROI(arg)
		if c.CheckCanBeat(ctx, arg.Img, oppROI) {
			return &maa.CustomRecognitionResult{Box: oppROI}, true
		}
		return nil, false
	})
}
