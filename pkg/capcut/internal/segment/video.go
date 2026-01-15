// Package segment/video 定义视频片段及其相关类
// 对应Python的 video_segment.py
package segment

import (
	"fmt"

	"novel-video-workflow/pkg/capcut/internal/types"

	"github.com/google/uuid"
)

// Mask 蒙版对象
// 对应Python的Mask类
type Mask struct {
	GlobalID     string `json:"id"`            // 蒙版全局id，由程序自动生成
	Name         string `json:"name"`          // 蒙版名称
	ResourceType string `json:"resource_type"` // 资源类型
	ResourceID   string `json:"resource_id"`   // 资源ID

	// 蒙版配置参数
	CenterX     float64 `json:"center_x"`     // 蒙版中心x坐标，以半素材宽为单位
	CenterY     float64 `json:"center_y"`     // 蒙版中心y坐标，以半素材高为单位
	Width       float64 `json:"width"`        // 宽度
	Height      float64 `json:"height"`       // 高度
	AspectRatio float64 `json:"aspect_ratio"` // 宽高比
	Rotation    float64 `json:"rotation"`     // 旋转角度
	Invert      bool    `json:"invert"`       // 是否反转
	Feather     float64 `json:"feather"`      // 羽化程度，0-1
	RoundCorner float64 `json:"round_corner"` // 矩形蒙版的圆角，0-1
}

// NewMask 创建新的蒙版对象
func NewMask(name, resourceType, resourceID string, cx, cy, w, h, ratio, rot, feather, roundCorner float64, inv bool) *Mask {
	return &Mask{
		GlobalID:     uuid.New().String(),
		Name:         name,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		CenterX:      cx,
		CenterY:      cy,
		Width:        w,
		Height:       h,
		AspectRatio:  ratio,
		Rotation:     rot,
		Invert:       inv,
		Feather:      feather,
		RoundCorner:  roundCorner,
	}
}

// ExportJSON 导出为JSON格式
func (m *Mask) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"config": map[string]interface{}{
			"aspectRatio": m.AspectRatio,
			"centerX":     m.CenterX,
			"centerY":     m.CenterY,
			"feather":     m.Feather,
			"height":      m.Height,
			"invert":      m.Invert,
			"rotation":    m.Rotation,
			"roundCorner": m.RoundCorner,
			"width":       m.Width,
		},
		"category":      "video",
		"category_id":   "",
		"category_name": "",
		"id":            m.GlobalID,
		"name":          m.Name,
		"platform":      "all",
		"position_info": "",
		"resource_type": m.ResourceType,
		"resource_id":   m.ResourceID,
		"type":          "mask",
		// 不导出path字段
	}
}

// VideoEffect 视频特效素材
// 对应Python的Video_effect类
type VideoEffect struct {
	Name            string        `json:"name"`              // 特效名称
	GlobalID        string        `json:"id"`                // 特效全局id，由程序自动生成
	EffectID        string        `json:"effect_id"`         // 某种特效id，由剪映本身提供
	ResourceID      string        `json:"resource_id"`       // 资源id，由剪映本身提供
	EffectType      string        `json:"type"`              // 特效类型："video_effect" 或 "face_effect"
	ApplyTargetType int           `json:"apply_target_type"` // 应用目标类型，0: 片段，2: 全局
	AdjustParams    []interface{} `json:"adjust_params"`     // 调整参数列表，暂时用interface{}
}

// NewVideoEffect 创建新的视频特效
func NewVideoEffect(name, effectID, resourceID, effectType string, applyTargetType int) *VideoEffect {
	return &VideoEffect{
		Name:            name,
		GlobalID:        uuid.New().String(),
		EffectID:        effectID,
		ResourceID:      resourceID,
		EffectType:      effectType,
		ApplyTargetType: applyTargetType,
		AdjustParams:    make([]interface{}, 0),
	}
}

// ExportJSON 导出为JSON格式
func (ve *VideoEffect) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"adjust_params":        ve.AdjustParams,
		"apply_target_type":    ve.ApplyTargetType,
		"apply_time_range":     nil,
		"category_id":          "", // 一律设为空
		"category_name":        "", // 一律设为空
		"common_keyframes":     []interface{}{},
		"disable_effect_faces": []interface{}{},
		"effect_id":            ve.EffectID,
		"formula_id":           "",
		"id":                   ve.GlobalID,
		"name":                 ve.Name,
		"platform":             "all",
		"render_index":         11000,
		"resource_id":          ve.ResourceID,
		"source_platform":      0,
		"time_range":           nil,
		"track_render_index":   0,
		"type":                 ve.EffectType,
		"value":                1.0,
		"version":              "",
		// 不导出path、request_id和algorithm_artifact_path字段
	}
}

