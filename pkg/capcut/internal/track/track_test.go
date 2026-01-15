package track

import (
	"encoding/json"
	"testing"

	"novel-video-workflow/pkg/capcut/internal/segment"
	"novel-video-workflow/pkg/capcut/internal/types"
)

func TestTrackType(t *testing.T) {
	// 测试轨道类型字符串表示
	testCases := []struct {
		trackType TrackType
		expected  string
	}{
		{TrackTypeVideo, "video"},
		{TrackTypeAudio, "audio"},
		{TrackTypeEffect, "effect"},
		{TrackTypeFilter, "filter"},
		{TrackTypeSticker, "sticker"},
		{TrackTypeText, "text"},
		{TrackTypeAdjust, "adjust"},
	}

	for _, tc := range testCases {
		if tc.trackType.String() != tc.expected {
			t.Errorf("Expected TrackType %v to be '%s', got '%s'", tc.trackType, tc.expected, tc.trackType.String())
		}
	}
}

func TestTrackTypeFromName(t *testing.T) {
	// 测试从名称获取轨道类型
	testCases := []struct {
		name     string
		expected TrackType
		hasError bool
	}{
		{"video", TrackTypeVideo, false},
		{"audio", TrackTypeAudio, false},
		{"text", TrackTypeText, false},
		{"invalid", 0, true},
	}

	for _, tc := range testCases {
		result, err := TrackTypeFromName(tc.name)
		if tc.hasError {
			if err == nil {
				t.Errorf("Expected error for name '%s', but got none", tc.name)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for name '%s': %v", tc.name, err)
			}
			if result != tc.expected {
				t.Errorf("Expected TrackType %v for name '%s', got %v", tc.expected, tc.name, result)
			}
		}
	}
}

func TestTrackTypeMeta(t *testing.T) {
	// 测试轨道类型元数据
	videoMeta := TrackTypeVideo.Meta()
	if videoMeta.RenderIndex != 0 {
		t.Errorf("Expected video render index 0, got %d", videoMeta.RenderIndex)
	}
	if !videoMeta.AllowModify {
		t.Error("Expected video to allow modify")
	}

	textMeta := TrackTypeText.Meta()
	if textMeta.RenderIndex != 15000 {
		t.Errorf("Expected text render index 15000, got %d", textMeta.RenderIndex)
	}
	if !textMeta.AllowModify {
		t.Error("Expected text to allow modify")
	}

	effectMeta := TrackTypeEffect.Meta()
	if effectMeta.RenderIndex != 10000 {
		t.Errorf("Expected effect render index 10000, got %d", effectMeta.RenderIndex)
	}
	if effectMeta.AllowModify {
		t.Error("Expected effect not to allow modify")
	}
}

func TestNewTrack(t *testing.T) {
	// 测试创建新轨道
	track := NewTrack(TrackTypeVideo, "主视频轨道", 0, false)

	if track.TrackType != TrackTypeVideo {
		t.Errorf("Expected track type video, got %v", track.TrackType)
	}

	if track.Name != "主视频轨道" {
		t.Errorf("Expected track name '主视频轨道', got '%s'", track.Name)
	}

	if track.TrackID == "" {
		t.Error("Expected non-empty track ID")
	}

	if track.RenderIndex != 0 {
		t.Errorf("Expected render index 0, got %d", track.RenderIndex)
	}

	if track.Mute != false {
		t.Error("Expected track not to be muted")
	}

	if len(track.Segments) != 0 {
		t.Errorf("Expected empty segments, got %d", len(track.Segments))
	}

	if len(track.PendingKeyframes) != 0 {
		t.Errorf("Expected empty pending keyframes, got %d", len(track.PendingKeyframes))
	}
}

func TestTrackWithDefaultRenderIndex(t *testing.T) {
	// 测试使用默认渲染索引
	track := NewTrack(TrackTypeText, "字幕轨道", 0, false)

	// 应该使用文本轨道的默认渲染索引15000
	if track.RenderIndex != 15000 {
		t.Errorf("Expected default render index 15000, got %d", track.RenderIndex)
	}
}

