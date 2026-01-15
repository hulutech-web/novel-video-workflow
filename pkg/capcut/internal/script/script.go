// Package script 定义剪映草稿文件系统
// 对应Python的 pyJianYingDraft/script_file.py
package script

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"

	"novel-video-workflow/pkg/capcut/internal/animation"
	"novel-video-workflow/pkg/capcut/internal/material"
	"novel-video-workflow/pkg/capcut/internal/segment"
	"novel-video-workflow/pkg/capcut/internal/template"
	"novel-video-workflow/pkg/capcut/internal/track"
	"novel-video-workflow/pkg/capcut/internal/types"
)

// ScriptMaterial 草稿文件中的素材信息部分
// 对应Python的Script_material类
type ScriptMaterial struct {
	// 基础素材
	Audios   []*material.AudioMaterial `json:"audios"`   // 音频素材列表
	Videos   []*material.VideoMaterial `json:"videos"`   // 视频素材列表
	Stickers []map[string]interface{}  `json:"stickers"` // 贴纸素材列表
	Texts    []map[string]interface{}  `json:"texts"`    // 文本素材列表

	// 效果素材
	AudioEffects []*segment.AudioEffect         `json:"audio_effects"` // 音频特效列表
	AudioFades   []*segment.AudioFade           `json:"audio_fades"`   // 音频淡入淡出效果列表
	Animations   []*animation.SegmentAnimations `json:"animations"`    // 动画素材列表
	VideoEffects []*segment.VideoEffect         `json:"video_effects"` // 视频特效列表

	// 其他素材
	Speeds      []*segment.Speed             `json:"speeds"`      // 变速列表
	Masks       []map[string]interface{}     `json:"masks"`       // 蒙版列表
	Transitions []*segment.Transition        `json:"transitions"` // 转场效果列表
	Filters     []interface{}                `json:"filters"`     // 滤镜/文本花字/文本气泡列表
	Canvases    []*segment.BackgroundFilling `json:"canvases"`    // 背景填充列表
}

// NewScriptMaterial 创建新的草稿素材管理器
func NewScriptMaterial() *ScriptMaterial {
	return &ScriptMaterial{
		Audios:       make([]*material.AudioMaterial, 0),
		Videos:       make([]*material.VideoMaterial, 0),
		Stickers:     make([]map[string]interface{}, 0),
		Texts:        make([]map[string]interface{}, 0),
		AudioEffects: make([]*segment.AudioEffect, 0),
		AudioFades:   make([]*segment.AudioFade, 0),
		Animations:   make([]*animation.SegmentAnimations, 0),
		VideoEffects: make([]*segment.VideoEffect, 0),
		Speeds:       make([]*segment.Speed, 0),
		Masks:        make([]map[string]interface{}, 0),
		Transitions:  make([]*segment.Transition, 0),
		Filters:      make([]interface{}, 0),
		Canvases:     make([]*segment.BackgroundFilling, 0),
	}
}

// Contains 检查素材是否已存在
// 对应Python的__contains__方法
func (sm *ScriptMaterial) Contains(item interface{}) bool {
	switch v := item.(type) {
	case *material.VideoMaterial:
		for _, video := range sm.Videos {
			if video.MaterialID == v.MaterialID {
				return true
			}
		}
	case *material.AudioMaterial:
		for _, audio := range sm.Audios {
			if audio.MaterialID == v.MaterialID {
				return true
			}
		}
	case *segment.AudioFade:
		for _, fade := range sm.AudioFades {
			if fade.FadeID == v.FadeID {
				return true
			}
		}
	case *segment.AudioEffect:
		for _, effect := range sm.AudioEffects {
			if effect.EffectID == v.EffectID {
				return true
			}
		}
	case *animation.SegmentAnimations:
		for _, ani := range sm.Animations {
			if ani.AnimationID == v.AnimationID {
				return true
			}
		}
	case *segment.VideoEffect:
		for _, effect := range sm.VideoEffects {
			if effect.GlobalID == v.GlobalID {
				return true
			}
		}
	case *segment.Transition:
		for _, transition := range sm.Transitions {
			if transition.GlobalID == v.GlobalID {
				return true
			}
		}
	case *segment.Filter:
		for _, filter := range sm.Filters {
			if f, ok := filter.(*segment.Filter); ok && f.GlobalID == v.GlobalID {
				return true
			}
		}
	default:
		return false
	}
	return false
}

