// Package track 定义轨道类及其元数据
// 对应Python的 track.py
package track

import (
	"fmt"
	"reflect"

	"novel-video-workflow/pkg/capcut/internal/keyframe"
	"novel-video-workflow/pkg/capcut/internal/segment"

	"github.com/google/uuid"
)

// TrackMeta 与轨道类型关联的轨道元数据
// 对应Python的Track_meta类
type TrackMeta struct {
	SegmentType reflect.Type `json:"segment_type"` // 与轨道关联的片段类型
	RenderIndex int          `json:"render_index"` // 默认渲染顺序，值越大越接近前景
	AllowModify bool         `json:"allow_modify"` // 当被导入时，是否允许修改
}

// TrackType 轨道类型枚举
// 对应Python的Track_type枚举
type TrackType int

const (
	TrackTypeVideo TrackType = iota
	TrackTypeAudio
	TrackTypeEffect
	TrackTypeFilter
	TrackTypeSticker
	TrackTypeText
	TrackTypeAdjust // 仅供导入时使用，不要尝试新建此类型的轨道
)

// trackTypeMetas 轨道类型元数据映射
var trackTypeMetas = map[TrackType]TrackMeta{
	TrackTypeVideo:   {reflect.TypeOf(&segment.VideoSegment{}), 0, true},
	TrackTypeAudio:   {reflect.TypeOf(&segment.AudioSegment{}), 0, true},
	TrackTypeEffect:  {reflect.TypeOf(&segment.EffectSegment{}), 10000, false}, // 特效轨道
	TrackTypeFilter:  {reflect.TypeOf(&segment.FilterSegment{}), 11000, false}, // 滤镜轨道
	TrackTypeSticker: {nil, 14000, false},                                      // TODO: 待实现Sticker_segment后设置
	TrackTypeText:    {reflect.TypeOf(&segment.TextSegment{}), 15000, true},    // 原本是14000，避免与sticker冲突改为15000
	TrackTypeAdjust:  {nil, 0, false},
}

// String 返回轨道类型的字符串表示
func (tt TrackType) String() string {
	switch tt {
	case TrackTypeVideo:
		return "video"
	case TrackTypeAudio:
		return "audio"
	case TrackTypeEffect:
		return "effect"
	case TrackTypeFilter:
		return "filter"
	case TrackTypeSticker:
		return "sticker"
	case TrackTypeText:
		return "text"
	case TrackTypeAdjust:
		return "adjust"
	default:
		return "unknown"
	}
}

// TrackTypeFromName 根据名称获取轨道类型
// 对应Python的Track_type.from_name方法
func TrackTypeFromName(name string) (TrackType, error) {
	switch name {
	case "video":
		return TrackTypeVideo, nil
	case "audio":
		return TrackTypeAudio, nil
	case "effect":
		return TrackTypeEffect, nil
	case "filter":
		return TrackTypeFilter, nil
	case "sticker":
		return TrackTypeSticker, nil
	case "text":
		return TrackTypeText, nil
	case "adjust":
		return TrackTypeAdjust, nil
	default:
		return TrackTypeVideo, fmt.Errorf("unknown track type: %s", name)
	}
}

// Meta 返回轨道类型的元数据
func (tt TrackType) Meta() TrackMeta {
	if meta, exists := trackTypeMetas[tt]; exists {
		return meta
	}
	return TrackMeta{}
}

// GetTrackMeta 获取轨道类型的元数据
func GetTrackMeta(trackType TrackType) TrackMeta {
	return trackType.Meta()
}

// BaseTrack 轨道基类接口
// 对应Python的Base_track抽象基类
type BaseTrack interface {
	GetTrackType() TrackType
	GetName() string
	GetTrackID() string
	GetRenderIndex() int
	ExportJSON() map[string]interface{}
}

// PendingKeyframe 待处理的关键帧
type PendingKeyframe struct {
	PropertyType string  `json:"property_type"` // 关键帧属性类型
	Time         float64 `json:"time"`          // 关键帧时间点（秒）
	Value        string  `json:"value"`         // 关键帧值
}

