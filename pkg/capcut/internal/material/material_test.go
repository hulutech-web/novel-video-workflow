package material

import (
	"encoding/json"
	"testing"
)

func TestCropSettings(t *testing.T) {
	// 测试默认裁剪设置
	defaultCrop := NewCropSettings()

	if defaultCrop.UpperLeftX != 0.0 {
		t.Errorf("Expected UpperLeftX 0.0, got %f", defaultCrop.UpperLeftX)
	}

	if defaultCrop.LowerRightX != 1.0 {
		t.Errorf("Expected LowerRightX 1.0, got %f", defaultCrop.LowerRightX)
	}

	if defaultCrop.LowerRightY != 1.0 {
		t.Errorf("Expected LowerRightY 1.0, got %f", defaultCrop.LowerRightY)
	}

	// 测试自定义裁剪设置
	customCrop := NewCropSettingsWithParams(0.1, 0.1, 0.9, 0.1, 0.1, 0.9, 0.9, 0.9)

	if customCrop.UpperLeftX != 0.1 {
		t.Errorf("Expected UpperLeftX 0.1, got %f", customCrop.UpperLeftX)
	}

	if customCrop.LowerRightX != 0.9 {
		t.Errorf("Expected LowerRightX 0.9, got %f", customCrop.LowerRightX)
	}

	// 测试JSON导出
	jsonData := customCrop.ExportJSON()
	if jsonData["upper_left_x"] != 0.1 {
		t.Errorf("Expected JSON upper_left_x 0.1, got %v", jsonData["upper_left_x"])
	}

	if jsonData["lower_right_x"] != 0.9 {
		t.Errorf("Expected JSON lower_right_x 0.9, got %v", jsonData["lower_right_x"])
	}
}

func TestVideoMaterialCreation(t *testing.T) {
	// 测试视频素材创建（使用远程URL）
	remoteURL := "https://example.com/video.mp4"
	materialName := "test_video.mp4"
	duration := 30.0
	width := 1920
	height := 1080

	material, err := NewVideoMaterial(
		MaterialTypeVideo,
		nil, // path
		nil, // replacePath
		&materialName,
		&remoteURL,
		nil, // cropSettings - 使用默认
		&duration,
		&width,
		&height,
	)

	if err != nil {
		t.Fatalf("Failed to create video material: %v", err)
	}

	if material.MaterialName != "test_video.mp4" {
		t.Errorf("Expected material name 'test_video.mp4', got '%s'", material.MaterialName)
	}

	if material.MaterialType != MaterialTypeVideo {
		t.Errorf("Expected material type 'video', got '%s'", material.MaterialType)
	}

	if material.Duration != 30000000 { // 30秒 = 30,000,000微秒
		t.Errorf("Expected duration 30000000, got %d", material.Duration)
	}

	if material.Width != 1920 {
		t.Errorf("Expected width 1920, got %d", material.Width)
	}

	if material.Height != 1080 {
		t.Errorf("Expected height 1080, got %d", material.Height)
	}

	if material.RemoteURL == nil || *material.RemoteURL != remoteURL {
		t.Errorf("Expected remote URL '%s', got %v", remoteURL, material.RemoteURL)
	}

	if material.Path != "" {
		t.Errorf("Expected empty path for remote material, got '%s'", material.Path)
	}
}

func TestPhotoMaterialCreation(t *testing.T) {
	// 测试图片素材创建
	remoteURL := "https://example.com/image.jpg"
	materialName := "test_image.jpg"

	material, err := NewVideoMaterial(
		MaterialTypePhoto,
		nil, // path
		nil, // replacePath
		&materialName,
		&remoteURL,
		nil, // cropSettings
		nil, // duration - 图片不需要
		nil, // width - 使用默认
		nil, // height - 使用默认
	)

	if err != nil {
		t.Fatalf("Failed to create photo material: %v", err)
	}

	if material.MaterialType != MaterialTypePhoto {
		t.Errorf("Expected material type 'photo', got '%s'", material.MaterialType)
	}

	if material.Duration != 10800000000 { // 3小时默认时长
		t.Errorf("Expected duration 10800000000, got %d", material.Duration)
	}

	if material.Width != 1920 { // 默认宽度
		t.Errorf("Expected default width 1920, got %d", material.Width)
	}

	if material.Height != 1080 { // 默认高度
		t.Errorf("Expected default height 1080, got %d", material.Height)
	}
}