// generateUUID 生成UUID字符串
func generateUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// ExportJSON 导出素材信息为JSON格式
// 对应Python的export_json方法
func (sm *ScriptMaterial) ExportJSON() map[string]interface{} {
	// 导出音频素材
	audios := make([]map[string]interface{}, len(sm.Audios))
	for i, audio := range sm.Audios {
		audios[i] = audio.ExportJSON()
	}

	// 导出视频素材
	videos := make([]map[string]interface{}, len(sm.Videos))
	for i, video := range sm.Videos {
		videos[i] = video.ExportJSON()
	}

	// 导出音频特效
	audioEffects := make([]map[string]interface{}, len(sm.AudioEffects))
	for i, effect := range sm.AudioEffects {
		audioEffects[i] = effect.ExportJSON()
	}

	// 导出音频淡入淡出
	audioFades := make([]map[string]interface{}, len(sm.AudioFades))
	for i, fade := range sm.AudioFades {
		audioFades[i] = fade.ExportJSON()
	}

	// 导出动画
	animations := make([]map[string]interface{}, len(sm.Animations))
	for i, ani := range sm.Animations {
		animations[i] = ani.ExportJSON()
	}

	// 导出视频特效
	videoEffects := make([]map[string]interface{}, len(sm.VideoEffects))
	for i, effect := range sm.VideoEffects {
		videoEffects[i] = effect.ExportJSON()
	}

	// 导出变速
	speeds := make([]map[string]interface{}, len(sm.Speeds))
	for i, speed := range sm.Speeds {
		speeds[i] = speed.ExportJSON()
	}

	// 导出转场
	transitions := make([]map[string]interface{}, len(sm.Transitions))
	for i, transition := range sm.Transitions {
		transitions[i] = transition.ExportJSON()
	}

	// 导出滤镜（接口类型，需要类型断言）
	filters := make([]map[string]interface{}, len(sm.Filters))
	for i, filter := range sm.Filters {
		if f, ok := filter.(interface{ ExportJSON() map[string]interface{} }); ok {
			filters[i] = f.ExportJSON()
		}
	}

	// 导出背景填充
	canvases := make([]map[string]interface{}, len(sm.Canvases))
	for i, canvas := range sm.Canvases {
		canvases[i] = canvas.ExportJSON()
	}

	result := map[string]interface{}{
		"ai_translates":          []interface{}{},
		"audio_balances":         []interface{}{},
		"audio_effects":          audioEffects,
		"audio_fades":            audioFades,
		"audio_track_indexes":    []interface{}{},
		"audios":                 audios,
		"beats":                  []interface{}{},
		"canvases":               canvases,
		"chromas":                []interface{}{},
		"color_curves":           []interface{}{},
		"digital_humans":         []interface{}{},
		"drafts":                 []interface{}{},
		"effects":                filters,
		"flowers":                []interface{}{},
		"green_screens":          []interface{}{},
		"handwrites":             []interface{}{},
		"hsl":                    []interface{}{},
		"images":                 []interface{}{},
		"log_color_wheels":       []interface{}{},
		"loudnesses":             []interface{}{},
		"manual_deformations":    []interface{}{},
		"material_animations":    animations,
		"material_colors":        []interface{}{},
		"multi_language_refs":    []interface{}{},
		"placeholders":           []interface{}{},
		"plugin_effects":         []interface{}{},
		"primary_color_wheels":   []interface{}{},
		"realtime_denoises":      []interface{}{},
		"shapes":                 []interface{}{},
		"smart_crops":            []interface{}{},
		"smart_relights":         []interface{}{},
		"sound_channel_mappings": []interface{}{},
		"speeds":                 speeds,
		"stickers":               sm.Stickers,
		"tail_leaders":           []interface{}{},
		"text_templates":         []interface{}{},
		"texts":                  sm.Texts,
		"time_marks":             []interface{}{},
		"transitions":            transitions,
		"video_effects":          videoEffects,
		"video_trackings":        []interface{}{},
		"videos":                 videos,
		"vocal_beautifys":        []interface{}{},
		"vocal_separations":      []interface{}{},
	}

	// 根据环境决定使用common_mask还是masks
	// TODO: 需要添加环境检测，暂时使用masks
	result["masks"] = sm.Masks

	return result
}

