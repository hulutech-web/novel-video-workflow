package script

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"novel-video-workflow/pkg/capcut/internal/animation"
	"novel-video-workflow/pkg/capcut/internal/material"
	"novel-video-workflow/pkg/capcut/internal/segment"
	"novel-video-workflow/pkg/capcut/internal/track"
)

// TestNewScriptMaterial 测试创建新的草稿素材管理器
func TestNewScriptMaterial(t *testing.T) {
	sm := NewScriptMaterial()

	if sm == nil {
		t.Fatal("创建ScriptMaterial失败")
	}

	if sm.Audios == nil || len(sm.Audios) != 0 {
		t.Error("音频素材列表初始化失败")
	}

	if sm.Videos == nil || len(sm.Videos) != 0 {
		t.Error("视频素材列表初始化失败")
	}

	if sm.Stickers == nil || len(sm.Stickers) != 0 {
		t.Error("贴纸素材列表初始化失败")
	}

	if sm.Texts == nil || len(sm.Texts) != 0 {
		t.Error("文本素材列表初始化失败")
	}

	if sm.AudioEffects == nil || len(sm.AudioEffects) != 0 {
		t.Error("音频特效列表初始化失败")
	}

	if sm.VideoEffects == nil || len(sm.VideoEffects) != 0 {
		t.Error("视频特效列表初始化失败")
	}

	if sm.Animations == nil || len(sm.Animations) != 0 {
		t.Error("动画列表初始化失败")
	}

	if sm.Transitions == nil || len(sm.Transitions) != 0 {
		t.Error("转场列表初始化失败")
	}
}

// TestScriptMaterialContains 测试素材包含检查
func TestScriptMaterialContains(t *testing.T) {
	sm := NewScriptMaterial()

	// 创建测试素材
	videoMaterial := &material.VideoMaterial{
		MaterialID: "test_video_123",
		Path:       "/path/to/video.mp4",
		CropSettings: &material.CropSettings{
			UpperLeftX:  0.0,
			UpperLeftY:  0.0,
			LowerRightX: 1.0,
			LowerRightY: 1.0,
		},
	}

	audioMaterial := &material.AudioMaterial{
		MaterialID: "test_audio_456",
		Path:       "/path/to/audio.mp3",
	}

	// 测试不包含的情况
	if sm.Contains(videoMaterial) {
		t.Error("期望不包含视频素材，但返回true")
	}

	if sm.Contains(audioMaterial) {
		t.Error("期望不包含音频素材，但返回true")
	}

	// 添加素材
	sm.Videos = append(sm.Videos, videoMaterial)
	sm.Audios = append(sm.Audios, audioMaterial)

	// 测试包含的情况
	if !sm.Contains(videoMaterial) {
		t.Error("期望包含视频素材，但返回false")
	}

	if !sm.Contains(audioMaterial) {
		t.Error("期望包含音频素材，但返回false")
	}

	// 测试不同ID的素材
	differentVideo := &material.VideoMaterial{
		MaterialID: "different_video_789",
		Path:       "/path/to/different.mp4",
		CropSettings: &material.CropSettings{
			UpperLeftX:  0.0,
			UpperLeftY:  0.0,
			LowerRightX: 1.0,
			LowerRightY: 1.0,
		},
	}

	if sm.Contains(differentVideo) {
		t.Error("期望不包含不同ID的视频素材，但返回true")
	}
}

