// Package animation 定义视频/文本动画相关类
// 对应Python的 pyJianYingDraft/animation.py
package animation

import (
	"fmt"
	"strings"

	"novel-video-workflow/pkg/capcut/internal/metadata"
	"novel-video-workflow/pkg/capcut/internal/types"

	"github.com/google/uuid"
)

// AnimationType 动画类型枚举
type AnimationType string

const (
	AnimationTypeIn    AnimationType = "in"    // 入场动画
	AnimationTypeOut   AnimationType = "out"   // 出场动画
	AnimationTypeGroup AnimationType = "group" // 组合动画
	AnimationTypeLoop  AnimationType = "loop"  // 循环动画
)

// Animation 一个视频/文本动画效果
// 对应Python的Animation类
type Animation struct {
	Name             string        `json:"name"`        // 动画名称，默认取为动画效果的名称
	EffectID         string        `json:"id"`          // 另一种动画id，由剪映本身提供
	AnimationType    AnimationType `json:"type"`        // 动画类型
	ResourceID       string        `json:"resource_id"` // 资源id，由剪映本身提供
	Start            int64         `json:"start"`       // 动画相对此片段开头的偏移，单位为微秒
	Duration         int64         `json:"duration"`    // 动画持续时间，单位为微秒
	IsVideoAnimation bool          `json:"-"`           // 是否为视频动画，在子类中定义
}

// NewAnimation 创建新的动画
func NewAnimation(animMeta metadata.AnimationMeta, start, duration int64, animType AnimationType, isVideo bool) *Animation {
	return &Animation{
		Name:             animMeta.Title,
		EffectID:         animMeta.EffectID,
		ResourceID:       animMeta.ResourceID,
		Start:            start,
		Duration:         duration,
		AnimationType:    animType,
		IsVideoAnimation: isVideo,
	}
}

// ExportJSON 导出为JSON格式
// 对应Python的export_json方法
func (a *Animation) ExportJSON() map[string]interface{} {
	panel := ""
	materialType := "sticker"
	if a.IsVideoAnimation {
		panel = "video"
		materialType = "video"
	}

	return map[string]interface{}{
		"anim_adjust_params": nil,
		"platform":           "all",
		"panel":              panel,
		"material_type":      materialType,
		"name":               a.Name,
		"id":                 a.EffectID,
		"type":               string(a.AnimationType),
		"resource_id":        a.ResourceID,
		"start":              a.Start,
		"duration":           a.Duration,
		// 不导出path和request_id
	}
}

// VideoAnimation 一个视频动画效果
// 对应Python的Video_animation类
type VideoAnimation struct {
	*Animation
}

// VideoAnimationInput 视频动画输入接口
type VideoAnimationInput interface {
	metadata.EffectEnumerable
}

// NewVideoAnimation 创建新的视频动画
func NewVideoAnimation(animType VideoAnimationInput, start, duration int64) (*VideoAnimation, error) {
	metaInterface := animType.GetMeta()
	meta, ok := metaInterface.(metadata.AnimationMeta)
	if !ok {
		return nil, fmt.Errorf("invalid animation meta type: %T", metaInterface)
	}

	var aType AnimationType
	switch animType.(type) {
	case metadata.IntroType:
		aType = AnimationTypeIn
	case metadata.OutroType:
		aType = AnimationTypeOut
	case metadata.GroupAnimationType:
		aType = AnimationTypeGroup
	case metadata.CapCutIntroType:
		aType = AnimationTypeIn
	case metadata.CapCutOutroType:
		aType = AnimationTypeOut
	case metadata.CapCutGroupAnimationType:
		aType = AnimationTypeGroup
	default:
		return nil, fmt.Errorf("unsupported video animation type: %T", animType)
	}

	animation := NewAnimation(meta, start, duration, aType, true)
	return &VideoAnimation{Animation: animation}, nil
}

// TextAnimation 一个文本动画效果
// 对应Python的Text_animation类
type TextAnimation struct {
	*Animation
}

// TextAnimationInput 文本动画输入接口
type TextAnimationInput interface {
	metadata.EffectEnumerable
}

