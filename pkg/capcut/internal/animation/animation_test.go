package animation

import (
	"encoding/json"
	"testing"

	"novel-video-workflow/pkg/capcut/internal/metadata"
)

// TestNewVideoAnimation 测试创建视频动画
func TestNewVideoAnimation(t *testing.T) {
	// 测试入场动画
	introType, err := metadata.FindIntroByName("渐显")
	if err != nil {
		t.Fatalf("查找入场动画失败: %v", err)
	}
	introAnim, err := NewVideoAnimation(introType, 0, 500000)

	if err != nil {
		t.Fatalf("创建入场动画失败: %v", err)
	}

	if introAnim.Animation.Name != "渐显" {
		t.Errorf("期望动画名称为 '渐显', 得到 '%s'", introAnim.Animation.Name)
	}

	if introAnim.Animation.AnimationType != AnimationTypeIn {
		t.Errorf("期望动画类型为 'in', 得到 '%s'", introAnim.Animation.AnimationType)
	}

	if !introAnim.Animation.IsVideoAnimation {
		t.Error("期望 IsVideoAnimation 为 true")
	}

	// 测试出场动画
	outroAnim, err := NewVideoAnimation(metadata.OutroType缩小, 1000000, 500000)
	if err != nil {
		t.Fatalf("创建出场动画失败: %v", err)
	}

	if outroAnim.Animation.AnimationType != AnimationTypeOut {
		t.Errorf("期望动画类型为 'out', 得到 '%s'", outroAnim.Animation.AnimationType)
	}

	// 测试组合动画
	groupAnim, err := NewVideoAnimation(metadata.GroupAnimationType三分割, 0, 2000000)
	if err != nil {
		t.Fatalf("创建组合动画失败: %v", err)
	}

	if groupAnim.Animation.AnimationType != AnimationTypeGroup {
		t.Errorf("期望动画类型为 'group', 得到 '%s'", groupAnim.Animation.AnimationType)
	}
}

// TestNewTextAnimation 测试创建文本动画
func TestNewTextAnimation(t *testing.T) {
	// 测试文本入场动画
	textIntroAnim, err := NewTextAnimation(metadata.TextIntro打字机, 0, 500000)
	if err != nil {
		t.Fatalf("创建文本入场动画失败: %v", err)
	}

	if textIntroAnim.Animation.Name != "打字机" {
		t.Errorf("期望动画名称为 '打字机', 得到 '%s'", textIntroAnim.Animation.Name)
	}

	if textIntroAnim.Animation.AnimationType != AnimationTypeIn {
		t.Errorf("期望动画类型为 'in', 得到 '%s'", textIntroAnim.Animation.AnimationType)
	}

	if textIntroAnim.Animation.IsVideoAnimation {
		t.Error("期望 IsVideoAnimation 为 false")
	}

	// 测试文本循环动画
	textLoopAnim, err := NewTextAnimation(metadata.TextLoopAnimType跳动, 0, 0)
	if err != nil {
		t.Fatalf("创建文本循环动画失败: %v", err)
	}

	if textLoopAnim.Animation.AnimationType != AnimationTypeLoop {
		t.Errorf("期望动画类型为 'loop', 得到 '%s'", textLoopAnim.Animation.AnimationType)
	}
}

// TestCapCutAnimations 测试CapCut动画类型
func TestCapCutAnimations(t *testing.T) {
	// 测试CapCut入场动画
	capCutIntroType, err := metadata.FindCapCutIntroByName("1998")
	if err != nil {
		t.Fatalf("查找CapCut入场动画失败: %v", err)
	}
	capCutIntro, err := NewVideoAnimation(capCutIntroType, 0, 500000)
	if err != nil {
		t.Fatalf("创建CapCut入场动画失败: %v", err)
	}

	if capCutIntro.Animation.Name != "1998" {
		t.Errorf("期望动画名称为 '1998', 得到 '%s'", capCutIntro.Animation.Name)
	}

	// 测试CapCut文本动画
	capCutTextIntroType, err := metadata.FindCapCutTextIntroByName("AI智能排版")
	if err != nil {
		t.Fatalf("查找CapCut文本动画失败: %v", err)
	}
	capCutTextIntro, err := NewTextAnimation(capCutTextIntroType, 0, 500000)
	if err != nil {
		t.Fatalf("创建CapCut文本动画失败: %v", err)
	}

	if capCutTextIntro.Animation.Name != "AI智能排版" {
		t.Errorf("期望动画名称为 'AI智能排版', 得到 '%s'", capCutTextIntro.Animation.Name)
	}
}