// TestScriptMaterialContainsEffects 测试特效素材包含检查
func TestScriptMaterialContainsEffects(t *testing.T) {
	sm := NewScriptMaterial()

	// 创建测试特效
	audioFade := &segment.AudioFade{
		FadeID: "fade_123",
	}

	audioEffect := &segment.AudioEffect{
		EffectID: "effect_456",
	}

	segmentAnimations := &animation.SegmentAnimations{
		AnimationID: "animation_789",
	}

	videoEffect := &segment.VideoEffect{
		GlobalID: "video_effect_101",
	}

	transition := &segment.Transition{
		GlobalID: "transition_202",
	}

	// 测试不包含的情况
	if sm.Contains(audioFade) {
		t.Error("期望不包含音频淡入淡出，但返回true")
	}

	if sm.Contains(audioEffect) {
		t.Error("期望不包含音频特效，但返回true")
	}

	if sm.Contains(segmentAnimations) {
		t.Error("期望不包含片段动画，但返回true")
	}

	if sm.Contains(videoEffect) {
		t.Error("期望不包含视频特效，但返回true")
	}

	if sm.Contains(transition) {
		t.Error("期望不包含转场，但返回true")
	}

	// 添加特效
	sm.AudioFades = append(sm.AudioFades, audioFade)
	sm.AudioEffects = append(sm.AudioEffects, audioEffect)
	sm.Animations = append(sm.Animations, segmentAnimations)
	sm.VideoEffects = append(sm.VideoEffects, videoEffect)
	sm.Transitions = append(sm.Transitions, transition)

	// 测试包含的情况
	if !sm.Contains(audioFade) {
		t.Error("期望包含音频淡入淡出，但返回false")
	}

	if !sm.Contains(audioEffect) {
		t.Error("期望包含音频特效，但返回false")
	}

	if !sm.Contains(segmentAnimations) {
		t.Error("期望包含片段动画，但返回false")
	}

	if !sm.Contains(videoEffect) {
		t.Error("期望包含视频特效，但返回false")
	}

	if !sm.Contains(transition) {
		t.Error("期望包含转场，但返回false")
	}
}

// TestScriptMaterialExportJSON 测试素材导出JSON
func TestScriptMaterialExportJSON(t *testing.T) {
	sm := NewScriptMaterial()

	// 添加一些测试素材
	videoMaterial := &material.VideoMaterial{
		MaterialID: "test_video_123",
		Path:       "/path/to/video.mp4",
		Width:      1920,
		Height:     1080,
		Duration:   5000000, // 5秒
		CropSettings: &material.CropSettings{
			UpperLeftX:  0.0,
			UpperLeftY:  0.0,
			LowerRightX: 1.0,
			LowerRightY: 1.0,
		},
	}
	sm.Videos = append(sm.Videos, videoMaterial)

	audioMaterial := &material.AudioMaterial{
		MaterialID: "test_audio_456",
		Path:       "/path/to/audio.mp3",
		Duration:   3000000, // 3秒
	}
	sm.Audios = append(sm.Audios, audioMaterial)

	// 导出JSON
	jsonData := sm.ExportJSON()

	// 验证基本结构
	if jsonData == nil {
		t.Fatal("导出的JSON数据为nil")
	}

	// 验证必要字段存在
	requiredFields := []string{
		"audios", "videos", "audio_effects", "audio_fades",
		"material_animations", "video_effects", "transitions",
		"speeds", "effects", "canvases", "masks",
	}

	for _, field := range requiredFields {
		if _, exists := jsonData[field]; !exists {
			t.Errorf("导出的JSON缺少必要字段: %s", field)
		}
	}

	// 验证视频素材数据
	if videos, ok := jsonData["videos"].([]map[string]interface{}); ok {
		if len(videos) != 1 {
			t.Errorf("期望视频素材数量为1，得到%d", len(videos))
		}
	} else {
		t.Error("视频素材数据格式不正确")
	}

	// 验证音频素材数据
	if audios, ok := jsonData["audios"].([]map[string]interface{}); ok {
		if len(audios) != 1 {
			t.Errorf("期望音频素材数量为1，得到%d", len(audios))
		}
	} else {
		t.Error("音频素材数据格式不正确")
	}

	// 验证空字段是否为空数组
	emptyFields := []string{"audio_effects", "audio_fades", "material_animations"}
	for _, field := range emptyFields {
		if fieldData, ok := jsonData[field].([]map[string]interface{}); ok {
			if len(fieldData) != 0 {
				t.Errorf("期望%s为空数组，但长度为%d", field, len(fieldData))
			}
		}
	}
}

