// Package template 定义与模板模式相关的类及函数
// 对应Python的 pyJianYingDraft/template_mode.py
package template

import (
	"fmt"
	"strings"

	"novel-video-workflow/pkg/capcut/internal/material"
	"novel-video-workflow/pkg/capcut/internal/segment"
	"novel-video-workflow/pkg/capcut/internal/track"
	"novel-video-workflow/pkg/capcut/internal/types"

	"github.com/google/uuid"
)

// ShrinkMode 处理替换素材时素材变短情况的方法
// 对应Python的Shrink_mode枚举
type ShrinkMode string

const (
	ShrinkModeCutHead      ShrinkMode = "cut_head"       // 裁剪头部，即后移片段起始点
	ShrinkModeCutTail      ShrinkMode = "cut_tail"       // 裁剪尾部，即前移片段终止点
	ShrinkModeCutTailAlign ShrinkMode = "cut_tail_align" // 裁剪尾部并消除间隙，即前移片段终止点，后续片段也依次前移
	ShrinkModeShrink       ShrinkMode = "shrink"         // 保持中间点不变，两端点向中间靠拢
)

// ExtendMode 处理替换素材时素材变长情况的方法
// 对应Python的Extend_mode枚举
type ExtendMode string

const (
	ExtendModeCutMaterialTail ExtendMode = "cut_material_tail" // 裁剪素材尾部，使得片段维持原长不变，此方法总是成功
	ExtendModeExtendHead      ExtendMode = "extend_head"       // 延伸头部，即尝试前移片段起始点，与前续片段重合时失败
	ExtendModeExtendTail      ExtendMode = "extend_tail"       // 延伸尾部，即尝试后移片段终止点，与后续片段重合时失败
	ExtendModePushTail        ExtendMode = "push_tail"         // 延伸尾部，若有必要则依次后移后续片段，此方法总是成功
)

// ImportedSegment 导入的片段
// 对应Python的ImportedSegment类
type ImportedSegment struct {
	*segment.BaseSegment
	RawData map[string]interface{} `json:"-"` // 原始json数据
}

// NewImportedSegment 创建导入的片段
func NewImportedSegment(jsonData map[string]interface{}) (*ImportedSegment, error) {
	// 提取基本属性
	materialID, ok := jsonData["material_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid material_id")
	}

	targetTimerangeData, ok := jsonData["target_timerange"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing or invalid target_timerange")
	}

	start, ok := targetTimerangeData["start"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid target_timerange.start")
	}

	duration, ok := targetTimerangeData["duration"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid target_timerange.duration")
	}

	targetTimerange := types.NewTimerange(int64(start), int64(duration))

	// 创建基础片段
	baseSegment := segment.NewBaseSegment(materialID, targetTimerange)

	// 复制原始数据
	rawData := make(map[string]interface{})
	for k, v := range jsonData {
		rawData[k] = v
	}

	return &ImportedSegment{
		BaseSegment: baseSegment,
		RawData:     rawData,
	}, nil
}

// ExportJSON 导出为JSON格式
// 对应Python的export_json方法
func (is *ImportedSegment) ExportJSON() map[string]interface{} {
	// 从原始数据开始
	jsonData := make(map[string]interface{})
	for k, v := range is.RawData {
		jsonData[k] = v
	}

	// 更新基本属性
	jsonData["material_id"] = is.MaterialID
	jsonData["target_timerange"] = map[string]interface{}{
		"start":    is.TargetTimerange.Start,
		"duration": is.TargetTimerange.Duration,
	}

	return jsonData
}

// ImportedMediaSegment 导入的视频/音频片段
// 对应Python的ImportedMediaSegment类
type ImportedMediaSegment struct {
	*ImportedSegment
	SourceTimerange *types.Timerange `json:"source_timerange"` // 片段取用的素材时间范围
}

// NewImportedMediaSegment 创建导入的媒体片段
func NewImportedMediaSegment(jsonData map[string]interface{}) (*ImportedMediaSegment, error) {
	// 先创建基础导入片段
	importedSegment, err := NewImportedSegment(jsonData)
	if err != nil {
		return nil, err
	}

	// 提取源时间范围
	sourceTimerangeData, ok := jsonData["source_timerange"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing or invalid source_timerange")
	}

	start, ok := sourceTimerangeData["start"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid source_timerange.start")
	}

	duration, ok := sourceTimerangeData["duration"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid source_timerange.duration")
	}

	sourceTimerange := types.NewTimerange(int64(start), int64(duration))

	return &ImportedMediaSegment{
		ImportedSegment: importedSegment,
		SourceTimerange: sourceTimerange,
	}, nil
}

