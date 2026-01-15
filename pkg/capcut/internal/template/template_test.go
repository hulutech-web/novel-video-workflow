package template

import (
	"encoding/json"
	"testing"

	"novel-video-workflow/pkg/capcut/internal/material"
	"novel-video-workflow/pkg/capcut/internal/track"
	"novel-video-workflow/pkg/capcut/internal/types"
)

// TestShrinkAndExtendModes 测试缩短和延长模式枚举
func TestShrinkAndExtendModes(t *testing.T) {
	// 测试缩短模式
	shrinkModes := []ShrinkMode{
		ShrinkModeCutHead,
		ShrinkModeCutTail,
		ShrinkModeCutTailAlign,
		ShrinkModeShrink,
	}

	expectedShrinkValues := []string{
		"cut_head",
		"cut_tail",
		"cut_tail_align",
		"shrink",
	}

	for i, mode := range shrinkModes {
		if string(mode) != expectedShrinkValues[i] {
			t.Errorf("期望缩短模式值为 '%s', 得到 '%s'", expectedShrinkValues[i], string(mode))
		}
	}

	// 测试延长模式
	extendModes := []ExtendMode{
		ExtendModeCutMaterialTail,
		ExtendModeExtendHead,
		ExtendModeExtendTail,
		ExtendModePushTail,
	}

	expectedExtendValues := []string{
		"cut_material_tail",
		"extend_head",
		"extend_tail",
		"push_tail",
	}

	for i, mode := range extendModes {
		if string(mode) != expectedExtendValues[i] {
			t.Errorf("期望延长模式值为 '%s', 得到 '%s'", expectedExtendValues[i], string(mode))
		}
	}
}

// TestNewImportedSegment 测试创建导入的片段
func TestNewImportedSegment(t *testing.T) {
	// 测试用JSON数据
	jsonData := map[string]interface{}{
		"material_id": "test_material_123",
		"target_timerange": map[string]interface{}{
			"start":    float64(1000000),
			"duration": float64(2000000),
		},
		"extra_field": "extra_value",
	}

	segment, err := NewImportedSegment(jsonData)
	if err != nil {
		t.Fatalf("创建导入片段失败: %v", err)
	}

	if segment.MaterialID != "test_material_123" {
		t.Errorf("期望素材ID为 'test_material_123', 得到 '%s'", segment.MaterialID)
	}

	if segment.TargetTimerange.Start != 1000000 {
		t.Errorf("期望开始时间为 1000000, 得到 %d", segment.TargetTimerange.Start)
	}

	if segment.TargetTimerange.Duration != 2000000 {
		t.Errorf("期望持续时间为 2000000, 得到 %d", segment.TargetTimerange.Duration)
	}

	// 验证原始数据被保存
	if segment.RawData["extra_field"] != "extra_value" {
		t.Errorf("期望原始数据包含 'extra_field': 'extra_value'")
	}
}

// TestImportedSegmentExportJSON 测试导入片段的JSON导出
func TestImportedSegmentExportJSON(t *testing.T) {
	jsonData := map[string]interface{}{
		"material_id": "test_material_123",
		"target_timerange": map[string]interface{}{
			"start":    float64(1000000),
			"duration": float64(2000000),
		},
		"render_index": float64(100),
		"visible":      true,
	}

	segment, err := NewImportedSegment(jsonData)
	if err != nil {
		t.Fatalf("创建导入片段失败: %v", err)
	}

	// 修改属性
	segment.MaterialID = "modified_material_456"
	segment.TargetTimerange.Start = 500000
	segment.TargetTimerange.Duration = 1500000

	exportedData := segment.ExportJSON()

	// 验证修改后的属性被正确导出
	if exportedData["material_id"] != "modified_material_456" {
		t.Errorf("期望导出的素材ID为 'modified_material_456', 得到 '%v'", exportedData["material_id"])
	}

	targetTimerange := exportedData["target_timerange"].(map[string]interface{})
	if targetTimerange["start"] != int64(500000) {
		t.Errorf("期望导出的开始时间为 500000, 得到 %v", targetTimerange["start"])
	}

	if targetTimerange["duration"] != int64(1500000) {
		t.Errorf("期望导出的持续时间为 1500000, 得到 %v", targetTimerange["duration"])
	}

	// 验证原始数据的其他字段被保留
	if exportedData["render_index"] != float64(100) {
		t.Errorf("期望导出的render_index为 100, 得到 %v", exportedData["render_index"])
	}

	if exportedData["visible"] != true {
		t.Errorf("期望导出的visible为 true, 得到 %v", exportedData["visible"])
	}
}