// TestNewScriptFile 测试创建新的草稿文件
func TestNewScriptFile(t *testing.T) {
	sf, err := NewScriptFile(1920, 1080, 30)
	if err != nil {
		t.Fatalf("创建ScriptFile失败: %v", err)
	}

	if sf.Width != 1920 {
		t.Errorf("期望宽度为1920，得到%d", sf.Width)
	}

	if sf.Height != 1080 {
		t.Errorf("期望高度为1080，得到%d", sf.Height)
	}

	if sf.FPS != 30 {
		t.Errorf("期望帧率为30，得到%d", sf.FPS)
	}

	if sf.Duration != 0 {
		t.Errorf("期望初始时长为0，得到%d", sf.Duration)
	}

	if sf.Materials == nil {
		t.Error("素材管理器未初始化")
	}

	if sf.Tracks == nil {
		t.Error("轨道映射未初始化")
	}

	if sf.ImportedMaterials == nil {
		t.Error("导入素材映射未初始化")
	}

	if sf.ImportedTracks == nil {
		t.Error("导入轨道列表未初始化")
	}

	if sf.Content == nil {
		t.Error("内容映射未初始化")
	}

	if sf.SavePath != nil {
		t.Error("保存路径应该为nil")
	}
}

// TestNewScriptFileWithDefaultFPS 测试使用默认帧率创建草稿文件
func TestNewScriptFileWithDefaultFPS(t *testing.T) {
	sf, err := NewScriptFile(1280, 720)
	if err != nil {
		t.Fatalf("创建ScriptFile失败: %v", err)
	}

	if sf.Width != 1280 {
		t.Errorf("期望宽度为1280，得到%d", sf.Width)
	}

	if sf.Height != 720 {
		t.Errorf("期望高度为720，得到%d", sf.Height)
	}

	if sf.FPS != 30 {
		t.Errorf("期望默认帧率为30，得到%d", sf.FPS)
	}
}

// TestScriptFileAddMaterial 测试添加素材
func TestScriptFileAddMaterial(t *testing.T) {
	sf, err := NewScriptFile(1920, 1080)
	if err != nil {
		t.Fatalf("创建ScriptFile失败: %v", err)
	}

	// 创建测试素材
	videoMaterial := &material.VideoMaterial{
		MaterialID: "test_video_123",
		Path:       "/path/to/video.mp4",
		CropSettings: &material.CropSettings{
			UpperLeftX:  0.0,
			UpperLeftY:  0.0,
			LowerRightX: 1.0,
			LowerRightY: 1.0,
		},
	}

	audioMaterial := &material.AudioMaterial{
		MaterialID: "test_audio_456",
		Path:       "/path/to/audio.mp3",
	}

	// 添加素材
	result := sf.AddMaterial(videoMaterial)
	if result != sf {
		t.Error("AddMaterial应该返回self")
	}

	if len(sf.Materials.Videos) != 1 {
		t.Errorf("期望视频素材数量为1，得到%d", len(sf.Materials.Videos))
	}

	if sf.Materials.Videos[0].MaterialID != "test_video_123" {
		t.Error("视频素材ID不匹配")
	}

	// 添加音频素材
	sf.AddMaterial(audioMaterial)
	if len(sf.Materials.Audios) != 1 {
		t.Errorf("期望音频素材数量为1，得到%d", len(sf.Materials.Audios))
	}

	if sf.Materials.Audios[0].MaterialID != "test_audio_456" {
		t.Error("音频素材ID不匹配")
	}

	// 测试重复添加
	sf.AddMaterial(videoMaterial)
	if len(sf.Materials.Videos) != 1 {
		t.Errorf("重复添加后，期望视频素材数量仍为1，得到%d", len(sf.Materials.Videos))
	}
}

