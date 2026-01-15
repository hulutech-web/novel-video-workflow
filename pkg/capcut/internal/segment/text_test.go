package segment

import (
	"encoding/json"
	"testing"

	"novel-video-workflow/pkg/capcut/internal/types"
)

func TestTextStyle(t *testing.T) {
	// 测试默认样式
	defaultStyle := NewTextStyle()
	if defaultStyle.Size != 8.0 {
		t.Errorf("Expected default size 8.0, got %f", defaultStyle.Size)
	}

	if defaultStyle.Color != [3]float64{1.0, 1.0, 1.0} {
		t.Errorf("Expected default color white, got %v", defaultStyle.Color)
	}

	if defaultStyle.Alpha != 1.0 {
		t.Errorf("Expected default alpha 1.0, got %f", defaultStyle.Alpha)
	}

	// 测试自定义样式
	customStyle := NewTextStyleWithParams(
		12.0, true, false, true,
		[3]float64{1.0, 0.0, 0.0}, 0.8,
		1, false, 5, 10,
	)

	if customStyle.Size != 12.0 {
		t.Errorf("Expected size 12.0, got %f", customStyle.Size)
	}

	if !customStyle.Bold {
		t.Error("Expected bold to be true")
	}

	if customStyle.Italic {
		t.Error("Expected italic to be false")
	}

	if !customStyle.Underline {
		t.Error("Expected underline to be true")
	}

	if customStyle.Color != [3]float64{1.0, 0.0, 0.0} {
		t.Errorf("Expected red color, got %v", customStyle.Color)
	}
}

func TestTextBorder(t *testing.T) {
	// 测试默认描边
	defaultBorder := NewTextBorderDefault()
	if defaultBorder.Alpha != 1.0 {
		t.Errorf("Expected alpha 1.0, got %f", defaultBorder.Alpha)
	}

	if defaultBorder.Color != [3]float64{0.0, 0.0, 0.0} {
		t.Errorf("Expected black color, got %v", defaultBorder.Color)
	}

	// 测试自定义描边
	customBorder := NewTextBorder(0.8, [3]float64{1.0, 0.0, 0.0}, 20.0)
	if customBorder.Alpha != 0.8 {
		t.Errorf("Expected alpha 0.8, got %f", customBorder.Alpha)
	}

	if customBorder.Color != [3]float64{1.0, 0.0, 0.0} {
		t.Errorf("Expected red color, got %v", customBorder.Color)
	}

	// 测试JSON导出
	jsonData := customBorder.ExportJSON()
	if jsonData["width"] != customBorder.Width {
		t.Errorf("Expected JSON width %f, got %v", customBorder.Width, jsonData["width"])
	}
}

func TestTextBackground(t *testing.T) {
	// 测试默认背景
	defaultBg := NewTextBackgroundDefault("#FF0000")
	if defaultBg.Color != "#FF0000" {
		t.Errorf("Expected color '#FF0000', got '%s'", defaultBg.Color)
	}

	if defaultBg.Alpha != 1.0 {
		t.Errorf("Expected alpha 1.0, got %f", defaultBg.Alpha)
	}

	// 测试自定义背景
	customBg := NewTextBackground("#00FF00", 2, 0.7, 5.0, 0.2, 0.3, 0.1, 0.9)
	if customBg.Color != "#00FF00" {
		t.Errorf("Expected color '#00FF00', got '%s'", customBg.Color)
	}

	if customBg.Alpha != 0.7 {
		t.Errorf("Expected alpha 0.7, got %f", customBg.Alpha)
	}

	// 测试JSON导出
	jsonData := customBg.ExportJSON()
	if jsonData["background_color"] != "#00FF00" {
		t.Errorf("Expected JSON background_color '#00FF00', got %v", jsonData["background_color"])
	}
}

func TestTextShadow(t *testing.T) {
	// 测试默认阴影
	defaultShadow := NewTextShadowDefault()
	if defaultShadow.HasShadow {
		t.Error("Expected HasShadow to be false by default")
	}

	if defaultShadow.Color != "#000000" {
		t.Errorf("Expected color '#000000', got '%s'", defaultShadow.Color)
	}

	// 测试自定义阴影
	customShadow := NewTextShadow(true, 0.8, 30.0, "#FF0000", 10.0, 0.6)
	if !customShadow.HasShadow {
		t.Error("Expected HasShadow to be true")
	}

	if customShadow.Alpha != 0.8 {
		t.Errorf("Expected alpha 0.8, got %f", customShadow.Alpha)
	}

	if customShadow.Angle != 30.0 {
		t.Errorf("Expected angle 30.0, got %f", customShadow.Angle)
	}
}

func TestTextSegment(t *testing.T) {
	// 测试简单文本片段
	timerange, _ := types.Trange("2s", "5s")
	textSegment := NewTextSegmentSimple("你好世界", timerange)

	if textSegment.Text != "你好世界" {
		t.Errorf("Expected text '你好世界', got '%s'", textSegment.Text)
	}

	if textSegment.Font != "思源黑体" {
		t.Errorf("Expected font '思源黑体', got '%s'", textSegment.Font)
	}

	if textSegment.Style == nil {
		t.Error("Expected non-nil style")
	}

	// 测试字符数统计
	if textSegment.GetWordCount() != 4 {
		t.Errorf("Expected word count 4, got %d", textSegment.GetWordCount())
	}

	// 测试行数统计
	if textSegment.GetLineCount() != 1 {
		t.Errorf("Expected line count 1, got %d", textSegment.GetLineCount())
	}
}

