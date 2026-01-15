package metadata

// Package metadata 定义各种特效/音效/滤镜等的元数据
// 对应Python的 pyJianYingDraft/metadata/ 包

import (
	"fmt"
	"strings"
)

// EffectParam 特效参数信息
// 对应Python的Effect_param类
type EffectParam struct {
	Name         string  `json:"name"`          // 参数名称
	DefaultValue float64 `json:"default_value"` // 默认值
	MinValue     float64 `json:"min_value"`     // 最小值
	MaxValue     float64 `json:"max_value"`     // 最大值
}

// NewEffectParam 创建新的特效参数
func NewEffectParam(name string, defaultValue, minValue, maxValue float64) EffectParam {
	return EffectParam{
		Name:         name,
		DefaultValue: defaultValue,
		MinValue:     minValue,
		MaxValue:     maxValue,
	}
}

// EffectParamInstance 特效参数实例
// 对应Python的Effect_param_instance类
type EffectParamInstance struct {
	EffectParam
	Index int     `json:"parameterIndex"` // 参数索引
	Value float64 `json:"value"`          // 当前值
}

// NewEffectParamInstance 创建特效参数实例
func NewEffectParamInstance(param EffectParam, index int, value float64) EffectParamInstance {
	return EffectParamInstance{
		EffectParam: param,
		Index:       index,
		Value:       value,
	}
}

// ExportJSON 导出为JSON格式
func (e EffectParamInstance) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"default_value":  e.DefaultValue,
		"max_value":      e.MaxValue,
		"min_value":      e.MinValue,
		"name":           e.Name,
		"parameterIndex": e.Index,
		"portIndex":      0,
		"value":          e.Value,
	}
}

// EffectMeta 特效元数据
// 对应Python的Effect_meta类
type EffectMeta struct {
	Name       string        `json:"name"`        // 效果名称
	IsVIP      bool          `json:"is_vip"`      // 是否为VIP特权
	ResourceID string        `json:"resource_id"` // 资源ID
	EffectID   string        `json:"effect_id"`   // 效果ID
	MD5        string        `json:"md5"`         // MD5值
	Params     []EffectParam `json:"params"`      // 效果的参数信息
}

// NewEffectMeta 创建新的特效元数据
func NewEffectMeta(name string, isVIP bool, resourceID, effectID, md5 string, params []EffectParam) EffectMeta {
	if params == nil {
		params = []EffectParam{}
	}
	return EffectMeta{
		Name:       name,
		IsVIP:      isVIP,
		ResourceID: resourceID,
		EffectID:   effectID,
		MD5:        md5,
		Params:     params,
	}
}

// ParseParams 解析参数列表(范围0~100), 返回参数实例列表
// 对应Python的parse_params方法
func (e EffectMeta) ParseParams(params []float64) ([]EffectParamInstance, error) {
	var ret []EffectParamInstance

	if params == nil {
		params = []float64{}
	}

	for i, param := range e.Params {
		val := param.DefaultValue
		if i < len(params) {
			inputV := params[i]
			if inputV < 0 || inputV > 100 {
				return nil, fmt.Errorf("invalid parameter value %f for %s", inputV, param.Name)
			}
			// 从0~100映射到实际值
			val = param.MinValue + (param.MaxValue-param.MinValue)*inputV/100.0
		}
		ret = append(ret, NewEffectParamInstance(param, i, val))
	}
	return ret, nil
}

// EffectEnumerable 特效枚举接口
// 对应Python的Effect_enum基类功能
type EffectEnumerable interface {
	GetName() string
	GetMeta() interface{} // 返回具体的元数据类型
}

// AnimationMetaProvider 动画元数据提供者接口
// 用于 animation 包中的类型匹配
type AnimationMetaProvider interface {
	GetAnimationMeta() AnimationMeta
}

// EffectEnum 特效枚举基础结构
type EffectEnum struct {
	name string
	meta interface{}
}

// NewEffectEnum 创建特效枚举项
func NewEffectEnum(name string, meta interface{}) EffectEnum {
	return EffectEnum{
		name: name,
		meta: meta,
	}
}

// GetName 获取特效名称
func (e EffectEnum) GetName() string {
	return e.name
}

// GetMeta 获取元数据
func (e EffectEnum) GetMeta() interface{} {
	return e.meta
}

// FindEffectByName 根据名称查找特效，忽略大小写、空格和下划线
// 对应Python Effect_enum.from_name方法
func FindEffectByName(effects []EffectEnumerable, name string) (EffectEnumerable, error) {
	normalizedName := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(name, " ", ""), "_", ""))

	for _, effect := range effects {
		effectName := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(effect.GetName(), " ", ""), "_", ""))
		if effectName == normalizedName {
			return effect, nil
		}
	}

	return nil, fmt.Errorf("effect named '%s' not found", name)
}

// EffectRegistry 特效注册表
// 用于统一管理所有特效类型的注册和获取
type EffectRegistry struct {
	effects map[string][]EffectEnumerable
}

// NewEffectRegistry 创建新的特效注册表
func NewEffectRegistry() *EffectRegistry {
	return &EffectRegistry{
		effects: make(map[string][]EffectEnumerable),
	}
}

// Register 注册特效到指定分类
func (r *EffectRegistry) Register(category string, effect EffectEnumerable) {
	if r.effects[category] == nil {
		r.effects[category] = make([]EffectEnumerable, 0)
	}
	r.effects[category] = append(r.effects[category], effect)
}

// GetAll 获取指定分类的所有特效
func (r *EffectRegistry) GetAll(category string) []EffectEnumerable {
	if effects, exists := r.effects[category]; exists {
		return effects
	}
	return []EffectEnumerable{}
}

// FindByName 在指定分类中根据名称查找特效
func (r *EffectRegistry) FindByName(category, name string) (EffectEnumerable, error) {
	effects := r.GetAll(category)
	return FindEffectByName(effects, name)
}

// GetAllCategories 获取所有分类名称
func (r *EffectRegistry) GetAllCategories() []string {
	var categories []string
	for category := range r.effects {
		categories = append(categories, category)
	}
	return categories
}

// 全局注册表实例
var globalRegistry = NewEffectRegistry()

// RegisterEffect 全局注册特效函数
func RegisterEffect(category string, effect EffectEnumerable) EffectEnumerable {
	globalRegistry.Register(category, effect)
	return effect
}

// GetAllEffects 获取指定分类的所有特效
func GetAllEffects(category string) []EffectEnumerable {
	return globalRegistry.GetAll(category)
}

// FindEffect 在指定分类中根据名称查找特效
func FindEffect(category, name string) (EffectEnumerable, error) {
	return globalRegistry.FindByName(category, name)
}
