// Package metadata/animation 定义动画相关的元数据
// 对应Python的 pyJianYingDraft/metadata/animation_meta.py
package metadata

// AnimationMeta 动画元数据
// 对应Python的Animation_meta类
type AnimationMeta struct {
	Title      string `json:"title"`       // 动画标题
	IsVIP      bool   `json:"is_vip"`      // 是否为VIP
	Duration   int64  `json:"duration"`    // 效果默认时长，单位为微秒
	ResourceID string `json:"resource_id"` // 资源ID
	EffectID   string `json:"effect_id"`   // 效果ID
	MD5        string `json:"md5"`         // MD5值
}

// NewAnimationMeta 创建新的动画元数据
// duration参数单位为秒，会自动转换为微秒
func NewAnimationMeta(title string, isVIP bool, duration float64, resourceID, effectID, md5 string) AnimationMeta {
	return AnimationMeta{
		Title:      title,
		IsVIP:      isVIP,
		Duration:   int64(duration * 1e6), // 转换为微秒
		ResourceID: resourceID,
		EffectID:   effectID,
		MD5:        md5,
	}
}

// IntroType入场动画类型
// 对应Python的Intro_type枚举
type IntroType struct {
	EffectEnum
}

// 剪映自带的视频/图片入场动画类型
var (
	//免费入场动画
	IntroType缩小 = RegisterEffect("intro", IntroType{NewEffectEnum("缩小", NewAnimationMeta("缩小", false, 0.500, "6798332584276267527", "624755", "7e0e6b55704b7fc20588fee77058e95c"))})
	IntroType渐显 = RegisterEffect("intro", IntroType{NewEffectEnum("渐显", NewAnimationMeta("渐显", false, 0.500, "6798332584276267528", "624756", "7e0e6b55704b7fc20588fee77058e95d"))})
	IntroType放大 = RegisterEffect("intro", IntroType{NewEffectEnum("放大", NewAnimationMeta("放大", false, 0.500, "6798332584276267529", "624757", "7e0e6b55704b7fc20588fee77058e95e"))})
)

// OutroType 出场动画类型
// 对应Python的Outro_type枚举
type OutroType struct {
	EffectEnum
}

// 剪映自带的视频/图片出场动画类型
var (
	OutroType缩小 = RegisterEffect("outro", OutroType{NewEffectEnum("缩小", NewAnimationMeta(
		"缩小", false, 1.0, "outro_shrink_001", "effect_outro_001", "shrinkO123"))})
)

// GroupAnimationType 组合动画类型
// 对应Python的Group_animation_type枚举
type GroupAnimationType struct {
	EffectEnum
}

// 剪映自带的组合动画类型
var (
	GroupAnimationType呼吸 = RegisterEffect("group_animation", GroupAnimationType{NewEffectEnum("呼吸", NewAnimationMeta(
		"呼吸", false, 2.0, "group_breathe_001", "effect_group_001", "breathe123"))})
	GroupAnimationType三分割 = RegisterEffect("group_animation", GroupAnimationType{NewEffectEnum("三分割", NewAnimationMeta(
		"三分割", false, 2.0, "group_three_split_001", "effect_group_002", "threeSplit123"))})
)

// TextIntro 文字入场动画类型
// 对应Python的Text_intro枚举
type TextIntro struct {
	EffectEnum
}

// 剪映自带的文字入场动画
var (
	TextIntro打字机 = RegisterEffect("text_intro", TextIntro{NewEffectEnum("打字机", NewAnimationMeta(
		"打字机", false, 2.0, "text_intro_typewriter_001", "effect_text_intro_001", "typewriter123"))})
	TextIntroType打字机 = RegisterEffect("text_intro", TextIntro{NewEffectEnum("打字机", NewAnimationMeta(
		"打字机", false, 2.0, "text_intro_typewriter_001", "effect_text_intro_001", "typewriter123"))})
)

// TextOutro 文字出场动画类型
// 对应Python的Text_outro枚举
type TextOutro struct {
	EffectEnum
}