// TestSegmentAnimations 测试片段动画序列
func TestSegmentAnimations(t *testing.T) {
	segmentAnims := NewSegmentAnimations()

	if segmentAnims.AnimationID == "" {
		t.Error("期望动画ID不为空")
	}

	if len(segmentAnims.Animations) != 0 {
		t.Error("期望初始动画列表为空")
	}

	// 添加视频入场动画
	introType, err := metadata.FindIntroByName("渐显")
	if err != nil {
		t.Fatalf("查找入场动画失败: %v", err)
	}
	err = segmentAnims.AddVideoAnimation(introType, 0, 500000)
	if err != nil {
		t.Fatalf("添加视频入场动画失败: %v", err)
	}

	if len(segmentAnims.Animations) != 1 {
		t.Errorf("期望动画数量为 1, 得到 %d", len(segmentAnims.Animations))
	}

	// 添加视频出场动画
	err = segmentAnims.AddVideoAnimation(metadata.OutroType缩小, 1500000, 500000)
	if err != nil {
		t.Fatalf("添加视频出场动画失败: %v", err)
	}

	if len(segmentAnims.Animations) != 2 {
		t.Errorf("期望动画数量为 2, 得到 %d", len(segmentAnims.Animations))
	}
}

// TestSegmentAnimationsConstraints 测试片段动画约束
func TestSegmentAnimationsConstraints(t *testing.T) {
	segmentAnims := NewSegmentAnimations()

	// 添加第一个入场动画
	introType, err := metadata.FindIntroByName("渐显")
	if err != nil {
		t.Fatalf("查找入场动画失败: %v", err)
	}
	err = segmentAnims.AddVideoAnimation(introType, 0, 500000)
	if err != nil {
		t.Fatalf("添加第一个入场动画失败: %v", err)
	}

	// 尝试添加第二个入场动画，应该失败
	introType2, err := metadata.FindIntroByName("放大")
	if err != nil {
		t.Fatalf("查找放大动画失败: %v", err)
	}
	err = segmentAnims.AddVideoAnimation(introType2, 0, 500000)
	if err == nil {
		t.Error("期望添加重复类型动画失败，但成功了")
	}

	// 测试组合动画约束
	groupSegmentAnims := NewSegmentAnimations()
	introType3, err := metadata.FindIntroByName("渐显")
	if err != nil {
		t.Fatalf("查找入场动画失败: %v", err)
	}
	err = groupSegmentAnims.AddVideoAnimation(introType3, 0, 500000)
	if err != nil {
		t.Fatalf("添加入场动画失败: %v", err)
	}

	// 尝试添加组合动画，应该失败
	err = groupSegmentAnims.AddVideoAnimation(metadata.GroupAnimationType三分割, 0, 2000000)
	if err == nil {
		t.Error("期望在已有动画时添加组合动画失败，但成功了")
	}

	// 测试先添加组合动画
	groupFirstSegmentAnims := NewSegmentAnimations()
	err = groupFirstSegmentAnims.AddVideoAnimation(metadata.GroupAnimationType三分割, 0, 2000000)
	if err != nil {
		t.Fatalf("添加组合动画失败: %v", err)
	}

	// 尝试再添加其他动画，应该失败
	introType4, err := metadata.FindIntroByName("渐显")
	if err != nil {
		t.Fatalf("查找入场动画失败: %v", err)
	}
	err = groupFirstSegmentAnims.AddVideoAnimation(introType4, 0, 500000)
	if err == nil {
		t.Error("期望在已有组合动画时添加其他动画失败，但成功了")
	}
}

