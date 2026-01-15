// Package segment/text 定义文本片段及其相关类
// 对应Python的 text_segment.py
package segment

import (
	"fmt"
	"strings"

	"novel-video-workflow/pkg/capcut/internal/types"

	"github.com/google/uuid"
)

// TextStyle 字体样式类
// 对应Python的Text_style类
type TextStyle struct {
	Size          float64    `json:"size"`           // 字体大小
	Bold          bool       `json:"bold"`           // 是否加粗
	Italic        bool       `json:"italic"`         // 是否斜体
	Underline     bool       `json:"underline"`      // 是否加下划线
	Color         [3]float64 `json:"color"`          // 字体颜色，RGB三元组，取值范围为[0, 1]
	Alpha         float64    `json:"alpha"`          // 字体不透明度
	Align         int        `json:"align"`          // 对齐方式：0: 左对齐, 1: 居中, 2: 右对齐
	Vertical      bool       `json:"vertical"`       // 是否为竖排文本
	LetterSpacing int        `json:"letter_spacing"` // 字符间距
	LineSpacing   int        `json:"line_spacing"`   // 行间距
}

// NewTextStyle 创建新的文本样式
func NewTextStyle() *TextStyle {
	return &TextStyle{
		Size:          8.0,
		Bold:          false,
		Italic:        false,
		Underline:     false,
		Color:         [3]float64{1.0, 1.0, 1.0}, // 白色
		Alpha:         1.0,
		Align:         0, // 左对齐
		Vertical:      false,
		LetterSpacing: 0,
		LineSpacing:   0,
	}
}

// NewTextStyleWithParams 创建带参数的文本样式
func NewTextStyleWithParams(size float64, bold, italic, underline bool, color [3]float64, alpha float64, align int, vertical bool, letterSpacing, lineSpacing int) *TextStyle {
	return &TextStyle{
		Size:          size,
		Bold:          bold,
		Italic:        italic,
		Underline:     underline,
		Color:         color,
		Alpha:         alpha,
		Align:         align,
		Vertical:      vertical,
		LetterSpacing: letterSpacing,
		LineSpacing:   lineSpacing,
	}
}

// TextBorder 文本描边的参数
// 对应Python的Text_border类
type TextBorder struct {
	Alpha float64    `json:"alpha"` // 描边不透明度
	Color [3]float64 `json:"color"` // 描边颜色，RGB三元组，取值范围为[0, 1]
	Width float64    `json:"width"` // 描边宽度
}

// NewTextBorder 创建新的文本描边
func NewTextBorder(alpha float64, color [3]float64, width float64) *TextBorder {
	return &TextBorder{
		Alpha: alpha,
		Color: color,
		Width: width / 100.0 * 0.2, // 映射到剪映的实际值
	}
}

// NewTextBorderDefault 创建默认的文本描边
func NewTextBorderDefault() *TextBorder {
	return NewTextBorder(1.0, [3]float64{0.0, 0.0, 0.0}, 40.0)
}

// ExportJSON 导出JSON数据，放置在素材content的styles中
func (tb *TextBorder) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"content": map[string]interface{}{
			"solid": map[string]interface{}{
				"alpha": tb.Alpha,
				"color": []float64{tb.Color[0], tb.Color[1], tb.Color[2]},
			},
		},
		"width": tb.Width,
	}
}

// TextBackground 文本背景参数
// 对应Python的Text_background类
type TextBackground struct {
	Style            int     `json:"style"`             // 背景样式
	Alpha            float64 `json:"alpha"`             // 背景不透明度
	Color            string  `json:"color"`             // 背景颜色，格式为'#RRGGBB'
	RoundRadius      float64 `json:"round_radius"`      // 背景圆角半径
	Height           float64 `json:"height"`            // 背景高度
	Width            float64 `json:"width"`             // 背景宽度
	HorizontalOffset float64 `json:"horizontal_offset"` // 背景水平偏移
	VerticalOffset   float64 `json:"vertical_offset"`   // 背景竖直偏移
}