// TestScriptFileAddTrack 测试添加轨道
func TestScriptFileAddTrack(t *testing.T) {
	sf, err := NewScriptFile(1920, 1080)
	if err != nil {
		t.Fatalf("创建ScriptFile失败: %v", err)
	}

	// 添加视频轨道
	trackName := "主视频轨道"
	result := sf.AddTrack(track.TrackTypeVideo, &trackName)
	if result != sf {
		t.Error("AddTrack应该返回self")
	}

	if len(sf.Tracks) != 1 {
		t.Errorf("期望轨道数量为1，得到%d", len(sf.Tracks))
	}

	if videoTrack, exists := sf.Tracks["主视频轨道"]; exists {
		if videoTrack.TrackType != track.TrackTypeVideo {
			t.Error("轨道类型不匹配")
		}
		if videoTrack.Name != "主视频轨道" {
			t.Error("轨道名称不匹配")
		}
	} else {
		t.Error("未找到添加的视频轨道")
	}

	// 添加音频轨道，不指定名称
	sf.AddTrack(track.TrackTypeAudio, nil)
	if len(sf.Tracks) != 2 {
		t.Errorf("期望轨道数量为2，得到%d", len(sf.Tracks))
	}

	if audioTrack, exists := sf.Tracks["audio"]; exists {
		if audioTrack.TrackType != track.TrackTypeAudio {
			t.Error("音频轨道类型不匹配")
		}
	} else {
		t.Error("未找到添加的音频轨道")
	}

	// 测试重复添加同名轨道
	sf.AddTrack(track.TrackTypeVideo, &trackName)
	if len(sf.Tracks) != 2 {
		t.Errorf("重复添加后，期望轨道数量仍为2，得到%d", len(sf.Tracks))
	}
}

// TestScriptFileAddTrackWithOptions 测试使用选项添加轨道
func TestScriptFileAddTrackWithOptions(t *testing.T) {
	sf, err := NewScriptFile(1920, 1080)
	if err != nil {
		t.Fatalf("创建ScriptFile失败: %v", err)
	}

	// 使用选项添加轨道
	trackName := "静音轨道"
	sf.AddTrack(track.TrackTypeVideo, &trackName,
		WithMute(true),
		WithRelativeIndex(5),
	)

	if videoTrack, exists := sf.Tracks["静音轨道"]; exists {
		if !videoTrack.Mute {
			t.Error("期望轨道为静音状态")
		}
		// 渲染层级应该是基础值 + 相对偏移
		expectedRenderIndex := track.GetTrackMeta(track.TrackTypeVideo).RenderIndex + 5
		if videoTrack.RenderIndex != expectedRenderIndex {
			t.Errorf("期望渲染层级为%d，得到%d", expectedRenderIndex, videoTrack.RenderIndex)
		}
	} else {
		t.Error("未找到添加的轨道")
	}

	// 使用绝对索引
	trackName2 := "绝对索引轨道"
	sf.AddTrack(track.TrackTypeAudio, &trackName2,
		WithAbsoluteIndex(1000),
	)

	if audioTrack, exists := sf.Tracks["绝对索引轨道"]; exists {
		if audioTrack.RenderIndex != 1000 {
			t.Errorf("期望绝对渲染层级为1000，得到%d", audioTrack.RenderIndex)
		}
	} else {
		t.Error("未找到添加的绝对索引轨道")
	}
}

