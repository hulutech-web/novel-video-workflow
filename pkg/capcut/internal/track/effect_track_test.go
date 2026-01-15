package track

import (
	"testing"

	"novel-video-workflow/pkg/capcut/internal/metadata"
	"novel-video-workflow/pkg/capcut/internal/segment"
	"novel-video-workflow/pkg/capcut/internal/types"
)

// getSegmentKeys 获取map的所有键
func getSegmentKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestEffectTrackCreation 测试特效轨道创建
func TestEffectTrackCreation(t *testing.T) {
	// 创建特效轨道
	effectTrack := NewTrack(TrackTypeEffect, "特效轨道1", 0, false)

	if effectTrack.TrackType != TrackTypeEffect {
		t.Errorf("期望轨道类型为TrackTypeEffect，得到%v", effectTrack.TrackType)
	}

	if effectTrack.Name != "特效轨道1" {
		t.Errorf("期望轨道名称为'特效轨道1'，得到'%s'", effectTrack.Name)
	}

	if effectTrack.RenderIndex != 10000 {
		t.Errorf("期望渲染索引为10000，得到%d", effectTrack.RenderIndex)
	}

	if effectTrack.Mute != false {
		t.Error("期望轨道未静音")
	}

	// 验证轨道接受的片段类型
	acceptedType := effectTrack.AcceptSegmentType()
	if acceptedType.String() != "*segment.EffectSegment" {
		t.Errorf("期望接受的片段类型为*segment.EffectSegment，得到%s", acceptedType.String())
	}
}

// TestFilterTrackCreation 测试滤镜轨道创建
func TestFilterTrackCreation(t *testing.T) {
	// 创建滤镜轨道
	filterTrack := NewTrack(TrackTypeFilter, "滤镜轨道1", 0, false)

	if filterTrack.TrackType != TrackTypeFilter {
		t.Errorf("期望轨道类型为TrackTypeFilter，得到%v", filterTrack.TrackType)
	}

	if filterTrack.Name != "滤镜轨道1" {
		t.Errorf("期望轨道名称为'滤镜轨道1'，得到'%s'", filterTrack.Name)
	}

	if filterTrack.RenderIndex != 11000 {
		t.Errorf("期望渲染索引为11000，得到%d", filterTrack.RenderIndex)
	}

	// 验证轨道接受的片段类型
	acceptedType := filterTrack.AcceptSegmentType()
	if acceptedType.String() != "*segment.FilterSegment" {
		t.Errorf("期望接受的片段类型为*segment.FilterSegment，得到%s", acceptedType.String())
	}
}

// TestAddEffectSegmentToTrack 测试向特效轨道添加特效片段
func TestAddEffectSegmentToTrack(t *testing.T) {
	// 创建特效轨道
	effectTrack := NewTrack(TrackTypeEffect, "测试特效轨道", 0, false)

	// 创建特效元数据
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

	// 创建特效片段
	timerange1 := types.NewTimerange(0, 5000000) // 0-5秒
	effectSegment1, err := segment.NewEffectSegment(effectMeta, timerange1, []float64{80.0})
	if err != nil {
		t.Fatalf("创建特效片段失败: %v", err)
	}

	// 添加特效片段到轨道
	err = effectTrack.AddSegment(effectSegment1)
	if err != nil {
		t.Fatalf("添加特效片段到轨道失败: %v", err)
	}

	// 验证片段已添加
	if len(effectTrack.Segments) != 1 {
		t.Errorf("期望轨道包含1个片段，得到%d个", len(effectTrack.Segments))
	}

	// 验证轨道结束时间
	expectedEndTime := int64(5000000)
	if effectTrack.EndTime() != expectedEndTime {
		t.Errorf("期望轨道结束时间为%d，得到%d", expectedEndTime, effectTrack.EndTime())
	}
}

