// Package segment/audio 定义音频片段及其相关类
// 对应Python的 audio_segment.py
package segment

import (
	"fmt"

	"novel-video-workflow/pkg/capcut/internal/types"

	"github.com/google/uuid"
)

// AudioFade 音频淡入淡出效果
// 对应Python的Audio_fade类
type AudioFade struct {
	FadeID      string `json:"id"`                // 淡入淡出效果的全局id，自动生成
	InDuration  int64  `json:"fade_in_duration"`  // 淡入时长，单位为微秒
	OutDuration int64  `json:"fade_out_duration"` // 淡出时长，单位为微秒
}

// NewAudioFade 创建新的音频淡入淡出效果
func NewAudioFade(inDuration, outDuration int64) *AudioFade {
	return &AudioFade{
		FadeID:      uuid.New().String(),
		InDuration:  inDuration,
		OutDuration: outDuration,
	}
}

// NewAudioFadeFromString 从字符串时间创建音频淡入淡出效果
func NewAudioFadeFromString(inDuration, outDuration interface{}) (*AudioFade, error) {
	inMicros, err := types.Tim(inDuration)
	if err != nil {
		return nil, fmt.Errorf("invalid in_duration: %w", err)
	}

	outMicros, err := types.Tim(outDuration)
	if err != nil {
		return nil, fmt.Errorf("invalid out_duration: %w", err)
	}

	return NewAudioFade(inMicros, outMicros), nil
}

// ExportJSON 导出为JSON格式
func (af *AudioFade) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"id":                af.FadeID,
		"fade_in_duration":  af.InDuration,
		"fade_out_duration": af.OutDuration,
		"fade_type":         0,
		"type":              "audio_fade",
	}
}

// AudioEffect 音频特效对象
// 对应Python的Audio_effect类
type AudioEffect struct {
	Name              string        `json:"name"`                // 特效名称
	EffectID          string        `json:"id"`                  // 特效全局id，由程序自动生成
	ResourceID        string        `json:"resource_id"`         // 资源id，由剪映本身提供
	CategoryID        string        `json:"category_id"`         // 分类ID："sound_effect", "tone", "speech_to_song"
	CategoryName      string        `json:"category_name"`       // 分类名称："场景音", "音色", "声音成曲"
	AudioAdjustParams []interface{} `json:"audio_adjust_params"` // 音频调整参数列表，暂时用interface{}
}

// NewAudioEffect 创建新的音频特效
func NewAudioEffect(name, resourceID, categoryID, categoryName string) *AudioEffect {
	return &AudioEffect{
		Name:              name,
		EffectID:          uuid.New().String(),
		ResourceID:        resourceID,
		CategoryID:        categoryID,
		CategoryName:      categoryName,
		AudioAdjustParams: make([]interface{}, 0),
	}
}

// AudioEffectCategory 音频特效分类
type AudioEffectCategory struct {
	ID   string
	Name string
}

// 预定义的音频特效分类
var (
	AudioEffectCategorySoundEffect  = AudioEffectCategory{"sound_effect", "场景音"}
	AudioEffectCategoryTone         = AudioEffectCategory{"tone", "音色"}
	AudioEffectCategorySpeechToSong = AudioEffectCategory{"speech_to_song", "声音成曲"}
	// CapCut版本的分类
	AudioEffectCategoryVoiceFilters       = AudioEffectCategory{"sound_effect", "Voice filters"}
	AudioEffectCategoryVoiceCharacters    = AudioEffectCategory{"tone", "Voice characters"}
	AudioEffectCategoryCapCutSpeechToSong = AudioEffectCategory{"speech_to_song", "Speech to song"}
)

// NewAudioEffectWithCategory 使用预定义分类创建音频特效
func NewAudioEffectWithCategory(name, resourceID string, category AudioEffectCategory) *AudioEffect {
	return NewAudioEffect(name, resourceID, category.ID, category.Name)
}

// ExportJSON 导出为JSON格式
func (ae *AudioEffect) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"audio_adjust_params": ae.AudioAdjustParams,
		"category_id":         ae.CategoryID,
		"category_name":       ae.CategoryName,
		"id":                  ae.EffectID,
		"is_ugc":              false,
		"name":                ae.Name,
		"production_path":     "",
		"resource_id":         ae.ResourceID,
		"speaker_id":          "",
		"sub_type":            1,
		"time_range": map[string]int64{
			"duration": 0,
			"start":    0,
		}, // 似乎并未用到
		"type": "audio_effect",
		// 不导出path和constant_material_id
	}
}

// AudioSegment 安放在轨道上的一个音频片段
// 对应Python的Audio_segment类
type AudioSegment struct {
	*MediaSegment
	MaterialInstance interface{}    `json:"-"`              // 音频素材实例，暂时用interface{}
	Fade             *AudioFade     `json:"fade,omitempty"` // 音频淡入淡出效果，可能为空
	Effects          []*AudioEffect `json:"effects"`        // 音频特效列表
}

