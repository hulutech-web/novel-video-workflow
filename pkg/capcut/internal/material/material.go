// Package material 定义素材管理系统
// 对应Python的 local_materials.py
package material

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// CropSettings 素材的裁剪设置，各属性均在0-1之间，注意素材的坐标原点在左上角
// 对应Python的Crop_settings类
type CropSettings struct {
	UpperLeftX  float64 `json:"upper_left_x"`  // 左上角X坐标
	UpperLeftY  float64 `json:"upper_left_y"`  // 左上角Y坐标
	UpperRightX float64 `json:"upper_right_x"` // 右上角X坐标
	UpperRightY float64 `json:"upper_right_y"` // 右上角Y坐标
	LowerLeftX  float64 `json:"lower_left_x"`  // 左下角X坐标
	LowerLeftY  float64 `json:"lower_left_y"`  // 左下角Y坐标
	LowerRightX float64 `json:"lower_right_x"` // 右下角X坐标
	LowerRightY float64 `json:"lower_right_y"` // 右下角Y坐标
}

// NewCropSettings 创建新的裁剪设置，默认参数表示不裁剪
func NewCropSettings() *CropSettings {
	return &CropSettings{
		UpperLeftX:  0.0,
		UpperLeftY:  0.0,
		UpperRightX: 1.0,
		UpperRightY: 0.0,
		LowerLeftX:  0.0,
		LowerLeftY:  1.0,
		LowerRightX: 1.0,
		LowerRightY: 1.0,
	}
}

// NewCropSettingsWithParams 创建带参数的裁剪设置
func NewCropSettingsWithParams(upperLeftX, upperLeftY, upperRightX, upperRightY,
	lowerLeftX, lowerLeftY, lowerRightX, lowerRightY float64) *CropSettings {
	return &CropSettings{
		UpperLeftX:  upperLeftX,
		UpperLeftY:  upperLeftY,
		UpperRightX: upperRightX,
		UpperRightY: upperRightY,
		LowerLeftX:  lowerLeftX,
		LowerLeftY:  lowerLeftY,
		LowerRightX: lowerRightX,
		LowerRightY: lowerRightY,
	}
}

// ExportJSON 导出为JSON格式
func (cs *CropSettings) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"upper_left_x":  cs.UpperLeftX,
		"upper_left_y":  cs.UpperLeftY,
		"upper_right_x": cs.UpperRightX,
		"upper_right_y": cs.UpperRightY,
		"lower_left_x":  cs.LowerLeftX,
		"lower_left_y":  cs.LowerLeftY,
		"lower_right_x": cs.LowerRightX,
		"lower_right_y": cs.LowerRightY,
	}
}

// MaterialType 素材类型
type MaterialType string

const (
	MaterialTypeVideo MaterialType = "video"
	MaterialTypePhoto MaterialType = "photo"
	MaterialTypeAudio MaterialType = "extract_music"
)

// VideoMaterial 本地视频素材（视频或图片），一份素材可以在多个片段中使用
// 对应Python的Video_material类
type VideoMaterial struct {
	MaterialID      string        `json:"id"`                // 素材全局id，自动生成
	LocalMaterialID string        `json:"local_material_id"` // 素材本地id，意义暂不明确
	MaterialName    string        `json:"material_name"`     // 素材名称
	Path            string        `json:"path"`              // 素材文件路径
	RemoteURL       *string       `json:"remote_url"`        // 远程URL地址
	Duration        int64         `json:"duration"`          // 素材时长，单位为微秒
	Height          int           `json:"height"`            // 素材高度
	Width           int           `json:"width"`             // 素材宽度
	CropSettings    *CropSettings `json:"crop_settings"`     // 素材裁剪设置
	MaterialType    MaterialType  `json:"type"`              // 素材类型：视频或图片
	ReplacePath     *string       `json:"replace_path"`      // 替换路径，如果设置了这个值，在导出json时会用这个路径替代原始path
}

