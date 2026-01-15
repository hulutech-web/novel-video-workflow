package keyframe

import (
	"encoding/json"
	"testing"
)

func TestKeyframe(t *testing.T) {
	// 测试关键帧创建
	timeOffset := int64(2000000) // 2秒
	value := 0.8

	kf := NewKeyframe(timeOffset, value)

	if kf.TimeOffset != timeOffset {
		t.Errorf("Expected time offset %d, got %d", timeOffset, kf.TimeOffset)
	}

	if len(kf.Values) != 1 {
		t.Errorf("Expected 1 value, got %d", len(kf.Values))
	}

	if kf.Values[0] != value {
		t.Errorf("Expected value %f, got %f", value, kf.Values[0])
	}

	if kf.KfID == "" {
		t.Error("Expected non-empty keyframe ID")
	}

	if len(kf.KfID) != 32 {
		t.Errorf("Expected keyframe ID length 32, got %d", len(kf.KfID))
	}
}

func TestKeyframeExportJSON(t *testing.T) {
	kf := NewKeyframe(3000000, 1.5) // 3秒，值1.5

	jsonData := kf.ExportJSON()

	// 验证必要字段
	requiredFields := []string{"curveType", "graphID", "left_control", "right_control", "id", "time_offset", "values"}
	for _, field := range requiredFields {
		if _, exists := jsonData[field]; !exists {
			t.Errorf("Missing required field '%s' in JSON export", field)
		}
	}

	// 验证具体值
	if jsonData["curveType"] != "Line" {
		t.Errorf("Expected curveType 'Line', got %v", jsonData["curveType"])
	}

	if jsonData["time_offset"] != int64(3000000) {
		t.Errorf("Expected time_offset 3000000, got %v", jsonData["time_offset"])
	}

	values := jsonData["values"].([]float64)
	if len(values) != 1 || values[0] != 1.5 {
		t.Errorf("Expected values [1.5], got %v", values)
	}

	// 验证可以序列化为JSON
	_, err := json.Marshal(jsonData)
	if err != nil {
		t.Errorf("Failed to marshal keyframe JSON: %v", err)
	}
}

func TestKeyframeProperty(t *testing.T) {
	// 测试属性有效性
	validProperties := []KeyframeProperty{
		KeyframePropertyPositionX,
		KeyframePropertyPositionY,
		KeyframePropertyRotation,
		KeyframePropertyScaleX,
		KeyframePropertyScaleY,
		KeyframePropertyUniformScale,
		KeyframePropertyAlpha,
		KeyframePropertySaturation,
		KeyframePropertyContrast,
		KeyframePropertyBrightness,
		KeyframePropertyVolume,
	}

	for _, prop := range validProperties {
		if !prop.IsValid() {
			t.Errorf("Property %s should be valid", prop)
		}
	}

	// 测试无效属性
	invalidProp := KeyframeProperty("INVALID_PROPERTY")
	if invalidProp.IsValid() {
		t.Error("Invalid property should not be valid")
	}
}

func TestKeyframePropertyFromString(t *testing.T) {
	// 测试从字符串创建属性
	testCases := []struct {
		input    string
		expected KeyframeProperty
	}{
		{"position_x", KeyframePropertyPositionX},
		{"position_y", KeyframePropertyPositionY},
		{"rotation", KeyframePropertyRotation},
		{"scale_x", KeyframePropertyScaleX},
		{"scale_y", KeyframePropertyScaleY},
		{"uniform_scale", KeyframePropertyUniformScale},
		{"alpha", KeyframePropertyAlpha},
		{"saturation", KeyframePropertySaturation},
		{"contrast", KeyframePropertyContrast},
		{"brightness", KeyframePropertyBrightness},
		{"volume", KeyframePropertyVolume},
	}

	for _, tc := range testCases {
		prop, err := KeyframePropertyFromString(tc.input)
		if err != nil {
			t.Errorf("Failed to parse property '%s': %v", tc.input, err)
		}
		if prop != tc.expected {
			t.Errorf("Expected property %s, got %s", tc.expected, prop)
		}
	}

	// 测试无效字符串
	_, err := KeyframePropertyFromString("invalid_property")
	if err == nil {
		t.Error("Expected error for invalid property string")
	}
}

func TestKeyframeList(t *testing.T) {
	// 测试关键帧列表创建
	kfl := NewKeyframeList(KeyframePropertyAlpha)

	if kfl.KeyframeProperty != KeyframePropertyAlpha {
		t.Errorf("Expected property %s, got %s", KeyframePropertyAlpha, kfl.KeyframeProperty)
	}

	if len(kfl.Keyframes) != 0 {
		t.Errorf("Expected empty keyframes list, got %d", len(kfl.Keyframes))
	}

	if kfl.ListID == "" {
		t.Error("Expected non-empty list ID")
	}
}

