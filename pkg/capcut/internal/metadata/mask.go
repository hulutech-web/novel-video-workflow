// Package metadata/mask 定义蒙版相关的元数据
// 对应Python的 pyJianYingDraft/metadata/mask_meta.py 和 capcut_mask_meta.py
package metadata

// MaskMeta 蒙版元数据
// 对应Python的Mask_meta类
type MaskMeta struct {
	Name               string  `json:"name"`                 // 蒙版名称
	ResourceType       string  `json:"resource_type"`        // 资源类型，与蒙版形状相关
	ResourceID         string  `json:"resource_id"`          // 资源ID
	EffectID           string  `json:"effect_id"`            // 效果ID
	MD5                string  `json:"md5"`                  // MD5值
	DefaultAspectRatio float64 `json:"default_aspect_ratio"` // 默认宽高比(宽高都是相对素材的比例)
}

// NewMaskMeta 创建新的蒙版元数据
func NewMaskMeta(name, resourceType, resourceID, effectID, md5 string, defaultAspectRatio float64) MaskMeta {
	return MaskMeta{
		Name:               name,
		ResourceType:       resourceType,
		ResourceID:         resourceID,
		EffectID:           effectID,
		MD5:                md5,
		DefaultAspectRatio: defaultAspectRatio,
	}
}

// MaskType 蒙版类型枚举
// 对应Python的Mask_type枚举
type MaskType struct {
	EffectEnum
}

// 剪映自带的蒙版类型
var (
	// 基础几何形状蒙版
	MaskType圆形 = RegisterEffect("mask", MaskType{NewEffectEnum("圆形", NewMaskMeta(
		"圆形", "circle", "mask_circle_001", "effect_mask_001", "abc123", 1.0))})
)

// GetAllMaskTypes 获取所有蒙版类型
func GetAllMaskTypes() []EffectEnumerable {
	return GetAllEffects("mask")
}

// CapCutMaskType CapCut特有蒙版类型
// 对应Python的CapCut_Mask_type枚举
type CapCutMaskType struct {
	EffectEnum
}

// CapCut特有的高级蒙版类型
var (
	// AI智能蒙版
	CapCutMaskTypeAI人物 = RegisterEffect("capcut_mask", CapCutMaskType{NewEffectEnum("AI人物", NewMaskMeta(
		"AI人物", "ai_person", "capcut_mask_ai_person_001", "capcut_effect_mask_001", "aiP123", 1.0))})
)

// GetAllCapCutMaskTypes 获取所有CapCut蒙版类型
func GetAllCapCutMaskTypes() []EffectEnumerable {
	return GetAllEffects("capcut_mask")
}

// FindMaskByName 根据名称查找蒙版类型
func FindMaskByName(name string) (EffectEnumerable, error) {
	// 先在基础蒙版中查找
	if mask, err := FindEffect("mask", name); err == nil {
		return mask, nil
	}

	// 再在CapCut蒙版中查找
	return FindEffect("capcut_mask", name)
}
