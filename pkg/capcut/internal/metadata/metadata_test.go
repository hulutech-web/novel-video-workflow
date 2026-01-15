package metadata

import (
	"encoding/json"
	"testing"
)

// TestAnimationMeta 测试动画元数据
func TestAnimationMeta(t *testing.T) {
	meta := NewAnimationMeta("测试动画", false, 1.5, "123456", "789012", "abcdef")

	if meta.Title != "测试动画" {
		t.Errorf("期望标题为 '测试动画', 得到 '%s'", meta.Title)
	}

	// 验证时间转换（1.5秒 = 1500000微秒）
	expectedDuration := int64(1500000)
	if meta.Duration != expectedDuration {
		t.Errorf("期望持续时间为 %d, 得到 %d", expectedDuration, meta.Duration)
	}
}

// TestEffectParam 测试特效参数
func TestEffectParam(t *testing.T) {
	param := NewEffectParam("test_param", 50.0, 0.0, 100.0)

	if param.Name != "test_param" {
		t.Errorf("期望参数名称为 'test_param', 得到 '%s'", param.Name)
	}

	if param.DefaultValue != 50.0 {
		t.Errorf("期望默认值为 50.0, 得到 %f", param.DefaultValue)
	}
}

// TestEffectParamInstance 测试特效参数实例
func TestEffectParamInstance(t *testing.T) {
	param := NewEffectParam("test_param", 50.0, 0.0, 100.0)
	instance := NewEffectParamInstance(param, 0, 75.0)

	if instance.Index != 0 {
		t.Errorf("期望参数索引为 0, 得到 %d", instance.Index)
	}

	if instance.Value != 75.0 {
		t.Errorf("期望参数值为 75.0, 得到 %f", instance.Value)
	}

	// 测试JSON导出
	jsonData := instance.ExportJSON()
	if jsonData["name"] != "test_param" {
		t.Errorf("JSON导出的name不正确")
	}
}

// TestEffectMeta 测试特效元数据
func TestEffectMeta(t *testing.T) {
	params := []EffectParam{
		NewEffectParam("intensity", 50.0, 0.0, 100.0),
		NewEffectParam("speed", 30.0, 10.0, 90.0),
	}

	meta := NewEffectMeta("测试特效", true, "effect123", "eff456", "hash789", params)

	if meta.Name != "测试特效" {
		t.Errorf("期望特效名称为 '测试特效', 得到 '%s'", meta.Name)
	}

	if len(meta.Params) != 2 {
		t.Errorf("期望参数数量为 2, 得到 %d", len(meta.Params))
	}

	// 测试参数解析
	testParams := []float64{80.0, 60.0}
	instances, err := meta.ParseParams(testParams)
	if err != nil {
		t.Errorf("参数解析失败: %v", err)
	}

	if len(instances) != 2 {
		t.Errorf("期望解析出 2 个参数实例, 得到 %d", len(instances))
	}
}

// TestMaskMeta 测试蒙版元数据
func TestMaskMeta(t *testing.T) {
	meta := NewMaskMeta("圆形蒙版", "circle", "mask001", "effect001", "hash123", 1.0)

	if meta.Name != "圆形蒙版" {
		t.Errorf("期望蒙版名称为 '圆形蒙版', 得到 '%s'", meta.Name)
	}

	if meta.ResourceType != "circle" {
		t.Errorf("期望资源类型为 'circle', 得到 '%s'", meta.ResourceType)
	}
}

// TestTransitionMeta 测试转场元数据
func TestTransitionMeta(t *testing.T) {
	meta := NewTransitionMeta("淡入淡出", false, "trans001", "effect001", "hash123", 1.5, true)

	if meta.Name != "淡入淡出" {
		t.Errorf("期望转场名称为 '淡入淡出', 得到 '%s'", meta.Name)
	}

	// 验证时间转换（1.5秒 = 1500000微秒）
	expectedDuration := int64(1500000)
	if meta.DefaultDuration != expectedDuration {
		t.Errorf("期望默认持续时间为 %d, 得到 %d", expectedDuration, meta.DefaultDuration)
	}
}

// TestFilterMeta 测试滤镜元数据
func TestFilterMeta(t *testing.T) {
	meta := NewFilterMeta("自然滤镜", false, "filter001", "effect001", "hash123", "人像", "自然肤色效果", 0.8)

	if meta.Category != "人像" {
		t.Errorf("期望滤镜分类为 '人像', 得到 '%s'", meta.Category)
	}

	if meta.Intensity != 0.8 {
		t.Errorf("期望滤镜强度为 0.8, 得到 %f", meta.Intensity)
	}
}

// TestFontMeta 测试字体元数据
func TestFontMeta(t *testing.T) {
	languages := []string{"zh-CN", "en-US"}
	meta := NewFontMeta("苹方", false, "font001", "PingFang SC", "normal", "normal", "系统", "现代简洁字体", "苹方字体", languages)

	if meta.FontFamily != "PingFang SC" {
		t.Errorf("期望字体族为 'PingFang SC', 得到 '%s'", meta.FontFamily)
	}

	if len(meta.Language) != 2 {
		t.Errorf("期望支持语言数量为 2, 得到 %d", len(meta.Language))
	}
}

