// Package segment 定义片段基类及部分比较通用的属性类
// 对应Python的 segment.py
package segment

import (
	"fmt"

	"novel-video-workflow/pkg/capcut/internal/animation"
	"novel-video-workflow/pkg/capcut/internal/keyframe"
	"novel-video-workflow/pkg/capcut/internal/types"

	"github.com/google/uuid"
)

// SegmentInterface 片段接口，所有片段类型都需要实现此接口
type SegmentInterface interface {
	Start() int64
	Duration() int64
	ExportJSON() map[string]interface{}
}

// BaseSegment 片段基类
// 对应Python的Base_segment类
type BaseSegment struct {
	SegmentID       string                       `json:"id"`               // 片段全局id，由程序自动生成
	MaterialID      string                       `json:"material_id"`      // 使用的素材id
	TargetTimerange *types.Timerange             `json:"target_timerange"` // 片段在轨道上的时间范围
	KeyframeManager *keyframe.KeyframeManager    `json:"-"`                // 关键帧管理器
	Animations      *animation.SegmentAnimations `json:"-"`                // 动画管理器
}

// NewBaseSegment 创建基础片段
func NewBaseSegment(materialID string, targetTimerange *types.Timerange) *BaseSegment {
	return &BaseSegment{
		SegmentID:       uuid.New().String(),
		MaterialID:      materialID,
		TargetTimerange: targetTimerange,
		KeyframeManager: keyframe.NewKeyframeManager(),
		Animations:      animation.NewSegmentAnimations(),
	}
}

// Start 片段开始时间，单位为微秒
func (bs *BaseSegment) Start() int64 {
	return bs.TargetTimerange.Start
}

// SetStart 设置片段开始时间
func (bs *BaseSegment) SetStart(value int64) {
	bs.TargetTimerange.Start = value
}

// Duration 片段持续时间，单位为微秒
func (bs *BaseSegment) Duration() int64 {
	return bs.TargetTimerange.Duration
}

// SetDuration 设置片段持续时间
func (bs *BaseSegment) SetDuration(value int64) {
	bs.TargetTimerange.Duration = value
}

// End 片段结束时间，单位为微秒
func (bs *BaseSegment) End() int64 {
	return bs.TargetTimerange.End()
}

// Overlaps 判断是否与另一个片段有重叠
func (bs *BaseSegment) Overlaps(other *BaseSegment) bool {
	if other == nil {
		return false
	}
	return bs.TargetTimerange.Overlaps(other.TargetTimerange)
}

// AddKeyframe 为指定属性添加关键帧
func (bs *BaseSegment) AddKeyframe(property keyframe.KeyframeProperty, timeOffset int64, value float64) {
	bs.KeyframeManager.AddKeyframe(property, timeOffset, value)
}

// AddKeyframeFromString 从字符串添加关键帧
func (bs *BaseSegment) AddKeyframeFromString(propertyName string, timeOffset int64, valueStr string) error {
	return bs.KeyframeManager.AddKeyframeFromString(propertyName, timeOffset, valueStr)
}

// GetKeyframeList 获取指定属性的关键帧列表
func (bs *BaseSegment) GetKeyframeList(property keyframe.KeyframeProperty) *keyframe.KeyframeList {
	return bs.KeyframeManager.GetKeyframeList(property)
}

// HasKeyframes 检查是否有关键帧
func (bs *BaseSegment) HasKeyframes() bool {
	return bs.KeyframeManager.HasKeyframes()
}

// GetID 获取片段ID
func (bs *BaseSegment) GetID() string {
	return bs.SegmentID
}

// GetTargetTimerange 获取目标时间范围
func (bs *BaseSegment) GetTargetTimerange() *types.Timerange {
	return bs.TargetTimerange
}

// GetMaterialRefs 获取素材引用列表
func (bs *BaseSegment) GetMaterialRefs() []string {
	if bs.MaterialID == "" {
		return []string{}
	}
	return []string{bs.MaterialID}
}

// TODO: Animation系统集成方法，待Template系统完成后重新实现
// AddVideoAnimation 添加视频动画
// func (bs *BaseSegment) AddVideoAnimation(meta *animation.AnimationMeta, animationType animation.AnimationType, start, duration int64) error {
// 	videoAnim := animation.NewVideoAnimation(meta, animationType, start, duration)
// 	return bs.Animations.AddVideoAnimation(videoAnim)
// }

// AddTextAnimation 添加文本动画
// func (bs *BaseSegment) AddTextAnimation(meta *animation.AnimationMeta, animationType animation.AnimationType, start, duration int64) error {
// 	textAnim := animation.NewTextAnimation(meta, animationType, start, duration)
// 	return bs.Animations.AddTextAnimation(textAnim)
// }

// HasAnimation 检查是否有指定类型的动画
// func (bs *BaseSegment) HasAnimation(animationType animation.AnimationType) bool {
// 	return bs.Animations.HasAnimation(animationType)
// }