// ExportJSON 导出为JSON格式
func (ims *ImportedMediaSegment) ExportJSON() map[string]interface{} {
	jsonData := ims.ImportedSegment.ExportJSON()

	// 添加源时间范围
	jsonData["source_timerange"] = map[string]interface{}{
		"start":    ims.SourceTimerange.Start,
		"duration": ims.SourceTimerange.Duration,
	}

	return jsonData
}

// ImportedTrack 模板模式下导入的轨道
// 对应Python的ImportedTrack类
type ImportedTrack struct {
	TrackType   track.TrackType        `json:"type"`         // 轨道类型
	Name        string                 `json:"name"`         // 轨道名称
	TrackID     string                 `json:"id"`           // 轨道ID
	RenderIndex int                    `json:"render_index"` // 渲染层级
	RawData     map[string]interface{} `json:"-"`            // 原始轨道数据
}

// NewImportedTrack 创建导入的轨道
func NewImportedTrack(jsonData map[string]interface{}) (*ImportedTrack, error) {
	// 提取轨道类型
	trackTypeName, ok := jsonData["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid track type")
	}

	trackType, err := track.TrackTypeFromName(trackTypeName)
	if err != nil {
		return nil, fmt.Errorf("invalid track type: %s", trackTypeName)
	}

	// 提取轨道名称
	name, ok := jsonData["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid track name")
	}

	// 提取轨道ID
	trackID, ok := jsonData["id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid track id")
	}

	// 计算渲染层级（从片段中获取最大值）
	renderIndex := 0
	if segments, ok := jsonData["segments"].([]interface{}); ok {
		for _, segData := range segments {
			if segMap, ok := segData.(map[string]interface{}); ok {
				if ri, ok := segMap["render_index"].(float64); ok {
					if int(ri) > renderIndex {
						renderIndex = int(ri)
					}
				}
			}
		}
	}

	// 复制原始数据
	rawData := make(map[string]interface{})
	for k, v := range jsonData {
		rawData[k] = v
	}

	return &ImportedTrack{
		TrackType:   trackType,
		Name:        name,
		TrackID:     trackID,
		RenderIndex: renderIndex,
		RawData:     rawData,
	}, nil
}

// ExportJSON 导出为JSON格式
// 对应Python的export_json方法
func (it *ImportedTrack) ExportJSON() map[string]interface{} {
	// 从原始数据开始
	jsonData := make(map[string]interface{})
	for k, v := range it.RawData {
		jsonData[k] = v
	}

	// 更新基本属性
	jsonData["name"] = it.Name
	jsonData["id"] = it.TrackID

	return jsonData
}

// EditableTrack 模板模式下导入且可修改的轨道(音视频及文本轨道)
// 对应Python的EditableTrack类
type EditableTrack struct {
	*ImportedTrack
	Segments []*ImportedSegment `json:"segments"` // 该轨道包含的片段列表
}

// NewEditableTrack 创建可编辑轨道
func NewEditableTrack(jsonData map[string]interface{}) (*EditableTrack, error) {
	// 先创建基础导入轨道
	importedTrack, err := NewImportedTrack(jsonData)
	if err != nil {
		return nil, err
	}

	return &EditableTrack{
		ImportedTrack: importedTrack,
		Segments:      make([]*ImportedSegment, 0),
	}, nil
}

// Len 返回轨道中片段的数量
func (et *EditableTrack) Len() int {
	return len(et.Segments)
}

// StartTime 返回轨道起始时间（微秒）
func (et *EditableTrack) StartTime() int64 {
	if len(et.Segments) == 0 {
		return 0
	}
	return et.Segments[0].TargetTimerange.Start
}

// EndTime 返回轨道结束时间（微秒）
func (et *EditableTrack) EndTime() int64 {
	if len(et.Segments) == 0 {
		return 0
	}
	lastSegment := et.Segments[len(et.Segments)-1]
	return lastSegment.TargetTimerange.Start + lastSegment.TargetTimerange.Duration
}