// TestTextAnimationConstraints 测试文本动画约束
func TestTextAnimationConstraints(t *testing.T) {
	segmentAnims := NewSegmentAnimations()

	// 添加文本入场动画
	err := segmentAnims.AddTextAnimation(metadata.TextIntro打字机, 0, 500000)
	if err != nil {
		t.Fatalf("添加文本入场动画失败: %v", err)
	}

	// 添加文本出场动画
	err = segmentAnims.AddTextAnimation(metadata.TextOutroType渐隐, 1500000, 500000)
	if err != nil {
		t.Fatalf("添加文本出场动画失败: %v", err)
	}

	// 添加文本循环动画
	err = segmentAnims.AddTextAnimation(metadata.TextLoopAnimType跳动, 0, 0)
	if err != nil {
		t.Fatalf("添加文本循环动画失败: %v", err)
	}

	if len(segmentAnims.Animations) != 3 {
		t.Errorf("期望动画数量为 3, 得到 %d", len(segmentAnims.Animations))
	}
}

// TestGetAnimationTimerange 测试获取动画时间范围
func TestGetAnimationTimerange(t *testing.T) {
	segmentAnims := NewSegmentAnimations()

	// 添加入场动画
	introType, err := metadata.FindIntroByName("渐显")
	if err != nil {
		t.Fatalf("查找入场动画失败: %v", err)
	}
	err = segmentAnims.AddVideoAnimation(introType, 100000, 500000)
	if err != nil {
		t.Fatalf("添加入场动画失败: %v", err)
	}

	// 获取入场动画时间范围
	timerange, err := segmentAnims.GetAnimationTimerange(AnimationTypeIn)
	if err != nil {
		t.Fatalf("获取动画时间范围失败: %v", err)
	}

	if timerange == nil {
		t.Fatal("期望获得时间范围，但得到nil")
	}

	if timerange.Start != 100000 {
		t.Errorf("期望开始时间为 100000, 得到 %d", timerange.Start)
	}

	if timerange.Duration != 500000 {
		t.Errorf("期望持续时间为 500000, 得到 %d", timerange.Duration)
	}

	// 获取不存在的动画类型
	timerange, err = segmentAnims.GetAnimationTimerange(AnimationTypeOut)
	if err != nil {
		t.Fatalf("获取不存在的动画时间范围失败: %v", err)
	}

	if timerange != nil {
		t.Error("期望获得nil，但得到了时间范围")
	}
}

// TestAnimationExportJSON 测试动画JSON导出
func TestAnimationExportJSON(t *testing.T) {
	// 测试视频动画JSON导出
	introType, err := metadata.FindIntroByName("渐显")
	if err != nil {
		t.Fatalf("查找入场动画失败: %v", err)
	}
	videoAnim, err := NewVideoAnimation(introType, 0, 500000)
	if err != nil {
		t.Fatalf("创建视频动画失败: %v", err)
	}

	videoJSON := videoAnim.Animation.ExportJSON()

	// 验证基本字段
	if videoJSON["name"] != "渐显" {
		t.Errorf("期望名称为 '渐显', 得到 '%v'", videoJSON["name"])
	}

	if videoJSON["type"] != "in" {
		t.Errorf("期望类型为 'in', 得到 '%v'", videoJSON["type"])
	}

	if videoJSON["panel"] != "video" {
		t.Errorf("期望panel为 'video', 得到 '%v'", videoJSON["panel"])
	}

	if videoJSON["material_type"] != "video" {
		t.Errorf("期望material_type为 'video', 得到 '%v'", videoJSON["material_type"])
	}

	if videoJSON["start"] != int64(0) {
		t.Errorf("期望开始时间为 0, 得到 %v", videoJSON["start"])
	}

	if videoJSON["duration"] != int64(500000) {
		t.Errorf("期望持续时间为 500000, 得到 %v", videoJSON["duration"])
	}

	// 测试文本动画JSON导出
	textAnim, err := NewTextAnimation(metadata.TextIntro打字机, 100000, 800000)
	if err != nil {
		t.Fatalf("创建文本动画失败: %v", err)
	}

	textJSON := textAnim.Animation.ExportJSON()

	if textJSON["panel"] != "" {
		t.Errorf("期望文本动画panel为空字符串, 得到 '%v'", textJSON["panel"])
	}

	if textJSON["material_type"] != "sticker" {
		t.Errorf("期望文本动画material_type为 'sticker', 得到 '%v'", textJSON["material_type"])
	}
}