// ScriptFile 剪映草稿文件，大部分接口定义在此
// 对应Python的Script_file类
type ScriptFile struct {
	SavePath *string                `json:"save_path,omitempty"` // 草稿文件保存路径，仅在模板模式下有效
	Content  map[string]interface{} `json:"content"`             // 草稿文件内容

	Width    int   `json:"width"`    // 视频的宽度，单位为像素
	Height   int   `json:"height"`   // 视频的高度，单位为像素
	FPS      int   `json:"fps"`      // 视频的帧率
	Duration int64 `json:"duration"` // 视频的总时长，单位为微秒

	Materials *ScriptMaterial         `json:"materials"` // 草稿文件中的素材信息部分
	Tracks    map[string]*track.Track `json:"tracks"`    // 轨道信息

	ImportedMaterials map[string][]map[string]interface{} `json:"imported_materials"` // 导入的素材信息
	ImportedTracks    []*track.Track                      `json:"imported_tracks"`    // 导入的轨道信息
}

const TemplateFile = "draft_content_template.json"

// NewScriptFile 创建一个剪映草稿
func NewScriptFile(width, height int, fps ...int) (*ScriptFile, error) {
	frameRate := 30
	if len(fps) > 0 {
		frameRate = fps[0]
	}

	sf := &ScriptFile{
		SavePath:          nil,
		Width:             width,
		Height:            height,
		FPS:               frameRate,
		Duration:          0,
		Materials:         NewScriptMaterial(),
		Tracks:            make(map[string]*track.Track),
		ImportedMaterials: make(map[string][]map[string]interface{}),
		ImportedTracks:    make([]*track.Track, 0),
	}

	// 加载模板文件
	templatePath := filepath.Join("internal", "script", TemplateFile)
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// 如果模板文件不存在，创建默认内容
		sf.Content = sf.createDefaultContent()
	} else {
		file, err := os.Open(templatePath)
		if err != nil {
			return nil, fmt.Errorf("无法打开模板文件: %v", err)
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&sf.Content); err != nil {
			return nil, fmt.Errorf("无法解析模板文件: %v", err)
		}
	}

	return sf, nil
}

// createDefaultContent 创建默认的草稿内容
func (sf *ScriptFile) createDefaultContent() map[string]interface{} {
	return map[string]interface{}{
		"version":   "1.0.0",
		"materials": map[string]interface{}{},
		"tracks":    []interface{}{},
		"fps":       sf.FPS,
		"duration":  sf.Duration,
		"canvas_config": map[string]interface{}{
			"width":  sf.Width,
			"height": sf.Height,
			"ratio":  "original",
		},
	}
}

