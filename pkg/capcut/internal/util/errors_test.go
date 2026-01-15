package util

import (
	"strings"
	"testing"
)

// TestTrackNotFoundError 测试轨道未找到错误
func TestTrackNotFoundError(t *testing.T) {
	condition := "type=video AND name=主视频"
	err := NewTrackNotFoundError(condition)

	// 测试错误消息
	expectedMsg := "track not found: " + condition
	if err.Error() != expectedMsg {
		t.Errorf("期望错误消息'%s'，得到'%s'", expectedMsg, err.Error())
	}

	// 测试错误类型检查
	if !IsTrackNotFound(err) {
		t.Error("IsTrackNotFound应该返回true")
	}

	// 测试其他错误类型检查返回false
	if IsAmbiguousTrack(err) {
		t.Error("IsAmbiguousTrack应该返回false")
	}
}

// TestAmbiguousTrackError 测试轨道模糊错误
func TestAmbiguousTrackError(t *testing.T) {
	condition := "name=视频轨道"
	count := 3
	err := NewAmbiguousTrackError(condition, count)

	// 测试错误消息
	expectedMsg := "ambiguous track: found 3 tracks matching condition: " + condition
	if err.Error() != expectedMsg {
		t.Errorf("期望错误消息'%s'，得到'%s'", expectedMsg, err.Error())
	}

	// 测试错误类型检查
	if !IsAmbiguousTrack(err) {
		t.Error("IsAmbiguousTrack应该返回true")
	}

	// 测试字段访问
	if err.Count != count {
		t.Errorf("期望Count为%d，得到%d", count, err.Count)
	}

	if err.Condition != condition {
		t.Errorf("期望Condition为'%s'，得到'%s'", condition, err.Condition)
	}
}

// TestSegmentOverlapError 测试片段重叠错误
func TestSegmentOverlapError(t *testing.T) {
	newStart, newEnd := int64(1000000), int64(5000000)
	existingStart, existingEnd := int64(3000000), int64(7000000)

	err := NewSegmentOverlapError(newStart, newEnd, existingStart, existingEnd)

	// 测试错误消息包含所有时间信息
	errMsg := err.Error()
	if !strings.Contains(errMsg, "1000000") || !strings.Contains(errMsg, "5000000") ||
		!strings.Contains(errMsg, "3000000") || !strings.Contains(errMsg, "7000000") {
		t.Errorf("错误消息应包含所有时间信息: %s", errMsg)
	}

	// 测试错误类型检查
	if !IsSegmentOverlap(err) {
		t.Error("IsSegmentOverlap应该返回true")
	}

	// 测试字段访问
	if err.NewSegmentStart != newStart {
		t.Errorf("期望NewSegmentStart为%d，得到%d", newStart, err.NewSegmentStart)
	}
}

// TestMaterialNotFoundError 测试素材未找到错误
func TestMaterialNotFoundError(t *testing.T) {
	condition := "path=/path/to/video.mp4"
	err := NewMaterialNotFoundError(condition)

	// 测试错误消息
	expectedMsg := "material not found: " + condition
	if err.Error() != expectedMsg {
		t.Errorf("期望错误消息'%s'，得到'%s'", expectedMsg, err.Error())
	}

	// 测试错误类型检查
	if !IsMaterialNotFound(err) {
		t.Error("IsMaterialNotFound应该返回true")
	}
}

// TestAmbiguousMaterialError 测试素材模糊错误
func TestAmbiguousMaterialError(t *testing.T) {
	condition := "type=video"
	count := 5
	err := NewAmbiguousMaterialError(condition, count)

	// 测试错误消息
	expectedMsg := "ambiguous material: found 5 materials matching condition: " + condition
	if err.Error() != expectedMsg {
		t.Errorf("期望错误消息'%s'，得到'%s'", expectedMsg, err.Error())
	}

	// 测试错误类型检查
	if !IsAmbiguousMaterial(err) {
		t.Error("IsAmbiguousMaterial应该返回true")
	}
}

// TestExtensionFailedError 测试延伸失败错误
func TestExtensionFailedError(t *testing.T) {
	reason := "新素材长度不足"
	segmentID := "segment123"
	materialID := "material456"

	err := NewExtensionFailedError(reason, segmentID, materialID)

	// 测试错误消息包含所有信息
	errMsg := err.Error()
	if !strings.Contains(errMsg, reason) || !strings.Contains(errMsg, segmentID) || !strings.Contains(errMsg, materialID) {
		t.Errorf("错误消息应包含所有信息: %s", errMsg)
	}

	// 测试错误类型检查
	if !IsExtensionFailed(err) {
		t.Error("IsExtensionFailed应该返回true")
	}

	// 测试字段访问
	if err.Reason != reason {
		t.Errorf("期望Reason为'%s'，得到'%s'", reason, err.Reason)
	}
}

