package segment

import (
	"testing"

	"novel-video-workflow/pkg/capcut/internal/metadata"
	"novel-video-workflow/pkg/capcut/internal/types"
)

// getKeys 获取map的所有键
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestNewEffectSegment 测试创建特效片段
func TestNewEffectSegment(t *testing.T) {
	// 创建测试用的特效元数据
	effectMeta := metadata.NewEffectMeta(
		"测试特效",
		false,
		"test_resource_123",
		"test_effect_456",
		"test_md5_hash",
		[]metadata.EffectParam{
			metadata.NewEffectParam("intensity", 1.0, 0.0, 1.0),
			metadata.NewEffectParam("brightness", 0.5, 0.0, 1.0),
		},
	)

	// 创建时间范围
	timerange := types.NewTimerange(1000000, 5000000) // 1秒开始，持续5秒

	// 测试参数
	params := []float64{80.0, 60.0} // 对应intensity=0.8, brightness=0.6

	// 创建特效片段
	effectSegment, err := NewEffectSegment(effectMeta, timerange, params)
	if err != nil {
		t.Fatalf("创建特效片段失败: %v", err)
	}

	// 验证基本属性
	if effectSegment.BaseSegment == nil {
		t.Error("BaseSegment未正确初始化")
	}

	if effectSegment.EffectInst == nil {
		t.Error("EffectInst未正确初始化")
	}

	// 验证特效实例属性
	if effectSegment.EffectInst.Name != "测试特效" {
		t.Errorf("期望特效名称为'测试特效'，得到'%s'", effectSegment.EffectInst.Name)
	}

	if effectSegment.EffectInst.EffectID != "test_effect_456" {
		t.Errorf("期望EffectID为'test_effect_456'，得到'%s'", effectSegment.EffectInst.EffectID)
	}

	if effectSegment.EffectInst.ResourceID != "test_resource_123" {
		t.Errorf("期望ResourceID为'test_resource_123'，得到'%s'", effectSegment.EffectInst.ResourceID)
	}

	if effectSegment.EffectInst.EffectType != "video_effect" {
		t.Errorf("期望EffectType为'video_effect'，得到'%s'", effectSegment.EffectInst.EffectType)
	}

	if effectSegment.EffectInst.ApplyTargetType != 2 {
		t.Errorf("期望ApplyTargetType为2，得到%d", effectSegment.EffectInst.ApplyTargetType)
	}

	// 验证时间范围
	if effectSegment.GetTargetTimerange().Start != 1000000 {
		t.Errorf("期望开始时间为1000000，得到%d", effectSegment.GetTargetTimerange().Start)
	}

	if effectSegment.GetTargetTimerange().Duration != 5000000 {
		t.Errorf("期望持续时间为5000000，得到%d", effectSegment.GetTargetTimerange().Duration)
	}
}

// TestNewEffectSegmentInvalidParams 测试无效参数的处理
func TestNewEffectSegmentInvalidParams(t *testing.T) {
	effectMeta := metadata.NewEffectMeta(
		"测试特效",
		false,
		"test_resource_123",
		"test_effect_456",
		"test_md5_hash",
		[]metadata.EffectParam{
			metadata.NewEffectParam("intensity", 1.0, 0.0, 1.0),
		},
	)

	timerange := types.NewTimerange(0, 3000000)

	// 测试超出范围的参数
	invalidParams := []float64{-10.0} // 负值应该被拒绝

	_, err := NewEffectSegment(effectMeta, timerange, invalidParams)
	if err == nil {
		t.Error("期望无效参数时返回错误")
	}

	// 测试超出范围的参数
	invalidParams2 := []float64{150.0} // 超过100应该被拒绝

	_, err = NewEffectSegment(effectMeta, timerange, invalidParams2)
	if err == nil {
		t.Error("期望超出范围参数时返回错误")
	}
}

