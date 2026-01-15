// Package metadata/transition 定义转场相关的元数据
// 对应Python的 pyJianYingDraft/metadata/transition_meta.py 和 capcut_transition_meta.py
package metadata

// TransitionMeta 转场元数据
// 对应Python的Transition_meta类
type TransitionMeta struct {
	Name            string `json:"name"`             // 转场名称
	IsVIP           bool   `json:"is_vip"`           // 是否为VIP特权
	ResourceID      string `json:"resource_id"`      // 资源ID
	EffectID        string `json:"effect_id"`        // 效果ID
	MD5             string `json:"md5"`              // MD5值
	DefaultDuration int64  `json:"default_duration"` // 默认持续时间，单位为微秒
	IsOverlap       bool   `json:"is_overlap"`       // 是否允许重叠
}

// NewTransitionMeta 创建新的转场元数据
// duration参数单位为秒，会自动转换为微秒
func NewTransitionMeta(name string, isVIP bool, resourceID, effectID, md5 string, duration float64, isOverlap bool) TransitionMeta {
	return TransitionMeta{
		Name:            name,
		IsVIP:           isVIP,
		ResourceID:      resourceID,
		EffectID:        effectID,
		MD5:             md5,
		DefaultDuration: int64(duration * 1e6), // 转换为微秒
		IsOverlap:       isOverlap,
	}
}

// TransitionType 转场类型枚举
// 对应Python的Transition_type枚举
type TransitionType struct {
	EffectEnum
}

// 剪映自带的转场类型 - 免费转场
var (
	// 基础转场
	TransitionType淡入淡出 = RegisterEffect("transition", TransitionType{NewEffectEnum("淡入淡出", NewTransitionMeta(
		"淡入淡出", false, "transition_fade_001", "effect_transition_001", "fade123", 1.0, true))})
)

// GetAllTransitionTypes 获取所有转场类型
func GetAllTransitionTypes() []EffectEnumerable {
	return GetAllEffects("transition")
}

// CapCutTransitionType CapCut特有转场类型
// 对应Python的CapCut_Transition_type枚举
type CapCutTransitionType struct {
	EffectEnum
}

// CapCut特有的高级转场类型
var (
	// AI智能转场
	CapCutTransitionTypeAI场景识别 = RegisterEffect("capcut_transition", CapCutTransitionType{NewEffectEnum("AI场景识别", NewTransitionMeta(
		"AI场景识别", true, "capcut_transition_ai_scene_001", "capcut_effect_transition_001", "aiSc123", 2.0, true))})
)

// GetAllCapCutTransitionTypes 获取所有CapCut转场类型
func GetAllCapCutTransitionTypes() []EffectEnumerable {
	return GetAllEffects("capcut_transition")
}

// FindTransitionByName 根据名称查找转场类型
func FindTransitionByName(name string) (EffectEnumerable, error) {
	// 先在基础转场中查找
	if transition, err := FindEffect("transition", name); err == nil {
		return transition, nil
	}

	// 再在CapCut转场中查找
	return FindEffect("capcut_transition", name)
}
