package segment

import (
	"encoding/json"
	"testing"

	"novel-video-workflow/pkg/capcut/internal/types"
)

func TestAudioSegment(t *testing.T) {
	sourceTimerange, _ := types.Trange("0s", "30s")
	targetTimerange, _ := types.Trange("5s", "10s")

	audioSegment := NewAudioSegment("audio_123", targetTimerange, sourceTimerange, 1.2, 0.8)

	// 验证基本属性
	if audioSegment.MaterialID != "audio_123" {
		t.Errorf("Expected MaterialID 'audio_123', got '%s'", audioSegment.MaterialID)
	}

	if audioSegment.Speed.Value != 1.2 {
		t.Errorf("Expected speed 1.2, got %f", audioSegment.Speed.Value)
	}

	if audioSegment.Volume != 0.8 {
		t.Errorf("Expected volume 0.8, got %f", audioSegment.Volume)
	}

	// 验证默认值
	if audioSegment.Fade != nil {
		t.Error("Expected Fade to be nil by default")
	}

	if len(audioSegment.Effects) != 0 {
		t.Errorf("Expected empty audio effects, got %d", len(audioSegment.Effects))
	}

	// 测试JSON导出
	jsonData := audioSegment.ExportJSON()
	if jsonData["material_id"] != "audio_123" {
		t.Errorf("Expected material_id 'audio_123', got %v", jsonData["material_id"])
	}

	if jsonData["volume"] != 0.8 {
		t.Errorf("Expected volume 0.8, got %v", jsonData["volume"])
	}

	// 验证JSON可以序列化
	_, err := json.Marshal(jsonData)
	if err != nil {
		t.Errorf("Failed to marshal AudioSegment JSON: %v", err)
	}
}

func TestAudioFade(t *testing.T) {
	// 测试淡入淡出
	inDuration, _ := types.Tim("2s")
	outDuration, _ := types.Tim("1.5s")
	fade := NewAudioFade(inDuration, outDuration)

	if fade.InDuration != 2000000 {
		t.Errorf("Expected InDuration 2000000, got %d", fade.InDuration)
	}

	if fade.OutDuration != 1500000 {
		t.Errorf("Expected OutDuration 1500000, got %d", fade.OutDuration)
	}

	if fade.FadeID == "" {
		t.Error("Expected non-empty FadeID")
	}

	// 测试JSON导出
	jsonData := fade.ExportJSON()
	if jsonData["fade_in_duration"] != int64(2000000) {
		t.Errorf("Expected JSON fade_in_duration 2000000, got %v", jsonData["fade_in_duration"])
	}

	if jsonData["fade_out_duration"] != int64(1500000) {
		t.Errorf("Expected JSON fade_out_duration 1500000, got %v", jsonData["fade_out_duration"])
	}

	if jsonData["type"] != "audio_fade" {
		t.Error("Expected JSON type 'audio_fade'")
	}
}

func TestAudioEffect(t *testing.T) {
	effect := NewAudioEffectWithCategory("echo_effect", "res_echo", AudioEffectCategorySoundEffect)

	if effect.Name != "echo_effect" {
		t.Errorf("Expected Name 'echo_effect', got '%s'", effect.Name)
	}

	if effect.ResourceID != "res_echo" {
		t.Errorf("Expected ResourceID 'res_echo', got '%s'", effect.ResourceID)
	}

	if effect.CategoryID != "sound_effect" {
		t.Errorf("Expected CategoryID 'sound_effect', got '%s'", effect.CategoryID)
	}

	if effect.EffectID == "" {
		t.Error("Expected non-empty EffectID")
	}

	// 测试JSON导出
	jsonData := effect.ExportJSON()
	if jsonData["name"] != "echo_effect" {
		t.Errorf("Expected JSON name 'echo_effect', got %v", jsonData["name"])
	}

	if jsonData["resource_id"] != "res_echo" {
		t.Errorf("Expected JSON resource_id 'res_echo', got %v", jsonData["resource_id"])
	}

	if jsonData["id"] != effect.EffectID {
		t.Errorf("Expected JSON id '%s', got %v", effect.EffectID, jsonData["id"])
	}
}