// TestEffectSegmentExportJSON 测试特效片段JSON导出
func TestEffectSegmentExportJSON(t *testing.T) {
	effectMeta := metadata.NewEffectMeta(
		"JSON测试特效",
		false,
		"json_resource_123",
		"json_effect_456",
		"json_md5_hash",
		[]metadata.EffectParam{
			metadata.NewEffectParam("opacity", 0.8, 0.0, 1.0),
		},
	)

	timerange := types.NewTimerange(2000000, 4000000)
	params := []float64{75.0}

	effectSegment, err := NewEffectSegment(effectMeta, timerange, params)
	if err != nil {
		t.Fatalf("创建特效片段失败: %v", err)
	}

	// 导出JSON
	jsonData := effectSegment.ExportJSON()

	// 验证基础字段
	if jsonData["id"] != effectSegment.GetID() {
		t.Error("JSON中的ID不匹配")
	}

	// 验证时间范围
	if targetTimerange, ok := jsonData["target_timerange"].(map[string]int64); ok {
		if start, exists := targetTimerange["start"]; !exists || start != 2000000 {
			t.Error("JSON中的开始时间不正确")
		}
		if duration, exists := targetTimerange["duration"]; !exists || duration != 4000000 {
			t.Error("JSON中的持续时间不正确")
		}
	} else {
		t.Error("JSON中缺少target_timerange字段或类型不正确")
	}

	// 验证特效实例
	if _, ok := jsonData["effect_inst"]; !ok {
		t.Error("JSON中缺少effect_inst字段")
	}
}

// TestNewFilterSegment 测试创建滤镜片段
func TestNewFilterSegment(t *testing.T) {
	// 创建测试用的滤镜元数据
	filterMeta := metadata.NewEffectMeta(
		"测试滤镜",
		false,
		"filter_resource_123",
		"filter_effect_456",
		"filter_md5_hash",
		[]metadata.EffectParam{}, // 滤镜通常没有复杂参数
	)

	// 创建时间范围
	timerange := types.NewTimerange(500000, 8000000) // 0.5秒开始，持续8秒

	// 滤镜强度
	intensity := 75.0

	// 创建滤镜片段
	filterSegment := NewFilterSegment(filterMeta, timerange, intensity)

	// 验证基本属性
	if filterSegment.BaseSegment == nil {
		t.Error("BaseSegment未正确初始化")
	}

	if filterSegment.Material == nil {
		t.Error("Material未正确初始化")
	}

	// 验证滤镜实例属性
	if filterSegment.Material.Name != "测试滤镜" {
		t.Errorf("期望滤镜名称为'测试滤镜'，得到'%s'", filterSegment.Material.Name)
	}

	if filterSegment.Material.EffectID != "filter_effect_456" {
		t.Errorf("期望EffectID为'filter_effect_456'，得到'%s'", filterSegment.Material.EffectID)
	}

	if filterSegment.Material.ResourceID != "filter_resource_123" {
		t.Errorf("期望ResourceID为'filter_resource_123'，得到'%s'", filterSegment.Material.ResourceID)
	}

	if filterSegment.Material.ApplyTargetType != 2 {
		t.Errorf("期望ApplyTargetType为2，得到%d", filterSegment.Material.ApplyTargetType)
	}

	// 验证强度（内部存储为0-1范围）
	expectedIntensity := 0.75 // 75% -> 0.75
	if filterSegment.Material.Intensity != expectedIntensity {
		t.Errorf("期望内部强度为%.2f，得到%.2f", expectedIntensity, filterSegment.Material.Intensity)
	}

	// 验证时间范围
	if filterSegment.GetTargetTimerange().Start != 500000 {
		t.Errorf("期望开始时间为500000，得到%d", filterSegment.GetTargetTimerange().Start)
	}

	if filterSegment.GetTargetTimerange().Duration != 8000000 {
		t.Errorf("期望持续时间为8000000，得到%d", filterSegment.GetTargetTimerange().Duration)
	}
}

