package segment

import (
	"encoding/json"
	"testing"

	"novel-video-workflow/pkg/capcut/internal/types"
)

func TestVideoSegment(t *testing.T) {
	sourceTimerange, _ := types.Trange("0s", "10s")
	targetTimerange, _ := types.Trange("2s", "5s")

	videoSegment := NewVideoSegment("video_123", sourceTimerange, targetTimerange, 1.0, 0.9, nil)

	// 验证基本属性
	if videoSegment.MaterialID != "video_123" {
		t.Errorf("Expected MaterialID 'video_123', got '%s'", videoSegment.MaterialID)
	}

	if videoSegment.Volume != 0.9 {
		t.Errorf("Expected volume 0.9, got %f", videoSegment.Volume)
	}

	// 测试JSON导出
	jsonData := videoSegment.ExportJSON()
	if jsonData["material_id"] != "video_123" {
		t.Errorf("Expected material_id 'video_123', got %v", jsonData["material_id"])
	}

	// 验证JSON可以序列化
	_, err := json.Marshal(jsonData)
	if err != nil {
		t.Errorf("Failed to marshal VideoSegment JSON: %v", err)
	}
}

func TestMask(t *testing.T) {
	mask := NewMask("mask_material_123", "image", "res_123", 0.0, 0.0, 1.0, 1.0, 1.0, 0.0, 0.0, 0.0, true)

	if mask.Name != "mask_material_123" {
		t.Errorf("Expected Name 'mask_material_123', got '%s'", mask.Name)
	}

	if mask.ResourceID != "res_123" {
		t.Errorf("Expected ResourceID 'res_123', got '%s'", mask.ResourceID)
	}

	if !mask.Invert {
		t.Error("Expected invert to be true")
	}

	if mask.Feather != 0.0 {
		t.Error("Expected feather to be 0.0")
	}

	// 测试JSON导出
	jsonData := mask.ExportJSON()
	if jsonData["name"] != "mask_material_123" {
		t.Errorf("Expected JSON name 'mask_material_123', got %v", jsonData["name"])
	}

	config := jsonData["config"].(map[string]interface{})
	if config["invert"] != true {
		t.Error("Expected JSON config invert to be true")
	}
}

func TestVideoEffect(t *testing.T) {
	effect := NewVideoEffect("cool_effect", "effect_123", "res_456", "video_effect", 0)

	if effect.EffectID != "effect_123" {
		t.Errorf("Expected EffectID 'effect_123', got '%s'", effect.EffectID)
	}

	if effect.ResourceID != "res_456" {
		t.Errorf("Expected ResourceID 'res_456', got '%s'", effect.ResourceID)
	}

	if effect.ApplyTargetType != 0 {
		t.Errorf("Expected ApplyTargetType 0, got %d", effect.ApplyTargetType)
	}

	// 测试JSON导出
	jsonData := effect.ExportJSON()
	if jsonData["effect_id"] != "effect_123" {
		t.Errorf("Expected JSON effect_id 'effect_123', got %v", jsonData["effect_id"])
	}

	if jsonData["resource_id"] != "res_456" {
		t.Errorf("Expected JSON resource_id 'res_456', got %v", jsonData["resource_id"])
	}
}

func TestFilter(t *testing.T) {
	filter := NewFilter("vintage_filter", "filter_456", "res_789", 0.8, 0)

	if filter.EffectID != "filter_456" {
		t.Errorf("Expected EffectID 'filter_456', got '%s'", filter.EffectID)
	}

	if filter.ResourceID != "res_789" {
		t.Errorf("Expected ResourceID 'res_789', got '%s'", filter.ResourceID)
	}

	if filter.Intensity != 0.8 {
		t.Errorf("Expected intensity 0.8, got %f", filter.Intensity)
	}

	// 测试JSON导出
	jsonData := filter.ExportJSON()
	if jsonData["effect_id"] != "filter_456" {
		t.Errorf("Expected JSON effect_id 'filter_456', got %v", jsonData["effect_id"])
	}

	if jsonData["intensity"] != 0.8 {
		t.Errorf("Expected JSON intensity 0.8, got %v", jsonData["intensity"])
	}
}

