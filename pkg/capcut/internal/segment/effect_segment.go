// Package segment 定义特效和滤镜片段类
// 对应Python的 pyJianYingDraft/effect_segment.py
package segment

import (
	"novel-video-workflow/pkg/capcut/internal/metadata"
	"novel-video-workflow/pkg/capcut/internal/types"
)

// EffectSegment 放置在独立特效轨道上的特效片段
// 对应Python的Effect_segment类
type EffectSegment struct {
	*BaseSegment
	EffectInst *VideoEffect `json:"effect_inst"` // 相应的特效素材，在放入轨道时自动添加到素材列表中
}

// NewEffectSegment 创建新的特效片段
// 对应Python的Effect_segment.__init__方法
func NewEffectSegment(effectMeta metadata.EffectMeta, targetTimerange *types.Timerange, params []float64) (*EffectSegment, error) {
	// 解析参数
	parsedParams, err := effectMeta.ParseParams(params)
	if err != nil {
		return nil, err
	}

	// 创建VideoEffect实例，作用域为全局(apply_target_type=2)
	effectInst := NewVideoEffectFromMeta(effectMeta, parsedParams, 2)

	// 创建基础片段
	baseSegment := NewBaseSegment(effectInst.GlobalID, targetTimerange)

	return &EffectSegment{
		BaseSegment: baseSegment,
		EffectInst:  effectInst,
	}, nil
}

// NewVideoEffectFromMeta 从元数据创建VideoEffect
func NewVideoEffectFromMeta(effectMeta metadata.EffectMeta, params []metadata.EffectParamInstance, applyTargetType int) *VideoEffect {
	// 确定特效类型
	effectType := "video_effect" // 默认为视频特效
	// TODO: 根据具体的特效元数据类型来确定是"video_effect"还是"face_effect"
	// 这里需要根据实际的元数据类型进行判断

	effect := NewVideoEffect(effectMeta.Name, effectMeta.EffectID, effectMeta.ResourceID, effectType, applyTargetType)

	// 转换参数格式
	adjustParams := make([]interface{}, len(params))
	for i, param := range params {
		adjustParams[i] = param.ExportJSON()
	}
	effect.AdjustParams = adjustParams

	return effect
}

// ExportJSON 导出为JSON格式
func (es *EffectSegment) ExportJSON() map[string]interface{} {
	result := es.BaseSegment.ExportJSON()

	// 添加特效片段特有的字段
	result["effect_inst"] = es.EffectInst.ExportJSON()

	// 确保包含基础字段
	result["id"] = es.SegmentID
	result["material_id"] = es.MaterialID
	result["target_timerange"] = es.TargetTimerange.ExportJSON()

	return result
}

// GetMaterialRefs 获取素材引用
func (es *EffectSegment) GetMaterialRefs() []string {
	refs := es.BaseSegment.GetMaterialRefs()
	refs = append(refs, es.EffectInst.GlobalID)
	return refs
}

// FilterSegment 放置在独立滤镜轨道上的滤镜片段
// 对应Python的Filter_segment类
type FilterSegment struct {
	*BaseSegment
	Material *Filter `json:"material"` // 相应的滤镜素材，在放入轨道时自动添加到素材列表中
}

// NewFilterSegment 创建新的滤镜片段
// 对应Python的Filter_segment.__init__方法
func NewFilterSegment(filterMeta metadata.EffectMeta, targetTimerange *types.Timerange, intensity float64) *FilterSegment {
	// 创建Filter实例，作用域为全局(apply_target_type=2)
	filterInst := NewFilterFromMeta(filterMeta, intensity, 2)

	// 创建基础片段
	baseSegment := NewBaseSegment(filterInst.GlobalID, targetTimerange)

	return &FilterSegment{
		BaseSegment: baseSegment,
		Material:    filterInst,
	}
}

// NewFilterFromMeta 从元数据创建Filter
func NewFilterFromMeta(filterMeta metadata.EffectMeta, intensity float64, applyTargetType int) *Filter {
	// 强度范围转换：输入0-100，内部存储0-1
	normalizedIntensity := intensity / 100.0
	if normalizedIntensity < 0 {
		normalizedIntensity = 0
	}
	if normalizedIntensity > 1 {
		normalizedIntensity = 1
	}

	return NewFilter(filterMeta.Name, filterMeta.EffectID, filterMeta.ResourceID, normalizedIntensity, applyTargetType)
}

// ExportJSON 导出为JSON格式
func (fs *FilterSegment) ExportJSON() map[string]interface{} {
	result := fs.BaseSegment.ExportJSON()

	// 添加滤镜片段特有的字段
	result["material"] = fs.Material.ExportJSON()

	// 确保包含基础字段
	result["id"] = fs.SegmentID
	result["material_id"] = fs.MaterialID
	result["target_timerange"] = fs.TargetTimerange.ExportJSON()

	return result
}

// GetMaterialRefs 获取素材引用
func (fs *FilterSegment) GetMaterialRefs() []string {
	refs := fs.BaseSegment.GetMaterialRefs()
	refs = append(refs, fs.Material.GlobalID)
	return refs
}

// SetIntensity 设置滤镜强度
func (fs *FilterSegment) SetIntensity(intensity float64) {
	// 强度范围转换：输入0-100，内部存储0-1
	normalizedIntensity := intensity / 100.0
	if normalizedIntensity < 0 {
		normalizedIntensity = 0
	}
	if normalizedIntensity > 1 {
		normalizedIntensity = 1
	}
	fs.Material.Intensity = normalizedIntensity
}

// GetIntensity 获取滤镜强度（返回0-100范围）
func (fs *FilterSegment) GetIntensity() float64 {
	return fs.Material.Intensity * 100.0
}