// ExportJSON 导出为JSON格式
func (et *EditableTrack) ExportJSON() map[string]interface{} {
	jsonData := et.ImportedTrack.ExportJSON()

	// 导出片段，为每个片段写入render_index
	segmentExports := make([]map[string]interface{}, len(et.Segments))
	for i, seg := range et.Segments {
		segmentExports[i] = seg.ExportJSON()
		segmentExports[i]["render_index"] = et.RenderIndex
	}
	jsonData["segments"] = segmentExports

	return jsonData
}

// ImportedTextTrack 模板模式下导入的文本轨道
// 对应Python的ImportedTextTrack类
type ImportedTextTrack struct {
	*EditableTrack
}

// NewImportedTextTrack 创建导入的文本轨道
func NewImportedTextTrack(jsonData map[string]interface{}) (*ImportedTextTrack, error) {
	// 先创建可编辑轨道
	editableTrack, err := NewEditableTrack(jsonData)
	if err != nil {
		return nil, err
	}

	// 创建文本片段
	if segments, ok := jsonData["segments"].([]interface{}); ok {
		for _, segData := range segments {
			if segMap, ok := segData.(map[string]interface{}); ok {
				segment, err := NewImportedSegment(segMap)
				if err != nil {
					return nil, fmt.Errorf("failed to create text segment: %v", err)
				}
				editableTrack.Segments = append(editableTrack.Segments, segment)
			}
		}
	}

	return &ImportedTextTrack{
		EditableTrack: editableTrack,
	}, nil
}

// ImportedMediaTrack 模板模式下导入的音频/视频轨道
// 对应Python的ImportedMediaTrack类
type ImportedMediaTrack struct {
	*EditableTrack
	MediaSegments []*ImportedMediaSegment `json:"-"` // 媒体片段列表（类型安全）
}

// NewImportedMediaTrack 创建导入的媒体轨道
func NewImportedMediaTrack(jsonData map[string]interface{}) (*ImportedMediaTrack, error) {
	// 先创建可编辑轨道
	editableTrack, err := NewEditableTrack(jsonData)
	if err != nil {
		return nil, err
	}

	mediaTrack := &ImportedMediaTrack{
		EditableTrack: editableTrack,
		MediaSegments: make([]*ImportedMediaSegment, 0),
	}

	// 创建媒体片段
	if segments, ok := jsonData["segments"].([]interface{}); ok {
		for _, segData := range segments {
			if segMap, ok := segData.(map[string]interface{}); ok {
				mediaSegment, err := NewImportedMediaSegment(segMap)
				if err != nil {
					return nil, fmt.Errorf("failed to create media segment: %v", err)
				}
				mediaTrack.MediaSegments = append(mediaTrack.MediaSegments, mediaSegment)
				mediaTrack.Segments = append(mediaTrack.Segments, mediaSegment.ImportedSegment)
			}
		}
	}

	return mediaTrack, nil
}

// CheckMaterialType 检查素材类型是否与轨道类型匹配
// 对应Python的check_material_type方法
func (imt *ImportedMediaTrack) CheckMaterialType(mat interface{}) bool {
	switch imt.TrackType {
	case track.TrackTypeVideo:
		_, ok := mat.(*material.VideoMaterial)
		return ok
	case track.TrackTypeAudio:
		_, ok := mat.(*material.AudioMaterial)
		return ok
	default:
		return false
	}
}