func TestKeyframeListAddKeyframe(t *testing.T) {
	kfl := NewKeyframeList(KeyframePropertyAlpha)

	// 添加关键帧（乱序添加测试排序功能）
	kfl.AddKeyframe(3000000, 0.5) // 3秒
	kfl.AddKeyframe(1000000, 1.0) // 1秒
	kfl.AddKeyframe(5000000, 0.0) // 5秒

	if len(kfl.Keyframes) != 3 {
		t.Errorf("Expected 3 keyframes, got %d", len(kfl.Keyframes))
	}

	// 验证排序
	expectedTimes := []int64{1000000, 3000000, 5000000}
	for i, expectedTime := range expectedTimes {
		if kfl.Keyframes[i].TimeOffset != expectedTime {
			t.Errorf("Expected keyframe %d time %d, got %d", i, expectedTime, kfl.Keyframes[i].TimeOffset)
		}
	}
}

func TestKeyframeListGetValueAt(t *testing.T) {
	kfl := NewKeyframeList(KeyframePropertyAlpha)

	// 测试空列表返回默认值
	defaultValue := kfl.GetValueAt(2000000)
	if defaultValue != 1.0 { // alpha默认值为1.0
		t.Errorf("Expected default value 1.0, got %f", defaultValue)
	}

	// 添加关键帧
	kfl.AddKeyframe(1000000, 0.0) // 1秒，值0.0
	kfl.AddKeyframe(3000000, 1.0) // 3秒，值1.0

	// 测试精确匹配
	if value := kfl.GetValueAt(1000000); value != 0.0 {
		t.Errorf("Expected exact match value 0.0, got %f", value)
	}

	// 测试线性插值
	if value := kfl.GetValueAt(2000000); value != 0.5 {
		t.Errorf("Expected interpolated value 0.5, got %f", value)
	}

	// 测试边界情况
	if value := kfl.GetValueAt(500000); value != 0.0 { // 在第一个关键帧之前
		t.Errorf("Expected boundary value 0.0, got %f", value)
	}

	if value := kfl.GetValueAt(4000000); value != 1.0 { // 在最后一个关键帧之后
		t.Errorf("Expected boundary value 1.0, got %f", value)
	}
}

func TestKeyframeListRemoveKeyframe(t *testing.T) {
	kfl := NewKeyframeList(KeyframePropertyAlpha)

	kfl.AddKeyframe(1000000, 0.0)
	kfl.AddKeyframe(2000000, 0.5)
	kfl.AddKeyframe(3000000, 1.0)

	// 移除中间的关键帧
	err := kfl.RemoveKeyframe(1)
	if err != nil {
		t.Errorf("Failed to remove keyframe: %v", err)
	}

	if len(kfl.Keyframes) != 2 {
		t.Errorf("Expected 2 keyframes after removal, got %d", len(kfl.Keyframes))
	}

	// 验证剩余关键帧
	if kfl.Keyframes[0].TimeOffset != 1000000 || kfl.Keyframes[1].TimeOffset != 3000000 {
		t.Error("Wrong keyframes remaining after removal")
	}

	// 测试无效索引
	err = kfl.RemoveKeyframe(10)
	if err == nil {
		t.Error("Expected error for invalid index")
	}
}

func TestKeyframeListExportJSON(t *testing.T) {
	kfl := NewKeyframeList(KeyframePropertyAlpha)
	kfl.AddKeyframe(1000000, 0.5)
	kfl.AddKeyframe(3000000, 1.0)

	jsonData := kfl.ExportJSON()

	// 验证必要字段
	requiredFields := []string{"id", "keyframe_list", "material_id", "property_type"}
	for _, field := range requiredFields {
		if _, exists := jsonData[field]; !exists {
			t.Errorf("Missing required field '%s' in JSON export", field)
		}
	}

	if jsonData["property_type"] != string(KeyframePropertyAlpha) {
		t.Errorf("Expected property_type '%s', got %v", KeyframePropertyAlpha, jsonData["property_type"])
	}

	keyframeList := jsonData["keyframe_list"].([]map[string]interface{})
	if len(keyframeList) != 2 {
		t.Errorf("Expected 2 keyframes in JSON, got %d", len(keyframeList))
	}

	// 验证可以序列化为JSON
	_, err := json.Marshal(jsonData)
	if err != nil {
		t.Errorf("Failed to marshal keyframe list JSON: %v", err)
	}
}

func TestParseValue(t *testing.T) {
	// 测试位置值解析
	value, err := ParseValue(KeyframePropertyPositionX, "0.5")
	if err != nil || value != 0.5 {
		t.Errorf("Failed to parse position value: %v, got %f", err, value)
	}

	// 测试位置值范围检查
	_, err = ParseValue(KeyframePropertyPositionX, "15")
	if err == nil {
		t.Error("Expected error for position value out of range")
	}

	// 测试旋转角度解析
	value, err = ParseValue(KeyframePropertyRotation, "45deg")
	if err != nil || value != 45.0 {
		t.Errorf("Failed to parse rotation value: %v, got %f", err, value)
	}

	// 测试百分比解析
	value, err = ParseValue(KeyframePropertyAlpha, "80%")
	if err != nil || value != 0.8 {
		t.Errorf("Failed to parse percentage value: %v, got %f", err, value)
	}

	// 测试正负号解析
	value, err = ParseValue(KeyframePropertyBrightness, "+0.3")
	if err != nil || value != 0.3 {
		t.Errorf("Failed to parse positive value: %v, got %f", err, value)
	}

	value, err = ParseValue(KeyframePropertyBrightness, "-0.2")
	if err != nil || value != -0.2 {
		t.Errorf("Failed to parse negative value: %v, got %f", err, value)
	}
}