func TestVideoMaterialValidation(t *testing.T) {
	// 测试缺少必要参数的情况
	_, err := NewVideoMaterial(
		MaterialTypeVideo,
		nil, // path
		nil, // replacePath
		nil, // materialName
		nil, // remoteURL - 都为空
		nil, // cropSettings
		nil, // duration
		nil, // width
		nil, // height
	)

	if err == nil {
		t.Error("Expected error when both path and remoteURL are empty")
	}

	// 测试远程URL但缺少materialName的情况
	remoteURL := "https://example.com/video.mp4"
	_, err = NewVideoMaterial(
		MaterialTypeVideo,
		nil, // path
		nil, // replacePath
		nil, // materialName - 使用远程URL时必须提供
		&remoteURL,
		nil, // cropSettings
		nil, // duration
		nil, // width
		nil, // height
	)

	if err == nil {
		t.Error("Expected error when using remoteURL without materialName")
	}
}

func TestVideoMaterialFromDict(t *testing.T) {
	// 测试从字典创建视频素材
	data := map[string]interface{}{
		"id":                "test_id_123",
		"local_material_id": "local_123",
		"material_name":     "test_video.mp4",
		"path":              "/path/to/video.mp4",
		"remote_url":        "https://example.com/video.mp4",
		"type":              "video",
		"duration":          float64(30000000),
		"width":             1920,
		"height":            1080,
		"crop": map[string]interface{}{
			"upper_left_x":  0.1,
			"upper_left_y":  0.1,
			"upper_right_x": 0.9,
			"upper_right_y": 0.1,
			"lower_left_x":  0.1,
			"lower_left_y":  0.9,
			"lower_right_x": 0.9,
			"lower_right_y": 0.9,
		},
	}

	material, err := NewVideoMaterialFromDict(data)
	if err != nil {
		t.Fatalf("Failed to create material from dict: %v", err)
	}

	if material.MaterialID != "test_id_123" {
		t.Errorf("Expected MaterialID 'test_id_123', got '%s'", material.MaterialID)
	}

	if material.MaterialName != "test_video.mp4" {
		t.Errorf("Expected MaterialName 'test_video.mp4', got '%s'", material.MaterialName)
	}

	if material.Duration != 30000000 {
		t.Errorf("Expected Duration 30000000, got %d", material.Duration)
	}

	if material.CropSettings.UpperLeftX != 0.1 {
		t.Errorf("Expected CropSettings.UpperLeftX 0.1, got %f", material.CropSettings.UpperLeftX)
	}
}

func TestVideoMaterialExportJSON(t *testing.T) {
	// 创建视频素材
	remoteURL := "https://example.com/video.mp4"
	materialName := "test_video.mp4"
	replacePath := "/replaced/path/video.mp4"
	duration := 25.0
	width := 1280
	height := 720

	material, err := NewVideoMaterial(
		MaterialTypeVideo,
		nil,
		&replacePath,
		&materialName,
		&remoteURL,
		nil,
		&duration,
		&width,
		&height,
	)

	if err != nil {
		t.Fatalf("Failed to create video material: %v", err)
	}

	// 测试JSON导出
	jsonData := material.ExportJSON()

	// 验证必要字段
	requiredFields := []string{
		"id", "material_name", "path", "remote_url", "type",
		"duration", "width", "height", "crop", "category_name",
	}

	for _, field := range requiredFields {
		if _, exists := jsonData[field]; !exists {
			t.Errorf("Missing required field '%s' in JSON export", field)
		}
	}

	// 验证具体值
	if jsonData["type"] != "video" {
		t.Errorf("Expected type 'video', got %v", jsonData["type"])
	}

	if jsonData["duration"] != int64(25000000) {
		t.Errorf("Expected duration 25000000, got %v", jsonData["duration"])
	}

	if jsonData["width"] != 1280 {
		t.Errorf("Expected width 1280, got %v", jsonData["width"])
	}

	if jsonData["height"] != 720 {
		t.Errorf("Expected height 720, got %v", jsonData["height"])
	}

	if jsonData["path"] != replacePath {
		t.Errorf("Expected path to use replacePath '%s', got %v", replacePath, jsonData["path"])
	}

	if jsonData["remote_url"] != remoteURL {
		t.Errorf("Expected remote_url '%s', got %v", remoteURL, jsonData["remote_url"])
	}

	// 验证可以序列化为JSON
	_, err = json.Marshal(jsonData)
	if err != nil {
		t.Errorf("Failed to marshal video material JSON: %v", err)
	}
}