// TestSegmentAnimationsExportJSON 测试片段动画序列JSON导出
func TestSegmentAnimationsExportJSON(t *testing.T) {
	segmentAnims := NewSegmentAnimations()

	// 添加多个动画
	introType, err := metadata.FindIntroByName("渐显")
	if err != nil {
		t.Fatalf("查找入场动画失败: %v", err)
	}
	err = segmentAnims.AddVideoAnimation(introType, 0, 500000)
	if err != nil {
		t.Fatalf("添加入场动画失败: %v", err)
	}

	err = segmentAnims.AddVideoAnimation(metadata.OutroType缩小, 1500000, 500000)
	if err != nil {
		t.Fatalf("添加出场动画失败: %v", err)
	}

	jsonData := segmentAnims.ExportJSON()

	// 验证基本字段
	if jsonData["id"] == "" {
		t.Error("期望ID不为空")
	}

	if jsonData["type"] != "sticker_animation" {
		t.Errorf("期望类型为 'sticker_animation', 得到 '%v'", jsonData["type"])
	}

	if jsonData["multi_language_current"] != "none" {
		t.Errorf("期望multi_language_current为 'none', 得到 '%v'", jsonData["multi_language_current"])
	}

	// 验证动画列表
	animations, ok := jsonData["animations"].([]map[string]interface{})
	if !ok {
		t.Fatal("期望animations为数组")
	}

	if len(animations) != 2 {
		t.Errorf("期望动画数量为 2, 得到 %d", len(animations))
	}

	// 验证第一个动画
	firstAnim := animations[0]
	if firstAnim["name"] != "渐显" {
		t.Errorf("期望第一个动画名称为 '渐显', 得到 '%v'", firstAnim["name"])
	}

	if firstAnim["type"] != "in" {
		t.Errorf("期望第一个动画类型为 'in', 得到 '%v'", firstAnim["type"])
	}

	// 验证第二个动画
	secondAnim := animations[1]
	if secondAnim["name"] != "缩小" {
		t.Errorf("期望第二个动画名称为 '缩小', 得到 '%v'", secondAnim["name"])
	}

	if secondAnim["type"] != "out" {
		t.Errorf("期望第二个动画类型为 'out', 得到 '%v'", secondAnim["type"])
	}
}

// TestJSONSerialization 测试JSON序列化兼容性
func TestJSONSerialization(t *testing.T) {
	segmentAnims := NewSegmentAnimations()

	// 添加动画
	introType, err := metadata.FindIntroByName("渐显")
	if err != nil {
		t.Fatalf("查找入场动画失败: %v", err)
	}
	err = segmentAnims.AddVideoAnimation(introType, 0, 500000)
	if err != nil {
		t.Fatalf("添加动画失败: %v", err)
	}

	// 导出JSON
	jsonData := segmentAnims.ExportJSON()

	// 序列化为JSON字符串
	jsonBytes, err := json.Marshal(jsonData)
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
	if unmarshaled["type"] != "sticker_animation" {
		t.Errorf("反序列化后类型不匹配")
	}

	animations := unmarshaled["animations"].([]interface{})
	if len(animations) != 1 {
		t.Errorf("反序列化后动画数量不匹配")
	}
}

// TestFindAnimationTypeByName 测试根据名称查找动画类型
func TestFindAnimationTypeByName(t *testing.T) {
	// 测试查找入场动画
	introType, err := metadata.FindIntroByName("渐显")
	if err != nil {
		t.Fatalf("查找入场动画失败: %v", err)
	}

	if introType.GetName() != "渐显" {
		t.Errorf("期望动画名称为 '渐显', 得到 '%s'", introType.GetName())
	}

	// 测试查找不存在的动画
	_, err = metadata.FindIntroByName("不存在的动画")
	if err == nil {
		t.Error("期望查找不存在的动画失败，但成功了")
	}

	// 测试查找文本动画
	textIntroType, err := metadata.FindTextIntroByName("打字机")
	if err != nil {
		t.Fatalf("查找文本入场动画失败: %v", err)
	}

	if textIntroType.GetName() != "打字机" {
		t.Errorf("期望动画名称为 '打字机', 得到 '%s'", textIntroType.GetName())
	}

	// 测试查找CapCut动画
	capCutIntroType, err := metadata.FindCapCutIntroByName("1998")
	if err != nil {
		t.Fatalf("查找CapCut入场动画失败: %v", err)
	}

	if capCutIntroType.GetName() != "1998" {
		t.Errorf("期望动画名称为 '1998', 得到 '%s'", capCutIntroType.GetName())
	}
}