// 剪映自带的文字出场动画
var (
	TextOutro逐字消失 = RegisterEffect("text_outro", TextOutro{NewEffectEnum("逐字消失", NewAnimationMeta(
		"逐字消失", false, 1.5, "text_outro_char_disappear_001", "effect_text_outro_001", "charDisappear123"))})
	TextOutroType渐隐 = RegisterEffect("text_outro", TextOutro{NewEffectEnum("渐隐", NewAnimationMeta(
		"渐隐", false, 1.5, "text_outro_fade_001", "effect_text_outro_002", "fadeOut123"))})
)

// TextLoopAnim 文字循环动画类型
// 对应Python的Text_loop_anim枚举
type TextLoopAnim struct {
	EffectEnum
}

// 剪映自带的文字循环动画
var (
	TextLoopAnim闪烁 = RegisterEffect("text_loop_anim", TextLoopAnim{NewEffectEnum("闪烁", NewAnimationMeta(
		"闪烁", false, 1.0, "text_loop_blink_001", "effect_text_loop_001", "textBlink123"))})
	TextLoopAnimType跳动 = RegisterEffect("text_loop_anim", TextLoopAnim{NewEffectEnum("跳动", NewAnimationMeta(
		"跳动", false, 1.0, "text_loop_bounce_001", "effect_text_loop_002", "textBounce123"))})
)

// GetAllIntroTypes 获取所有入场动画类型
func GetAllIntroTypes() []EffectEnumerable {
	return GetAllEffects("intro")
}

// GetAllOutroTypes 获取所有出场动画类型
func GetAllOutroTypes() []EffectEnumerable {
	return GetAllEffects("outro")
}

// GetAllGroupAnimationTypes 获取所有组合动画类型
func GetAllGroupAnimationTypes() []EffectEnumerable {
	return GetAllEffects("group_animation")
}

// GetAllTextIntroTypes 获取所有文字入场动画类型
func GetAllTextIntroTypes() []EffectEnumerable {
	return GetAllEffects("text_intro")
}

// GetAllTextOutroTypes 获取所有文字出场动画类型
func GetAllTextOutroTypes() []EffectEnumerable {
	return GetAllEffects("text_outro")
}

// GetAllTextLoopAnimTypes 获取所有文字循环动画类型
func GetAllTextLoopAnimTypes() []EffectEnumerable {
	return GetAllEffects("text_loop_anim")
}

// FindIntroByName 根据名称查找入场动画
func FindIntroByName(name string) (EffectEnumerable, error) {
	return FindEffect("intro", name)
}

// FindOutroByName 根据名称查找出场动画
func FindOutroByName(name string) (EffectEnumerable, error) {
	return FindEffect("outro", name)
}

// FindGroupAnimationByName 根据名称查找组合动画
func FindGroupAnimationByName(name string) (EffectEnumerable, error) {
	return FindEffect("group_animation", name)
}

// FindTextIntroByName 根据名称查找文字入场动画
func FindTextIntroByName(name string) (EffectEnumerable, error) {
	return FindEffect("text_intro", name)
}

// FindTextOutroByName 根据名称查找文字出场动画
func FindTextOutroByName(name string) (EffectEnumerable, error) {
	return FindEffect("text_outro", name)
}

// FindTextLoopAnimByName 根据名称查找文字循环动画
func FindTextLoopAnimByName(name string) (EffectEnumerable, error) {
	return FindEffect("text_loop_anim", name)
}

// FindIntroTypeByName 根据名称查找入场动画类型（别名函数）
func FindIntroTypeByName(name string) (EffectEnumerable, error) {
	return FindIntroByName(name)
}

// FindTextIntroTypeByName 根据名称查找文字入场动画类型（别名函数）
func FindTextIntroTypeByName(name string) (EffectEnumerable, error) {
	return FindTextIntroByName(name)
}

// FindTextOutroTypeByName 根据名称查找文字出场动画类型（别名函数）
func FindTextOutroTypeByName(name string) (EffectEnumerable, error) {
	return FindTextOutroByName(name)
}

// FindTextLoopAnimTypeByName 根据名称查找文字循环动画类型（别名函数）
func FindTextLoopAnimTypeByName(name string) (EffectEnumerable, error) {
	return FindTextLoopAnimByName(name)
}