func TestAudioMaterialCreation(t *testing.T) {
	// 测试音频素材创建
	remoteURL := "https://example.com/audio.mp3"
	materialName := "test_audio.mp3"
	duration := 180.0 // 3分钟

	material, err := NewAudioMaterial(
		nil, // path
		nil, // replacePath
		&materialName,
		&remoteURL,
		&duration,
	)

	if err != nil {
		t.Fatalf("Failed to create audio material: %v", err)
	}

	if material.MaterialName != "test_audio.mp3" {
		t.Errorf("Expected material name 'test_audio.mp3', got '%s'", material.MaterialName)
	}

	if material.Duration != 180000000 { // 180秒 = 180,000,000微秒
		t.Errorf("Expected duration 180000000, got %d", material.Duration)
	}

	if material.RemoteURL == nil || *material.RemoteURL != remoteURL {
		t.Errorf("Expected remote URL '%s', got %v", remoteURL, material.RemoteURL)
	}

	if material.HasAudioEffect != false {
		t.Error("Expected HasAudioEffect to be false by default")
	}
}

func TestAudioMaterialFromURL(t *testing.T) {
	// 测试从URL自动生成音频素材名称
	remoteURL := "https://example.com/path/to/song.wav?param=123"

	material, err := NewAudioMaterial(
		nil, // path
		nil, // replacePath
		nil, // materialName - 自动从URL生成
		&remoteURL,
		nil, // duration - 使用默认
	)

	if err != nil {
		t.Fatalf("Failed to create audio material: %v", err)
	}

	// 应该从URL中提取文件名并转换为mp3
	expectedName := "song.mp3"
	if material.MaterialName != expectedName {
		t.Errorf("Expected auto-generated name '%s', got '%s'", expectedName, material.MaterialName)
	}

	if material.Duration != 180000000 { // 默认3分钟
		t.Errorf("Expected default duration 180000000, got %d", material.Duration)
	}
}

func TestAudioMaterialFromDict(t *testing.T) {
	// 测试从字典创建音频素材
	data := map[string]interface{}{
		"id":         "audio_id_456",
		"name":       "background_music.mp3",
		"path":       "/path/to/audio.mp3",
		"remote_url": "https://example.com/audio.mp3",
		"duration":   float64(240000000), // 4分钟
	}

	material, err := NewAudioMaterialFromDict(data)
	if err != nil {
		t.Fatalf("Failed to create audio material from dict: %v", err)
	}

	if material.MaterialID != "audio_id_456" {
		t.Errorf("Expected MaterialID 'audio_id_456', got '%s'", material.MaterialID)
	}

	if material.MaterialName != "background_music.mp3" {
		t.Errorf("Expected MaterialName 'background_music.mp3', got '%s'", material.MaterialName)
	}

	if material.Duration != 240000000 {
		t.Errorf("Expected Duration 240000000, got %d", material.Duration)
	}
}

