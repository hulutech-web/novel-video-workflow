package segment

import (
	"encoding/json"
	"testing"

	"novel-video-workflow/pkg/capcut/internal/types"
)

func TestBaseSegment(t *testing.T) {
	// 测试基础片段创建
	timerange, _ := types.Trange("1s", "3s")
	segment := NewBaseSegment("material_123", timerange)

	// 验证基本属性
	if segment.MaterialID != "material_123" {
		t.Errorf("Expected MaterialID 'material_123', got '%s'", segment.MaterialID)
	}

	if segment.TargetTimerange.Start != 1000000 {
		t.Errorf("Expected start time 1000000, got %d", segment.TargetTimerange.Start)
	}

	if segment.TargetTimerange.Duration != 3000000 {
		t.Errorf("Expected duration 3000000, got %d", segment.TargetTimerange.Duration)
	}

	// 测试属性访问器
	if segment.Start() != 1000000 {
		t.Errorf("Expected Start() 1000000, got %d", segment.Start())
	}

	if segment.Duration() != 3000000 {
		t.Errorf("Expected Duration() 3000000, got %d", segment.Duration())
	}

	if segment.End() != 4000000 {
		t.Errorf("Expected End() 4000000, got %d", segment.End())
	}

	// 测试设置属性
	segment.SetStart(2000000)
	if segment.Start() != 2000000 {
		t.Errorf("Expected Start() after SetStart 2000000, got %d", segment.Start())
	}

	segment.SetDuration(5000000)
	if segment.Duration() != 5000000 {
		t.Errorf("Expected Duration() after SetDuration 5000000, got %d", segment.Duration())
	}

	// 测试重叠检测
	otherTimerange, _ := types.Trange("3s", "2s")
	other := NewBaseSegment("other_material", otherTimerange)
	if !segment.Overlaps(other) {
		t.Error("Expected segments to overlap")
	}

	nonOverlapTimerange, _ := types.Trange("10s", "2s")
	nonOverlapping := NewBaseSegment("non_overlap", nonOverlapTimerange)
	if segment.Overlaps(nonOverlapping) {
		t.Error("Expected segments not to overlap")
	}
}

func TestSpeed(t *testing.T) {
	speed := NewSpeed(1.5)

	if speed.Value != 1.5 {
		t.Errorf("Expected speed 1.5, got %f", speed.Value)
	}

	if speed.GlobalID == "" {
		t.Error("Expected non-empty GlobalID")
	}

	// 测试JSON导出
	jsonData := speed.ExportJSON()
	expectedFields := []string{"curve_speed", "id", "mode", "speed", "type"}
	for _, field := range expectedFields {
		if _, exists := jsonData[field]; !exists {
			t.Errorf("Expected field '%s' in JSON export", field)
		}
	}

	if jsonData["speed"] != 1.5 {
		t.Errorf("Expected JSON speed 1.5, got %v", jsonData["speed"])
	}

	if jsonData["type"] != "speed" {
		t.Errorf("Expected JSON type 'speed', got %v", jsonData["type"])
	}
}

func TestClipSettings(t *testing.T) {
	// 测试默认设置
	defaultClip := NewClipSettings()
	if defaultClip.Alpha != 1.0 {
		t.Errorf("Expected default alpha 1.0, got %f", defaultClip.Alpha)
	}

	if defaultClip.FlipHorizontal || defaultClip.FlipVertical {
		t.Error("Expected default flip values to be false")
	}

	// 测试自定义设置
	customClip := &ClipSettings{
		Alpha:          0.8,
		FlipHorizontal: true,
		FlipVertical:   false,
		Rotation:       45.0,
		ScaleX:         1.2,
		ScaleY:         0.8,
		TransformX:     0.1,
		TransformY:     -0.2,
	}

	// 测试JSON导出
	jsonData := customClip.ExportJSON()
	if jsonData["alpha"] != 0.8 {
		t.Errorf("Expected alpha 0.8, got %v", jsonData["alpha"])
	}

	flip := jsonData["flip"].(map[string]bool)
	if flip["horizontal"] != true {
		t.Error("Expected horizontal flip to be true")
	}

	if flip["vertical"] != false {
		t.Error("Expected vertical flip to be false")
	}

	if jsonData["rotation"] != 45.0 {
		t.Errorf("Expected rotation 45.0, got %v", jsonData["rotation"])
	}
}