// TestAddFilterSegmentToTrack 测试向滤镜轨道添加滤镜片段
func TestAddFilterSegmentToTrack(t *testing.T) {
	// 创建滤镜轨道
	filterTrack := NewTrack(TrackTypeFilter, "测试滤镜轨道", 0, false)

	// 创建滤镜元数据
	filterMeta := metadata.NewEffectMeta(
		"测试滤镜",
		false,
		"filter_resource_123",
		"filter_effect_456",
		"filter_md5_hash",
		[]metadata.EffectParam{},
	)

	// 创建滤镜片段
	timerange1 := types.NewTimerange(1000000, 8000000) // 1-9秒
	filterSegment1 := segment.NewFilterSegment(filterMeta, timerange1, 75.0)

	// 添加滤镜片段到轨道
	err := filterTrack.AddSegment(filterSegment1)
	if err != nil {
		t.Fatalf("添加滤镜片段到轨道失败: %v", err)
	}

	// 验证片段已添加
	if len(filterTrack.Segments) != 1 {
		t.Errorf("期望轨道包含1个片段，得到%d个", len(filterTrack.Segments))
	}

	// 验证轨道结束时间
	expectedEndTime := int64(9000000) // 1000000 + 8000000
	if filterTrack.EndTime() != expectedEndTime {
		t.Errorf("期望轨道结束时间为%d，得到%d", expectedEndTime, filterTrack.EndTime())
	}
}

// TestEffectSegmentOverlapPrevention 测试特效片段重叠检测
func TestEffectSegmentOverlapPrevention(t *testing.T) {
	effectTrack := NewTrack(TrackTypeEffect, "重叠测试轨道", 0, false)

	effectMeta := metadata.NewEffectMeta(
		"重叠测试特效",
		false,
		"overlap_resource",
		"overlap_effect",
		"overlap_md5",
		[]metadata.EffectParam{},
	)

	// 添加第一个特效片段 (0-5秒)
	timerange1 := types.NewTimerange(0, 5000000)
	effectSegment1, err := segment.NewEffectSegment(effectMeta, timerange1, []float64{})
	if err != nil {
		t.Fatalf("创建第一个特效片段失败: %v", err)
	}

	err = effectTrack.AddSegment(effectSegment1)
	if err != nil {
		t.Fatalf("添加第一个特效片段失败: %v", err)
	}

	// 尝试添加重叠的特效片段 (3-8秒)
	timerange2 := types.NewTimerange(3000000, 5000000)
	effectSegment2, err := segment.NewEffectSegment(effectMeta, timerange2, []float64{})
	if err != nil {
		t.Fatalf("创建第二个特效片段失败: %v", err)
	}

	err = effectTrack.AddSegment(effectSegment2)
	if err == nil {
		t.Error("期望添加重叠片段时返回错误")
	}

	// 验证只有一个片段
	if len(effectTrack.Segments) != 1 {
		t.Errorf("期望轨道包含1个片段，得到%d个", len(effectTrack.Segments))
	}

	// 添加非重叠的特效片段 (6-10秒)
	timerange3 := types.NewTimerange(6000000, 4000000)
	effectSegment3, err := segment.NewEffectSegment(effectMeta, timerange3, []float64{})
	if err != nil {
		t.Fatalf("创建第三个特效片段失败: %v", err)
	}

	err = effectTrack.AddSegment(effectSegment3)
	if err != nil {
		t.Fatalf("添加非重叠特效片段失败: %v", err)
	}

	// 验证现在有两个片段
	if len(effectTrack.Segments) != 2 {
		t.Errorf("期望轨道包含2个片段，得到%d个", len(effectTrack.Segments))
	}
}

// TestWrongSegmentTypeRejection 测试轨道拒绝错误类型的片段
func TestWrongSegmentTypeRejection(t *testing.T) {
	// 创建特效轨道
	effectTrack := NewTrack(TrackTypeEffect, "类型测试轨道", 0, false)

	// 尝试添加滤镜片段到特效轨道（应该被拒绝）
	filterMeta := metadata.NewEffectMeta(
		"错误类型滤镜",
		false,
		"wrong_resource",
		"wrong_effect",
		"wrong_md5",
		[]metadata.EffectParam{},
	)

	timerange := types.NewTimerange(0, 3000000)
	filterSegment := segment.NewFilterSegment(filterMeta, timerange, 50.0)

	err := effectTrack.AddSegment(filterSegment)
	if err == nil {
		t.Error("期望添加错误类型片段时返回错误")
	}

	// 验证片段未被添加
	if len(effectTrack.Segments) != 0 {
		t.Errorf("期望轨道包含0个片段，得到%d个", len(effectTrack.Segments))
	}
}

