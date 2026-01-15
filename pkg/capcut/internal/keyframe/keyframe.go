// Package keyframe 定义关键帧动画系统
// 对应Python的keyframe.py
package keyframe

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Keyframe 一个关键帧（关键点），目前只支持线性插值
// 对应Python的Keyframe类
type Keyframe struct {
	KfID       string    `json:"id"`          // 关键帧全局id，自动生成
	TimeOffset int64     `json:"time_offset"` // 相对于素材起始点的时间偏移量（微秒）
	Values     []float64 `json:"values"`      // 关键帧的值，似乎一般只有一个元素
}

// NewKeyframe 创建新的关键帧
func NewKeyframe(timeOffset int64, value float64) *Keyframe {
	return &Keyframe{
		KfID:       strings.ReplaceAll(uuid.New().String(), "-", ""),
		TimeOffset: timeOffset,
		Values:     []float64{value},
	}
}

// ExportJSON 导出为JSON格式
func (kf *Keyframe) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		// 默认值
		"curveType":     "Line",
		"graphID":       "",
		"left_control":  map[string]float64{"x": 0.0, "y": 0.0},
		"right_control": map[string]float64{"x": 0.0, "y": 0.0},
		// 自定义属性
		"id":          kf.KfID,
		"time_offset": kf.TimeOffset,
		"values":      kf.Values,
	}
}

// KeyframeProperty 关键帧所控制的属性类型
// 对应Python的Keyframe_property枚举
type KeyframeProperty string

const (
	// 位置相关
	KeyframePropertyPositionX KeyframeProperty = "KFTypePositionX" // 右移为正，此处的数值应该为`剪映中显示的值` / `草稿宽度`，也即单位是半个画布宽
	KeyframePropertyPositionY KeyframeProperty = "KFTypePositionY" // 上移为正，此处的数值应该为`剪映中显示的值` / `草稿高度`，也即单位是半个画布高
	KeyframePropertyRotation  KeyframeProperty = "KFTypeRotation"  // 顺时针旋转的**角度**

	// 缩放相关
	KeyframePropertyScaleX       KeyframeProperty = "KFTypeScaleX"  // 单独控制X轴缩放比例(1.0为不缩放)，与`uniform_scale`互斥
	KeyframePropertyScaleY       KeyframeProperty = "KFTypeScaleY"  // 单独控制Y轴缩放比例(1.0为不缩放)，与`uniform_scale`互斥
	KeyframePropertyUniformScale KeyframeProperty = "UNIFORM_SCALE" // 同时控制X轴及Y轴缩放比例(1.0为不缩放)，与`scale_x`和`scale_y`互斥

	// 视觉效果相关
	KeyframePropertyAlpha      KeyframeProperty = "KFTypeAlpha"      // 不透明度，1.0为完全不透明，仅对`Video_segment`有效
	KeyframePropertySaturation KeyframeProperty = "KFTypeSaturation" // 饱和度，0.0为原始饱和度，范围为-1.0到1.0，仅对`Video_segment`有效
	KeyframePropertyContrast   KeyframeProperty = "KFTypeContrast"   // 对比度，0.0为原始对比度，范围为-1.0到1.0，仅对`Video_segment`有效
	KeyframePropertyBrightness KeyframeProperty = "KFTypeBrightness" // 亮度，0.0为原始亮度，范围为-1.0到1.0，仅对`Video_segment`有效

	// 音频相关
	KeyframePropertyVolume KeyframeProperty = "KFTypeVolume" // 音量，1.0为原始音量，仅对`Audio_segment`和`Video_segment`有效
)

// String 实现Stringer接口
func (kp KeyframeProperty) String() string {
	return string(kp)
}

// IsValid 检查关键帧属性是否有效
func (kp KeyframeProperty) IsValid() bool {
	switch kp {
	case KeyframePropertyPositionX, KeyframePropertyPositionY, KeyframePropertyRotation,
		KeyframePropertyScaleX, KeyframePropertyScaleY, KeyframePropertyUniformScale,
		KeyframePropertyAlpha, KeyframePropertySaturation, KeyframePropertyContrast,
		KeyframePropertyBrightness, KeyframePropertyVolume:
		return true
	default:
		return false
	}
}

// KeyframePropertyFromString 从字符串创建关键帧属性
func KeyframePropertyFromString(s string) (KeyframeProperty, error) {
	property := KeyframeProperty(s)
	if property.IsValid() {
		return property, nil
	}

	// 尝试从简化名称匹配
	switch s {
	case "position_x":
		return KeyframePropertyPositionX, nil
	case "position_y":
		return KeyframePropertyPositionY, nil
	case "rotation":
		return KeyframePropertyRotation, nil
	case "scale_x":
		return KeyframePropertyScaleX, nil
	case "scale_y":
		return KeyframePropertyScaleY, nil
	case "uniform_scale":
		return KeyframePropertyUniformScale, nil
	case "alpha":
		return KeyframePropertyAlpha, nil
	case "saturation":
		return KeyframePropertySaturation, nil
	case "contrast":
		return KeyframePropertyContrast, nil
	case "brightness":
		return KeyframePropertyBrightness, nil
	case "volume":
		return KeyframePropertyVolume, nil
	default:
		return "", fmt.Errorf("unsupported keyframe property type: %s", s)
	}
}