// TestAudioEffectMeta 测试音频特效元数据
func TestAudioEffectMeta(t *testing.T) {
	params := []EffectParam{
		NewEffectParam("intensity", 50.0, 0.0, 100.0),
	}

	meta := NewAudioEffectMeta("雨声", false, "audio001", "effect001", "hash123", "环境", "自然雨声效果", params)

	if meta.Category != "环境" {
		t.Errorf("期望音效分类为 '环境', 得到 '%s'", meta.Category)
	}

	if len(meta.Params) != 1 {
		t.Errorf("期望音效参数数量为 1, 得到 %d", len(meta.Params))
	}
}

// TestFindEffectByName 测试按名称查找特效
func TestFindEffectByName(t *testing.T) {
	// 测试入场动画查找
	intro, err := FindIntroByName("缩小")
	if err != nil {
		t.Errorf("查找入场动画失败: %v", err)
	}

	if intro.GetName() != "缩小" {
		t.Errorf("期望找到名称为 '缩小' 的动画, 得到 '%s'", intro.GetName())
	}

	// 测试蒙版查找
	mask, err := FindMaskByName("圆形")
	if err != nil {
		t.Errorf("查找蒙版失败: %v", err)
	}

	if mask.GetName() != "圆形" {
		t.Errorf("期望找到名称为 '圆形' 的蒙版, 得到 '%s'", mask.GetName())
	}
}

// TestGetAllTypes 测试获取所有类型
func TestGetAllTypes(t *testing.T) {
	// 测试获取所有入场动画
	intros := GetAllIntroTypes()
	if len(intros) == 0 {
		t.Errorf("期望至少有一个入场动画")
	}

	// 测试获取所有蒙版
	masks := GetAllMaskTypes()
	if len(masks) == 0 {
		t.Errorf("期望至少有一个蒙版")
	}

	// 测试获取所有转场
	transitions := GetAllTransitionTypes()
	if len(transitions) == 0 {
		t.Errorf("期望至少有一个转场")
	}

	// 测试获取所有滤镜
	filters := GetAllFilterTypes()
	if len(filters) == 0 {
		t.Errorf("期望至少有一个滤镜")
	}
}

// TestCategoryFilters 测试分类过滤
func TestCategoryFilters(t *testing.T) {
	// 测试滤镜分类过滤
	portraitFilters := GetFiltersByCategory("人像")
	if len(portraitFilters) == 0 {
		t.Errorf("期望至少有一个人像滤镜")
	}

	// 测试字体分类过滤
	systemFonts := GetFontsByCategory("系统")
	if len(systemFonts) == 0 {
		t.Errorf("期望至少有一个系统字体")
	}

	// 测试音频特效分类过滤
	envEffects := GetAudioEffectsByCategory("环境")
	if len(envEffects) == 0 {
		t.Errorf("期望至少有一个环境音效")
	}
}

// TestCapCutMetadata 测试CapCut特有元数据
func TestCapCutMetadata(t *testing.T) {
	// 测试CapCut入场动画
	capCutIntros := GetAllCapCutIntroTypes()
	if len(capCutIntros) == 0 {
		t.Errorf("期望至少有一个CapCut入场动画")
	}

	// 测试CapCut蒙版
	capCutMasks := GetAllCapCutMaskTypes()
	if len(capCutMasks) == 0 {
		t.Errorf("期望至少有一个CapCut蒙版")
	}

	// 测试CapCut语音特效
	capCutVoiceFilters := GetAllCapCutVoiceFiltersEffectTypes()
	if len(capCutVoiceFilters) == 0 {
		t.Errorf("期望至少有一个CapCut语音滤镜")
	}
}

// TestJSONSerialization 测试JSON序列化
func TestJSONSerialization(t *testing.T) {
	// 测试动画元数据JSON序列化
	animMeta := NewAnimationMeta("测试动画", true, 2.0, "res123", "eff456", "hash789")
	animJSON, err := json.Marshal(animMeta)
	if err != nil {
		t.Errorf("动画元数据JSON序列化失败: %v", err)
	}

	var animDeserialized AnimationMeta
	err = json.Unmarshal(animJSON, &animDeserialized)
	if err != nil {
		t.Errorf("动画元数据JSON反序列化失败: %v", err)
	}

	if animDeserialized.Title != "测试动画" {
		t.Errorf("反序列化后标题不匹配")
	}
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	params := []EffectParam{
		NewEffectParam("intensity", 50.0, 0.0, 100.0),
	}
	meta := NewEffectMeta("测试特效", false, "effect123", "eff456", "hash789", params)

	// 测试超出范围的参数值
	invalidParams := []float64{150.0} // 超出100的范围
	_, err := meta.ParseParams(invalidParams)
	if err == nil {
		t.Errorf("期望解析超出范围的参数时返回错误")
	}

	// 测试不存在的特效查找
	_, err = FindIntroByName("不存在的动画")
	if err == nil {
		t.Errorf("期望查找不存在的动画时返回错误")
	}
}