func TestAudioSegmentWithEffects(t *testing.T) {
	// 创建一个包含淡入淡出和音效的音频片段
	sourceTimerange, _ := types.Trange("0s", "60s")
	targetTimerange, _ := types.Trange("10s", "20s")

	audioSegment := NewAudioSegment("music_track", targetTimerange, sourceTimerange, 1.0, 0.9)

	// 添加淡入淡出
	inDuration, _ := types.Tim("3s")
	outDuration, _ := types.Tim("2s")
	audioSegment.Fade = NewAudioFade(inDuration, outDuration)

	// 添加音频特效
	audioSegment.AddEffect("reverb", "reverb_resource", AudioEffectCategorySoundEffect)
	audioSegment.AddEffect("echo", "echo_resource", AudioEffectCategoryTone)

	// 测试JSON导出
	jsonData := audioSegment.ExportJSON()

	// 验证淡入淡出效果
	if jsonData["fade"] == nil {
		t.Error("Expected fade to be included in JSON")
	}

	// 验证音频特效
	effects := jsonData["effects"].([]interface{})
	if len(effects) != 2 {
		t.Errorf("Expected 2 audio effects, got %d", len(effects))
	}

	// 验证extra_material_refs包含所有效果的ID
	extraRefs := jsonData["extra_material_refs"].([]string)
	expectedRefs := 3 // speed + 2 effects
	if len(extraRefs) != expectedRefs {
		t.Errorf("Expected %d extra material refs, got %d", expectedRefs, len(extraRefs))
	}

	// 验证JSON可以序列化
	_, err := json.Marshal(jsonData)
	if err != nil {
		t.Errorf("Failed to marshal AudioSegment with effects JSON: %v", err)
	}
}

func TestAudioSegmentVolumeSettings(t *testing.T) {
	// 测试不同音量设置
	testCases := []struct {
		volume   float64
		expected float64
	}{
		{0.0, 0.0}, // 静音
		{0.5, 0.5}, // 半音量
		{1.0, 1.0}, // 正常音量
		{1.5, 1.5}, // 增强音量
		{2.0, 2.0}, // 双倍音量
	}

	for _, tc := range testCases {
		timerange, _ := types.Trange("0s", "5s")
		audioSegment := NewAudioSegment("test_audio", timerange, nil, 1.0, tc.volume)

		if audioSegment.Volume != tc.expected {
			t.Errorf("Expected volume %f, got %f", tc.expected, audioSegment.Volume)
		}

		jsonData := audioSegment.ExportJSON()
		if jsonData["volume"] != tc.expected {
			t.Errorf("Expected JSON volume %f, got %v", tc.expected, jsonData["volume"])
		}
	}
}

func TestAudioSegmentSpeedSettings(t *testing.T) {
	// 测试不同速度设置
	testCases := []struct {
		speed    float64
		expected float64
	}{
		{0.5, 0.5}, // 半速
		{1.0, 1.0}, // 正常速度
		{1.5, 1.5}, // 1.5倍速
		{2.0, 2.0}, // 双倍速
		{4.0, 4.0}, // 四倍速
	}

	for _, tc := range testCases {
		timerange, _ := types.Trange("0s", "10s")
		audioSegment := NewAudioSegment("test_audio", timerange, nil, tc.speed, 1.0)

		if audioSegment.Speed.Value != tc.expected {
			t.Errorf("Expected speed %f, got %f", tc.expected, audioSegment.Speed.Value)
		}

		jsonData := audioSegment.ExportJSON()
		if jsonData["speed"] != tc.expected {
			t.Errorf("Expected JSON speed %f, got %v", tc.expected, jsonData["speed"])
		}
	}
}