// LoadTemplate 从JSON文件加载草稿模板
// 对应Python的load_template静态方法
func LoadTemplate(jsonPath string) (*ScriptFile, error) {
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("JSON文件 '%s' 不存在", jsonPath)
	}

	file, err := os.Open(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("无法打开JSON文件: %v", err)
	}
	defer file.Close()

	var content map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&content); err != nil {
		return nil, fmt.Errorf("无法解析JSON文件: %v", err)
	}

	sf := &ScriptFile{
		SavePath:          &jsonPath,
		Content:           content,
		Materials:         NewScriptMaterial(),
		Tracks:            make(map[string]*track.Track),
		ImportedMaterials: make(map[string][]map[string]interface{}),
		ImportedTracks:    make([]*track.Track, 0),
	}

	// 提取基本属性
	if fps, ok := content["fps"].(float64); ok {
		sf.FPS = int(fps)
	}
	if duration, ok := content["duration"].(float64); ok {
		sf.Duration = int64(duration)
	}

	// 提取画布配置
	if canvasConfig, ok := content["canvas_config"].(map[string]interface{}); ok {
		if width, ok := canvasConfig["width"].(float64); ok {
			sf.Width = int(width)
		}
		if height, ok := canvasConfig["height"].(float64); ok {
			sf.Height = int(height)
		}
	}

	// 导入素材
	if materials, ok := content["materials"].(map[string]interface{}); ok {
		for key, value := range materials {
			if materialList, ok := value.([]interface{}); ok {
				materialMaps := make([]map[string]interface{}, len(materialList))
				for i, item := range materialList {
					if itemMap, ok := item.(map[string]interface{}); ok {
						materialMaps[i] = itemMap
					}
				}
				sf.ImportedMaterials[key] = materialMaps
			}
		}
	}

	// 导入轨道
	if tracks, ok := content["tracks"].([]interface{}); ok {
		for _, trackData := range tracks {
			if trackMap, ok := trackData.(map[string]interface{}); ok {
				// 转换 ImportedMaterials 为正确的类型
				materialsMap := make(map[string]interface{})
				for key, value := range sf.ImportedMaterials {
					interfaceSlice := make([]interface{}, len(value))
					for i, item := range value {
						interfaceSlice[i] = item
					}
					materialsMap[key] = interfaceSlice
				}

				importedTrack, err := template.ImportTrack(trackMap, materialsMap)
				if err != nil {
					return nil, fmt.Errorf("导入轨道失败: %v", err)
				}
				sf.ImportedTracks = append(sf.ImportedTracks, importedTrack)
			}
		}
	}

	return sf, nil
}

// AddMaterial 向草稿文件中添加一个素材
// 对应Python的add_material方法
func (sf *ScriptFile) AddMaterial(mat interface{}) *ScriptFile {
	if sf.Materials.Contains(mat) {
		return sf // 素材已存在
	}

	switch material := mat.(type) {
	case *material.VideoMaterial:
		sf.Materials.Videos = append(sf.Materials.Videos, material)
	case *material.AudioMaterial:
		sf.Materials.Audios = append(sf.Materials.Audios, material)
	default:
		// TODO: 可以添加日志记录不支持的素材类型
	}

	return sf
}

// AddTrack 向草稿文件中添加一个指定类型、指定名称的轨道
// 对应Python的add_track方法
func (sf *ScriptFile) AddTrack(trackType track.TrackType, trackName *string, options ...TrackOption) *ScriptFile {
	config := &TrackConfig{
		Mute:          false,
		RelativeIndex: 0,
		AbsoluteIndex: nil,
	}

	// 应用选项
	for _, option := range options {
		option(config)
	}

	// 处理轨道名称
	finalTrackName := ""
	if trackName != nil {
		finalTrackName = *trackName
	} else {
		// 检查是否已存在同类型轨道
		for _, existingTrack := range sf.Tracks {
			if existingTrack.TrackType == trackType {
				// 已存在同类型轨道且未指定名称，返回错误或使用默认处理
				return sf
			}
		}
		finalTrackName = trackType.String()
	}

	// 检查是否已存在同名轨道
	if _, exists := sf.Tracks[finalTrackName]; exists {
		return sf
	}

	// 计算渲染层级
	renderIndex := track.GetTrackMeta(trackType).RenderIndex + config.RelativeIndex
	if config.AbsoluteIndex != nil {
		renderIndex = *config.AbsoluteIndex
	}

	// 创建轨道
	newTrack := track.NewTrack(trackType, finalTrackName, renderIndex, config.Mute)
	sf.Tracks[finalTrackName] = newTrack

	return sf
}

// TrackConfig 轨道配置
type TrackConfig struct {
	Mute          bool
	RelativeIndex int
	AbsoluteIndex *int
}

// TrackOption 轨道选项函数类型
type TrackOption func(*TrackConfig)