// TestNewImportedMediaSegment 测试创建导入的媒体片段
func TestNewImportedMediaSegment(t *testing.T) {
	jsonData := map[string]interface{}{
		"material_id": "media_material_789",
		"target_timerange": map[string]interface{}{
			"start":    float64(500000),
			"duration": float64(1000000),
		},
		"source_timerange": map[string]interface{}{
			"start":    float64(0),
			"duration": float64(1000000),
		},
		"speed":  float64(1.0),
		"volume": float64(0.8),
	}

	mediaSegment, err := NewImportedMediaSegment(jsonData)
	if err != nil {
		t.Fatalf("创建导入媒体片段失败: %v", err)
	}

	if mediaSegment.MaterialID != "media_material_789" {
		t.Errorf("期望素材ID为 'media_material_789', 得到 '%s'", mediaSegment.MaterialID)
	}

	if mediaSegment.SourceTimerange.Start != 0 {
		t.Errorf("期望源开始时间为 0, 得到 %d", mediaSegment.SourceTimerange.Start)
	}

	if mediaSegment.SourceTimerange.Duration != 1000000 {
		t.Errorf("期望源持续时间为 1000000, 得到 %d", mediaSegment.SourceTimerange.Duration)
	}
}

// TestImportedMediaSegmentExportJSON 测试导入媒体片段的JSON导出
func TestImportedMediaSegmentExportJSON(t *testing.T) {
	jsonData := map[string]interface{}{
		"material_id": "media_material_789",
		"target_timerange": map[string]interface{}{
			"start":    float64(500000),
			"duration": float64(1000000),
		},
		"source_timerange": map[string]interface{}{
			"start":    float64(0),
			"duration": float64(1000000),
		},
	}

	mediaSegment, err := NewImportedMediaSegment(jsonData)
	if err != nil {
		t.Fatalf("创建导入媒体片段失败: %v", err)
	}

	// 修改源时间范围
	mediaSegment.SourceTimerange.Start = 100000
	mediaSegment.SourceTimerange.Duration = 800000

	exportedData := mediaSegment.ExportJSON()

	// 验证源时间范围被正确导出
	sourceTimerange := exportedData["source_timerange"].(map[string]interface{})
	if sourceTimerange["start"] != int64(100000) {
		t.Errorf("期望导出的源开始时间为 100000, 得到 %v", sourceTimerange["start"])
	}

	if sourceTimerange["duration"] != int64(800000) {
		t.Errorf("期望导出的源持续时间为 800000, 得到 %v", sourceTimerange["duration"])
	}
}

// TestNewImportedTrack 测试创建导入的轨道
func TestNewImportedTrack(t *testing.T) {
	jsonData := map[string]interface{}{
		"type": "video",
		"name": "测试视频轨道",
		"id":   "track_123",
		"segments": []interface{}{
			map[string]interface{}{
				"render_index": float64(100),
			},
			map[string]interface{}{
				"render_index": float64(200),
			},
		},
	}

	importedTrack, err := NewImportedTrack(jsonData)
	if err != nil {
		t.Fatalf("创建导入轨道失败: %v", err)
	}

	if importedTrack.TrackType != track.TrackTypeVideo {
		t.Errorf("期望轨道类型为 Video, 得到 %v", importedTrack.TrackType)
	}

	if importedTrack.Name != "测试视频轨道" {
		t.Errorf("期望轨道名称为 '测试视频轨道', 得到 '%s'", importedTrack.Name)
	}

	if importedTrack.TrackID != "track_123" {
		t.Errorf("期望轨道ID为 'track_123', 得到 '%s'", importedTrack.TrackID)
	}

	// 验证渲染层级为片段中的最大值
	if importedTrack.RenderIndex != 200 {
		t.Errorf("期望渲染层级为 200, 得到 %d", importedTrack.RenderIndex)
	}
}