// TestFilterSegmentIntensityMethods 测试滤镜强度设置和获取方法
func TestFilterSegmentIntensityMethods(t *testing.T) {
	filterMeta := metadata.NewEffectMeta(
		"强度测试滤镜",
		false,
		"intensity_resource",
		"intensity_effect",
		"intensity_md5",
		[]metadata.EffectParam{},
	)

	timerange := types.NewTimerange(0, 1000000)
	filterSegment := NewFilterSegment(filterMeta, timerange, 50.0)

	// 测试初始强度
	if filterSegment.GetIntensity() != 50.0 {
		t.Errorf("期望初始强度为50.0，得到%.1f", filterSegment.GetIntensity())
	}

	// 测试设置强度
	filterSegment.SetIntensity(80.0)
	if filterSegment.GetIntensity() != 80.0 {
		t.Errorf("期望设置强度后为80.0，得到%.1f", filterSegment.GetIntensity())
	}

	// 测试边界值
	filterSegment.SetIntensity(0.0)
	if filterSegment.GetIntensity() != 0.0 {
		t.Errorf("期望最小强度为0.0，得到%.1f", filterSegment.GetIntensity())
	}

	filterSegment.SetIntensity(100.0)
	if filterSegment.GetIntensity() != 100.0 {
		t.Errorf("期望最大强度为100.0，得到%.1f", filterSegment.GetIntensity())
	}

	// 测试超出范围的值被限制
	filterSegment.SetIntensity(-10.0)
	if filterSegment.GetIntensity() != 0.0 {
		t.Errorf("期望负值被限制为0.0，得到%.1f", filterSegment.GetIntensity())
	}

	filterSegment.SetIntensity(150.0)
	if filterSegment.GetIntensity() != 100.0 {
		t.Errorf("期望超出范围值被限制为100.0，得到%.1f", filterSegment.GetIntensity())
	}
}

// TestFilterSegmentExportJSON 测试滤镜片段JSON导出
func TestFilterSegmentExportJSON(t *testing.T) {
	filterMeta := metadata.NewEffectMeta(
		"JSON滤镜",
		false,
		"json_filter_resource",
		"json_filter_effect",
		"json_filter_md5",
		[]metadata.EffectParam{},
	)

	timerange := types.NewTimerange(1500000, 6000000)
	filterSegment := NewFilterSegment(filterMeta, timerange, 90.0)

	// 导出JSON
	jsonData := filterSegment.ExportJSON()

	// 验证基础字段
	if jsonData["id"] != filterSegment.GetID() {
		t.Error("JSON中的ID不匹配")
	}

	// 验证时间范围
	if targetTimerange, ok := jsonData["target_timerange"].(map[string]int64); ok {
		if start, exists := targetTimerange["start"]; !exists || start != 1500000 {
			t.Error("JSON中的开始时间不正确")
		}
		if duration, exists := targetTimerange["duration"]; !exists || duration != 6000000 {
			t.Error("JSON中的持续时间不正确")
		}
	} else {
		t.Error("JSON中缺少target_timerange字段或类型不正确")
	}

	// 验证滤镜材质
	if _, ok := jsonData["material"]; !ok {
		t.Error("JSON中缺少material字段")
	}
}

// TestEffectSegmentGetMaterialRefs 测试特效片段的素材引用
func TestEffectSegmentGetMaterialRefs(t *testing.T) {
	effectMeta := metadata.NewEffectMeta(
		"引用测试特效",
		false,
		"ref_resource",
		"ref_effect",
		"ref_md5",
		[]metadata.EffectParam{},
	)

	timerange := types.NewTimerange(0, 1000000)
	effectSegment, err := NewEffectSegment(effectMeta, timerange, []float64{})
	if err != nil {
		t.Fatalf("创建特效片段失败: %v", err)
	}

	refs := effectSegment.GetMaterialRefs()

	// 应该包含特效实例的GlobalID
	found := false
	for _, ref := range refs {
		if ref == effectSegment.EffectInst.GlobalID {
			found = true
			break
		}
	}

	if !found {
		t.Error("素材引用列表中未找到特效实例的GlobalID")
	}
}