func TestKeyframeManager(t *testing.T) {
	// 测试关键帧管理器创建
	km := NewKeyframeManager()

	if km.HasKeyframes() {
		t.Error("New keyframe manager should not have keyframes")
	}

	// 添加关键帧
	km.AddKeyframe(KeyframePropertyAlpha, 1000000, 0.5)
	km.AddKeyframe(KeyframePropertyRotation, 2000000, 45.0)

	if !km.HasKeyframes() {
		t.Error("Keyframe manager should have keyframes after adding")
	}

	// 获取关键帧列表
	alphaList := km.GetKeyframeList(KeyframePropertyAlpha)
	if alphaList == nil {
		t.Error("Expected alpha keyframe list to exist")
	}

	if len(alphaList.Keyframes) != 1 {
		t.Errorf("Expected 1 alpha keyframe, got %d", len(alphaList.Keyframes))
	}

	// 获取所有关键帧列表
	allLists := km.GetAllKeyframeLists()
	if len(allLists) != 2 {
		t.Errorf("Expected 2 keyframe lists, got %d", len(allLists))
	}
}

func TestKeyframeManagerFromString(t *testing.T) {
	km := NewKeyframeManager()

	// 从字符串添加关键帧
	err := km.AddKeyframeFromString("alpha", 1000000, "80%")
	if err != nil {
		t.Errorf("Failed to add keyframe from string: %v", err)
	}

	err = km.AddKeyframeFromString("rotation", 2000000, "45deg")
	if err != nil {
		t.Errorf("Failed to add rotation keyframe from string: %v", err)
	}

	// 验证添加结果
	alphaList := km.GetKeyframeList(KeyframePropertyAlpha)
	if alphaList == nil || len(alphaList.Keyframes) != 1 {
		t.Error("Alpha keyframe not added correctly")
	}

	if alphaList.Keyframes[0].Values[0] != 0.8 {
		t.Errorf("Expected alpha value 0.8, got %f", alphaList.Keyframes[0].Values[0])
	}

	// 测试无效属性
	err = km.AddKeyframeFromString("invalid_property", 1000000, "1.0")
	if err == nil {
		t.Error("Expected error for invalid property")
	}
}

func TestKeyframeManagerExportJSON(t *testing.T) {
	km := NewKeyframeManager()

	km.AddKeyframe(KeyframePropertyAlpha, 1000000, 0.5)
	km.AddKeyframe(KeyframePropertyRotation, 2000000, 45.0)

	jsonData := km.ExportJSON()

	if len(jsonData) != 2 {
		t.Errorf("Expected 2 keyframe lists in JSON, got %d", len(jsonData))
	}

	// 验证可以序列化为JSON
	_, err := json.Marshal(jsonData)
	if err != nil {
		t.Errorf("Failed to marshal keyframe manager JSON: %v", err)
	}
}

func TestKeyframeManagerClear(t *testing.T) {
	km := NewKeyframeManager()

	km.AddKeyframe(KeyframePropertyAlpha, 1000000, 0.5)
	km.AddKeyframe(KeyframePropertyRotation, 2000000, 45.0)

	if !km.HasKeyframes() {
		t.Error("Manager should have keyframes before clear")
	}

	km.Clear()

	if km.HasKeyframes() {
		t.Error("Manager should not have keyframes after clear")
	}

	if len(km.GetAllKeyframeLists()) != 0 {
		t.Error("All keyframe lists should be cleared")
	}
}

func TestDefaultValues(t *testing.T) {
	// 测试不同属性的默认值
	testCases := []struct {
		property        KeyframeProperty
		expectedDefault float64
	}{
		{KeyframePropertyPositionX, 0.0},
		{KeyframePropertyPositionY, 0.0},
		{KeyframePropertyRotation, 0.0},
		{KeyframePropertyScaleX, 1.0},
		{KeyframePropertyScaleY, 1.0},
		{KeyframePropertyUniformScale, 1.0},
		{KeyframePropertyAlpha, 1.0},
		{KeyframePropertyVolume, 1.0},
		{KeyframePropertySaturation, 0.0},
		{KeyframePropertyContrast, 0.0},
		{KeyframePropertyBrightness, 0.0},
	}

	for _, tc := range testCases {
		kfl := NewKeyframeList(tc.property)
		defaultValue := kfl.GetValueAt(1000000) // 任意时间点
		if defaultValue != tc.expectedDefault {
			t.Errorf("Property %s expected default %f, got %f", tc.property, tc.expectedDefault, defaultValue)
		}
	}
}