// TestImportedTrackExportJSON 测试导入轨道的JSON导出
func TestImportedTrackExportJSON(t *testing.T) {
	jsonData := map[string]interface{}{
		"type":      "audio",
		"name":      "原始音频轨道",
		"id":        "track_456",
		"attribute": float64(1),
		"extra":     "extra_data",
	}

	importedTrack, err := NewImportedTrack(jsonData)
	if err != nil {
		t.Fatalf("创建导入轨道失败: %v", err)
	}

	// 修改名称
	importedTrack.Name = "修改后的音频轨道"

	exportedData := importedTrack.ExportJSON()

	// 验证修改后的名称被导出
	if exportedData["name"] != "修改后的音频轨道" {
		t.Errorf("期望导出的轨道名称为 '修改后的音频轨道', 得到 '%v'", exportedData["name"])
	}

	// 验证原始数据的其他字段被保留
	if exportedData["attribute"] != float64(1) {
		t.Errorf("期望导出的attribute为 1, 得到 %v", exportedData["attribute"])
	}

	if exportedData["extra"] != "extra_data" {
		t.Errorf("期望导出的extra为 'extra_data', 得到 '%v'", exportedData["extra"])
	}
}

// TestNewImportedTextTrack 测试创建导入的文本轨道
func TestNewImportedTextTrack(t *testing.T) {
	jsonData := map[string]interface{}{
		"type": "text",
		"name": "文本轨道",
		"id":   "text_track_789",
		"segments": []interface{}{
			map[string]interface{}{
				"material_id": "text_material_1",
				"target_timerange": map[string]interface{}{
					"start":    float64(0),
					"duration": float64(3000000),
				},
			},
			map[string]interface{}{
				"material_id": "text_material_2",
				"target_timerange": map[string]interface{}{
					"start":    float64(3000000),
					"duration": float64(2000000),
				},
			},
		},
	}

	textTrack, err := NewImportedTextTrack(jsonData)
	if err != nil {
		t.Fatalf("创建导入文本轨道失败: %v", err)
	}

	if textTrack.TrackType != track.TrackTypeText {
		t.Errorf("期望轨道类型为 Text, 得到 %v", textTrack.TrackType)
	}

	if len(textTrack.Segments) != 2 {
		t.Errorf("期望片段数量为 2, 得到 %d", len(textTrack.Segments))
	}

	if textTrack.Len() != 2 {
		t.Errorf("期望Len()返回 2, 得到 %d", textTrack.Len())
	}

	// 验证起始和结束时间
	if textTrack.StartTime() != 0 {
		t.Errorf("期望起始时间为 0, 得到 %d", textTrack.StartTime())
	}

	if textTrack.EndTime() != 5000000 {
		t.Errorf("期望结束时间为 5000000, 得到 %d", textTrack.EndTime())
	}
}

// TestNewImportedMediaTrack 测试创建导入的媒体轨道
func TestNewImportedMediaTrack(t *testing.T) {
	jsonData := map[string]interface{}{
		"type": "video",
		"name": "视频轨道",
		"id":   "video_track_101",
		"segments": []interface{}{
			map[string]interface{}{
				"material_id": "video_material_1",
				"target_timerange": map[string]interface{}{
					"start":    float64(0),
					"duration": float64(2000000),
				},
				"source_timerange": map[string]interface{}{
					"start":    float64(0),
					"duration": float64(2000000),
				},
			},
		},
	}

	mediaTrack, err := NewImportedMediaTrack(jsonData)
	if err != nil {
		t.Fatalf("创建导入媒体轨道失败: %v", err)
	}

	if mediaTrack.TrackType != track.TrackTypeVideo {
		t.Errorf("期望轨道类型为 Video, 得到 %v", mediaTrack.TrackType)
	}

	if len(mediaTrack.MediaSegments) != 1 {
		t.Errorf("期望媒体片段数量为 1, 得到 %d", len(mediaTrack.MediaSegments))
	}

	if len(mediaTrack.Segments) != 1 {
		t.Errorf("期望片段数量为 1, 得到 %d", len(mediaTrack.Segments))
	}
}