// NewVideoMaterial 创建新的视频素材
func NewVideoMaterial(materialType MaterialType, path, replacePath, materialName, remoteURL *string,
	cropSettings *CropSettings, duration *float64, width, height *int) (*VideoMaterial, error) {

	// 确保至少提供了path或remoteURL
	if (path == nil || *path == "") && (remoteURL == nil || *remoteURL == "") {
		return nil, fmt.Errorf("必须提供 path 或 remoteURL 中的至少一个参数")
	}

	var finalPath string
	var finalRemoteURL *string
	var finalMaterialName string

	// 处理远程URL情况
	if remoteURL != nil && *remoteURL != "" {
		if materialName == nil || *materialName == "" {
			return nil, fmt.Errorf("使用 remoteURL 参数时必须指定 materialName")
		}
		finalRemoteURL = remoteURL
		finalPath = "" // 远程资源没有本地路径
		finalMaterialName = *materialName
	} else {
		// 处理本地文件情况
		absPath, err := filepath.Abs(*path)
		if err != nil {
			return nil, fmt.Errorf("无法获取绝对路径: %w", err)
		}
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("找不到文件: %s", absPath)
		}
		finalPath = absPath
		finalRemoteURL = nil

		if materialName != nil && *materialName != "" {
			finalMaterialName = *materialName
		} else {
			finalMaterialName = filepath.Base(absPath)
		}
	}

	// 生成素材ID
	materialID := generateMaterialID(finalMaterialName)

	// 设置默认裁剪设置
	if cropSettings == nil {
		cropSettings = NewCropSettings()
	}

	// 创建素材对象
	material := &VideoMaterial{
		MaterialID:      materialID,
		LocalMaterialID: "",
		MaterialName:    finalMaterialName,
		Path:            finalPath,
		RemoteURL:       finalRemoteURL,
		CropSettings:    cropSettings,
		MaterialType:    materialType,
		ReplacePath:     replacePath,
	}

	// 如果是photo类型，跳过媒体信息获取
	if materialType == MaterialTypePhoto {
		material.Duration = 10800000000 // 静态图片默认3小时（微秒）
		// 使用默认宽高，实际使用时可能需要获取真实尺寸
		if width != nil {
			material.Width = *width
		} else {
			material.Width = 1920
		}
		if height != nil {
			material.Height = *height
		} else {
			material.Height = 1080
		}
		return material, nil
	}

	// 如果外部提供了所有参数，直接使用
	if duration != nil && width != nil && height != nil {
		material.Duration = int64(*duration * 1e6) // 转换为微秒
		material.Width = *width
		material.Height = *height
		return material, nil
	}

	// TODO: 实现ffprobe媒体信息获取
	// 目前使用默认值，实际项目中需要调用ffprobe获取真实信息
	if duration != nil {
		material.Duration = int64(*duration * 1e6)
	} else {
		material.Duration = 30000000 // 默认30秒
	}

	if width != nil {
		material.Width = *width
	} else {
		material.Width = 1920
	}

	if height != nil {
		material.Height = *height
	} else {
		material.Height = 1080
	}

	return material, nil
}

// NewVideoMaterialFromDict 从字典创建视频素材对象
func NewVideoMaterialFromDict(data map[string]interface{}) (*VideoMaterial, error) {
	material := &VideoMaterial{}

	// 基本字符串字段
	if id, ok := data["id"].(string); ok {
		material.MaterialID = id
	}

	if localID, ok := data["local_material_id"].(string); ok {
		material.LocalMaterialID = localID
	}

	if name, ok := data["material_name"].(string); ok {
		material.MaterialName = name
	}

	if path, ok := data["path"].(string); ok {
		material.Path = path
	}

	if remoteURL, ok := data["remote_url"].(string); ok && remoteURL != "" {
		material.RemoteURL = &remoteURL
	}

	if matType, ok := data["type"].(string); ok {
		material.MaterialType = MaterialType(matType)
	}

	// 数值字段
	if duration, ok := data["duration"]; ok {
		switch v := duration.(type) {
		case float64:
			material.Duration = int64(v)
		case int:
			material.Duration = int64(v)
		case int64:
			material.Duration = v
		}
	}

	if width, ok := data["width"]; ok {
		switch v := width.(type) {
		case float64:
			material.Width = int(v)
		case int:
			material.Width = v
		}
	}

	if height, ok := data["height"]; ok {
		switch v := height.(type) {
		case float64:
			material.Height = int(v)
		case int:
			material.Height = v
		}
	}

	// 裁剪设置
	if cropData, ok := data["crop"].(map[string]interface{}); ok {
		material.CropSettings = &CropSettings{}
		if val, exists := cropData["upper_left_x"]; exists {
			if f, ok := val.(float64); ok {
				material.CropSettings.UpperLeftX = f
			}
		}
		if val, exists := cropData["upper_left_y"]; exists {
			if f, ok := val.(float64); ok {
				material.CropSettings.UpperLeftY = f
			}
		}
		if val, exists := cropData["upper_right_x"]; exists {
			if f, ok := val.(float64); ok {
				material.CropSettings.UpperRightX = f
			}
		}
		if val, exists := cropData["upper_right_y"]; exists {
			if f, ok := val.(float64); ok {
				material.CropSettings.UpperRightY = f
			}
		}
		if val, exists := cropData["lower_left_x"]; exists {
			if f, ok := val.(float64); ok {
				material.CropSettings.LowerLeftX = f
			}
		}
		if val, exists := cropData["lower_left_y"]; exists {
			if f, ok := val.(float64); ok {
				material.CropSettings.LowerLeftY = f
			}
		}
		if val, exists := cropData["lower_right_x"]; exists {
			if f, ok := val.(float64); ok {
				material.CropSettings.LowerRightX = f
			}
		}
		if val, exists := cropData["lower_right_y"]; exists {
			if f, ok := val.(float64); ok {
				material.CropSettings.LowerRightY = f
			}
		}
	} else {
		material.CropSettings = NewCropSettings()
	}

	return material, nil
}

