// Package metadata/audio_effect 定义音频特效相关的元数据
// 对应Python的 pyJianYingDraft/metadata/audio_effect_meta.py 和 capcut_audio_effect_meta.py
package metadata

// AudioEffectMeta 音频特效元数据
// 对应Python的Audio_effect_meta类（虽然Python中没有明确定义，但隐含存在）
type AudioEffectMeta struct {
	Name        string        `json:"name"`        // 音效名称
	IsVIP       bool          `json:"is_vip"`      // 是否为VIP特权
	ResourceID  string        `json:"resource_id"` // 资源ID
	EffectID    string        `json:"effect_id"`   // 效果ID
	MD5         string        `json:"md5"`         // MD5值
	Category    string        `json:"category"`    // 音效分类
	Description string        `json:"description"` // 音效描述
	Params      []EffectParam `json:"params"`      // 音效参数
}

// NewAudioEffectMeta 创建新的音频特效元数据
func NewAudioEffectMeta(name string, isVIP bool, resourceID, effectID, md5, category, description string, params []EffectParam) AudioEffectMeta {
	if params == nil {
		params = []EffectParam{}
	}
	return AudioEffectMeta{
		Name:        name,
		IsVIP:       isVIP,
		ResourceID:  resourceID,
		EffectID:    effectID,
		MD5:         md5,
		Category:    category,
		Description: description,
		Params:      params,
	}
}

// AudioSceneEffectType 音频场景特效类型
// 对应Python的Audio_scene_effect_type枚举
type AudioSceneEffectType struct {
	EffectEnum
}

// 剪映自带的音频场景特效类型
var (
	// === 环境音效 ===
	AudioSceneEffectType雨声 = RegisterEffect("audio_scene", AudioSceneEffectType{NewEffectEnum("雨声", NewAudioEffectMeta(
		"雨声", false, "audio_scene_rain_001", "effect_audio_scene_001", "rain123", "环境",
		"自然雨声效果", []EffectParam{
			NewEffectParam("intensity", 50.0, 0.0, 100.0),  // 强度
			NewEffectParam("frequency", 30.0, 10.0, 100.0), // 频率
		}))})
)

// ToneEffectType 音调特效类型
// 对应Python的Tone_effect_type枚举
type ToneEffectType struct {
	EffectEnum
}

// 音调调节特效类型
var (
	// === 基础音调调节 ===
	ToneEffectType升调 = RegisterEffect("tone_effect", ToneEffectType{NewEffectEnum("升调", NewAudioEffectMeta(
		"升调", false, "audio_tone_pitch_up_001", "effect_audio_tone_001", "pitchUp123", "音调",
		"提高音调", []EffectParam{
			NewEffectParam("pitch", 20.0, 0.0, 100.0), // 音调偏移
		}))})
)

// SpeechToSongType 语音转歌声特效类型
// 对应Python的Speech_to_song_type枚举
type SpeechToSongType struct {
	EffectEnum
}

// 语音转歌声特效类型
var (
	SpeechToSongType流行 = RegisterEffect("speech_to_song", SpeechToSongType{NewEffectEnum("流行", NewAudioEffectMeta(
		"流行", true, "audio_s2s_pop_001", "effect_audio_s2s_001", "pop901", "语音转歌声",
		"流行音乐风格", []EffectParam{
			NewEffectParam("melody", 70.0, 0.0, 100.0),  // 旋律强度
			NewEffectParam("harmony", 50.0, 0.0, 100.0), // 和声
			NewEffectParam("rhythm", 60.0, 0.0, 100.0),  // 节奏
		}))})
)

// GetAllAudioSceneEffectTypes 获取所有音频场景特效类型
func GetAllAudioSceneEffectTypes() []EffectEnumerable {
	return GetAllEffects("audio_scene")
}

// GetAllToneEffectTypes 获取所有音调特效类型
func GetAllToneEffectTypes() []EffectEnumerable {
	return GetAllEffects("tone_effect")
}

// GetAllSpeechToSongTypes 获取所有语音转歌声类型
func GetAllSpeechToSongTypes() []EffectEnumerable {
	return GetAllEffects("speech_to_song")
}

// GetAudioEffectsByCategory 根据分类获取音频特效
func GetAudioEffectsByCategory(category string) []EffectEnumerable {
	var result []EffectEnumerable

	// 合并所有音频特效
	allEffects := append(GetAllAudioSceneEffectTypes(), GetAllToneEffectTypes()...)
	allEffects = append(allEffects, GetAllSpeechToSongTypes()...)

	for _, effect := range allEffects {
		if meta, ok := effect.GetMeta().(AudioEffectMeta); ok {
			if meta.Category == category {
				result = append(result, effect)
			}
		}
	}

	return result
}

// GetAllAudioEffectCategories 获取所有音频特效分类
func GetAllAudioEffectCategories() []string {
	return []string{
		"环境",
		"空间",
		"音调",
		"变声",
		"语音转歌声",
	}
}

// FindAudioSceneEffectByName 根据名称查找音频场景特效
func FindAudioSceneEffectByName(name string) (EffectEnumerable, error) {
	return FindEffect("audio_scene", name)
}

// FindToneEffectByName 根据名称查找音调特效
func FindToneEffectByName(name string) (EffectEnumerable, error) {
	return FindEffect("tone_effect", name)
}

// FindSpeechToSongByName 根据名称查找语音转歌声特效
func FindSpeechToSongByName(name string) (EffectEnumerable, error) {
	return FindEffect("speech_to_song", name)
}