func TestAddPendingKeyframe(t *testing.T) {
	// 测试添加待处理关键帧
	track := NewTrack(TrackTypeVideo, "test_track", 100, false)

	track.AddPendingKeyframe("alpha", 2.5, "80%")
	track.AddPendingKeyframe("rotation", 5.0, "45deg")

	if len(track.PendingKeyframes) != 2 {
		t.Errorf("Expected 2 pending keyframes, got %d", len(track.PendingKeyframes))
	}

	firstKf := track.PendingKeyframes[0]
	if firstKf.PropertyType != "alpha" {
		t.Errorf("Expected property type 'alpha', got '%s'", firstKf.PropertyType)
	}
	if firstKf.Time != 2.5 {
		t.Errorf("Expected time 2.5, got %f", firstKf.Time)
	}
	if firstKf.Value != "80%" {
		t.Errorf("Expected value '80%%', got '%s'", firstKf.Value)
	}
}

func TestProcessPendingKeyframes(t *testing.T) {
	// 测试处理待处理关键帧
	track := NewTrack(TrackTypeVideo, "test_track", 0, false)

	track.AddPendingKeyframe("alpha", 1.0, "50%")
	track.AddPendingKeyframe("volume", 2.0, "75%")

	err := track.ProcessPendingKeyframes()
	if err != nil {
		t.Errorf("Unexpected error processing keyframes: %v", err)
	}

	// 处理后应该清空待处理关键帧
	if len(track.PendingKeyframes) != 0 {
		t.Errorf("Expected empty pending keyframes after processing, got %d", len(track.PendingKeyframes))
	}
}

func TestEndTime(t *testing.T) {
	// 测试轨道结束时间
	track := NewTrack(TrackTypeVideo, "test_track", 0, false)

	// 空轨道的结束时间应该是0
	if track.EndTime() != 0 {
		t.Errorf("Expected end time 0 for empty track, got %d", track.EndTime())
	}

	// 添加一些片段
	timerange1, _ := types.Trange("1s", "3s")
	timerange2, _ := types.Trange("5s", "2s")

	segment1 := segment.NewVideoSegment("video1", nil, timerange1, 1.0, 1.0, nil)
	segment2 := segment.NewVideoSegment("video2", nil, timerange2, 1.0, 1.0, nil)

	track.AddSegment(segment1)
	track.AddSegment(segment2)

	// 轨道结束时间应该是最后一个片段的结束时间
	expectedEndTime := int64(7000000) // 5s + 2s = 7s
	if track.EndTime() != expectedEndTime {
		t.Errorf("Expected end time %d, got %d", expectedEndTime, track.EndTime())
	}
}

func TestAddSegment(t *testing.T) {
	// 测试添加片段
	track := NewTrack(TrackTypeVideo, "video_track", 0, false)

	timerange, _ := types.Trange("1s", "3s")
	videoSegment := segment.NewVideoSegment("video_material", nil, timerange, 1.0, 1.0, nil)

	err := track.AddSegment(videoSegment)
	if err != nil {
		t.Errorf("Unexpected error adding segment: %v", err)
	}

	if len(track.Segments) != 1 {
		t.Errorf("Expected 1 segment, got %d", len(track.Segments))
	}
}

func TestAddSegmentOverlap(t *testing.T) {
	// 测试添加重叠片段
	track := NewTrack(TrackTypeVideo, "video_track", 0, false)

	timerange1, _ := types.Trange("1s", "3s")
	timerange2, _ := types.Trange("2s", "3s") // 与第一个片段重叠

	segment1 := segment.NewVideoSegment("video1", nil, timerange1, 1.0, 1.0, nil)
	segment2 := segment.NewVideoSegment("video2", nil, timerange2, 1.0, 1.0, nil)

	// 添加第一个片段应该成功
	err := track.AddSegment(segment1)
	if err != nil {
		t.Errorf("Unexpected error adding first segment: %v", err)
	}

	// 添加重叠片段应该失败
	err = track.AddSegment(segment2)
	if err == nil {
		t.Error("Expected error when adding overlapping segment, but got none")
	}
}