// TestCheckMaterialType 测试素材类型检查
func TestCheckMaterialType(t *testing.T) {
	// 创建视频轨道
	videoTrackData := map[string]interface{}{
		"type": "video",
		"name": "视频轨道",
		"id":   "video_track_test",
	}

	videoTrack, err := NewImportedMediaTrack(videoTrackData)
	if err != nil {
		t.Fatalf("创建视频轨道失败: %v", err)
	}

	// 创建音频轨道
	audioTrackData := map[string]interface{}{
		"type": "audio",
		"name": "音频轨道",
		"id":   "audio_track_test",
	}

	audioTrack, err := NewImportedMediaTrack(audioTrackData)
	if err != nil {
		t.Fatalf("创建音频轨道失败: %v", err)
	}

	// 创建测试素材
	videoMaterial := &material.VideoMaterial{}
	audioMaterial := &material.AudioMaterial{}

	// 测试视频轨道
	if !videoTrack.CheckMaterialType(videoMaterial) {
		t.Error("期望视频轨道接受视频素材")
	}

	if videoTrack.CheckMaterialType(audioMaterial) {
		t.Error("期望视频轨道拒绝音频素材")
	}

	// 测试音频轨道
	if !audioTrack.CheckMaterialType(audioMaterial) {
		t.Error("期望音频轨道接受音频素材")
	}

	if audioTrack.CheckMaterialType(videoMaterial) {
		t.Error("期望音频轨道拒绝视频素材")
	}
}

// TestProcessTimerange 测试时间范围处理
func TestProcessTimerange(t *testing.T) {
	// 创建测试媒体轨道
	jsonData := map[string]interface{}{
		"type": "video",
		"name": "测试轨道",
		"id":   "test_track",
		"segments": []interface{}{
			map[string]interface{}{
				"material_id": "material_1",
				"target_timerange": map[string]interface{}{
					"start":    float64(0),
					"duration": float64(2000000),
				},
				"source_timerange": map[string]interface{}{
					"start":    float64(0),
					"duration": float64(2000000),
				},
			},
			map[string]interface{}{
				"material_id": "material_2",
				"target_timerange": map[string]interface{}{
					"start":    float64(3000000),
					"duration": float64(1000000),
				},
				"source_timerange": map[string]interface{}{
					"start":    float64(0),
					"duration": float64(1000000),
				},
			},
		},
	}

	mediaTrack, err := NewImportedMediaTrack(jsonData)
	if err != nil {
		t.Fatalf("创建媒体轨道失败: %v", err)
	}

	// 测试缩短处理 - cut_tail
	shortTimerange := types.NewTimerange(0, 1500000) // 比原来短500000微秒
	err = mediaTrack.ProcessTimerange(0, shortTimerange, ShrinkModeCutTail, nil)
	if err != nil {
		t.Fatalf("处理缩短时间范围失败: %v", err)
	}

	// 验证第一个片段的持续时间被缩短
	if mediaTrack.MediaSegments[0].TargetTimerange.Duration != 1500000 {
		t.Errorf("期望第一个片段持续时间为 1500000, 得到 %d", mediaTrack.MediaSegments[0].TargetTimerange.Duration)
	}

	// 测试延长处理 - extend_tail
	longTimerange := types.NewTimerange(0, 2500000) // 比原来长1000000微秒
	err = mediaTrack.ProcessTimerange(0, longTimerange, ShrinkModeCutTail, []ExtendMode{ExtendModeExtendTail})
	if err != nil {
		t.Fatalf("处理延长时间范围失败: %v", err)
	}

	// 验证第一个片段的持续时间被延长
	if mediaTrack.MediaSegments[0].TargetTimerange.Duration != 2500000 {
		t.Errorf("期望第一个片段持续时间为 2500000, 得到 %d", mediaTrack.MediaSegments[0].TargetTimerange.Duration)
	}
}