// NewTextBackground 创建新的文本背景
func NewTextBackground(color string, style int, alpha, roundRadius, height, width, horizontalOffset, verticalOffset float64) *TextBackground {
	// 转换样式值 (1,2) -> (0,2)
	mappedStyle := 0
	if style == 2 {
		mappedStyle = 2
	}

	return &TextBackground{
		Style:            mappedStyle,
		Alpha:            alpha,
		Color:            color,
		RoundRadius:      roundRadius,
		Height:           height,
		Width:            width,
		HorizontalOffset: horizontalOffset*2 - 1, // 转换范围从[0,1]到[-1,1]
		VerticalOffset:   verticalOffset*2 - 1,   // 转换范围从[0,1]到[-1,1]
	}
}

// NewTextBackgroundDefault 创建默认的文本背景
func NewTextBackgroundDefault(color string) *TextBackground {
	return NewTextBackground(color, 1, 1.0, 0.0, 0.14, 0.14, 0.5, 0.5)
}

// ExportJSON 生成子JSON数据，在Text_segment导出时合并到其中
func (tbg *TextBackground) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"background_style":             tbg.Style,
		"background_color":             tbg.Color,
		"background_alpha":             tbg.Alpha,
		"background_round_radius":      tbg.RoundRadius,
		"background_height":            tbg.Height,
		"background_width":             tbg.Width,
		"background_horizontal_offset": tbg.HorizontalOffset,
		"background_vertical_offset":   tbg.VerticalOffset,
	}
}

// TextShadow 文本阴影参数
// 对应Python的Text_shadow类
type TextShadow struct {
	HasShadow bool    `json:"has_shadow"` // 是否启用阴影
	Alpha     float64 `json:"alpha"`      // 阴影不透明度
	Angle     float64 `json:"angle"`      // 阴影角度
	Color     string  `json:"color"`      // 阴影颜色，格式为'#RRGGBB'
	Distance  float64 `json:"distance"`   // 阴影距离
	Smoothing float64 `json:"smoothing"`  // 阴影平滑度
}

// NewTextShadow 创建新的文本阴影
func NewTextShadow(hasShadow bool, alpha, angle float64, color string, distance, smoothing float64) *TextShadow {
	return &TextShadow{
		HasShadow: hasShadow,
		Alpha:     alpha,
		Angle:     angle,
		Color:     color,
		Distance:  distance,
		Smoothing: smoothing,
	}
}

// NewTextShadowDefault 创建默认的文本阴影
func NewTextShadowDefault() *TextShadow {
	return NewTextShadow(false, 0.9, -45.0, "#000000", 5.0, 0.45)
}

// TextBubble 文本气泡效果
// 对应Python的TextBubble类（简化版本）
type TextBubble struct {
	GlobalID   string `json:"id"`          // 全局ID
	EffectID   string `json:"effect_id"`   // 效果ID
	ResourceID string `json:"resource_id"` // 资源ID
	Name       string `json:"name"`        // 名称
}

// NewTextBubble 创建新的文本气泡效果
func NewTextBubble(effectID, resourceID, name string) *TextBubble {
	return &TextBubble{
		GlobalID:   uuid.New().String(),
		EffectID:   effectID,
		ResourceID: resourceID,
		Name:       name,
	}
}

// ExportJSON 导出为JSON格式
func (tb *TextBubble) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"id":          tb.GlobalID,
		"effect_id":   tb.EffectID,
		"resource_id": tb.ResourceID,
		"name":        tb.Name,
		"type":        "text_bubble",
		"platform":    "all",
		// 其他字段可以根据需要添加
	}
}

// TextEffect 花字效果
// 对应Python的TextEffect类（简化版本）
type TextEffect struct {
	GlobalID   string `json:"id"`          // 全局ID
	EffectID   string `json:"effect_id"`   // 效果ID
	ResourceID string `json:"resource_id"` // 资源ID
	Name       string `json:"name"`        // 名称
}