// GetAnimationTimerange 获取指定类型动画的时间范围
// func (bs *BaseSegment) GetAnimationTimerange(animationType animation.AnimationType) (int64, int64, bool) {
// 	return bs.Animations.GetAnimationTimerange(animationType)
// }

// RemoveAnimation 移除指定类型的动画
// func (bs *BaseSegment) RemoveAnimation(animationType animation.AnimationType) bool {
// 	return bs.Animations.RemoveAnimation(animationType)
// }

// ClearAnimations 清空所有动画
// func (bs *BaseSegment) ClearAnimations() {
// 	bs.Animations.ClearAnimations()
// }

// GetAnimationCount 获取动画数量
// func (bs *BaseSegment) GetAnimationCount() int {
// 	return bs.Animations.GetAnimationCount()
// }

// ExportJSON 返回通用于各种片段的属性
func (bs *BaseSegment) ExportJSON() map[string]interface{} {
	result := map[string]interface{}{
		"enable_adjust":               true,
		"enable_color_correct_adjust": false,
		"enable_color_curves":         true,
		"enable_color_match_adjust":   false,
		"enable_color_wheels":         true,
		"enable_lut":                  true,
		"enable_smart_color_adjust":   false,
		"last_nonzero_volume":         1.0,
		"reverse":                     false,
		"track_attribute":             0,
		"track_render_index":          0,
		"visible":                     true,
		// 自定义字段
		"id":               bs.SegmentID,
		"material_id":      bs.MaterialID,
		"target_timerange": bs.TargetTimerange.ExportJSON(),
		"common_keyframes": bs.KeyframeManager.ExportJSON(),
		"keyframe_refs":    []interface{}{}, // 意义不明
	}

	// 如果有动画，添加动画数据
	if len(bs.Animations.Animations) > 0 {
		result["animations"] = bs.Animations.ExportJSON()
	}

	return result
}

// Speed 播放速度对象，目前只支持固定速度
// 对应Python的Speed类
type Speed struct {
	GlobalID string  `json:"id"`    // 全局id，由程序自动生成
	Value    float64 `json:"speed"` // 播放速度
}

// NewSpeed 创建新的播放速度对象
func NewSpeed(speed float64) *Speed {
	return &Speed{
		GlobalID: uuid.New().String(),
		Value:    speed,
	}
}

// ExportJSON 导出为JSON格式
func (s *Speed) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"curve_speed": nil,
		"id":          s.GlobalID,
		"mode":        0,
		"speed":       s.Value,
		"type":        "speed",
	}
}

// ClipSettings 素材片段的图像调节设置
// 对应Python的Clip_settings类
type ClipSettings struct {
	Alpha          float64 `json:"alpha"`           // 图像不透明度，0-1
	FlipHorizontal bool    `json:"flip_horizontal"` // 是否水平翻转
	FlipVertical   bool    `json:"flip_vertical"`   // 是否垂直翻转
	Rotation       float64 `json:"rotation"`        // 顺时针旋转的角度，可正可负
	ScaleX         float64 `json:"scale_x"`         // 水平缩放比例
	ScaleY         float64 `json:"scale_y"`         // 垂直缩放比例
	TransformX     float64 `json:"transform_x"`     // 水平位移，单位为半个画布宽
	TransformY     float64 `json:"transform_y"`     // 垂直位移，单位为半个画布高
}

// NewClipSettings 创建新的图像调节设置，默认不作任何图像变换
func NewClipSettings() *ClipSettings {
	return &ClipSettings{
		Alpha:          1.0,
		FlipHorizontal: false,
		FlipVertical:   false,
		Rotation:       0.0,
		ScaleX:         1.0,
		ScaleY:         1.0,
		TransformX:     0.0,
		TransformY:     0.0,
	}
}

// NewClipSettingsWithParams 创建带参数的图像调节设置
func NewClipSettingsWithParams(alpha, rotation, scaleX, scaleY, transformX, transformY float64, flipH, flipV bool) *ClipSettings {
	return &ClipSettings{
		Alpha:          alpha,
		FlipHorizontal: flipH,
		FlipVertical:   flipV,
		Rotation:       rotation,
		ScaleX:         scaleX,
		ScaleY:         scaleY,
		TransformX:     transformX,
		TransformY:     transformY,
	}
}

// ExportJSON 导出为JSON格式
func (cs *ClipSettings) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"alpha": cs.Alpha,
		"flip": map[string]bool{
			"horizontal": cs.FlipHorizontal,
			"vertical":   cs.FlipVertical,
		},
		"rotation": cs.Rotation,
		"scale": map[string]float64{
			"x": cs.ScaleX,
			"y": cs.ScaleY,
		},
		"transform": map[string]float64{
			"x": cs.TransformX,
			"y": cs.TransformY,
		},
	}
}

// MediaSegment 媒体片段基类
// 对应Python的Media_segment类
type MediaSegment struct {
	*BaseSegment
	SourceTimerange   *types.Timerange `json:"source_timerange"`    // 截取的素材片段的时间范围，对贴纸而言不存在
	Speed             *Speed           `json:"-"`                   // 播放速度设置，在JSON中体现为speed字段
	Volume            float64          `json:"volume"`              // 音量
	ExtraMaterialRefs []string         `json:"extra_material_refs"` // 附加的素材id列表，用于链接动画/特效等
}