// WithMute 设置轨道静音
func WithMute(mute bool) TrackOption {
	return func(c *TrackConfig) {
		c.Mute = mute
	}
}

// WithRelativeIndex 设置相对图层位置
func WithRelativeIndex(index int) TrackOption {
	return func(c *TrackConfig) {
		c.RelativeIndex = index
	}
}

// WithAbsoluteIndex 设置绝对图层位置
func WithAbsoluteIndex(index int) TrackOption {
	return func(c *TrackConfig) {
		c.AbsoluteIndex = &index
	}
}

// GetTrack 获取指定类型的轨道
// 对应Python的get_track方法
func (sf *ScriptFile) GetTrack(segmentType string, trackName *string) (*track.Track, error) {
	// 指定轨道名称
	if trackName != nil {
		if targetTrack, exists := sf.Tracks[*trackName]; exists {
			return targetTrack, nil
		}
		return nil, fmt.Errorf("不存在名为 '%s' 的轨道", *trackName)
	}

	// 寻找唯一的同类型的轨道
	var matchingTracks []*track.Track
	for _, t := range sf.Tracks {
		// TODO: 需要根据segmentType匹配轨道类型，这里暂时简化处理
		matchingTracks = append(matchingTracks, t)
	}

	if len(matchingTracks) == 0 {
		return nil, fmt.Errorf("不存在接受 '%s' 的轨道", segmentType)
	}
	if len(matchingTracks) > 1 {
		return nil, fmt.Errorf("存在多个接受 '%s' 的轨道，请指定轨道名称", segmentType)
	}

	return matchingTracks[0], nil
}

// GetTrackAndImportedTrack 获取指定类型的所有轨道（包括普通轨道和导入的轨道）
// 对应Python的_get_track_and_imported_track方法
func (sf *ScriptFile) GetTrackAndImportedTrack(segmentType string, trackName *string) ([]*track.Track, error) {
	var resultTracks []*track.Track

	// 如果指定了轨道名称
	if trackName != nil {
		// 在普通轨道中查找
		if t, exists := sf.Tracks[*trackName]; exists {
			resultTracks = append(resultTracks, t)
		}
		// 在导入的轨道中查找
		for _, t := range sf.ImportedTracks {
			if t.Name == *trackName {
				resultTracks = append(resultTracks, t)
			}
		}
		if len(resultTracks) == 0 {
			return nil, fmt.Errorf("不存在名为 '%s' 的轨道", *trackName)
		}
	} else {
		// TODO: 根据segmentType查找匹配的轨道类型
		// 这里需要更复杂的类型匹配逻辑
		for _, t := range sf.Tracks {
			resultTracks = append(resultTracks, t)
		}
		for _, t := range sf.ImportedTracks {
			resultTracks = append(resultTracks, t)
		}

		if len(resultTracks) == 0 {
			return nil, fmt.Errorf("不存在接受 '%s' 的轨道", segmentType)
		}
		if len(resultTracks) > 1 {
			return nil, fmt.Errorf("存在多个接受 '%s' 的轨道，请指定轨道名称", segmentType)
		}
	}

	return resultTracks, nil
}

// AddSegment 向指定轨道中添加一个片段
// 对应Python的add_segment方法
func (sf *ScriptFile) AddSegment(seg interface{}, trackName *string) error {
	// TODO: 根据片段类型找到合适的轨道
	// 这里需要实现复杂的片段类型匹配逻辑

	// 简化实现：假设我们有一个通用的片段接口
	if segment, ok := seg.(interface{ GetTargetTimerange() *types.Timerange }); ok {
		timerange := segment.GetTargetTimerange()
		if timerange != nil {
			endTime := timerange.Start + timerange.Duration
			if endTime > sf.Duration {
				sf.Duration = endTime
			}
		}
	}

	// TODO: 实现具体的片段添加逻辑
	// 需要根据片段类型添加相关素材到materials中

	return nil
}