// TestScriptFileGetTrack 测试获取轨道
func TestScriptFileGetTrack(t *testing.T) {
	sf, err := NewScriptFile(1920, 1080)
	if err != nil {
		t.Fatalf("创建ScriptFile失败: %v", err)
	}

	// 添加测试轨道
	trackName := "测试轨道"
	sf.AddTrack(track.TrackTypeVideo, &trackName)

	// 通过名称获取轨道
	foundTrack, err := sf.GetTrack("video", &trackName)
	if err != nil {
		t.Fatalf("获取轨道失败: %v", err)
	}

	if foundTrack.Name != "测试轨道" {
		t.Error("获取的轨道名称不匹配")
	}

	// 测试获取不存在的轨道
	nonExistentName := "不存在的轨道"
	_, err = sf.GetTrack("video", &nonExistentName)
	if err == nil {
		t.Error("期望获取不存在的轨道时返回错误")
	}
}

// TestScriptFileDumpsAndDump 测试JSON导出功能
func TestScriptFileDumpsAndDump(t *testing.T) {
	sf, err := NewScriptFile(1920, 1080, 25)
	if err != nil {
		t.Fatalf("创建ScriptFile失败: %v", err)
	}

	// 添加一些内容
	videoMaterial := &material.VideoMaterial{
		MaterialID: "test_video_123",
		Path:       "/path/to/video.mp4",
		CropSettings: &material.CropSettings{
			UpperLeftX:  0.0,
			UpperLeftY:  0.0,
			LowerRightX: 1.0,
			LowerRightY: 1.0,
		},
	}
	sf.AddMaterial(videoMaterial)

	trackName := "主轨道"
	sf.AddTrack(track.TrackTypeVideo, &trackName)

	sf.Duration = 10000000 // 10秒

	// 测试Dumps方法
	jsonStr, err := sf.Dumps()
	if err != nil {
		t.Fatalf("Dumps失败: %v", err)
	}

	if jsonStr == "" {
		t.Error("导出的JSON字符串为空")
	}

	// 验证JSON格式是否正确
	var parsedJSON map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsedJSON); err != nil {
		t.Fatalf("导出的JSON格式不正确: %v", err)
	}

	// 验证基本字段
	if fps, ok := parsedJSON["fps"].(float64); !ok || int(fps) != 25 {
		t.Error("FPS字段不正确")
	}

	if duration, ok := parsedJSON["duration"].(float64); !ok || int64(duration) != 10000000 {
		t.Error("Duration字段不正确")
	}

	if canvasConfig, ok := parsedJSON["canvas_config"].(map[string]interface{}); ok {
		if width, ok := canvasConfig["width"].(float64); !ok || int(width) != 1920 {
			t.Error("Canvas width不正确")
		}
		if height, ok := canvasConfig["height"].(float64); !ok || int(height) != 1080 {
			t.Error("Canvas height不正确")
		}
	} else {
		t.Error("Canvas config不存在")
	}

	// 测试Dump方法
	tempFile := filepath.Join(os.TempDir(), "test_script.json")
	defer os.Remove(tempFile)

	if err := sf.Dump(tempFile); err != nil {
		t.Fatalf("Dump失败: %v", err)
	}

	// 验证文件是否创建
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("Dump未创建文件")
	}

	// 读取文件内容并验证
	fileContent, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("读取Dump文件失败: %v", err)
	}

	var fileJSON map[string]interface{}
	if err := json.Unmarshal(fileContent, &fileJSON); err != nil {
		t.Fatalf("Dump文件JSON格式不正确: %v", err)
	}

	if fps, ok := fileJSON["fps"].(float64); !ok || int(fps) != 25 {
		t.Error("Dump文件中FPS字段不正确")
	}
}

// TestScriptFileSave 测试保存功能
func TestScriptFileSave(t *testing.T) {
	// 测试没有保存路径的情况
	sf, err := NewScriptFile(1920, 1080)
	if err != nil {
		t.Fatalf("创建ScriptFile失败: %v", err)
	}

	if err := sf.Save(); err == nil {
		t.Error("期望Save在没有保存路径时返回错误")
	}

	// 测试有保存路径的情况
	tempFile := filepath.Join(os.TempDir(), "test_save.json")
	defer os.Remove(tempFile)

	sf.SavePath = &tempFile
	if err := sf.Save(); err != nil {
		t.Fatalf("Save失败: %v", err)
	}

	// 验证文件是否创建
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("Save未创建文件")
	}
}