// NewMediaSegment 创建媒体片段
func NewMediaSegment(materialID string, sourceTimerange, targetTimerange *types.Timerange, speed, volume float64) *MediaSegment {
	speedObj := NewSpeed(speed)
	return &MediaSegment{
		BaseSegment:       NewBaseSegment(materialID, targetTimerange),
		SourceTimerange:   sourceTimerange,
		Speed:             speedObj,
		Volume:            volume,
		ExtraMaterialRefs: []string{speedObj.GlobalID},
	}
}

// ExportJSON 返回通用于音频和视频片段的默认属性
func (ms *MediaSegment) ExportJSON() map[string]interface{} {
	result := ms.BaseSegment.ExportJSON()

	// 添加媒体片段特有的字段
	if ms.SourceTimerange != nil {
		result["source_timerange"] = ms.SourceTimerange.ExportJSON()
	} else {
		result["source_timerange"] = nil
	}

	result["speed"] = ms.Speed.Value
	result["volume"] = ms.Volume
	result["extra_material_refs"] = ms.ExtraMaterialRefs

	return result
}

// VisualSegment 视觉片段基类，用于处理所有可见片段（视频、贴纸、文本）的共同属性和行为
// 对应Python的Visual_segment类
type VisualSegment struct {
	*MediaSegment
	ClipSettings       *ClipSettings `json:"clip_settings"`       // 图像调节设置，其效果可被关键帧覆盖
	UniformScale       bool          `json:"uniform_scale"`       // 是否锁定XY轴缩放比例
	AnimationsInstance interface{}   `json:"animations_instance"` // 动画实例，可能为空，暂时用interface{}
}

// NewVisualSegment 创建视觉片段
func NewVisualSegment(materialID string, sourceTimerange, targetTimerange *types.Timerange, speed, volume float64, clipSettings *ClipSettings) *VisualSegment {
	if clipSettings == nil {
		clipSettings = NewClipSettings()
	}

	return &VisualSegment{
		MediaSegment:       NewMediaSegment(materialID, sourceTimerange, targetTimerange, speed, volume),
		ClipSettings:       clipSettings,
		UniformScale:       true,
		AnimationsInstance: nil,
	}
}

// AddKeyframe 为给定属性创建一个关键帧，并自动加入到关键帧列表中
func (vs *VisualSegment) AddKeyframe(property string, timeOffset interface{}, value float64) error {
	var offsetMicros int64
	var err error

	switch v := timeOffset.(type) {
	case string:
		offsetMicros, err = types.Tim(v)
		if err != nil {
			return fmt.Errorf("invalid time offset: %w", err)
		}
	case int64:
		offsetMicros = v
	case int:
		offsetMicros = int64(v)
	default:
		return fmt.Errorf("unsupported time offset type: %T", timeOffset)
	}

	// 处理uniform_scale逻辑
	if (property == "scale_x" || property == "scale_y") && vs.UniformScale {
		vs.UniformScale = false
	} else if property == "uniform_scale" {
		if !vs.UniformScale {
			return fmt.Errorf("已设置 scale_x 或 scale_y 时，不能再设置 uniform_scale")
		}
		// uniform_scale实际上是通过scale_x实现的
		property = "scale_x"
	}

	// 将属性名转换为关键帧属性类型
	keyframeProp, err := keyframe.KeyframePropertyFromString(property)
	if err != nil {
		return fmt.Errorf("unsupported keyframe property: %w", err)
	}

	// 添加关键帧到管理器
	vs.KeyframeManager.AddKeyframe(keyframeProp, offsetMicros, value)
	return nil
}

// ExportJSON 导出通用于所有视觉片段的JSON数据
func (vs *VisualSegment) ExportJSON() map[string]interface{} {
	result := vs.MediaSegment.ExportJSON()

	// 添加视觉片段特有的字段
	result["clip"] = vs.ClipSettings.ExportJSON()
	result["uniform_scale"] = map[string]interface{}{
		"on":    vs.UniformScale,
		"value": 1.0,
	}

	return result
}

// String 返回片段的字符串表示
func (bs *BaseSegment) String() string {
	return fmt.Sprintf("Segment{ID: %s, Material: %s, Time: %s}",
		bs.SegmentID, bs.MaterialID, bs.TargetTimerange.String())
}

// SegmentType 片段类型枚举
type SegmentType int

const (
	SegmentTypeVideo SegmentType = iota
	SegmentTypeAudio
	SegmentTypeText
	SegmentTypeSticker
	SegmentTypeEffect
)

// String 返回片段类型的字符串表示
func (st SegmentType) String() string {
	switch st {
	case SegmentTypeVideo:
		return "Video"
	case SegmentTypeAudio:
		return "Audio"
	case SegmentTypeText:
		return "Text"
	case SegmentTypeSticker:
		return "Sticker"
	case SegmentTypeEffect:
		return "Effect"
	default:
		return "Unknown"
	}
}