// TestProcessTimerangeShrinkModes 测试不同的缩短模式
func TestProcessTimerangeShrinkModes(t *testing.T) {
	// 创建测试媒体轨道
	createTestTrack := func() *ImportedMediaTrack {
		jsonData := map[string]interface{}{
			"type": "video",
			"name": "测试轨道",
			"id":   "test_track",
			"segments": []interface{}{
				map[string]interface{}{
					"material_id": "material_1",
					"target_timerange": map[string]interface{}{
						"start":    float64(1000000),
						"duration": float64(2000000),
					},
					"source_timerange": map[string]interface{}{
						"start":    float64(0),
						"duration": float64(2000000),
					},
				},
				map[string]interface{}{
					"material_id": "material_2",
					"target_timerange": map[string]interface{}{
						"start":    float64(4000000),
						"duration": float64(1000000),
					},
					"source_timerange": map[string]interface{}{
						"start":    float64(0),
						"duration": float64(1000000),
					},
				},
			},
		}

		mediaTrack, _ := NewImportedMediaTrack(jsonData)
		return mediaTrack
	}

	// 测试 cut_head 模式
	track1 := createTestTrack()
	shortTimerange := types.NewTimerange(0, 1500000) // 缩短500000微秒
	err := track1.ProcessTimerange(0, shortTimerange, ShrinkModeCutHead, nil)
	if err != nil {
		t.Fatalf("cut_head模式失败: %v", err)
	}

	// 验证起始时间后移
	if track1.MediaSegments[0].TargetTimerange.Start != 1500000 {
		t.Errorf("cut_head: 期望起始时间为 1500000, 得到 %d", track1.MediaSegments[0].TargetTimerange.Start)
	}

	// 测试 cut_tail_align 模式
	track2 := createTestTrack()
	err = track2.ProcessTimerange(0, shortTimerange, ShrinkModeCutTailAlign, nil)
	if err != nil {
		t.Fatalf("cut_tail_align模式失败: %v", err)
	}

	// 验证后续片段前移
	if track2.MediaSegments[1].TargetTimerange.Start != 3500000 {
		t.Errorf("cut_tail_align: 期望后续片段起始时间为 3500000, 得到 %d", track2.MediaSegments[1].TargetTimerange.Start)
	}

	// 测试 shrink 模式
	track3 := createTestTrack()
	err = track3.ProcessTimerange(0, shortTimerange, ShrinkModeShrink, nil)
	if err != nil {
		t.Fatalf("shrink模式失败: %v", err)
	}

	// 验证起始时间和持续时间都改变
	if track3.MediaSegments[0].TargetTimerange.Start != 1250000 { // 1000000 + 500000/2
		t.Errorf("shrink: 期望起始时间为 1250000, 得到 %d", track3.MediaSegments[0].TargetTimerange.Start)
	}

	if track3.MediaSegments[0].TargetTimerange.Duration != 1500000 {
		t.Errorf("shrink: 期望持续时间为 1500000, 得到 %d", track3.MediaSegments[0].TargetTimerange.Duration)
	}
}