func TestAudioMaterialExportJSON(t *testing.T) {
	// 创建音频素材（使用远程URL避免文件不存在问题）
	remoteURL := "https://example.com/audio.mp3"
	materialName := "test_audio.mp3"
	duration := 120.0

	material, err := NewAudioMaterial(
		nil, // path
		nil, // replacePath
		&materialName,
		&remoteURL,
		&duration,
	)

	if err != nil {
		t.Fatalf("Failed to create audio material: %v", err)
	}

	// 设置音频效果标志
	material.HasAudioEffect = true

	// 测试JSON导出
	jsonData := material.ExportJSON()

	// 验证必要字段
	requiredFields := []string{
		"id", "name", "path", "type", "duration",
		"category_name", "check_flag", "local_material_id",
	}

	for _, field := range requiredFields {
		if _, exists := jsonData[field]; !exists {
			t.Errorf("Missing required field '%s' in JSON export", field)
		}
	}

	// 验证具体值
	if jsonData["type"] != "extract_music" {
		t.Errorf("Expected type 'extract_music', got %v", jsonData["type"])
	}

	if jsonData["duration"] != int64(120000000) {
		t.Errorf("Expected duration 120000000, got %v", jsonData["duration"])
	}

	if jsonData["check_flag"] != 3 { // 有音频效果时为3
		t.Errorf("Expected check_flag 3 (with audio effect), got %v", jsonData["check_flag"])
	}

	if jsonData["category_name"] != "local" {
		t.Errorf("Expected category_name 'local', got %v", jsonData["category_name"])
	}

	if jsonData["remote_url"] != remoteURL {
		t.Errorf("Expected remote_url '%s', got %v", remoteURL, jsonData["remote_url"])
	}

	// 验证可以序列化为JSON
	_, err = json.Marshal(jsonData)
	if err != nil {
		t.Errorf("Failed to marshal audio material JSON: %v", err)
	}
}

func TestMaterialIDGeneration(t *testing.T) {
	// 测试素材ID生成的一致性
	materialName := "test_material.mp4"

	id1 := generateMaterialID(materialName)
	id2 := generateMaterialID(materialName)

	if id1 != id2 {
		t.Errorf("Expected consistent material ID generation, got '%s' and '%s'", id1, id2)
	}

	if len(id1) != 32 { // UUID without hyphens
		t.Errorf("Expected material ID length 32, got %d", len(id1))
	}

	// 不同名称应该生成不同ID
	differentName := "different_material.mp4"
	differentID := generateMaterialID(differentName)

	if id1 == differentID {
		t.Error("Expected different IDs for different material names")
	}
}

func TestMaterialInterface(t *testing.T) {
	// 测试MaterialInterface接口实现

	// 视频素材
	remoteURL := "https://example.com/video.mp4"
	materialName := "test_video.mp4"
	duration := 30.0

	videoMaterial, err := NewVideoMaterial(
		MaterialTypeVideo,
		nil,
		nil,
		&materialName,
		&remoteURL,
		nil,
		&duration,
		nil,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create video material: %v", err)
	}

	// 测试接口方法
	var material MaterialInterface = videoMaterial

	if material.GetMaterialID() == "" {
		t.Error("Expected non-empty material ID")
	}

	if material.GetMaterialName() != materialName {
		t.Errorf("Expected material name '%s', got '%s'", materialName, material.GetMaterialName())
	}

	if material.GetDuration() != int64(30000000) {
		t.Errorf("Expected duration 30000000, got %d", material.GetDuration())
	}

	// 音频素材
	audioMaterial, err := NewAudioMaterial(
		nil,
		nil,
		&materialName,
		&remoteURL,
		&duration,
	)

	if err != nil {
		t.Fatalf("Failed to create audio material: %v", err)
	}

	var audioInterface MaterialInterface = audioMaterial

	if audioInterface.GetMaterialID() == "" {
		t.Error("Expected non-empty audio material ID")
	}

	if audioInterface.GetMaterialName() != materialName {
		t.Errorf("Expected audio material name '%s', got '%s'", materialName, audioInterface.GetMaterialName())
	}

	if audioInterface.GetDuration() != int64(30000000) {
		t.Errorf("Expected audio duration 30000000, got %d", audioInterface.GetDuration())
	}
}