// Filter 滤镜素材
// 对应Python的Filter类
type Filter struct {
	GlobalID        string  `json:"id"`                // 滤镜全局id，由程序自动生成
	Name            string  `json:"name"`              // 滤镜名称
	EffectID        string  `json:"effect_id"`         // 效果ID
	ResourceID      string  `json:"resource_id"`       // 资源ID
	Intensity       float64 `json:"intensity"`         // 滤镜强度（滤镜的唯一参数）
	ApplyTargetType int     `json:"apply_target_type"` // 应用目标类型，0: 片段，2: 全局
}

// NewFilter 创建新的滤镜
func NewFilter(name, effectID, resourceID string, intensity float64, applyTargetType int) *Filter {
	return &Filter{
		GlobalID:        uuid.New().String(),
		Name:            name,
		EffectID:        effectID,
		ResourceID:      resourceID,
		Intensity:       intensity,
		ApplyTargetType: applyTargetType,
	}
}

// ExportJSON 导出为JSON格式
func (f *Filter) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"adjust_params":           []interface{}{},
		"algorithm_artifact_path": "",
		"apply_target_type":       f.ApplyTargetType,
		"bloom_params":            nil,
		"category_id":             "", // 一律设为空
		"category_name":           "", // 一律设为空
		"color_match_info": map[string]string{
			"source_feature_path": "",
			"target_feature_path": "",
			"target_image_path":   "",
		},
		"common_keyframes":   []interface{}{},
		"effect_id":          f.EffectID,
		"formula_id":         "",
		"id":                 f.GlobalID,
		"intensity":          f.Intensity,
		"is_ai_generate":     false,
		"name":               f.Name,
		"platform":           "all",
		"render_index":       11000,
		"resource_id":        f.ResourceID,
		"source_platform":    0,
		"time_range":         nil,
		"track_render_index": 0,
		"type":               "filter",
		"value":              1.0,
		"version":            "",
		// 不导出其他字段
	}
}

// Transition 转场效果
// 对应Python的Transition类
type Transition struct {
	GlobalID   string `json:"id"`          // 转场全局id，由程序自动生成
	Name       string `json:"name"`        // 转场名称
	EffectID   string `json:"effect_id"`   // 效果ID
	ResourceID string `json:"resource_id"` // 资源ID
	Duration   int64  `json:"duration"`    // 转场持续时间，单位为微秒
}

// NewTransition 创建新的转场效果
func NewTransition(name, effectID, resourceID string, duration int64) *Transition {
	return &Transition{
		GlobalID:   uuid.New().String(),
		Name:       name,
		EffectID:   effectID,
		ResourceID: resourceID,
		Duration:   duration,
	}
}

// ExportJSON 导出为JSON格式
func (t *Transition) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"category_id":      "",
		"category_name":    "",
		"check_flag":       0,
		"default_duration": t.Duration,
		"duration":         t.Duration,
		"effect_id":        t.EffectID,
		"id":               t.GlobalID,
		"name":             t.Name,
		"platform":         "all",
		"render_index":     11000,
		"resource_id":      t.ResourceID,
		"type":             "transition",
		// 不导出其他字段
	}
}

// BackgroundFilling 背景填充
// 对应Python的BackgroundFilling类
type BackgroundFilling struct {
	GlobalID string  `json:"id"`        // 全局ID
	FillType string  `json:"fill_type"` // 填充类型："canvas_blur" 或 "canvas_color"
	Blur     float64 `json:"blur"`      // 模糊值
	Color    string  `json:"color"`     // 颜色值
}

// NewBackgroundFilling 创建新的背景填充
func NewBackgroundFilling(fillType string, blur float64, color string) *BackgroundFilling {
	return &BackgroundFilling{
		GlobalID: uuid.New().String(),
		FillType: fillType,
		Blur:     blur,
		Color:    color,
	}
}

// ExportJSON 导出为JSON格式
func (bf *BackgroundFilling) ExportJSON() map[string]interface{} {
	result := map[string]interface{}{
		"id":   bf.GlobalID,
		"type": bf.FillType,
	}

	switch bf.FillType {
	case "canvas_blur":
		result["blur"] = bf.Blur
	case "canvas_color":
		result["color"] = bf.Color
	}

	return result
}

// VideoSegment 视频片段
// 对应Python的Video_segment类
type VideoSegment struct {
	*VisualSegment
	MaterialInstance  interface{}        `json:"-"`                            // 视频素材实例，暂时用interface{}
	Mask              *Mask              `json:"mask,omitempty"`               // 蒙版，可能为空
	Effects           []*VideoEffect     `json:"effects"`                      // 视频特效列表
	Filters           []*Filter          `json:"filters"`                      // 滤镜列表
	Transition        *Transition        `json:"transition,omitempty"`         // 转场效果，可能为空
	BackgroundFilling *BackgroundFilling `json:"background_filling,omitempty"` // 背景填充，可能为空
}