// Track 非模板模式下的轨道
// 对应Python的Track[Seg_type]泛型类
type Track struct {
	TrackType        TrackType                  `json:"type"`              // 轨道类型
	Name             string                     `json:"name"`              // 轨道名称
	TrackID          string                     `json:"id"`                // 轨道全局ID
	RenderIndex      int                        `json:"render_index"`      // 渲染顺序，值越大越接近前景
	Mute             bool                       `json:"mute"`              // 是否静音
	Segments         []segment.SegmentInterface `json:"segments"`          // 该轨道包含的片段列表
	PendingKeyframes []PendingKeyframe          `json:"pending_keyframes"` // 待处理的关键帧列表
}

// NewTrack 创建新的轨道
func NewTrack(trackType TrackType, name string, renderIndex int, mute bool) *Track {
	// 如果未指定渲染索引，使用默认值
	if renderIndex == 0 {
		renderIndex = trackType.Meta().RenderIndex
	}

	return &Track{
		TrackType:        trackType,
		Name:             name,
		TrackID:          uuid.New().String(),
		RenderIndex:      renderIndex,
		Mute:             mute,
		Segments:         make([]segment.SegmentInterface, 0),
		PendingKeyframes: make([]PendingKeyframe, 0),
	}
}

// GetTrackType 实现BaseTrack接口
func (t *Track) GetTrackType() TrackType {
	return t.TrackType
}

// GetName 实现BaseTrack接口
func (t *Track) GetName() string {
	return t.Name
}

// GetTrackID 实现BaseTrack接口
func (t *Track) GetTrackID() string {
	return t.TrackID
}

// GetRenderIndex 实现BaseTrack接口
func (t *Track) GetRenderIndex() int {
	return t.RenderIndex
}

// AddPendingKeyframe 添加待处理的关键帧
func (t *Track) AddPendingKeyframe(propertyType string, time float64, value string) {
	kf := PendingKeyframe{
		PropertyType: propertyType,
		Time:         time,
		Value:        value,
	}
	t.PendingKeyframes = append(t.PendingKeyframes, kf)
}

// ProcessPendingKeyframes 处理所有待处理的关键帧
func (t *Track) ProcessPendingKeyframes() error {
	if len(t.PendingKeyframes) == 0 {
		return nil
	}

	// 遍历所有待处理的关键帧
	for _, kfInfo := range t.PendingKeyframes {
		propertyType := kfInfo.PropertyType
		time := kfInfo.Time
		value := kfInfo.Value

		// 将时间转换为微秒
		targetTime := int64(time * 1e6)

		// 找到时间点对应的片段
		var targetSegment segment.SegmentInterface
		for _, seg := range t.Segments {
			if seg.Start() <= targetTime && targetTime <= seg.Start()+seg.Duration() {
				targetSegment = seg
				break
			}
		}

		if targetSegment == nil {
			fmt.Printf("警告：在轨道 %s 的时间点 %.2fs 找不到对应的片段，跳过此关键帧\n", t.Name, time)
			continue
		}

		// 计算时间偏移量
		offsetTime := targetTime - targetSegment.Start()

		// 尝试将片段转换为BaseSegment类型以访问关键帧功能
		if baseSeg, ok := targetSegment.(*segment.BaseSegment); ok {
			err := baseSeg.AddKeyframeFromString(propertyType, offsetTime, value)
			if err != nil {
				fmt.Printf("添加关键帧失败: %v\n", err)
				continue
			}
		} else {
			// 尝试其他片段类型
			switch seg := targetSegment.(type) {
			case *segment.VideoSegment:
				if vs := seg.VisualSegment; vs != nil {
					// 解析值
					keyframeProp, err := keyframe.KeyframePropertyFromString(propertyType)
					if err != nil {
						fmt.Printf("不支持的属性类型: %v\n", err)
						continue
					}
					floatValue, err := keyframe.ParseValue(keyframeProp, value)
					if err != nil {
						fmt.Printf("解析值失败: %v\n", err)
						continue
					}
					err = vs.AddKeyframe(propertyType, offsetTime, floatValue)
					if err != nil {
						fmt.Printf("添加视频片段关键帧失败: %v\n", err)
					}
				}
			case *segment.TextSegment:
				if vs := seg.VisualSegment; vs != nil {
					// 解析值
					keyframeProp, err := keyframe.KeyframePropertyFromString(propertyType)
					if err != nil {
						fmt.Printf("不支持的属性类型: %v\n", err)
						continue
					}
					floatValue, err := keyframe.ParseValue(keyframeProp, value)
					if err != nil {
						fmt.Printf("解析值失败: %v\n", err)
						continue
					}
					err = vs.AddKeyframe(propertyType, offsetTime, floatValue)
					if err != nil {
						fmt.Printf("添加文本片段关键帧失败: %v\n", err)
					}
				}
			default:
				fmt.Printf("不支持的片段类型进行关键帧添加: %T\n", targetSegment)
			}
		}

		fmt.Printf("成功添加关键帧: %s 在 %.2fs\n", propertyType, time)
	}

	// 清空待处理的关键帧
	t.PendingKeyframes = t.PendingKeyframes[:0]

	return nil
}