func TestTransition(t *testing.T) {
	transitionDuration, _ := types.Tim("0.5s")
	transition := NewTransition("fade_transition", "transition_789", "res_abc", transitionDuration)

	if transition.EffectID != "transition_789" {
		t.Errorf("Expected EffectID 'transition_789', got '%s'", transition.EffectID)
	}

	if transition.ResourceID != "res_abc" {
		t.Errorf("Expected ResourceID 'res_abc', got '%s'", transition.ResourceID)
	}

	if transition.Duration != 500000 {
		t.Errorf("Expected duration 500000, got %d", transition.Duration)
	}

	// 测试JSON导出
	jsonData := transition.ExportJSON()
	if jsonData["effect_id"] != "transition_789" {
		t.Errorf("Expected JSON effect_id 'transition_789', got %v", jsonData["effect_id"])
	}

	if jsonData["duration"] != int64(500000) {
		t.Errorf("Expected JSON duration 500000, got %v", jsonData["duration"])
	}
}

func TestBackgroundFilling(t *testing.T) {
	// 测试颜色填充
	colorFill := NewBackgroundFilling("canvas_color", 0.0, "#FF8040C8")

	if colorFill.FillType != "canvas_color" {
		t.Errorf("Expected FillType 'canvas_color', got '%s'", colorFill.FillType)
	}

	if colorFill.Color != "#FF8040C8" {
		t.Errorf("Expected color '#FF8040C8', got '%s'", colorFill.Color)
	}

	// 测试模糊填充
	blurFill := NewBackgroundFilling("canvas_blur", 15.0, "")

	if blurFill.FillType != "canvas_blur" {
		t.Errorf("Expected FillType 'canvas_blur', got '%s'", blurFill.FillType)
	}

	if blurFill.Blur != 15.0 {
		t.Errorf("Expected blur 15.0, got %f", blurFill.Blur)
	}

	// 测试JSON导出
	colorJSON := colorFill.ExportJSON()
	if colorJSON["type"] != "canvas_color" {
		t.Errorf("Expected JSON type 'canvas_color', got %v", colorJSON["type"])
	}

	blurJSON := blurFill.ExportJSON()
	if blurJSON["type"] != "canvas_blur" {
		t.Errorf("Expected JSON type 'canvas_blur', got %v", blurJSON["type"])
	}

	if blurJSON["blur"] != 15.0 {
		t.Errorf("Expected JSON blur 15.0, got %v", blurJSON["blur"])
	}
}

func TestVideoSegmentWithEffects(t *testing.T) {
	// 创建一个包含各种效果的视频片段
	sourceTimerange, _ := types.Trange("0s", "8s")
	targetTimerange, _ := types.Trange("1s", "6s")

	videoSegment := NewVideoSegment("video_with_effects", sourceTimerange, targetTimerange, 1.0, 1.0, nil)

	// 添加蒙版
	videoSegment.AddMask("circle", "circle_mask", "image", "mask_resource", 0.0, 0.0, 1.0, 0.0, 0.0, false, nil, nil)

	// 添加视频特效
	videoSegment.AddEffect("cool_effect", "effect_123", "effect_res", "video_effect", 0)

	// 添加滤镜
	videoSegment.AddFilter("vintage_filter", "filter_456", "filter_res", 0.6, 0)

	// 添加转场
	transitionDuration, _ := types.Tim("1s")
	videoSegment.AddTransition("fade_transition", "trans_789", "trans_res", transitionDuration)

	// 添加背景填充
	videoSegment.SetBackgroundFilling("canvas_color", 0.0, "#64C896FF")

	// 测试JSON导出
	jsonData := videoSegment.ExportJSON()

	// 验证各种效果都被包含在JSON中
	if jsonData["mask"] == nil {
		t.Error("Expected mask to be included in JSON")
	}

	effects := jsonData["effects"].([]interface{})
	if len(effects) != 1 {
		t.Errorf("Expected 1 video effect, got %d", len(effects))
	}

	filters := jsonData["filters"].([]interface{})
	if len(filters) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(filters))
	}

	if jsonData["transition"] == nil {
		t.Error("Expected transition to be included in JSON")
	}

	if jsonData["background_filling"] == nil {
		t.Error("Expected background_filling to be included in JSON")
	}

	// 验证JSON可以序列化
	_, err := json.Marshal(jsonData)
	if err != nil {
		t.Errorf("Failed to marshal VideoSegment with effects JSON: %v", err)
	}
}
