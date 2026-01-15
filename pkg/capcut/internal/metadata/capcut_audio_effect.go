// Package metadata/capcut_audio_effect 定义CapCut特有的音频特效相关元数据
// 对应Python的 pyJianYingDraft/metadata/capcut_audio_effect_meta.py
package metadata

// CapCutVoiceFiltersEffectType CapCut语音滤镜特效类型
// 对应Python的CapCut_Voice_filters_effect_type枚举
type CapCutVoiceFiltersEffectType struct {
	EffectEnum
}

// CapCut特有的语音滤镜特效类型
var (
	// AI智能语音处理
	CapCutVoiceFiltersEffectTypeAI降噪 = CapCutVoiceFiltersEffectType{NewEffectEnum("AI降噪", NewAudioEffectMeta(
		"AI降噪", true, "capcut_voice_ai_denoise_001", "capcut_effect_voice_001", "aiDenoise123", "AI语音",
		"智能识别并消除背景噪音", []EffectParam{
			NewEffectParam("strength", 80.0, 0.0, 100.0),      // 降噪强度
			NewEffectParam("preservation", 90.0, 50.0, 100.0), // 语音保真度
		}))}
)

// CapCutVoiceCharactersEffectType CapCut语音角色特效类型
// 对应Python的CapCut_Voice_characters_effect_type枚举
type CapCutVoiceCharactersEffectType struct {
	EffectEnum
}

// CapCut特有的语音角色特效类型
var (
	// AI智能角色声音
	CapCutVoiceCharactersEffectTypeAI小萝莉 = CapCutVoiceCharactersEffectType{NewEffectEnum("AI小萝莉", NewAudioEffectMeta(
		"AI小萝莉", true, "capcut_voice_ai_loli_001", "capcut_effect_voice_char_001", "aiLoli123", "AI角色",
		"AI生成的小萝莉声音", []EffectParam{
			NewEffectParam("cuteness", 90.0, 50.0, 100.0), // 可爱度
			NewEffectParam("pitch", 80.0, 60.0, 100.0),    // 音调
			NewEffectParam("sweetness", 85.0, 0.0, 100.0), // 甜美度
		}))}
)

// CapCutSpeechToSongEffectType CapCut语音转歌声特效类型
// 对应Python的CapCut_Speech_to_song_effect_type枚举
type CapCutSpeechToSongEffectType struct {
	EffectEnum
}

// CapCut特有的语音转歌声特效类型
var (
	// AI智能转换
	CapCutSpeechToSongEffectTypeAI流行风 = CapCutSpeechToSongEffectType{NewEffectEnum("AI流行风", NewAudioEffectMeta(
		"AI流行风", true, "capcut_s2s_ai_pop_001", "capcut_effect_s2s_001", "aiPopStyle123", "AI转歌声",
		"AI智能转换为流行音乐风格", []EffectParam{
			NewEffectParam("melody_strength", 80.0, 0.0, 100.0), // 旋律强度
			NewEffectParam("rhythm_sync", 85.0, 50.0, 100.0),    // 节奏同步
			NewEffectParam("harmony_depth", 70.0, 0.0, 100.0),   // 和声深度
			NewEffectParam("auto_tune", 75.0, 0.0, 100.0),       // 自动调音
		}))}
)

// GetAllCapCutVoiceFiltersEffectTypes 获取所有CapCut语音滤镜特效类型
func GetAllCapCutVoiceFiltersEffectTypes() []EffectEnumerable {
	return []EffectEnumerable{
		CapCutVoiceFiltersEffectTypeAI降噪,
	}
}

// GetAllCapCutVoiceCharactersEffectTypes 获取所有CapCut语音角色特效类型
func GetAllCapCutVoiceCharactersEffectTypes() []EffectEnumerable {
	return []EffectEnumerable{
		CapCutVoiceCharactersEffectTypeAI小萝莉,
	}
}

// GetAllCapCutSpeechToSongEffectTypes 获取所有CapCut语音转歌声特效类型
func GetAllCapCutSpeechToSongEffectTypes() []EffectEnumerable {
	return []EffectEnumerable{
		CapCutSpeechToSongEffectTypeAI流行风,
	}
}

// GetCapCutAudioEffectsByCategory 根据分类获取CapCut音频特效
func GetCapCutAudioEffectsByCategory(category string) []EffectEnumerable {
	var result []EffectEnumerable

	// 合并所有CapCut音频特效
	allEffects := append(GetAllCapCutVoiceFiltersEffectTypes(), GetAllCapCutVoiceCharactersEffectTypes()...)
	allEffects = append(allEffects, GetAllCapCutSpeechToSongEffectTypes()...)

	for _, effect := range allEffects {
		if meta, ok := effect.GetMeta().(AudioEffectMeta); ok {
			if meta.Category == category {
				result = append(result, effect)
			}
		}
	}

	return result
}

// GetAllCapCutAudioEffectCategories 获取所有CapCut音频特效分类
func GetAllCapCutAudioEffectCategories() []string {
	return []string{
		"AI语音",
		"专业语音",
		"创意语音",
		"AI角色",
		"虚拟角色",
		"动漫角色",
		"AI转歌声",
		"专业转歌声",
	}
}

// FindCapCutVoiceFilterByName 根据名称查找CapCut语音滤镜特效
func FindCapCutVoiceFilterByName(name string) (EffectEnumerable, error) {
	return FindEffectByName(GetAllCapCutVoiceFiltersEffectTypes(), name)
}

// FindCapCutVoiceCharacterByName 根据名称查找CapCut语音角色特效
func FindCapCutVoiceCharacterByName(name string) (EffectEnumerable, error) {
	return FindEffectByName(GetAllCapCutVoiceCharactersEffectTypes(), name)
}

// FindCapCutSpeechToSongByName 根据名称查找CapCut语音转歌声特效
func FindCapCutSpeechToSongByName(name string) (EffectEnumerable, error) {
	return FindEffectByName(GetAllCapCutSpeechToSongEffectTypes(), name)
}