// ProcessTimerange 处理素材替换的时间范围变更
// 对应Python的process_timerange方法
func (imt *ImportedMediaTrack) ProcessTimerange(segIndex int, srcTimerange *types.Timerange, shrinkMode ShrinkMode, extendModes []ExtendMode) error {
	if segIndex < 0 || segIndex >= len(imt.MediaSegments) {
		return fmt.Errorf("segment index %d out of range", segIndex)
	}

	seg := imt.MediaSegments[segIndex]
	newDuration := srcTimerange.Duration
	oldDuration := seg.TargetTimerange.Duration

	// 计算时长差
	deltaDuration := newDuration - oldDuration
	if deltaDuration < 0 {
		deltaDuration = -deltaDuration
	}

	// 时长变短
	if newDuration < oldDuration {
		switch shrinkMode {
		case ShrinkModeCutHead:
			seg.TargetTimerange.Start += deltaDuration
		case ShrinkModeCutTail:
			seg.TargetTimerange.Duration -= deltaDuration
		case ShrinkModeCutTailAlign:
			seg.TargetTimerange.Duration -= deltaDuration
			// 后续片段也依次前移相应值（保持间隙）
			for i := segIndex + 1; i < len(imt.MediaSegments); i++ {
				imt.MediaSegments[i].TargetTimerange.Start -= deltaDuration
			}
		case ShrinkModeShrink:
			seg.TargetTimerange.Duration -= deltaDuration
			seg.TargetTimerange.Start += deltaDuration / 2
		default:
			return fmt.Errorf("unsupported shrink mode: %s", shrinkMode)
		}
	} else if newDuration > oldDuration {
		// 时长变长
		successFlag := false
		prevSegEnd := int64(0)
		if segIndex > 0 {
			prevSeg := imt.MediaSegments[segIndex-1]
			prevSegEnd = prevSeg.TargetTimerange.Start + prevSeg.TargetTimerange.Duration
		}

		nextSegStart := int64(1e15)
		if segIndex < len(imt.MediaSegments)-1 {
			nextSegStart = imt.MediaSegments[segIndex+1].TargetTimerange.Start
		}

		for _, mode := range extendModes {
			switch mode {
			case ExtendModeExtendHead:
				if seg.TargetTimerange.Start-deltaDuration >= prevSegEnd {
					seg.TargetTimerange.Start -= deltaDuration
					successFlag = true
				}
			case ExtendModeExtendTail:
				if seg.TargetTimerange.Start+seg.TargetTimerange.Duration+deltaDuration <= nextSegStart {
					seg.TargetTimerange.Duration += deltaDuration
					successFlag = true
				}
			case ExtendModePushTail:
				shiftDuration := int64(0)
				newEnd := seg.TargetTimerange.Start + seg.TargetTimerange.Duration + deltaDuration
				if newEnd > nextSegStart {
					shiftDuration = newEnd - nextSegStart
				}
				seg.TargetTimerange.Duration += deltaDuration
				if shiftDuration > 0 {
					// 有必要时后移后续片段
					for i := segIndex + 1; i < len(imt.MediaSegments); i++ {
						imt.MediaSegments[i].TargetTimerange.Start += shiftDuration
					}
				}
				successFlag = true
			case ExtendModeCutMaterialTail:
				srcTimerange.Duration = seg.TargetTimerange.Duration
				successFlag = true
			default:
				return fmt.Errorf("unsupported extend mode: %s", mode)
			}

			if successFlag {
				break
			}
		}

		if !successFlag {
			return fmt.Errorf("failed to extend segment to %d μs, tried methods: %v", newDuration, extendModes)
		}
	}

	// 写入素材时间范围
	seg.SourceTimerange = srcTimerange

	return nil
}

// ImportTrack 导入轨道
// 对应Python的import_track函数
func ImportTrack(jsonData map[string]interface{}, importedMaterials map[string]interface{}) (*track.Track, error) {
	trackTypeName, ok := jsonData["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid track type")
	}

	trackType, err := track.TrackTypeFromName(trackTypeName)
	if err != nil {
		return nil, fmt.Errorf("invalid track type: %s", trackTypeName)
	}

	name, ok := jsonData["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid track name")
	}

	// 计算渲染层级
	renderIndex := 0
	if segments, ok := jsonData["segments"].([]interface{}); ok {
		for _, segData := range segments {
			if segMap, ok := segData.(map[string]interface{}); ok {
				if ri, ok := segMap["render_index"].(float64); ok {
					if int(ri) > renderIndex {
						renderIndex = int(ri)
					}
				}
			}
		}
	}

	// 获取静音状态
	mute := false
	if attribute, ok := jsonData["attribute"].(float64); ok {
		mute = attribute != 0
	}

	// 创建新的Track实例
	newTrack := track.NewTrack(trackType, name, renderIndex, mute)

	// 设置track_id，使用原始ID
	if trackID, ok := jsonData["id"].(string); ok {
		newTrack.TrackID = trackID
	} else {
		newTrack.TrackID = strings.ReplaceAll(uuid.New().String(), "-", "")
	}

	// 如果轨道类型允许修改且有导入的素材，导入所有片段
	trackMeta := track.GetTrackMeta(trackType)
	if trackMeta.AllowModify && importedMaterials != nil {
		// 这里可以根据需要实现具体的片段导入逻辑
		// 由于涉及复杂的素材匹配和片段创建，暂时返回基本轨道
		// 在实际使用中，可以根据具体需求扩展这部分功能
	}

	return newTrack, nil
}
