// Package metadata/capcut_animation 定义CapCut特有的动画相关元数据
// 对应Python的 pyJianYingDraft/metadata/capcut_animation_meta.py 和 capcut_text_animation_meta.py
package metadata

// CapCutIntroType CapCut特有入场动画类型
// 对应Python的CapCut_Intro_type枚举
type CapCutIntroType struct {
	EffectEnum
}

// CapCut特有的高级入场动画类型
var (
	// 剪映自带的画面特效类型
	CapCutIntroType_1998 = CapCutIntroType{NewEffectEnum("1998", NewAnimationMeta(
		"1998", false, 1.5, "6981791065204331044", "1183068", "d53096e8139dd33f7a2be6adcd7ce56b"))}
	CapCutIntroTypeFadeIn = CapCutIntroType{NewEffectEnum("Fade In", NewAnimationMeta(
		"Fade In", true, 1.0, "capcut_intro_fade_001", "capcut_effect_intro_002", "fadeIn123"))}
)

// CapCutOutroType CapCut特有出场动画类型
// 对应Python的CapCut_Outro_type枚举
type CapCutOutroType struct {
	EffectEnum
}

// CapCut特有的高级出场动画类型
var (
	// AI智能动画
	CapCutOutroTypeAI人物消散 = CapCutOutroType{NewEffectEnum("AI人物消散", NewAnimationMeta(
		"AI人物消散", true, 2.0, "capcut_outro_ai_person_001", "capcut_effect_outro_001", "aiPersonOutro123"))}
)

// CapCutGroupAnimationType CapCut特有组合动画类型
// 对应Python的CapCut_Group_animation_type枚举
type CapCutGroupAnimationType struct {
	EffectEnum
}

// CapCut特有的高级组合动画类型
var (
	// AI驱动动画
	CapCutGroupAnimationTypeAI节拍同步 = CapCutGroupAnimationType{NewEffectEnum("AI节拍同步", NewAnimationMeta(
		"AI节拍同步", true, 0.0, "capcut_group_ai_beat_sync_001", "capcut_effect_group_001", "aiBeatSync123"))}
	CapCutGroupAnimationTypeRotation = CapCutGroupAnimationType{NewEffectEnum("Rotation", NewAnimationMeta(
		"Rotation", true, 1.5, "capcut_group_rotation_001", "capcut_effect_group_002", "rotationCapCut123"))}
)

// CapCutTextIntro CapCut特有文字入场动画类型
// 对应Python的CapCut_Text_intro枚举
type CapCutTextIntro struct {
	EffectEnum
}

// CapCut特有的高级文字入场动画
var (
	// AI文字动画
	CapCutTextIntroAI智能排版 = CapCutTextIntro{NewEffectEnum("AI智能排版", NewAnimationMeta(
		"AI智能排版", true, 2.0, "capcut_text_intro_ai_layout_001", "capcut_effect_text_intro_001", "aiLayoutIntro123"))}
	CapCutTextIntroTypeTypewriter = CapCutTextIntro{NewEffectEnum("Typewriter", NewAnimationMeta(
		"Typewriter", true, 2.0, "capcut_text_intro_typewriter_001", "capcut_effect_text_intro_002", "typewriterCapCut123"))}
)

// CapCutTextOutro CapCut特有文字出场动画类型
// 对应Python的CapCut_Text_outro枚举
type CapCutTextOutro struct {
	EffectEnum
}

// CapCut特有的高级文字出场动画
var (
	// AI文字动画
	CapCutTextOutroAI智能消散 = CapCutTextOutro{NewEffectEnum("AI智能消散", NewAnimationMeta(
		"AI智能消散", true, 2.0, "capcut_text_outro_ai_dissolve_001", "capcut_effect_text_outro_001", "aiDissolveOutro123"))}
)

// CapCutTextLoopAnim CapCut特有文字循环动画类型
// 对应Python的CapCut_Text_loop_anim枚举
type CapCutTextLoopAnim struct {
	EffectEnum
}

