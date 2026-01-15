// Package metadata/video_effect_meta
// 对应Python的 pyJianYingDraft/metadata/video_effect_meta.py
package metadata

// VideoSceneEffectType 视频场景效果类型
// 对应Python的CapCut_Intro_type枚举
type VideoSceneEffectType struct {
	EffectEnum
}

// 记录剪映自带的视频特效
var (
	// 剪映自带的画面特效类型
	CapCutIntroTypeAI人物识别 = RegisterEffect("video_scene", VideoSceneEffectType{NewEffectEnum("AI人物识别", NewAnimationMeta(
		"AI人物识别", true, 2.0, "capcut_intro_ai_person_001", "capcut_effect_intro_001", "aiPersonIntro123"))})
)

// VideoCharacterEffectType 视频角色效果类型
// 对应Python的CapCut_Outro_type枚举
type VideoCharacterEffectType struct {
	EffectEnum
}

var (
	// 剪映自带的画面特效类型
	CapCutIntroTypeAI人物识别1 = RegisterEffect("video_character", VideoCharacterEffectType{NewEffectEnum("AI人物识别", NewAnimationMeta(
		"AI人物识别", true, 2.0, "capcut_intro_ai_person_001", "capcut_effect_intro_001", "aiPersonIntro123"))})
)

// GetAllVideoSceneEffectType 视频场景效果类型
func GetAllVideoSceneEffectType() []EffectEnumerable {
	return GetAllEffects("video_scene")
}

// GetAllVideoCharacterEffectType 视频角色效果类型
func GetAllVideoCharacterEffectType() []EffectEnumerable {
	return GetAllEffects("video_character")
}

// FindVideoSceneEffectByName 根据名称查找频场景效果
func FindVideoSceneEffectByName(name string) (EffectEnumerable, error) {
	return FindEffect("video_scene", name)
}

// FindVideoCharacterEffectByName 根据名称查找视频角色效果类型
func FindVideoCharacterEffectByName(name string) (EffectEnumerable, error) {
	return FindEffect("video_character", name)
}