// TestLoadTemplate 测试加载模板功能
func TestLoadTemplate(t *testing.T) {
	// 创建临时测试文件
	tempFile := filepath.Join(os.TempDir(), "test_template.json")
	defer os.Remove(tempFile)

	// 创建测试JSON内容
	testContent := map[string]interface{}{
		"fps":      float64(24),
		"duration": float64(5000000),
		"canvas_config": map[string]interface{}{
			"width":  float64(1280),
			"height": float64(720),
		},
		"materials": map[string]interface{}{
			"videos": []interface{}{
				map[string]interface{}{
					"id":   "test_video_1",
					"path": "/test/video.mp4",
				},
			},
		},
		"tracks": []interface{}{
			map[string]interface{}{
				"type": "video",
				"name": "主轨道",
				"id":   "track_1",
			},
		},
	}

	// 写入测试文件
	jsonBytes, err := json.Marshal(testContent)
	if err != nil {
		t.Fatalf("创建测试JSON失败: %v", err)
	}

	if err := os.WriteFile(tempFile, jsonBytes, 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	// 测试加载模板
	sf, err := LoadTemplate(tempFile)
	if err != nil {
		t.Fatalf("LoadTemplate失败: %v", err)
	}

	if sf.FPS != 24 {
		t.Errorf("期望FPS为24，得到%d", sf.FPS)
	}

	if sf.Duration != 5000000 {
		t.Errorf("期望Duration为5000000，得到%d", sf.Duration)
	}

	if sf.Width != 1280 {
		t.Errorf("期望Width为1280，得到%d", sf.Width)
	}

	if sf.Height != 720 {
		t.Errorf("期望Height为720，得到%d", sf.Height)
	}

	if sf.SavePath == nil || *sf.SavePath != tempFile {
		t.Error("SavePath设置不正确")
	}

	// 验证导入的素材
	if videos, exists := sf.ImportedMaterials["videos"]; exists {
		if len(videos) != 1 {
			t.Errorf("期望导入1个视频素材，得到%d", len(videos))
		}
	} else {
		t.Error("未找到导入的视频素材")
	}

	// 测试加载不存在的文件
	_, err = LoadTemplate("/path/to/nonexistent/file.json")
	if err == nil {
		t.Error("期望加载不存在的文件时返回错误")
	}
}

// TestScriptFileInspectMaterial 测试素材检查功能
func TestScriptFileInspectMaterial(t *testing.T) {
	sf, err := NewScriptFile(1920, 1080)
	if err != nil {
		t.Fatalf("创建ScriptFile失败: %v", err)
	}

	// 添加测试素材
	sf.ImportedMaterials["stickers"] = []map[string]interface{}{
		{
			"resource_id": "sticker_123",
			"name":        "测试贴纸",
		},
	}

	sf.ImportedMaterials["effects"] = []map[string]interface{}{
		{
			"type":        "text_shape",
			"effect_id":   "effect_456",
			"resource_id": "bubble_789",
			"name":        "文字气泡",
		},
		{
			"type":        "text_effect",
			"resource_id": "flower_101",
			"name":        "花字效果",
		},
	}

	// 测试InspectMaterial方法（这个方法主要是打印，我们只验证它不会panic）
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("InspectMaterial发生panic: %v", r)
		}
	}()

	sf.InspectMaterial()
}