// KeyframeList 关键帧列表，记录与某个特定属性相关的一系列关键帧
// 对应Python的Keyframe_list类
type KeyframeList struct {
	ListID           string           `json:"id"`            // 关键帧列表全局id，自动生成
	KeyframeProperty KeyframeProperty `json:"property_type"` // 关键帧对应的属性
	Keyframes        []*Keyframe      `json:"keyframe_list"` // 关键帧列表
	MaterialID       string           `json:"material_id"`   // 素材ID，通常为空字符串
}

// NewKeyframeList 为给定的关键帧属性初始化关键帧列表
func NewKeyframeList(keyframeProperty KeyframeProperty) *KeyframeList {
	return &KeyframeList{
		ListID:           strings.ReplaceAll(uuid.New().String(), "-", ""),
		KeyframeProperty: keyframeProperty,
		Keyframes:        make([]*Keyframe, 0),
		MaterialID:       "",
	}
}

// AddKeyframe 给定时间偏移量及关键值，向此关键帧列表中添加一个关键帧
func (kfl *KeyframeList) AddKeyframe(timeOffset int64, value float64) {
	keyframe := NewKeyframe(timeOffset, value)
	kfl.Keyframes = append(kfl.Keyframes, keyframe)

	// 按时间偏移量排序
	sort.Slice(kfl.Keyframes, func(i, j int) bool {
		return kfl.Keyframes[i].TimeOffset < kfl.Keyframes[j].TimeOffset
	})
}

// RemoveKeyframe 移除指定索引的关键帧
func (kfl *KeyframeList) RemoveKeyframe(index int) error {
	if index < 0 || index >= len(kfl.Keyframes) {
		return fmt.Errorf("index out of range: %d", index)
	}

	kfl.Keyframes = append(kfl.Keyframes[:index], kfl.Keyframes[index+1:]...)
	return nil
}

// GetKeyframeAt 获取指定时间偏移量的关键帧
func (kfl *KeyframeList) GetKeyframeAt(timeOffset int64) *Keyframe {
	for _, kf := range kfl.Keyframes {
		if kf.TimeOffset == timeOffset {
			return kf
		}
	}
	return nil
}

// GetValueAt 获取指定时间偏移量的插值结果
// 如果没有关键帧，返回默认值
// 如果只有一个关键帧，返回该关键帧的值
// 如果有多个关键帧，使用线性插值
func (kfl *KeyframeList) GetValueAt(timeOffset int64) float64 {
	if len(kfl.Keyframes) == 0 {
		return kfl.getDefaultValue()
	}

	if len(kfl.Keyframes) == 1 {
		return kfl.Keyframes[0].Values[0]
	}

	// 找到时间范围
	var before, after *Keyframe
	for _, kf := range kfl.Keyframes {
		if kf.TimeOffset == timeOffset {
			return kf.Values[0] // 精确匹配
		}

		if kf.TimeOffset < timeOffset {
			before = kf
		} else if after == nil {
			after = kf
			break
		}
	}

	// 边界情况处理
	if before == nil {
		return kfl.Keyframes[0].Values[0] // 在第一个关键帧之前
	}
	if after == nil {
		return kfl.Keyframes[len(kfl.Keyframes)-1].Values[0] // 在最后一个关键帧之后
	}

	// 线性插值
	ratio := float64(timeOffset-before.TimeOffset) / float64(after.TimeOffset-before.TimeOffset)
	return before.Values[0] + ratio*(after.Values[0]-before.Values[0])
}

// getDefaultValue 获取属性的默认值
func (kfl *KeyframeList) getDefaultValue() float64 {
	switch kfl.KeyframeProperty {
	case KeyframePropertyPositionX, KeyframePropertyPositionY:
		return 0.0 // 位置默认居中
	case KeyframePropertyRotation:
		return 0.0 // 旋转默认为0度
	case KeyframePropertyScaleX, KeyframePropertyScaleY, KeyframePropertyUniformScale:
		return 1.0 // 缩放默认为1.0（不缩放）
	case KeyframePropertyAlpha, KeyframePropertyVolume:
		return 1.0 // 透明度和音量默认为1.0（完全不透明/原始音量）
	case KeyframePropertySaturation, KeyframePropertyContrast, KeyframePropertyBrightness:
		return 0.0 // 饱和度、对比度、亮度默认为0.0（原始值）
	default:
		return 0.0
	}
}