// TestEffectFilterTrackTypeFromName 测试从名称获取特效和滤镜轨道类型
func TestEffectFilterTrackTypeFromName(t *testing.T) {
	// 测试特效轨道类型
	effectType, err := TrackTypeFromName("effect")
	if err != nil {
		t.Fatalf("获取effect轨道类型失败: %v", err)
	}
	if effectType != TrackTypeEffect {
		t.Errorf("期望TrackTypeEffect，得到%v", effectType)
	}

	// 测试滤镜轨道类型
	filterType, err := TrackTypeFromName("filter")
	if err != nil {
		t.Fatalf("获取filter轨道类型失败: %v", err)
	}
	if filterType != TrackTypeFilter {
		t.Errorf("期望TrackTypeFilter，得到%v", filterType)
	}
}

// TestEffectFilterTrackTypeString 测试特效和滤镜轨道类型字符串表示
func TestEffectFilterTrackTypeString(t *testing.T) {
	if TrackTypeEffect.String() != "effect" {
		t.Errorf("期望TrackTypeEffect字符串为'effect'，得到'%s'", TrackTypeEffect.String())
	}

	if TrackTypeFilter.String() != "filter" {
		t.Errorf("期望TrackTypeFilter字符串为'filter'，得到'%s'", TrackTypeFilter.String())
	}
}

// TestGetTrackMeta 测试获取轨道元数据
func TestGetTrackMeta(t *testing.T) {
	// 测试特效轨道元数据
	effectMeta := GetTrackMeta(TrackTypeEffect)
	if effectMeta.RenderIndex != 10000 {
		t.Errorf("期望特效轨道渲染索引为10000，得到%d", effectMeta.RenderIndex)
	}
	if effectMeta.AllowModify != false {
		t.Error("期望特效轨道不允许修改")
	}

	// 测试滤镜轨道元数据
	filterMeta := GetTrackMeta(TrackTypeFilter)
	if filterMeta.RenderIndex != 11000 {
		t.Errorf("期望滤镜轨道渲染索引为11000，得到%d", filterMeta.RenderIndex)
	}
	if filterMeta.AllowModify != false {
		t.Error("期望滤镜轨道不允许修改")
	}
}

// TestTrackExportJSONWithEffects 测试包含特效的轨道JSON导出
func TestTrackExportJSONWithEffects(t *testing.T) {
	effectTrack := NewTrack(TrackTypeEffect, "JSON导出测试", 10000, false)

	// 添加一个特效片段
	effectMeta := metadata.NewEffectMeta(
		"JSON测试特效",
		false,
		"json_resource",
		"json_effect",
		"json_md5",
		[]metadata.EffectParam{
			metadata.NewEffectParam("brightness", 0.5, 0.0, 1.0),
		},
	)

	timerange := types.NewTimerange(2000000, 6000000)
	effectSegment, err := segment.NewEffectSegment(effectMeta, timerange, []float64{70.0})
	if err != nil {
		t.Fatalf("创建特效片段失败: %v", err)
	}

	err = effectTrack.AddSegment(effectSegment)
	if err != nil {
		t.Fatalf("添加特效片段失败: %v", err)
	}

	// 导出JSON
	jsonData := effectTrack.ExportJSON()

	// 验证基础轨道信息
	if jsonData["type"] != "effect" {
		t.Errorf("期望轨道类型为'effect'，得到'%v'", jsonData["type"])
	}

	if jsonData["name"] != "JSON导出测试" {
		t.Errorf("期望轨道名称为'JSON导出测试'，得到'%v'", jsonData["name"])
	}

	if jsonData["render_index"] != 10000 {
		t.Errorf("期望渲染索引为10000，得到%v", jsonData["render_index"])
	}

	// 验证片段列表
	if segments, ok := jsonData["segments"].([]interface{}); ok {
		if len(segments) != 1 {
			t.Errorf("期望1个片段，得到%d个", len(segments))
		}
		// 验证第一个片段是map类型
		if _, ok := segments[0].(map[string]interface{}); !ok {
			t.Errorf("第一个片段不是map类型，而是: %T", segments[0])
		}
	} else {
		t.Errorf("JSON中segments字段格式不正确，类型为: %T", jsonData["segments"])
	}
}