// TestScriptFileJSONSerialization 测试JSON序列化兼容性
func TestScriptFileJSONSerialization(t *testing.T) {
	sf, err := NewScriptFile(1920, 1080, 30)
	if err != nil {
		t.Fatalf("创建ScriptFile失败: %v", err)
	}

	// 添加复杂数据
	videoMaterial := &material.VideoMaterial{
		MaterialID: "complex_video_123",
		Path:       "/path/to/complex.mp4",
		Width:      1920,
		Height:     1080,
		Duration:   10000000,
		CropSettings: &material.CropSettings{
			UpperLeftX:  0.0,
			UpperLeftY:  0.0,
			LowerRightX: 1.0,
			LowerRightY: 1.0,
		},
	}
	sf.AddMaterial(videoMaterial)

	audioMaterial := &material.AudioMaterial{
		MaterialID: "complex_audio_456",
		Path:       "/path/to/complex.mp3",
		Duration:   8000000,
	}
	sf.AddMaterial(audioMaterial)

	trackName1 := "复杂视频轨道"
	trackName2 := "复杂音频轨道"
	sf.AddTrack(track.TrackTypeVideo, &trackName1, WithRelativeIndex(2))
	sf.AddTrack(track.TrackTypeAudio, &trackName2, WithMute(true))

	sf.Duration = 15000000

	// 导出JSON
	jsonStr, err := sf.Dumps()
	if err != nil {
		t.Fatalf("JSON导出失败: %v", err)
	}

	// 解析JSON验证结构
	var parsedData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsedData); err != nil {
		t.Fatalf("JSON解析失败: %v", err)
	}

	// 验证平台信息
	if platform, ok := parsedData["platform"].(map[string]interface{}); ok {
		if appID, ok := platform["app_id"].(float64); !ok || int(appID) != 359289 {
			t.Error("平台app_id不正确")
		}
		if appSource, ok := platform["app_source"].(string); !ok || appSource != "cc" {
			t.Error("平台app_source不正确")
		}
	} else {
		t.Error("缺少平台信息")
	}

	// 验证素材信息
	if materials, ok := parsedData["materials"].(map[string]interface{}); ok {
		if videos, ok := materials["videos"].([]interface{}); ok {
			if len(videos) != 1 {
				t.Errorf("期望1个视频素材，得到%d", len(videos))
			}
		} else {
			t.Error("视频素材格式不正确")
		}

		if audios, ok := materials["audios"].([]interface{}); ok {
			if len(audios) != 1 {
				t.Errorf("期望1个音频素材，得到%d", len(audios))
			}
		} else {
			t.Error("音频素材格式不正确")
		}
	} else {
		t.Error("缺少素材信息")
	}

	// 验证轨道信息
	if tracks, ok := parsedData["tracks"].([]interface{}); ok {
		if len(tracks) != 2 {
			t.Errorf("期望2个轨道，得到%d", len(tracks))
		}
	} else {
		t.Error("轨道信息格式不正确")
	}
}

// TestScriptFileErrorHandling 测试错误处理
func TestScriptFileErrorHandling(t *testing.T) {
	// 测试创建ScriptFile时的各种边界情况

	// 正常创建
	sf, err := NewScriptFile(1920, 1080, 30)
	if err != nil {
		t.Fatalf("正常创建失败: %v", err)
	}

	// 测试零值参数
	sf2, err := NewScriptFile(0, 0, 0)
	if err != nil {
		t.Fatalf("零值参数创建失败: %v", err)
	}

	if sf2.Width != 0 || sf2.Height != 0 || sf2.FPS != 0 {
		t.Error("零值参数未正确设置")
	}

	// 测试负值参数
	sf3, err := NewScriptFile(-100, -200, -10)
	if err != nil {
		t.Fatalf("负值参数创建失败: %v", err)
	}

	if sf3.Width != -100 || sf3.Height != -200 || sf3.FPS != -10 {
		t.Error("负值参数未正确设置")
	}

	// 测试添加nil素材
	sf.AddMaterial(nil)
	// 应该不会panic，只是不添加

	// 测试获取不存在类型的轨道
	_, err = sf.GetTrack("nonexistent_type", nil)
	if err == nil {
		t.Error("期望获取不存在类型的轨道时返回错误")
	}
}