// NewAudioSegment 创建新的音频片段
func NewAudioSegment(materialID string, targetTimerange *types.Timerange, sourceTimerange *types.Timerange, speed, volume float64) *AudioSegment {
	// 处理参数逻辑，类似Python版本
	var finalSourceTimerange *types.Timerange
	var finalSpeed float64
	var finalTargetTimerange *types.Timerange

	if sourceTimerange != nil && speed != 0 {
		// 如果同时指定了源时间范围和速度，重新计算目标时间范围
		newDuration := int64(float64(sourceTimerange.Duration)/speed + 0.5)
		finalTargetTimerange = types.NewTimerange(targetTimerange.Start, newDuration)
		finalSourceTimerange = sourceTimerange
		finalSpeed = speed
	} else if sourceTimerange != nil {
		// 只指定了源时间范围，计算速度
		finalSpeed = float64(sourceTimerange.Duration) / float64(targetTimerange.Duration)
		finalSourceTimerange = sourceTimerange
		finalTargetTimerange = targetTimerange
	} else {
		// 没有指定源时间范围
		if speed == 0 {
			finalSpeed = 1.0
		} else {
			finalSpeed = speed
		}
		sourceDuration := int64(float64(targetTimerange.Duration)*finalSpeed + 0.5)
		finalSourceTimerange = types.NewTimerange(0, sourceDuration)
		finalTargetTimerange = targetTimerange
	}

	return &AudioSegment{
		MediaSegment:     NewMediaSegment(materialID, finalSourceTimerange, finalTargetTimerange, finalSpeed, volume),
		MaterialInstance: nil, // TODO: 待实现Audio_material后设置
		Fade:             nil,
		Effects:          make([]*AudioEffect, 0),
	}
}

// NewAudioSegmentSimple 创建简单的音频片段
func NewAudioSegmentSimple(materialID string, targetTimerange *types.Timerange, volume float64) *AudioSegment {
	return NewAudioSegment(materialID, targetTimerange, nil, 1.0, volume)
}

// AddEffect 为音频片段添加一个作用于整个片段的音频效果
func (as *AudioSegment) AddEffect(name, resourceID string, category AudioEffectCategory, effectID ...string) error {
	// 检查是否已经存在相同分类的音效
	for _, effect := range as.Effects {
		if effect.CategoryID == category.ID {
			return fmt.Errorf("当前音频片段已经有此类型 (%s) 的音效了", category.Name)
		}
	}

	effect := NewAudioEffectWithCategory(name, resourceID, category)

	// 如果提供了自定义effect_id，使用它
	if len(effectID) > 0 && effectID[0] != "" {
		effect.EffectID = effectID[0]
	}

	as.Effects = append(as.Effects, effect)
	as.ExtraMaterialRefs = append(as.ExtraMaterialRefs, effect.EffectID)

	return nil
}

// AddFade 为音频片段添加淡入淡出效果
func (as *AudioSegment) AddFade(inDuration, outDuration interface{}) error {
	if as.Fade != nil {
		return fmt.Errorf("当前片段已存在淡入淡出效果")
	}

	fade, err := NewAudioFadeFromString(inDuration, outDuration)
	if err != nil {
		return fmt.Errorf("创建淡入淡出效果失败: %w", err)
	}

	as.Fade = fade
	as.ExtraMaterialRefs = append(as.ExtraMaterialRefs, fade.FadeID)

	return nil
}

// RemoveEffect 移除指定分类的音频特效
func (as *AudioSegment) RemoveEffect(categoryID string) bool {
	for i, effect := range as.Effects {
		if effect.CategoryID == categoryID {
			// 从特效列表中移除
			as.Effects = append(as.Effects[:i], as.Effects[i+1:]...)

			// 从额外素材引用列表中移除
			for j, ref := range as.ExtraMaterialRefs {
				if ref == effect.EffectID {
					as.ExtraMaterialRefs = append(as.ExtraMaterialRefs[:j], as.ExtraMaterialRefs[j+1:]...)
					break
				}
			}

			return true
		}
	}
	return false
}

// RemoveFade 移除淡入淡出效果
func (as *AudioSegment) RemoveFade() bool {
	if as.Fade == nil {
		return false
	}

	// 从额外素材引用列表中移除
	for i, ref := range as.ExtraMaterialRefs {
		if ref == as.Fade.FadeID {
			as.ExtraMaterialRefs = append(as.ExtraMaterialRefs[:i], as.ExtraMaterialRefs[i+1:]...)
			break
		}
	}

	as.Fade = nil
	return true
}

// HasEffect 检查是否包含指定分类的音频特效
func (as *AudioSegment) HasEffect(categoryID string) bool {
	for _, effect := range as.Effects {
		if effect.CategoryID == categoryID {
			return true
		}
	}
	return false
}

// GetEffect 获取指定分类的音频特效
func (as *AudioSegment) GetEffect(categoryID string) *AudioEffect {
	for _, effect := range as.Effects {
		if effect.CategoryID == categoryID {
			return effect
		}
	}
	return nil
}

// ExportJSON 导出音频片段的JSON数据
func (as *AudioSegment) ExportJSON() map[string]interface{} {
	result := as.MediaSegment.ExportJSON()

	// 添加音频片段特有的字段
	result["type"] = "audio"

	// 如果有淡入淡出效果，添加相关信息
	if as.Fade != nil {
		result["fade"] = as.Fade.ExportJSON()
	}

	// 添加特效列表
	effectsJSON := make([]interface{}, len(as.Effects))
	for i, effect := range as.Effects {
		effectsJSON[i] = effect.ExportJSON()
	}
	result["effects"] = effectsJSON

	return result
}

// String 返回音频片段的字符串表示
func (as *AudioSegment) String() string {
	fadeInfo := "NoFade"
	if as.Fade != nil {
		fadeInfo = fmt.Sprintf("Fade(%s,%s)",
			types.FormatDuration(as.Fade.InDuration),
			types.FormatDuration(as.Fade.OutDuration))
	}

	return fmt.Sprintf("AudioSegment{%s, %s, Effects: %d}",
		as.BaseSegment.String(), fadeInfo, len(as.Effects))
}