// TestImportTrack 测试导入轨道函数
func TestImportTrack(t *testing.T) {
	jsonData := map[string]interface{}{
		"type": "video",
		"name": "导入的视频轨道",
		"id":   "imported_track_123",
		"segments": []interface{}{
			map[string]interface{}{
				"render_index": float64(150),
			},
		},
		"attribute": float64(0), // 非静音
	}

	importedMaterials := map[string]interface{}{
		"videos": []interface{}{},
		"audios": []interface{}{},
	}

	newTrack, err := ImportTrack(jsonData, importedMaterials)
	if err != nil {
		t.Fatalf("导入轨道失败: %v", err)
	}

	if newTrack.TrackType != track.TrackTypeVideo {
		t.Errorf("期望轨道类型为 Video, 得到 %v", newTrack.TrackType)
	}

	if newTrack.Name != "导入的视频轨道" {
		t.Errorf("期望轨道名称为 '导入的视频轨道', 得到 '%s'", newTrack.Name)
	}

	if newTrack.TrackID != "imported_track_123" {
		t.Errorf("期望轨道ID为 'imported_track_123', 得到 '%s'", newTrack.TrackID)
	}

	if newTrack.RenderIndex != 150 {
		t.Errorf("期望渲染层级为 150, 得到 %d", newTrack.RenderIndex)
	}

	if newTrack.Mute != false {
		t.Errorf("期望静音状态为 false, 得到 %v", newTrack.Mute)
	}
}

// TestJSONSerialization 测试JSON序列化兼容性
func TestJSONSerialization(t *testing.T) {
	// 测试导入片段的JSON序列化
	segmentData := map[string]interface{}{
		"material_id": "test_material",
		"target_timerange": map[string]interface{}{
			"start":    float64(1000000),
			"duration": float64(2000000),
		},
		"render_index": float64(100),
		"visible":      true,
	}

	segment, err := NewImportedSegment(segmentData)
	if err != nil {
		t.Fatalf("创建导入片段失败: %v", err)
	}

	exportedData := segment.ExportJSON()

	// 序列化为JSON字符串
	jsonBytes, err := json.Marshal(exportedData)
	if err != nil {
		t.Fatalf("JSON序列化失败: %v", err)
	}

	// 反序列化验证
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("JSON反序列化失败: %v", err)
	}

	// 验证关键字段
	if unmarshaled["material_id"] != "test_material" {
		t.Errorf("反序列化后素材ID不匹配")
	}

	if unmarshaled["render_index"] != float64(100) {
		t.Errorf("反序列化后渲染层级不匹配")
	}
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	// 测试缺少必要字段的情况
	invalidData := map[string]interface{}{
		"material_id": "test_material",
		// 缺少 target_timerange
	}

	_, err := NewImportedSegment(invalidData)
	if err == nil {
		t.Error("期望创建导入片段失败，但成功了")
	}

	// 测试无效的轨道类型
	invalidTrackData := map[string]interface{}{
		"type": "invalid_type",
		"name": "测试轨道",
		"id":   "test_track",
	}

	_, err = NewImportedTrack(invalidTrackData)
	if err == nil {
		t.Error("期望创建导入轨道失败，但成功了")
	}

	// 测试超出范围的片段索引
	validTrackData := map[string]interface{}{
		"type": "video",
		"name": "测试轨道",
		"id":   "test_track",
		"segments": []interface{}{
			map[string]interface{}{
				"material_id": "material_1",
				"target_timerange": map[string]interface{}{
					"start":    float64(0),
					"duration": float64(1000000),
				},
				"source_timerange": map[string]interface{}{
					"start":    float64(0),
					"duration": float64(1000000),
				},
			},
		},
	}

	mediaTrack, err := NewImportedMediaTrack(validTrackData)
	if err != nil {
		t.Fatalf("创建媒体轨道失败: %v", err)
	}

	// 测试超出范围的索引
	timerange := types.NewTimerange(0, 500000)
	err = mediaTrack.ProcessTimerange(10, timerange, ShrinkModeCutTail, nil)
	if err == nil {
		t.Error("期望超出范围的索引处理失败，但成功了")
	}

	// 测试不支持的模式
	err = mediaTrack.ProcessTimerange(0, timerange, ShrinkMode("invalid_mode"), nil)
	if err == nil {
		t.Error("期望不支持的缩短模式失败，但成功了")
	}
}