// EndTime 轨道结束时间，微秒
func (t *Track) EndTime() int64 {
	if len(t.Segments) == 0 {
		return 0
	}

	var maxEndTime int64 = 0
	for _, seg := range t.Segments {
		endTime := seg.Start() + seg.Duration()
		if endTime > maxEndTime {
			maxEndTime = endTime
		}
	}

	return maxEndTime
}

// AcceptSegmentType 返回该轨道允许的片段类型
func (t *Track) AcceptSegmentType() reflect.Type {
	return t.TrackType.Meta().SegmentType
}

// AddSegment 向轨道中添加一个片段，添加的片段必须匹配轨道类型且不与现有片段重叠
func (t *Track) AddSegment(seg segment.SegmentInterface) error {
	// 检查片段类型是否匹配轨道类型
	acceptedType := t.AcceptSegmentType()
	if acceptedType != nil {
		segmentType := reflect.TypeOf(seg)
		if segmentType != acceptedType {
			return fmt.Errorf("新片段类型 (%s) 与轨道类型 (%s) 不匹配", segmentType, acceptedType)
		}
	}

	// 检查片段是否重叠
	for _, existingSeg := range t.Segments {
		if t.segmentsOverlap(existingSeg, seg) {
			return fmt.Errorf("新片段与现有片段重叠 [start: %d, end: %d]",
				seg.Start(), seg.Start()+seg.Duration())
		}
	}

	t.Segments = append(t.Segments, seg)
	return nil
}

// segmentsOverlap 检查两个片段是否重叠
func (t *Track) segmentsOverlap(seg1, seg2 segment.SegmentInterface) bool {
	start1, end1 := seg1.Start(), seg1.Start()+seg1.Duration()
	start2, end2 := seg2.Start(), seg2.Start()+seg2.Duration()

	// 如果一个片段的结束时间小于等于另一个片段的开始时间，则不重叠
	return !(end1 <= start2 || end2 <= start1)
}

// ExportJSON 导出轨道为JSON格式
func (t *Track) ExportJSON() map[string]interface{} {
	// 导出所有片段的JSON，并为每个片段设置render_index
	segmentExports := make([]interface{}, len(t.Segments))
	for i, seg := range t.Segments {
		segmentJSON := seg.ExportJSON()
		segmentJSON["render_index"] = t.RenderIndex
		segmentExports[i] = segmentJSON
	}

	return map[string]interface{}{
		"attribute":       t.getMuteAttribute(),
		"flag":            0,
		"id":              t.TrackID,
		"is_default_name": len(t.Name) == 0,
		"name":            t.Name,
		"render_index":    t.RenderIndex,
		"segments":        segmentExports,
		"type":            t.TrackType.String(),
	}
}

// getMuteAttribute 获取静音属性值
func (t *Track) getMuteAttribute() int {
	if t.Mute {
		return 1
	}
	return 0
}

// String 返回轨道的字符串表示
func (t *Track) String() string {
	return fmt.Sprintf("Track{Type: %s, Name: %s, ID: %s, Segments: %d}",
		t.TrackType.String(), t.Name, t.TrackID, len(t.Segments))
}