// NewTextAnimation 创建新的文本动画
func NewTextAnimation(animType TextAnimationInput, start, duration int64) (*TextAnimation, error) {
	metaInterface := animType.GetMeta()
	meta, ok := metaInterface.(metadata.AnimationMeta)
	if !ok {
		return nil, fmt.Errorf("invalid animation meta type: %T", metaInterface)
	}

	var aType AnimationType
	switch animType.(type) {
	case metadata.TextIntro:
		aType = AnimationTypeIn
	case metadata.TextOutro:
		aType = AnimationTypeOut
	case metadata.TextLoopAnim:
		aType = AnimationTypeLoop
	case metadata.CapCutTextIntro:
		aType = AnimationTypeIn
	case metadata.CapCutTextOutro:
		aType = AnimationTypeOut
	case metadata.CapCutTextLoopAnim:
		aType = AnimationTypeLoop
	default:
		return nil, fmt.Errorf("unsupported text animation type: %T", animType)
	}

	animation := NewAnimation(meta, start, duration, aType, false)
	return &TextAnimation{Animation: animation}, nil
}

// SegmentAnimations 附加于某素材上的一系列动画
// 对应Python的Segment_animations类
//
// 对视频片段：入场、出场或组合动画；对文本片段：入场、出场或循环动画
type SegmentAnimations struct {
	AnimationID string       `json:"id"`         // 系列动画的全局id，自动生成
	Animations  []*Animation `json:"animations"` // 动画列表
}

// NewSegmentAnimations 创建新的片段动画序列
func NewSegmentAnimations() *SegmentAnimations {
	return &SegmentAnimations{
		AnimationID: strings.ReplaceAll(uuid.New().String(), "-", ""),
		Animations:  make([]*Animation, 0),
	}
}

// GetAnimationTimerange 获取指定类型的动画的时间范围
// 对应Python的get_animation_trange方法
func (sa *SegmentAnimations) GetAnimationTimerange(animType AnimationType) (*types.Timerange, error) {
	for _, animation := range sa.Animations {
		if animation.AnimationType == animType {
			return types.NewTimerange(animation.Start, animation.Duration), nil
		}
	}
	return nil, nil
}

// AddAnimation 添加动画
// 对应Python的add_animation方法
func (sa *SegmentAnimations) AddAnimation(animation *Animation) error {
	// 不允许添加超过一个同类型的动画（如两个入场动画）
	for _, existingAnim := range sa.Animations {
		if existingAnim.AnimationType == animation.AnimationType {
			return fmt.Errorf("当前片段已存在类型为 '%s' 的动画", animation.AnimationType)
		}
	}

	if animation.IsVideoAnimation {
		// 不允许组合动画与出入场动画同时出现
		for _, existingAnim := range sa.Animations {
			if existingAnim.AnimationType == AnimationTypeGroup {
				return fmt.Errorf("当前片段已存在组合动画, 此时不能添加其它动画")
			}
		}
		if animation.AnimationType == AnimationTypeGroup && len(sa.Animations) > 0 {
			return fmt.Errorf("当前片段已存在动画时, 不能添加组合动画")
		}
	} else {
		// 文本动画的循环动画限制
		if animation.AnimationType == AnimationTypeLoop {
			for _, existingAnim := range sa.Animations {
				if existingAnim.AnimationType == AnimationTypeLoop {
					return fmt.Errorf("当前片段已存在循环动画, 若希望同时使用循环动画和入出场动画, 请先添加出入场动画再添加循环动画")
				}
			}
		}
	}

	sa.Animations = append(sa.Animations, animation)
	return nil
}

// AddVideoAnimation 添加视频动画
func (sa *SegmentAnimations) AddVideoAnimation(animType VideoAnimationInput, start, duration int64) error {
	videoAnim, err := NewVideoAnimation(animType, start, duration)
	if err != nil {
		return err
	}
	return sa.AddAnimation(videoAnim.Animation)
}

// AddTextAnimation 添加文本动画
func (sa *SegmentAnimations) AddTextAnimation(animType TextAnimationInput, start, duration int64) error {
	textAnim, err := NewTextAnimation(animType, start, duration)
	if err != nil {
		return err
	}
	return sa.AddAnimation(textAnim.Animation)
}

// ExportJSON 导出为JSON格式
// 对应Python的export_json方法
func (sa *SegmentAnimations) ExportJSON() map[string]interface{} {
	animations := make([]map[string]interface{}, len(sa.Animations))
	for i, animation := range sa.Animations {
		animations[i] = animation.ExportJSON()
	}

	return map[string]interface{}{
		"id":                     sa.AnimationID,
		"type":                   "sticker_animation",
		"multi_language_current": "none",
		"animations":             animations,
	}
}