// NewTextEffect 创建新的花字效果
func NewTextEffect(effectID, resourceID, name string) *TextEffect {
	return &TextEffect{
		GlobalID:   uuid.New().String(),
		EffectID:   effectID,
		ResourceID: resourceID,
		Name:       name,
	}
}

// ExportJSON 导出为JSON格式
func (te *TextEffect) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"id":          te.GlobalID,
		"effect_id":   te.EffectID,
		"resource_id": te.ResourceID,
		"name":        te.Name,
		"type":        "text_effect",
		"platform":    "all",
		// 其他字段可以根据需要添加
	}
}

// TextStyleRange 多样式文本的样式范围
// 对应Python的TextStyleRange类
type TextStyleRange struct {
	Start  int         `json:"start"`  // 起始位置
	End    int         `json:"end"`    // 结束位置
	Style  *TextStyle  `json:"style"`  // 文本样式
	Border *TextBorder `json:"border"` // 描边（可选）
	Font   string      `json:"font"`   // 字体
}

// NewTextStyleRange 创建新的文本样式范围
func NewTextStyleRange(start, end int, style *TextStyle, border *TextBorder, font string) *TextStyleRange {
	return &TextStyleRange{
		Start:  start,
		End:    end,
		Style:  style,
		Border: border,
		Font:   font,
	}
}

// TextSegment 文本片段
// 对应Python的Text_segment类
type TextSegment struct {
	*VisualSegment
	Text        string            `json:"text"`                  // 文本内容
	Font        string            `json:"font"`                  // 字体
	Style       *TextStyle        `json:"style"`                 // 基础文本样式
	Border      *TextBorder       `json:"border,omitempty"`      // 描边（可选）
	Background  *TextBackground   `json:"background,omitempty"`  // 背景（可选）
	Shadow      *TextShadow       `json:"shadow,omitempty"`      // 阴影（可选）
	Bubble      *TextBubble       `json:"bubble,omitempty"`      // 气泡效果（可选）
	Effect      *TextEffect       `json:"effect,omitempty"`      // 花字效果（可选）
	FixedWidth  int               `json:"fixed_width"`           // 固定宽度
	FixedHeight int               `json:"fixed_height"`          // 固定高度
	TextStyles  []*TextStyleRange `json:"text_styles,omitempty"` // 多样式文本（可选）
}

// NewTextSegment 创建新的文本片段
func NewTextSegment(text string, targetTimerange *types.Timerange, font string, style *TextStyle, clipSettings *ClipSettings) *TextSegment {
	if style == nil {
		style = NewTextStyle()
	}

	// 生成虚拟的material_id用于文本
	textMaterialID := uuid.New().String()

	return &TextSegment{
		VisualSegment: NewVisualSegment(textMaterialID, nil, targetTimerange, 1.0, 1.0, clipSettings),
		Text:          text,
		Font:          font,
		Style:         style,
		Border:        nil,
		Background:    nil,
		Shadow:        nil,
		Bubble:        nil,
		Effect:        nil,
		FixedWidth:    -1, // -1表示自动宽度
		FixedHeight:   -1, // -1表示自动高度
		TextStyles:    nil,
	}
}

// NewTextSegmentSimple 创建简单的文本片段
func NewTextSegmentSimple(text string, targetTimerange *types.Timerange) *TextSegment {
	return NewTextSegment(text, targetTimerange, "思源黑体", nil, nil)
}

// SetBorder 设置文本描边
func (ts *TextSegment) SetBorder(alpha float64, color [3]float64, width float64) *TextSegment {
	ts.Border = NewTextBorder(alpha, color, width)
	return ts
}

// SetBackground 设置文本背景
func (ts *TextSegment) SetBackground(color string, style int, alpha, roundRadius, height, width, horizontalOffset, verticalOffset float64) *TextSegment {
	ts.Background = NewTextBackground(color, style, alpha, roundRadius, height, width, horizontalOffset, verticalOffset)
	return ts
}