// ExportJSON 导出为JSON格式
func (kfl *KeyframeList) ExportJSON() map[string]interface{} {
	keyframeList := make([]map[string]interface{}, 0, len(kfl.Keyframes))
	for _, kf := range kfl.Keyframes {
		keyframeList = append(keyframeList, kf.ExportJSON())
	}

	return map[string]interface{}{
		"id":            kfl.ListID,
		"keyframe_list": keyframeList,
		"material_id":   kfl.MaterialID,
		"property_type": string(kfl.KeyframeProperty),
	}
}

// ParseValue 解析字符串值为float64
// 支持各种格式：百分比、角度、位置等
func ParseValue(propertyType KeyframeProperty, value string) (float64, error) {
	switch propertyType {
	case KeyframePropertyPositionX, KeyframePropertyPositionY:
		// 位置值，范围[-10, 10]
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid position value: %s", value)
		}
		if floatValue < -10 || floatValue > 10 {
			return 0, fmt.Errorf("position value %f out of range [-10, 10]", floatValue)
		}
		return floatValue, nil

	case KeyframePropertyRotation:
		// 旋转角度
		if strings.HasSuffix(value, "deg") {
			return strconv.ParseFloat(value[:len(value)-3], 64)
		}
		return strconv.ParseFloat(value, 64)

	case KeyframePropertyAlpha, KeyframePropertyVolume:
		// 透明度和音量，支持百分比格式
		if strings.HasSuffix(value, "%") {
			percentValue, err := strconv.ParseFloat(value[:len(value)-1], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid percentage value: %s", value)
			}
			return percentValue / 100.0, nil
		}
		return strconv.ParseFloat(value, 64)

	case KeyframePropertySaturation, KeyframePropertyContrast, KeyframePropertyBrightness:
		// 饱和度、对比度、亮度，支持正负号
		if strings.HasPrefix(value, "+") {
			return strconv.ParseFloat(value[1:], 64)
		} else if strings.HasPrefix(value, "-") {
			negValue, err := strconv.ParseFloat(value[1:], 64)
			if err != nil {
				return 0, err
			}
			return -negValue, nil
		}
		return strconv.ParseFloat(value, 64)

	default:
		// 其他属性直接转换为float
		return strconv.ParseFloat(value, 64)
	}
}

// KeyframeManager 关键帧管理器，管理所有属性的关键帧列表
type KeyframeManager struct {
	keyframeLists map[KeyframeProperty]*KeyframeList
}

// NewKeyframeManager 创建新的关键帧管理器
func NewKeyframeManager() *KeyframeManager {
	return &KeyframeManager{
		keyframeLists: make(map[KeyframeProperty]*KeyframeList),
	}
}

// AddKeyframe 添加关键帧到指定属性
func (km *KeyframeManager) AddKeyframe(property KeyframeProperty, timeOffset int64, value float64) {
	if !property.IsValid() {
		return
	}

	list, exists := km.keyframeLists[property]
	if !exists {
		list = NewKeyframeList(property)
		km.keyframeLists[property] = list
	}

	list.AddKeyframe(timeOffset, value)
}

// AddKeyframeFromString 从字符串添加关键帧
func (km *KeyframeManager) AddKeyframeFromString(propertyName string, timeOffset int64, value string) error {
	property, err := KeyframePropertyFromString(propertyName)
	if err != nil {
		return err
	}

	floatValue, err := ParseValue(property, value)
	if err != nil {
		return err
	}

	km.AddKeyframe(property, timeOffset, floatValue)
	return nil
}

// GetKeyframeList 获取指定属性的关键帧列表
func (km *KeyframeManager) GetKeyframeList(property KeyframeProperty) *KeyframeList {
	return km.keyframeLists[property]
}

// GetAllKeyframeLists 获取所有关键帧列表
func (km *KeyframeManager) GetAllKeyframeLists() []*KeyframeList {
	lists := make([]*KeyframeList, 0, len(km.keyframeLists))
	for _, list := range km.keyframeLists {
		lists = append(lists, list)
	}
	return lists
}

// RemoveKeyframeList 移除指定属性的关键帧列表
func (km *KeyframeManager) RemoveKeyframeList(property KeyframeProperty) {
	delete(km.keyframeLists, property)
}

// Clear 清空所有关键帧列表
func (km *KeyframeManager) Clear() {
	km.keyframeLists = make(map[KeyframeProperty]*KeyframeList)
}

// HasKeyframes 检查是否有关键帧
func (km *KeyframeManager) HasKeyframes() bool {
	return len(km.keyframeLists) > 0
}

// ExportJSON 导出所有关键帧列表为JSON格式
func (km *KeyframeManager) ExportJSON() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(km.keyframeLists))
	for _, list := range km.keyframeLists {
		result = append(result, list.ExportJSON())
	}
	return result
}
