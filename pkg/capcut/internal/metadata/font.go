// Package metadata/font 定义字体相关的元数据
// 对应Python的 pyJianYingDraft/metadata/font_meta.py
package metadata

// FontMeta 字体元数据
// 对应Python的Font_meta类（虽然Python中没有明确定义，但隐含存在）
type FontMeta struct {
	Name        string   `json:"name"`         // 字体名称
	IsVIP       bool     `json:"is_vip"`       // 是否为VIP特权
	ResourceID  string   `json:"resource_id"`  // 资源ID
	FontFamily  string   `json:"font_family"`  // 字体族名称
	FontWeight  string   `json:"font_weight"`  // 字体粗细 (normal, bold, etc.)
	FontStyle   string   `json:"font_style"`   // 字体样式 (normal, italic, etc.)
	Category    string   `json:"category"`     // 字体分类
	Language    []string `json:"language"`     // 支持的语言
	Description string   `json:"description"`  // 字体描述
	PreviewText string   `json:"preview_text"` // 预览文本
}

// NewFontMeta 创建新的字体元数据
func NewFontMeta(name string, isVIP bool, resourceID, fontFamily, fontWeight, fontStyle, category, description, previewText string, language []string) FontMeta {
	if language == nil {
		language = []string{"zh-CN"} // 默认支持中文
	}
	return FontMeta{
		Name:        name,
		IsVIP:       isVIP,
		ResourceID:  resourceID,
		FontFamily:  fontFamily,
		FontWeight:  fontWeight,
		FontStyle:   fontStyle,
		Category:    category,
		Language:    language,
		Description: description,
		PreviewText: previewText,
	}
}

// FontType 字体类型枚举
// 对应Python的Font_type枚举
type FontType struct {
	EffectEnum
}

// 剪映自带的字体类型 - 按分类组织
var (
	// === 系统字体 ===
	FontType默认 = RegisterEffect("font", FontType{NewEffectEnum("默认", NewFontMeta(
		"默认", false, "font_system_default_001", "PingFang SC", "normal", "normal", "系统",
		"系统默认字体，适用于各种场景", "默认字体", []string{"zh-CN", "en-US"}))})
)

// GetAllFontTypes 获取所有字体类型
func GetAllFontTypes() []EffectEnumerable {
	return GetAllEffects("font")
}

// GetFontsByCategory 根据分类获取字体类型
func GetFontsByCategory(category string) []EffectEnumerable {
	var result []EffectEnumerable
	allFonts := GetAllFontTypes()

	for _, font := range allFonts {
		if meta, ok := font.GetMeta().(FontMeta); ok {
			if meta.Category == category {
				result = append(result, font)
			}
		}
	}

	return result
}

// GetFontsByLanguage 根据语言获取字体类型
func GetFontsByLanguage(language string) []EffectEnumerable {
	var result []EffectEnumerable
	allFonts := GetAllFontTypes()

	for _, font := range allFonts {
		if meta, ok := font.GetMeta().(FontMeta); ok {
			for _, lang := range meta.Language {
				if lang == language {
					result = append(result, font)
					break
				}
			}
		}
	}

	return result
}

// GetAllFontCategories 获取所有字体分类
func GetAllFontCategories() []string {
	return []string{
		"系统",
		"创意",
		"标题",
		"英文",
		"艺术",
		"特殊",
	}
}

// GetSupportedLanguages 获取所有支持的语言
func GetSupportedLanguages() []string {
	return []string{
		"zh-CN", // 简体中文
		"en-US", // 英文
	}
}

// FindFontByName 根据名称查找字体类型
func FindFontByName(name string) (EffectEnumerable, error) {
	return FindEffect("font", name)
}