// TestFilterSegmentGetMaterialRefs 测试滤镜片段的素材引用
func TestFilterSegmentGetMaterialRefs(t *testing.T) {
	filterMeta := metadata.NewEffectMeta(
		"引用测试滤镜",
		false,
		"ref_filter_resource",
		"ref_filter_effect",
		"ref_filter_md5",
		[]metadata.EffectParam{},
	)

	timerange := types.NewTimerange(0, 1000000)
	filterSegment := NewFilterSegment(filterMeta, timerange, 50.0)

	refs := filterSegment.GetMaterialRefs()

	// 应该包含滤镜材质的GlobalID
	found := false
	for _, ref := range refs {
		if ref == filterSegment.Material.GlobalID {
			found = true
			break
		}
	}

	if !found {
		t.Error("素材引用列表中未找到滤镜材质的GlobalID")
	}
}

// TestEffectSegmentWithComplexParams 测试复杂参数的特效片段
func TestEffectSegmentWithComplexParams(t *testing.T) {
	effectMeta := metadata.NewEffectMeta(
		"复杂特效",
		true, // VIP特效
		"complex_resource",
		"complex_effect",
		"complex_md5",
		[]metadata.EffectParam{
			metadata.NewEffectParam("brightness", 0.5, 0.0, 1.0),
			metadata.NewEffectParam("contrast", 0.3, 0.0, 1.0),
			metadata.NewEffectParam("saturation", 0.8, 0.0, 1.0),
			metadata.NewEffectParam("hue", 0.0, -1.0, 1.0),
		},
	)

	timerange := types.NewTimerange(0, 10000000)
	params := []float64{70.0, 40.0, 90.0, 25.0} // 对应各个参数的百分比值

	effectSegment, err := NewEffectSegment(effectMeta, timerange, params)
	if err != nil {
		t.Fatalf("创建复杂特效片段失败: %v", err)
	}

	// 验证参数数量
	if len(effectSegment.EffectInst.AdjustParams) != 4 {
		t.Errorf("期望4个调整参数，得到%d个", len(effectSegment.EffectInst.AdjustParams))
	}

	// 导出JSON并验证结构
	jsonData := effectSegment.ExportJSON()
	if effectInst, ok := jsonData["effect_inst"].(map[string]interface{}); ok {
		if adjustParams, ok := effectInst["adjust_params"].([]interface{}); ok {
			if len(adjustParams) != 4 {
				t.Errorf("JSON中期望4个调整参数，得到%d个", len(adjustParams))
			}
		} else {
			t.Error("JSON中adjust_params字段格式不正确")
		}
	} else {
		t.Error("JSON中effect_inst字段格式不正确")
	}
}

// TestFilterSegmentEdgeCases 测试滤镜片段的边界情况
func TestFilterSegmentEdgeCases(t *testing.T) {
	filterMeta := metadata.NewEffectMeta(
		"边界测试滤镜",
		false,
		"edge_resource",
		"edge_effect",
		"edge_md5",
		[]metadata.EffectParam{},
	)

	// 测试极短时间范围
	shortTimerange := types.NewTimerange(0, 1) // 1微秒
	shortFilter := NewFilterSegment(filterMeta, shortTimerange, 100.0)

	if shortFilter.GetTargetTimerange().Duration != 1 {
		t.Error("极短时间范围处理不正确")
	}

	// 测试极长时间范围
	longTimerange := types.NewTimerange(0, 3600000000) // 1小时
	longFilter := NewFilterSegment(filterMeta, longTimerange, 0.0)

	if longFilter.GetTargetTimerange().Duration != 3600000000 {
		t.Error("极长时间范围处理不正确")
	}

	// 验证JSON导出在边界情况下不会出错
	shortJSON := shortFilter.ExportJSON()
	if shortJSON == nil {
		t.Error("极短滤镜片段JSON导出失败")
	}

	longJSON := longFilter.ExportJSON()
	if longJSON == nil {
		t.Error("极长滤镜片段JSON导出失败")
	}
}