func TestAddSegmentNonOverlapping(t *testing.T) {
	// 测试添加不重叠的片段
	track := NewTrack(TrackTypeVideo, "video_track", 0, false)

	timerange1, _ := types.Trange("1s", "2s")
	timerange2, _ := types.Trange("4s", "2s") // 不重叠

	segment1 := segment.NewVideoSegment("video1", nil, timerange1, 1.0, 1.0, nil)
	segment2 := segment.NewVideoSegment("video2", nil, timerange2, 1.0, 1.0, nil)

	err := track.AddSegment(segment1)
	if err != nil {
		t.Errorf("Unexpected error adding first segment: %v", err)
	}

	err = track.AddSegment(segment2)
	if err != nil {
		t.Errorf("Unexpected error adding second segment: %v", err)
	}

	if len(track.Segments) != 2 {
		t.Errorf("Expected 2 segments, got %d", len(track.Segments))
	}
}

func TestTrackExportJSON(t *testing.T) {
	// 测试轨道JSON导出
	track := NewTrack(TrackTypeVideo, "主轨道", 100, true)

	// 添加一个片段
	timerange, _ := types.Trange("2s", "5s")
	videoSegment := segment.NewVideoSegment("video_material", nil, timerange, 1.0, 0.8, nil)
	track.AddSegment(videoSegment)

	jsonData := track.ExportJSON()

	// 验证JSON结构
	expectedFields := []string{"attribute", "flag", "id", "is_default_name", "name", "segments", "type"}
	for _, field := range expectedFields {
		if _, exists := jsonData[field]; !exists {
			t.Errorf("Missing field '%s' in JSON export", field)
		}
	}

	// 验证具体值
	if jsonData["attribute"] != 1 { // mute = true
		t.Errorf("Expected attribute 1, got %v", jsonData["attribute"])
	}

	if jsonData["name"] != "主轨道" {
		t.Errorf("Expected name '主轨道', got %v", jsonData["name"])
	}

	if jsonData["type"] != "video" {
		t.Errorf("Expected type 'video', got %v", jsonData["type"])
	}

	if jsonData["is_default_name"] != false {
		t.Error("Expected is_default_name to be false")
	}

	// 验证片段数组
	segments := jsonData["segments"].([]interface{})
	if len(segments) != 1 {
		t.Errorf("Expected 1 segment in JSON, got %d", len(segments))
	}

	// 验证片段包含render_index
	segmentJSON := segments[0].(map[string]interface{})
	if segmentJSON["render_index"] != 100 {
		t.Errorf("Expected segment render_index 100, got %v", segmentJSON["render_index"])
	}
}

func TestTrackJSONSerialization(t *testing.T) {
	// 测试JSON序列化
	track := NewTrack(TrackTypeAudio, "音频轨道", 0, false)

	// 添加音频片段
	timerange, _ := types.Trange("0s", "10s")
	audioSegment := segment.NewAudioSegment("audio_material", timerange, nil, 1.0, 1.0)
	track.AddSegment(audioSegment)

	jsonData := track.ExportJSON()

	// 验证可以序列化为JSON
	_, err := json.Marshal(jsonData)
	if err != nil {
		t.Errorf("Failed to marshal track JSON: %v", err)
	}
}

func TestTrackString(t *testing.T) {
	// 测试字符串表示
	track := NewTrack(TrackTypeText, "字幕", 15000, false)

	str := track.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}

	// 字符串应该包含关键信息
	expectedContents := []string{"text", "字幕", track.TrackID, "0"}
	for _, content := range expectedContents {
		if !containsString(str, content) {
			t.Errorf("Expected string to contain '%s', got '%s'", content, str)
		}
	}
}

// containsString 检查字符串是否包含子字符串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