// ExportJSON 导出为JSON格式
func (vm *VideoMaterial) ExportJSON() map[string]interface{} {
	pathValue := vm.Path
	if vm.ReplacePath != nil {
		pathValue = *vm.ReplacePath
	}

	var remoteURLValue interface{} = nil
	if vm.RemoteURL != nil {
		remoteURLValue = *vm.RemoteURL
	}

	return map[string]interface{}{
		"audio_fade":        nil,
		"category_id":       "",
		"category_name":     "local",
		"check_flag":        63487,
		"crop":              vm.CropSettings.ExportJSON(),
		"crop_ratio":        "free",
		"crop_scale":        1.0,
		"duration":          vm.Duration,
		"height":            vm.Height,
		"id":                vm.MaterialID,
		"local_material_id": vm.LocalMaterialID,
		"material_id":       vm.MaterialID,
		"material_name":     vm.MaterialName,
		"media_path":        "",
		"path":              pathValue,
		"remote_url":        remoteURLValue,
		"type":              string(vm.MaterialType),
		"width":             vm.Width,
	}
}

// AudioMaterial 本地音频素材
// 对应Python的Audio_material类
type AudioMaterial struct {
	MaterialID     string  `json:"id"`               // 素材全局id，自动生成
	MaterialName   string  `json:"name"`             // 素材名称
	Path           string  `json:"path"`             // 素材文件路径
	RemoteURL      *string `json:"remote_url"`       // 远程URL地址
	ReplacePath    *string `json:"replace_path"`     // 替换路径
	HasAudioEffect bool    `json:"has_audio_effect"` // 是否有音频效果
	Duration       int64   `json:"duration"`         // 素材时长，单位为微秒
}

// NewAudioMaterial 创建新的音频素材
func NewAudioMaterial(path, replacePath, materialName, remoteURL *string, duration *float64) (*AudioMaterial, error) {
	// 确保至少提供了path或remoteURL
	if (path == nil || *path == "") && (remoteURL == nil || *remoteURL == "") {
		return nil, fmt.Errorf("必须提供 path 或 remoteURL 中的至少一个参数")
	}

	var finalPath string
	var finalRemoteURL *string
	var finalMaterialName string

	// 处理路径和名称
	if path != nil && *path != "" {
		absPath, err := filepath.Abs(*path)
		if err != nil {
			return nil, fmt.Errorf("无法获取绝对路径: %w", err)
		}
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("找不到文件: %s", absPath)
		}
		finalPath = absPath

		if materialName != nil && *materialName != "" {
			finalMaterialName = *materialName
		} else {
			finalMaterialName = filepath.Base(absPath)
		}
	}

	if remoteURL != nil && *remoteURL != "" {
		finalRemoteURL = remoteURL
		if materialName == nil || *materialName == "" {
			// 从URL中获取文件名
			urlPath := strings.Split(*remoteURL, "?")[0] // 去除查询参数
			originalFilename := filepath.Base(urlPath)
			nameWithoutExt := strings.TrimSuffix(originalFilename, filepath.Ext(originalFilename))
			finalMaterialName = nameWithoutExt + ".mp3" // 使用原始文件名+固定mp3扩展名
		} else {
			finalMaterialName = *materialName
		}
	}

	// 生成素材ID
	materialID := generateMaterialID(finalMaterialName)

	// 创建音频素材对象
	material := &AudioMaterial{
		MaterialID:     materialID,
		MaterialName:   finalMaterialName,
		Path:           finalPath,
		RemoteURL:      finalRemoteURL,
		ReplacePath:    replacePath,
		HasAudioEffect: false,
	}

	// 设置时长
	if duration != nil {
		material.Duration = int64(*duration * 1e6) // 转换为微秒
	} else {
		// TODO: 实现ffprobe音频信息获取
		// 目前使用默认值，实际项目中需要调用ffprobe获取真实时长
		material.Duration = 180000000 // 默认3分钟
	}

	return material, nil
}