// NewVideoSegment 创建新的视频片段
func NewVideoSegment(materialID string, sourceTimerange, targetTimerange *types.Timerange, speed, volume float64, clipSettings *ClipSettings) *VideoSegment {
	return &VideoSegment{
		VisualSegment:     NewVisualSegment(materialID, sourceTimerange, targetTimerange, speed, volume, clipSettings),
		MaterialInstance:  nil, // TODO: 待实现Video_material后设置
		Mask:              nil,
		Effects:           make([]*VideoEffect, 0),
		Filters:           make([]*Filter, 0),
		Transition:        nil,
		BackgroundFilling: nil,
	}
}

// AddMask 添加蒙版
func (vs *VideoSegment) AddMask(maskType, name, resourceType, resourceID string, centerX, centerY, size, rotation, feather float64, invert bool, rectWidth, roundCorner *float64) *VideoSegment {
	width := size
	height := size
	aspectRatio := 1.0

	// 处理矩形蒙版的特殊参数
	if maskType == "rectangle" && rectWidth != nil {
		width = *rectWidth
		aspectRatio = width / height
	}

	roundCornerValue := 0.0
	if roundCorner != nil {
		roundCornerValue = *roundCorner / 100.0 // 转换为0-1范围
	}

	vs.Mask = NewMask(name, resourceType, resourceID, centerX, centerY, width, height, aspectRatio, rotation, feather, roundCornerValue, invert)

	return vs
}

// AddEffect 添加视频特效
func (vs *VideoSegment) AddEffect(name, effectID, resourceID, effectType string, applyTargetType int) *VideoSegment {
	effect := NewVideoEffect(name, effectID, resourceID, effectType, applyTargetType)
	vs.Effects = append(vs.Effects, effect)

	// 将特效ID添加到额外素材引用列表
	vs.ExtraMaterialRefs = append(vs.ExtraMaterialRefs, effect.GlobalID)

	return vs
}

// AddFilter 添加滤镜
func (vs *VideoSegment) AddFilter(name, effectID, resourceID string, intensity float64, applyTargetType int) *VideoSegment {
	filter := NewFilter(name, effectID, resourceID, intensity, applyTargetType)
	vs.Filters = append(vs.Filters, filter)

	// 将滤镜ID添加到额外素材引用列表
	vs.ExtraMaterialRefs = append(vs.ExtraMaterialRefs, filter.GlobalID)

	return vs
}

// AddTransition 添加转场效果
func (vs *VideoSegment) AddTransition(name, effectID, resourceID string, duration int64) *VideoSegment {
	vs.Transition = NewTransition(name, effectID, resourceID, duration)

	// 将转场ID添加到额外素材引用列表
	vs.ExtraMaterialRefs = append(vs.ExtraMaterialRefs, vs.Transition.GlobalID)

	return vs
}

// SetBackgroundFilling 设置背景填充
func (vs *VideoSegment) SetBackgroundFilling(fillType string, blur float64, color string) *VideoSegment {
	vs.BackgroundFilling = NewBackgroundFilling(fillType, blur, color)
	return vs
}

// ExportJSON 导出视频片段的JSON数据
func (vs *VideoSegment) ExportJSON() map[string]interface{} {
	result := vs.VisualSegment.ExportJSON()

	// 添加视频片段特有的字段
	result["type"] = "video"

	// 如果有蒙版，添加蒙版信息
	if vs.Mask != nil {
		result["mask"] = vs.Mask.ExportJSON()
	}

	// 添加特效列表
	effectsJSON := make([]interface{}, len(vs.Effects))
	for i, effect := range vs.Effects {
		effectsJSON[i] = effect.ExportJSON()
	}
	result["effects"] = effectsJSON

	// 添加滤镜列表
	filtersJSON := make([]interface{}, len(vs.Filters))
	for i, filter := range vs.Filters {
		filtersJSON[i] = filter.ExportJSON()
	}
	result["filters"] = filtersJSON

	// 如果有转场，添加转场信息
	if vs.Transition != nil {
		result["transition"] = vs.Transition.ExportJSON()
	}

	// 如果有背景填充，添加背景填充信息
	if vs.BackgroundFilling != nil {
		result["background_filling"] = vs.BackgroundFilling.ExportJSON()
	}

	return result
}

// String 返回视频片段的字符串表示
func (vs *VideoSegment) String() string {
	return fmt.Sprintf("VideoSegment{%s, Effects: %d, Filters: %d}",
		vs.BaseSegment.String(), len(vs.Effects), len(vs.Filters))
}
