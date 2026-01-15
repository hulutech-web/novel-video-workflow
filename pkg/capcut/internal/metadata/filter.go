// Package metadata/filter 定义滤镜相关的元数据
// 对应Python的 pyJianYingDraft/metadata/filter_meta.py
package metadata

// FilterMeta 滤镜元数据
// 对应Python的Filter_meta类（虽然Python中没有明确定义，但隐含存在）
type FilterMeta struct {
	Name        string  `json:"name"`        // 滤镜名称
	IsVIP       bool    `json:"is_vip"`      // 是否为VIP特权
	ResourceID  string  `json:"resource_id"` // 资源ID
	EffectID    string  `json:"effect_id"`   // 效果ID
	MD5         string  `json:"md5"`         // MD5值
	Category    string  `json:"category"`    // 滤镜分类
	Intensity   float64 `json:"intensity"`   // 默认强度 (0.0-1.0)
	Description string  `json:"description"` // 滤镜描述
}

// NewFilterMeta 创建新的滤镜元数据
func NewFilterMeta(name string, isVIP bool, resourceID, effectID, md5, category, description string, intensity float64) FilterMeta {
	return FilterMeta{
		Name:        name,
		IsVIP:       isVIP,
		ResourceID:  resourceID,
		EffectID:    effectID,
		MD5:         md5,
		Category:    category,
		Intensity:   intensity,
		Description: description,
	}
}

// FilterType 滤镜类型枚举
// 对应Python的Filter_type枚举
type FilterType struct {
	EffectEnum
}

// 剪映自带的滤镜类型 - 按分类组织
var (
	// === 人像滤镜 ===
	FilterType自然 = RegisterEffect("filter", FilterType{NewEffectEnum("自然", NewFilterMeta(
		"自然", false, "filter_portrait_natural_001", "effect_filter_001", "nat123", "人像", "自然肤色，适合日常拍摄", 0.8))})
)

// GetAllFilterTypes 获取所有滤镜类型
func GetAllFilterTypes() []EffectEnumerable {
	return GetAllEffects("filter")
}

// GetFiltersByCategory 根据分类获取滤镜类型
func GetFiltersByCategory(category string) []EffectEnumerable {
	var result []EffectEnumerable
	allFilters := GetAllFilterTypes()

	for _, filter := range allFilters {
		if meta, ok := filter.GetMeta().(FilterMeta); ok {
			if meta.Category == category {
				result = append(result, filter)
			}
		}
	}

	return result
}

// GetAllFilterCategories 获取所有滤镜分类
func GetAllFilterCategories() []string {
	return []string{
		"人像",
		"风景",
		"电影",
		"美食",
		"日系",
		"韩系",
		"创意",
	}
}

// FindFilterByName 根据名称查找滤镜类型
func FindFilterByName(name string) (EffectEnumerable, error) {
	return FindEffect("filter", name)
}