func TestTextSegmentWithStyle(t *testing.T) {
	// 测试带样式的文本片段
	timerange, _ := types.Trange("1s", "3s")
	style := NewTextStyleWithParams(
		14.0, true, true, false,
		[3]float64{0.0, 1.0, 0.0}, 0.9,
		1, false, 2, 5,
	)
	clipSettings := NewClipSettingsWithParams(0.8, 15.0, 1.2, 1.2, 0.1, -0.05, false, false)

	textSegment := NewTextSegment("Hello World", timerange, "Arial", style, clipSettings)

	if textSegment.Text != "Hello World" {
		t.Errorf("Expected text 'Hello World', got '%s'", textSegment.Text)
	}

	if textSegment.Font != "Arial" {
		t.Errorf("Expected font 'Arial', got '%s'", textSegment.Font)
	}

	if textSegment.Style.Size != 14.0 {
		t.Errorf("Expected style size 14.0, got %f", textSegment.Style.Size)
	}

	if !textSegment.Style.Bold {
		t.Error("Expected style bold to be true")
	}

	if textSegment.ClipSettings.Alpha != 0.8 {
		t.Errorf("Expected clip alpha 0.8, got %f", textSegment.ClipSettings.Alpha)
	}
}

func TestTextSegmentWithEffects(t *testing.T) {
	// 测试带各种效果的文本片段
	timerange, _ := types.Trange("0s", "4s")
	textSegment := NewTextSegmentSimple("测试文本", timerange)

	// 添加描边
	textSegment.SetBorder(0.9, [3]float64{1.0, 0.0, 0.0}, 30.0)

	// 添加背景
	textSegment.SetBackground("#0080FF", 1, 0.8, 2.0, 0.15, 0.2, 0.3, 0.7)

	// 添加阴影
	textSegment.SetShadow(true, 0.7, -30.0, "#800080", 8.0, 0.5)

	// 设置固定尺寸
	textSegment.SetFixedSize(1920, 200)

	// 验证效果
	if textSegment.Border == nil {
		t.Error("Expected border to be set")
	}

	if textSegment.Background == nil {
		t.Error("Expected background to be set")
	}

	if textSegment.Shadow == nil {
		t.Error("Expected shadow to be set")
	}

	if textSegment.FixedWidth != 1920 {
		t.Errorf("Expected fixed width 1920, got %d", textSegment.FixedWidth)
	}

	if textSegment.FixedHeight != 200 {
		t.Errorf("Expected fixed height 200, got %d", textSegment.FixedHeight)
	}

	// 测试JSON导出
	jsonData := textSegment.ExportJSON()
	if jsonData["type"] != "text" {
		t.Errorf("Expected type 'text', got %v", jsonData["type"])
	}

	if jsonData["text"] != "测试文本" {
		t.Errorf("Expected text '测试文本', got %v", jsonData["text"])
	}

	// 验证JSON可以序列化
	_, err := json.Marshal(jsonData)
	if err != nil {
		t.Errorf("Failed to marshal TextSegment JSON: %v", err)
	}
}

func TestTextBubbleAndEffect(t *testing.T) {
	// 测试气泡效果
	bubble := NewTextBubble("bubble_123", "bubble_res", "Cool Bubble")
	if bubble.EffectID != "bubble_123" {
		t.Errorf("Expected EffectID 'bubble_123', got '%s'", bubble.EffectID)
	}

	if bubble.Name != "Cool Bubble" {
		t.Errorf("Expected Name 'Cool Bubble', got '%s'", bubble.Name)
	}

	// 测试添加到文本片段
	timerange, _ := types.Trange("1s", "2s")
	textSegment := NewTextSegmentSimple("气泡文本", timerange)
	textSegment.SetBubble("bubble_456", "bubble_res2", "Another Bubble")

	if textSegment.Bubble == nil {
		t.Error("Expected bubble to be set")
	}

	if textSegment.Bubble.EffectID != "bubble_456" {
		t.Errorf("Expected bubble EffectID 'bubble_456', got '%s'", textSegment.Bubble.EffectID)
	}
}

func TestMultilineText(t *testing.T) {
	// 测试多行文本
	multilineText := "Line 1\nLine 2\nLine 3"
	timerange, _ := types.Trange("0s", "3s")
	textSegment := NewTextSegmentSimple(multilineText, timerange)

	if textSegment.GetLineCount() != 3 {
		t.Errorf("Expected line count 3, got %d", textSegment.GetLineCount())
	}

	// 测试字符数统计（不包括\n）
	expectedWordCount := len([]rune(multilineText))
	if textSegment.GetWordCount() != expectedWordCount {
		t.Errorf("Expected word count %d, got %d", expectedWordCount, textSegment.GetWordCount())
	}
}