// SetShadow 设置文本阴影
func (ts *TextSegment) SetShadow(hasShadow bool, alpha, angle float64, color string, distance, smoothing float64) *TextSegment {
	ts.Shadow = NewTextShadow(hasShadow, alpha, angle, color, distance, smoothing)
	return ts
}

// SetBubble 设置气泡效果
func (ts *TextSegment) SetBubble(effectID, resourceID, name string) *TextSegment {
	ts.Bubble = NewTextBubble(effectID, resourceID, name)
	return ts
}

// SetEffect 设置花字效果
func (ts *TextSegment) SetEffect(effectID, resourceID, name string) *TextSegment {
	ts.Effect = NewTextEffect(effectID, resourceID, name)
	return ts
}

// SetFixedSize 设置固定尺寸
func (ts *TextSegment) SetFixedSize(width, height int) *TextSegment {
	ts.FixedWidth = width
	ts.FixedHeight = height
	return ts
}

// AddTextStyle 添加多样式文本范围
func (ts *TextSegment) AddTextStyle(start, end int, style *TextStyle, border *TextBorder, font string) *TextSegment {
	if ts.TextStyles == nil {
		ts.TextStyles = make([]*TextStyleRange, 0)
	}

	styleRange := NewTextStyleRange(start, end, style, border, font)
	ts.TextStyles = append(ts.TextStyles, styleRange)

	return ts
}

// GetWordCount 获取文本字符数
func (ts *TextSegment) GetWordCount() int {
	return len([]rune(ts.Text)) // 使用rune来正确计算中文字符数
}

// GetLineCount 获取文本行数
func (ts *TextSegment) GetLineCount() int {
	return len(strings.Split(ts.Text, "\n"))
}

// CreateFromTemplate 从模板文本片段创建新的文本片段（简化版本）
func CreateFromTemplate(text string, targetTimerange *types.Timerange, template *TextSegment) *TextSegment {
	newSegment := NewTextSegment(text, targetTimerange, template.Font, template.Style, template.ClipSettings)

	// 复制其他属性
	if template.Border != nil {
		newSegment.Border = &TextBorder{
			Alpha: template.Border.Alpha,
			Color: template.Border.Color,
			Width: template.Border.Width,
		}
	}

	if template.Background != nil {
		newSegment.Background = &TextBackground{
			Style:            template.Background.Style,
			Alpha:            template.Background.Alpha,
			Color:            template.Background.Color,
			RoundRadius:      template.Background.RoundRadius,
			Height:           template.Background.Height,
			Width:            template.Background.Width,
			HorizontalOffset: template.Background.HorizontalOffset,
			VerticalOffset:   template.Background.VerticalOffset,
		}
	}

	newSegment.FixedWidth = template.FixedWidth
	newSegment.FixedHeight = template.FixedHeight

	return newSegment
}

// ExportMaterial 导出为素材格式
func (ts *TextSegment) ExportMaterial() map[string]interface{} {
	material := map[string]interface{}{
		"id":   ts.MaterialID,
		"type": "text",
		"content": map[string]interface{}{
			"text": ts.Text,
			"font": ts.Font,
		},
		"fixed_width":  ts.FixedWidth,
		"fixed_height": ts.FixedHeight,
	}

	// 添加样式信息
	if ts.Style != nil {
		styleData := map[string]interface{}{
			"size":           ts.Style.Size,
			"bold":           ts.Style.Bold,
			"italic":         ts.Style.Italic,
			"underline":      ts.Style.Underline,
			"color":          []float64{ts.Style.Color[0], ts.Style.Color[1], ts.Style.Color[2]},
			"alpha":          ts.Style.Alpha,
			"align":          ts.Style.Align,
			"vertical":       ts.Style.Vertical,
			"letter_spacing": ts.Style.LetterSpacing,
			"line_spacing":   ts.Style.LineSpacing,
		}
		material["style"] = styleData
	}

	// 添加描边信息
	if ts.Border != nil {
		material["border"] = ts.Border.ExportJSON()
	}

	// 添加背景信息
	if ts.Background != nil {
		backgroundData := ts.Background.ExportJSON()
		for k, v := range backgroundData {
			material[k] = v
		}
	}

	return material
}