// TestDraftNotFoundError 测试草稿未找到错误
func TestDraftNotFoundError(t *testing.T) {
	// 测试按路径创建的错误
	path := "/path/to/draft"
	err := NewDraftNotFoundError(path)

	expectedMsg := "draft not found at path: " + path
	if err.Error() != expectedMsg {
		t.Errorf("期望错误消息'%s'，得到'%s'", expectedMsg, err.Error())
	}

	// 测试错误类型检查
	if !IsDraftNotFound(err) {
		t.Error("IsDraftNotFound应该返回true")
	}

	// 测试按名称创建的错误
	name := "测试草稿"
	err2 := NewDraftNotFoundErrorByName(name)

	expectedMsg2 := "draft not found: " + name
	if err2.Error() != expectedMsg2 {
		t.Errorf("期望错误消息'%s'，得到'%s'", expectedMsg2, err2.Error())
	}
}

// TestAutomationError 测试自动化错误
func TestAutomationError(t *testing.T) {
	operation := "export_video"
	reason := "剪映窗口未响应"

	err := NewAutomationError(operation, reason)

	// 测试错误消息
	errMsg := err.Error()
	if !strings.Contains(errMsg, operation) || !strings.Contains(errMsg, reason) {
		t.Errorf("错误消息应包含操作和原因: %s", errMsg)
	}

	// 测试错误类型检查
	if !IsAutomationError(err) {
		t.Error("IsAutomationError应该返回true")
	}
}

// TestExportTimeoutError 测试导出超时错误
func TestExportTimeoutError(t *testing.T) {
	duration := int64(300) // 5分钟
	filePath := "/output/video.mp4"

	err := NewExportTimeoutError(duration, filePath)

	// 测试错误消息
	errMsg := err.Error()
	if !strings.Contains(errMsg, "300") || !strings.Contains(errMsg, filePath) {
		t.Errorf("错误消息应包含持续时间和文件路径: %s", errMsg)
	}

	// 测试错误类型检查
	if !IsExportTimeout(err) {
		t.Error("IsExportTimeout应该返回true")
	}

	// 测试字段访问
	if err.Duration != duration {
		t.Errorf("期望Duration为%d，得到%d", duration, err.Duration)
	}

	if err.FilePath != filePath {
		t.Errorf("期望FilePath为'%s'，得到'%s'", filePath, err.FilePath)
	}
}

// TestValidationError 测试数据验证错误
func TestValidationError(t *testing.T) {
	field := "duration"
	value := -100
	reason := "持续时间不能为负数"

	err := NewValidationError(field, value, reason)

	// 测试错误消息
	errMsg := err.Error()
	if !strings.Contains(errMsg, field) || !strings.Contains(errMsg, reason) {
		t.Errorf("错误消息应包含字段和原因: %s", errMsg)
	}

	// 测试错误类型检查
	if !IsValidationError(err) {
		t.Error("IsValidationError应该返回true")
	}

	// 测试字段访问
	if err.Field != field {
		t.Errorf("期望Field为'%s'，得到'%s'", field, err.Field)
	}

	if err.Value != value {
		t.Errorf("期望Value为%v，得到%v", value, err.Value)
	}
}

// TestJSONProcessingError 测试JSON处理错误
func TestJSONProcessingError(t *testing.T) {
	operation := "unmarshal"
	data := `{"invalid": json}`
	reason := "unexpected token"

	err := NewJSONProcessingError(operation, data, reason)

	// 测试错误消息
	errMsg := err.Error()
	if !strings.Contains(errMsg, operation) || !strings.Contains(errMsg, data) || !strings.Contains(errMsg, reason) {
		t.Errorf("错误消息应包含所有信息: %s", errMsg)
	}

	// 测试错误类型检查
	if !IsJSONProcessingError(err) {
		t.Error("IsJSONProcessingError应该返回true")
	}
}

// TestTypeConversionError 测试类型转换错误
func TestTypeConversionError(t *testing.T) {
	sourceType := "string"
	targetType := "int"
	value := "not_a_number"

	err := NewTypeConversionError(sourceType, targetType, value)

	// 测试错误消息
	errMsg := err.Error()
	if !strings.Contains(errMsg, sourceType) || !strings.Contains(errMsg, targetType) {
		t.Errorf("错误消息应包含源类型和目标类型: %s", errMsg)
	}

	// 测试错误类型检查
	if !IsTypeConversionError(err) {
		t.Error("IsTypeConversionError应该返回true")
	}

	// 测试字段访问
	if err.SourceType != sourceType {
		t.Errorf("期望SourceType为'%s'，得到'%s'", sourceType, err.SourceType)
	}

	if err.TargetType != targetType {
		t.Errorf("期望TargetType为'%s'，得到'%s'", targetType, err.TargetType)
	}
}