func TestMediaSegment(t *testing.T) {
	sourceTimerange, _ := types.Trange("0s", "5s")
	targetTimerange, _ := types.Trange("2s", "3s")

	segment := NewMediaSegment("media_123", sourceTimerange, targetTimerange, 1.2, 0.8)

	if segment.Speed.Value != 1.2 {
		t.Errorf("Expected speed 1.2, got %f", segment.Speed.Value)
	}

	if segment.Volume != 0.8 {
		t.Errorf("Expected volume 0.8, got %f", segment.Volume)
	}

	if len(segment.ExtraMaterialRefs) != 1 {
		t.Errorf("Expected 1 extra material ref, got %d", len(segment.ExtraMaterialRefs))
	}

	// 测试JSON导出
	jsonData := segment.ExportJSON()
	if jsonData["volume"] != 0.8 {
		t.Errorf("Expected JSON volume 0.8, got %v", jsonData["volume"])
	}

	if jsonData["speed"] != 1.2 {
		t.Errorf("Expected JSON speed 1.2, got %v", jsonData["speed"])
	}
}

func TestVisualSegment(t *testing.T) {
	sourceTimerange, _ := types.Trange("0s", "4s")
	targetTimerange, _ := types.Trange("1s", "3s")
	clipSettings := &ClipSettings{
		Alpha:      0.9,
		Rotation:   30.0,
		ScaleX:     1.1,
		ScaleY:     1.1,
		TransformX: 0.05,
		TransformY: -0.1,
	}

	segment := NewVisualSegment("visual_123", sourceTimerange, targetTimerange, 1.0, 1.0, clipSettings)

	if segment.ClipSettings.Alpha != 0.9 {
		t.Errorf("Expected clip alpha 0.9, got %f", segment.ClipSettings.Alpha)
	}

	if !segment.UniformScale {
		t.Error("Expected uniform scale to be true by default")
	}

	// 测试JSON导出
	jsonData := segment.ExportJSON()
	clip := jsonData["clip"].(map[string]interface{})
	if clip["alpha"] != 0.9 {
		t.Errorf("Expected clip alpha 0.9, got %v", clip["alpha"])
	}

	uniformScale := jsonData["uniform_scale"].(map[string]interface{})
	if uniformScale["on"] != true {
		t.Error("Expected uniform_scale.on to be true")
	}
}

func TestSegmentJSONExport(t *testing.T) {
	// 创建一个完整的视觉片段并测试JSON导出
	sourceTimerange, _ := types.Trange("0s", "10s")
	targetTimerange, _ := types.Trange("5s", "8s")

	segment := NewVisualSegment("test_material", sourceTimerange, targetTimerange, 1.0, 0.8, nil)

	jsonData := segment.ExportJSON()

	// 验证JSON结构包含所有必需字段
	requiredFields := []string{
		"id", "material_id", "target_timerange", "source_timerange",
		"speed", "volume", "clip", "uniform_scale", "enable_adjust",
		"visible", "common_keyframes", "extra_material_refs",
	}

	for _, field := range requiredFields {
		if _, exists := jsonData[field]; !exists {
			t.Errorf("Missing required field '%s' in JSON export", field)
		}
	}

	// 验证可以序列化为JSON
	_, err := json.Marshal(jsonData)
	if err != nil {
		t.Errorf("Failed to marshal JSON: %v", err)
	}
}

func TestSegmentIDGeneration(t *testing.T) {
	// 测试每个片段都有唯一的ID
	timerange1, _ := types.Trange("0s", "1s")
	timerange2, _ := types.Trange("1s", "1s")
	segment1 := NewBaseSegment("material1", timerange1)
	segment2 := NewBaseSegment("material2", timerange2)

	if segment1.SegmentID == segment2.SegmentID {
		t.Error("Expected different segment IDs for different segments")
	}

	if len(segment1.SegmentID) != 36 {
		t.Errorf("Expected segment ID length 36, got %d", len(segment1.SegmentID))
	}
}