// Dumps 将草稿文件内容导出为JSON字符串
// 对应Python的dumps方法
func (sf *ScriptFile) Dumps() (string, error) {
	// 更新基本信息
	sf.Content["fps"] = float64(sf.FPS)
	sf.Content["duration"] = sf.Duration
	sf.Content["canvas_config"] = map[string]interface{}{
		"width":  sf.Width,
		"height": sf.Height,
		"ratio":  "original",
	}
	sf.Content["materials"] = sf.Materials.ExportJSON()

	// 设置ID和名称
	if _, exists := sf.Content["id"].(*string); !exists {
		sf.Content["id"] = generateUUID() // 使用辅助函数生成UUID
	}
	if _, exists := sf.Content["name"].(*string); !exists {
		sf.Content["name"] = ""
	}
	if _, exists := sf.Content["cover"].(*string); !exists {
		sf.Content["cover"] = nil
	}

	// 添加额外的关键字段
	if _, exists := sf.Content["color_space"].(*int); !exists {
		sf.Content["color_space"] = 0
	}
	if _, exists := sf.Content["create_time"].(*int64); !exists {
		sf.Content["create_time"] = int64(0)
	}
	if _, exists := sf.Content["extra_info"].(*string); !exists {
		sf.Content["extra_info"] = nil
	}
	if _, exists := sf.Content["group_container"].(*string); !exists {
		sf.Content["group_container"] = nil
	}
	if _, exists := sf.Content["mutable_config"].(*string); !exists {
		sf.Content["mutable_config"] = nil
	}
	if _, exists := sf.Content["retouch_cover"].(*string); !exists {
		sf.Content["retouch_cover"] = nil
	}
	if _, exists := sf.Content["source"].(*string); !exists {
		sf.Content["source"] = "default"
	}
	if _, exists := sf.Content["static_cover_image_path"].(*string); !exists {
		sf.Content["static_cover_image_path"] = ""
	}
	if _, exists := sf.Content["update_time"].(*int64); !exists {
		sf.Content["update_time"] = int64(0)
	}
	if _, exists := sf.Content["version"].(*string); !exists {
		sf.Content["version"] = "360000"
	}
	if _, exists := sf.Content["new_version"].(*string); !exists {
		sf.Content["new_version"] = "57.0.0"
	}
	if _, exists := sf.Content["relationships"].(*[]interface{}); !exists {
		relationshipList := make([]interface{}, 0)
		sf.Content["relationships"] = relationshipList
	}
	if _, exists := sf.Content["render_index_track_mode_on"].(*bool); !exists {
		sf.Content["render_index_track_mode_on"] = true
	}

	// 添加config配置
	if _, exists := sf.Content["config"].(*map[string]interface{}); !exists {
		config := map[string]interface{}{
			"adjust_max_index":              1,
			"attachment_info":               []interface{}{},
			"combination_max_index":         1,
			"export_range":                  nil,
			"extract_audio_last_index":      1,
			"lyrics_recognition_id":         "",
			"lyrics_sync":                   true,
			"lyrics_taskinfo":               []interface{}{},
			"maintrack_adsorb":              true,
			"material_save_mode":            0,
			"original_sound_last_index":     1,
			"record_audio_last_index":       1,
			"sticker_max_index":             1,
			"subtitle_recognition_id":       "",
			"subtitle_sync":                 true,
			"subtitle_taskinfo":             []interface{}{},
			"video_mute":                    false,
			"zoom_info_params":              nil,
		}
		sf.Content["config"] = config
	}

	// 添加keyframes配置
	if _, exists := sf.Content["keyframes"].(*map[string]interface{}); !exists {
		keyframes := map[string]interface{}{
			"adjusts":   []interface{}{},
			"audios":    []interface{}{},
			"filters":   []interface{}{},
			"handwrites": []interface{}{},
			"stickers":  []interface{}{},
			"texts":     []interface{}{},
			"videos":    []interface{}{},
		}
		sf.Content["keyframes"] = keyframes
	}

	// 设置平台信息
	platformInfo := map[string]interface{}{
		"app_id":       3704,  // 注意：手工项目中app_id是3704而不是359289
		"app_version":  "3.4.1",  // 注意：手工项目中版本号是3.4.1而不是6.5.0
		"os":           "mac",
		"os_version":   "26.2",
	}
	sf.Content["last_modified_platform"] = platformInfo
	sf.Content["platform"] = platformInfo

	// 合并导入的素材
	if materials, ok := sf.Content["materials"].(map[string]interface{}); ok {
		for materialType, materialList := range sf.ImportedMaterials {
			if existingList, exists := materials[materialType]; exists {
				if existingSlice, ok := existingList.([]interface{}); ok {
					for _, item := range materialList {
						existingSlice = append(existingSlice, item)
					}
					materials[materialType] = existingSlice
				}
			} else {
				interfaceList := make([]interface{}, len(materialList))
				for i, item := range materialList {
					interfaceList[i] = item
				}
				materials[materialType] = interfaceList
			}
		}
	}

	// 对轨道排序并导出
	var trackList []*track.Track
	for _, t := range sf.Tracks {
		trackList = append(trackList, t)
	}
	trackList = append(trackList, sf.ImportedTracks...)

	// 按渲染层级排序
	sort.Slice(trackList, func(i, j int) bool {
		return trackList[i].RenderIndex < trackList[j].RenderIndex
	})

	// 导出轨道
	tracks := make([]map[string]interface{}, len(trackList))
	for i, t := range trackList {
		tracks[i] = t.ExportJSON()
	}
	sf.Content["tracks"] = tracks

	// 序列化为JSON
	jsonBytes, err := json.MarshalIndent(sf.Content, "", "    ")
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}

	return string(jsonBytes), nil
}