// TestConfigurationError 测试配置错误
func TestConfigurationError(t *testing.T) {
	component := "video_encoder"
	setting := "bitrate"
	reason := "值超出允许范围"

	err := NewConfigurationError(component, setting, reason)

	// 测试错误消息
	errMsg := err.Error()
	if !strings.Contains(errMsg, component) || !strings.Contains(errMsg, setting) || !strings.Contains(errMsg, reason) {
		t.Errorf("错误消息应包含所有信息: %s", errMsg)
	}

	// 测试错误类型检查
	if !IsConfigurationError(err) {
		t.Error("IsConfigurationError应该返回true")
	}
}

// TestErrorTypeChecking 测试错误类型检查的准确性
func TestErrorTypeChecking(t *testing.T) {
	// 创建不同类型的错误
	trackNotFound := NewTrackNotFoundError("test")
	materialNotFound := NewMaterialNotFoundError("test")
	segmentOverlap := NewSegmentOverlapError(0, 100, 50, 150)

	// 测试类型检查函数的准确性
	if IsTrackNotFound(materialNotFound) {
		t.Error("IsTrackNotFound不应该对MaterialNotFoundError返回true")
	}

	if IsMaterialNotFound(trackNotFound) {
		t.Error("IsMaterialNotFound不应该对TrackNotFoundError返回true")
	}

	if IsSegmentOverlap(trackNotFound) {
		t.Error("IsSegmentOverlap不应该对TrackNotFoundError返回true")
	}

	// 测试正确的类型检查
	if !IsTrackNotFound(trackNotFound) {
		t.Error("IsTrackNotFound应该对TrackNotFoundError返回true")
	}

	if !IsMaterialNotFound(materialNotFound) {
		t.Error("IsMaterialNotFound应该对MaterialNotFoundError返回true")
	}

	if !IsSegmentOverlap(segmentOverlap) {
		t.Error("IsSegmentOverlap应该对SegmentOverlapError返回true")
	}
}

// TestErrorChaining 测试错误链
func TestErrorChaining(t *testing.T) {
	// 创建一个基础错误
	baseErr := NewValidationError("field", "value", "invalid format")

	// 创建一个包装错误
	wrapperErr := NewJSONProcessingError("parse", "data", baseErr.Error())

	// 测试错误消息传播
	if !strings.Contains(wrapperErr.Error(), "invalid format") {
		t.Error("包装错误应该包含基础错误的信息")
	}

	// 测试错误类型检查
	if !IsJSONProcessingError(wrapperErr) {
		t.Error("包装错误应该保持其类型")
	}

	if IsValidationError(wrapperErr) {
		t.Error("包装错误不应该被识别为基础错误类型")
	}
}

// TestErrorWithNilValues 测试包含nil值的错误
func TestErrorWithNilValues(t *testing.T) {
	// 测试空字符串字段
	err1 := NewTrackNotFoundError("")
	if err1.Error() == "" {
		t.Error("即使条件为空，错误消息也不应该为空")
	}

	// 测试零值
	err2 := NewAmbiguousTrackError("", 0)
	if !strings.Contains(err2.Error(), "0") {
		t.Error("错误消息应该包含零值计数")
	}

	// 测试nil值在类型转换中的处理
	err3 := NewTypeConversionError("", "", nil)
	errMsg := err3.Error()
	if errMsg == "" {
		t.Error("即使参数为空/nil，错误消息也不应该为空")
	}
}

// TestErrorInterface 测试所有错误都实现了error接口
func TestErrorInterface(t *testing.T) {
	errors := []error{
		NewTrackNotFoundError("test"),
		NewAmbiguousTrackError("test", 1),
		NewSegmentOverlapError(0, 100, 50, 150),
		NewMaterialNotFoundError("test"),
		NewAmbiguousMaterialError("test", 1),
		NewExtensionFailedError("test", "seg", "mat"),
		NewDraftNotFoundError("test"),
		NewAutomationError("test", "test"),
		NewExportTimeoutError(60, "test"),
		NewValidationError("test", "test", "test"),
		NewJSONProcessingError("test", "test", "test"),
		NewTypeConversionError("test", "test", "test"),
		NewConfigurationError("test", "test", "test"),
	}

	for i, err := range errors {
		// 测试Error()方法
		if err.Error() == "" {
			t.Errorf("错误%d的Error()方法返回空字符串", i)
		}

		// 测试错误可以被赋值给error接口
		var genericErr error = err
		if genericErr == nil {
			t.Errorf("错误%d不能被赋值给error接口", i)
		}
	}
}