// NewAudioMaterialFromDict 从字典创建音频素材对象
func NewAudioMaterialFromDict(data map[string]interface{}) (*AudioMaterial, error) {
	material := &AudioMaterial{}

	// 基本字符串字段
	if id, ok := data["id"].(string); ok {
		material.MaterialID = id
	}

	if name, ok := data["name"].(string); ok {
		material.MaterialName = name
	}

	if path, ok := data["path"].(string); ok {
		material.Path = path
	}

	if remoteURL, ok := data["remote_url"].(string); ok && remoteURL != "" {
		material.RemoteURL = &remoteURL
	}

	// 数值字段
	if duration, ok := data["duration"]; ok {
		switch v := duration.(type) {
		case float64:
			material.Duration = int64(v)
		case int:
			material.Duration = int64(v)
		case int64:
			material.Duration = v
		}
	}

	material.HasAudioEffect = false // 默认没有音频效果

	return material, nil
}

// ExportJSON 导出为JSON格式
func (am *AudioMaterial) ExportJSON() map[string]interface{} {
	pathValue := am.Path
	if am.ReplacePath != nil {
		pathValue = *am.ReplacePath
	}

	var remoteURLValue interface{} = nil
	if am.RemoteURL != nil {
		remoteURLValue = *am.RemoteURL
	}

	checkFlag := 1
	if am.HasAudioEffect {
		checkFlag = 3
	}

	return map[string]interface{}{
		"app_id":                    0,
		"category_id":               "",
		"category_name":             "local",
		"check_flag":                checkFlag,
		"copyright_limit_type":      "none",
		"duration":                  am.Duration,
		"effect_id":                 "",
		"formula_id":                "",
		"id":                        am.MaterialID,
		"intensifies_path":          "",
		"is_ai_clone_tone":          false,
		"is_text_edit_overdub":      false,
		"is_ugc":                    false,
		"local_material_id":         am.MaterialID,
		"music_id":                  am.MaterialID,
		"name":                      am.MaterialName,
		"path":                      pathValue,
		"remote_url":                remoteURLValue,
		"query":                     "",
		"request_id":                "",
		"resource_id":               "",
		"search_id":                 "",
		"source_from":               "",
		"source_platform":           0,
		"team_id":                   "",
		"text_id":                   "",
		"tone_category_id":          "",
		"tone_category_name":        "",
		"tone_effect_id":            "",
		"tone_effect_name":          "",
		"tone_platform":             "",
		"tone_second_category_id":   "",
		"tone_second_category_name": "",
		"tone_speaker":              "",
		"tone_type":                 "",
		"type":                      string(MaterialTypeAudio),
		"video_id":                  "",
		"wave_points":               []interface{}{},
	}
}

// generateMaterialID 生成素材ID（基于UUID3）
func generateMaterialID(materialName string) string {
	// 使用DNS命名空间和素材名称生成UUID3
	namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8") // DNS命名空间
	id := uuid.NewMD5(namespace, []byte(materialName))
	return strings.ReplaceAll(id.String(), "-", "")
}

// MaterialInterface 素材接口，所有素材类型都应该实现
type MaterialInterface interface {
	ExportJSON() map[string]interface{}
	GetMaterialID() string
	GetMaterialName() string
	GetDuration() int64
}

// GetMaterialID 实现MaterialInterface接口
func (vm *VideoMaterial) GetMaterialID() string {
	return vm.MaterialID
}

// GetMaterialName 实现MaterialInterface接口
func (vm *VideoMaterial) GetMaterialName() string {
	return vm.MaterialName
}

// GetDuration 实现MaterialInterface接口
func (vm *VideoMaterial) GetDuration() int64 {
	return vm.Duration
}

// GetMaterialID 实现MaterialInterface接口
func (am *AudioMaterial) GetMaterialID() string {
	return am.MaterialID
}

// GetMaterialName 实现MaterialInterface接口
func (am *AudioMaterial) GetMaterialName() string {
	return am.MaterialName
}

// GetDuration 实现MaterialInterface接口
func (am *AudioMaterial) GetDuration() int64 {
	return am.Duration
}