// CapCut特有的高级文字循环动画
var (
	// AI文字动画
	CapCutTextLoopAnimAI节拍跟随 = CapCutTextLoopAnim{NewEffectEnum("AI节拍跟随", NewAnimationMeta(
		"AI节拍跟随", true, 0.0, "capcut_text_loop_ai_beat_001", "capcut_effect_text_loop_001", "aiBeatLoop123"))}
)

// init 初始化函数，注册所有CapCut动画类型
func init() {
	// 注册CapCut入场动画
	RegisterEffect("capcut_intro", CapCutIntroType_1998)
	RegisterEffect("capcut_intro", CapCutIntroTypeFadeIn)

	// 注册CapCut出场动画
	RegisterEffect("capcut_outro", CapCutOutroTypeAI人物消散)

	// 注册CapCut组合动画
	RegisterEffect("capcut_group_animation", CapCutGroupAnimationTypeAI节拍同步)
	RegisterEffect("capcut_group_animation", CapCutGroupAnimationTypeRotation)

	// 注册CapCut文字入场动画
	RegisterEffect("capcut_text_intro", CapCutTextIntroAI智能排版)
	RegisterEffect("capcut_text_intro", CapCutTextIntroTypeTypewriter)

	// 注册CapCut文字出场动画
	RegisterEffect("capcut_text_outro", CapCutTextOutroAI智能消散)

	// 注册CapCut文字循环动画
	RegisterEffect("capcut_text_loop_anim", CapCutTextLoopAnimAI节拍跟随)
}

// GetAllCapCutIntroTypes 获取所有CapCut入场动画类型
func GetAllCapCutIntroTypes() []EffectEnumerable {
	return GetAllEffects("capcut_intro")
}

// GetAllCapCutOutroTypes 获取所有CapCut出场动画类型
func GetAllCapCutOutroTypes() []EffectEnumerable {
	return GetAllEffects("capcut_outro")
}

// GetAllCapCutGroupAnimationTypes 获取所有CapCut组合动画类型
func GetAllCapCutGroupAnimationTypes() []EffectEnumerable {
	return GetAllEffects("capcut_group_animation")
}

// GetAllCapCutTextIntroTypes 获取所有CapCut文字入场动画类型
func GetAllCapCutTextIntroTypes() []EffectEnumerable {
	return GetAllEffects("capcut_text_intro")
}

// GetAllCapCutTextOutroTypes 获取所有CapCut文字出场动画类型
func GetAllCapCutTextOutroTypes() []EffectEnumerable {
	return GetAllEffects("capcut_text_outro")
}

// GetAllCapCutTextLoopAnimTypes 获取所有CapCut文字循环动画类型
func GetAllCapCutTextLoopAnimTypes() []EffectEnumerable {
	return GetAllEffects("capcut_text_loop_anim")
}

// FindCapCutIntroByName 根据名称查找CapCut入场动画
func FindCapCutIntroByName(name string) (EffectEnumerable, error) {
	return FindEffect("capcut_intro", name)
}

// FindCapCutOutroByName 根据名称查找CapCut出场动画
func FindCapCutOutroByName(name string) (EffectEnumerable, error) {
	return FindEffect("capcut_outro", name)
}

// FindCapCutGroupAnimationByName 根据名称查找CapCut组合动画
func FindCapCutGroupAnimationByName(name string) (EffectEnumerable, error) {
	return FindEffect("capcut_group_animation", name)
}

// FindCapCutTextIntroByName 根据名称查找CapCut文字入场动画
func FindCapCutTextIntroByName(name string) (EffectEnumerable, error) {
	return FindEffect("capcut_text_intro", name)
}

// FindCapCutTextOutroByName 根据名称查找CapCut文字出场动画
func FindCapCutTextOutroByName(name string) (EffectEnumerable, error) {
	return FindEffect("capcut_text_outro", name)
}

// FindCapCutTextLoopAnimByName 根据名称查找CapCut文字循环动画
func FindCapCutTextLoopAnimByName(name string) (EffectEnumerable, error) {
	return FindEffect("capcut_text_loop_anim", name)
}

// FindCapCutIntroTypeByName 根据名称查找CapCut入场动画类型（别名函数）
func FindCapCutIntroTypeByName(name string) (EffectEnumerable, error) {
	return FindCapCutIntroByName(name)
}