// ExportJSON 导出文本片段的JSON数据
func (ts *TextSegment) ExportJSON() map[string]interface{} {
	result := ts.VisualSegment.ExportJSON()

	// 添加文本片段特有的字段
	result["type"] = "text"
	result["text"] = ts.Text
	result["font"] = ts.Font

	// 添加样式信息
	if ts.Style != nil {
		result["style"] = map[string]interface{}{
			"size":           ts.Style.Size,
			"bold":           ts.Style.Bold,
			"italic":         ts.Style.Italic,
			"underline":      ts.Style.Underline,
			"color":          []float64{ts.Style.Color[0], ts.Style.Color[1], ts.Style.Color[2]},
			"alpha":          ts.Style.Alpha,
			"align":          ts.Style.Align,
			"vertical":       ts.Style.Vertical,
			"letter_spacing": ts.Style.LetterSpacing,
			"line_spacing":   ts.Style.LineSpacing,
		}
	}

	// 添加其他效果
	if ts.Border != nil {
		result["border"] = ts.Border.ExportJSON()
	}

	if ts.Background != nil {
		backgroundData := ts.Background.ExportJSON()
		for k, v := range backgroundData {
			result[k] = v
		}
	}

	if ts.Shadow != nil {
		result["shadow"] = map[string]interface{}{
			"has_shadow": ts.Shadow.HasShadow,
			"alpha":      ts.Shadow.Alpha,
			"angle":      ts.Shadow.Angle,
			"color":      ts.Shadow.Color,
			"distance":   ts.Shadow.Distance,
			"smoothing":  ts.Shadow.Smoothing,
		}
	}

	if ts.Bubble != nil {
		result["bubble"] = ts.Bubble.ExportJSON()
	}

	if ts.Effect != nil {
		result["text_effect"] = ts.Effect.ExportJSON()
	}

	result["fixed_width"] = ts.FixedWidth
	result["fixed_height"] = ts.FixedHeight

	// 添加多样式文本
	if ts.TextStyles != nil && len(ts.TextStyles) > 0 {
		stylesJSON := make([]interface{}, len(ts.TextStyles))
		for i, styleRange := range ts.TextStyles {
			styleJSON := map[string]interface{}{
				"start": styleRange.Start,
				"end":   styleRange.End,
				"font":  styleRange.Font,
			}

			if styleRange.Style != nil {
				styleJSON["style"] = map[string]interface{}{
					"size":           styleRange.Style.Size,
					"bold":           styleRange.Style.Bold,
					"italic":         styleRange.Style.Italic,
					"underline":      styleRange.Style.Underline,
					"color":          []float64{styleRange.Style.Color[0], styleRange.Style.Color[1], styleRange.Style.Color[2]},
					"alpha":          styleRange.Style.Alpha,
					"align":          styleRange.Style.Align,
					"vertical":       styleRange.Style.Vertical,
					"letter_spacing": styleRange.Style.LetterSpacing,
					"line_spacing":   styleRange.Style.LineSpacing,
				}
			}

			if styleRange.Border != nil {
				styleJSON["border"] = styleRange.Border.ExportJSON()
			}

			stylesJSON[i] = styleJSON
		}
		result["text_styles"] = stylesJSON
	}

	return result
}

// String 返回文本片段的字符串表示
func (ts *TextSegment) String() string {
	textPreview := ts.Text
	if len(textPreview) > 20 {
		textPreview = textPreview[:17] + "..."
	}

	return fmt.Sprintf("TextSegment{%s, Text: \"%s\", Font: %s}",
		ts.BaseSegment.String(), textPreview, ts.Font)
}