// Dump 将草稿文件内容写入文件
// 对应Python的dump方法
func (sf *ScriptFile) Dump(filePath string) error {
	jsonStr, err := sf.Dumps()
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("无法创建文件: %v", err)
	}
	defer file.Close()

	_, err = io.WriteString(file, jsonStr)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// Save 保存草稿文件至打开时的路径，仅在模板模式下可用
// 对应Python的save方法
func (sf *ScriptFile) Save() error {
	if sf.SavePath == nil {
		return fmt.Errorf("没有设置保存路径，可能不在模板模式下")
	}
	return sf.Dump(*sf.SavePath)
}

// InspectMaterial 输出草稿中导入的贴纸、文本气泡以及花字素材的元数据
// 对应Python的inspect_material方法
func (sf *ScriptFile) InspectMaterial() {
	fmt.Println("贴纸素材:")
	if stickers, ok := sf.ImportedMaterials["stickers"]; ok {
		for _, sticker := range stickers {
			resourceID := ""
			name := ""
			if rid, ok := sticker["resource_id"].(string); ok {
				resourceID = rid
			}
			if n, ok := sticker["name"].(string); ok {
				name = n
			}
			fmt.Printf("\tResource id: %s '%s'\n", resourceID, name)
		}
	}

	fmt.Println("文字气泡效果:")
	if effects, ok := sf.ImportedMaterials["effects"]; ok {
		for _, effect := range effects {
			if effectType, ok := effect["type"].(string); ok && effectType == "text_shape" {
				effectID := ""
				resourceID := ""
				name := ""
				if eid, ok := effect["effect_id"].(string); ok {
					effectID = eid
				}
				if rid, ok := effect["resource_id"].(string); ok {
					resourceID = rid
				}
				if n, ok := effect["name"].(string); ok {
					name = n
				}
				fmt.Printf("\tEffect id: %s ,Resource id: %s '%s'\n", effectID, resourceID, name)
			}
		}
	}

	fmt.Println("花字效果:")
	if effects, ok := sf.ImportedMaterials["effects"]; ok {
		for _, effect := range effects {
			if effectType, ok := effect["type"].(string); ok && effectType == "text_effect" {
				resourceID := ""
				name := ""
				if rid, ok := effect["resource_id"].(string); ok {
					resourceID = rid
				}
				if n, ok := effect["name"].(string); ok {
					name = n
				}
				fmt.Printf("\tResource id: %s '%s'\n", resourceID, name)
			}
		}
	}
}
