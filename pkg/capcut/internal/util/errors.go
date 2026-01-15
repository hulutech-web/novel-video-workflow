// Package util 提供自定义错误类型
// 对应Python的 pyJianYingDraft/exceptions.py
package util

import "fmt"

// TrackNotFoundError 未找到满足条件的轨道
// 对应Python的TrackNotFound异常
type TrackNotFoundError struct {
	Condition string
}

func (e *TrackNotFoundError) Error() string {
	return fmt.Sprintf("track not found: %s", e.Condition)
}

// NewTrackNotFoundError 创建轨道未找到错误
func NewTrackNotFoundError(condition string) *TrackNotFoundError {
	return &TrackNotFoundError{Condition: condition}
}

// AmbiguousTrackError 找到多个满足条件的轨道
// 对应Python的AmbiguousTrack异常
type AmbiguousTrackError struct {
	Condition string
	Count     int
}

func (e *AmbiguousTrackError) Error() string {
	return fmt.Sprintf("ambiguous track: found %d tracks matching condition: %s", e.Count, e.Condition)
}

// NewAmbiguousTrackError 创建轨道模糊错误
func NewAmbiguousTrackError(condition string, count int) *AmbiguousTrackError {
	return &AmbiguousTrackError{Condition: condition, Count: count}
}

// SegmentOverlapError 新片段与已有的轨道片段重叠
// 对应Python的SegmentOverlap异常
type SegmentOverlapError struct {
	NewSegmentStart int64
	NewSegmentEnd   int64
	ExistingStart   int64
	ExistingEnd     int64
}

func (e *SegmentOverlapError) Error() string {
	return fmt.Sprintf("segment overlap: new segment [%d, %d] overlaps with existing segment [%d, %d]",
		e.NewSegmentStart, e.NewSegmentEnd, e.ExistingStart, e.ExistingEnd)
}

// NewSegmentOverlapError 创建片段重叠错误
func NewSegmentOverlapError(newStart, newEnd, existingStart, existingEnd int64) *SegmentOverlapError {
	return &SegmentOverlapError{
		NewSegmentStart: newStart,
		NewSegmentEnd:   newEnd,
		ExistingStart:   existingStart,
		ExistingEnd:     existingEnd,
	}
}

// MaterialNotFoundError 未找到满足条件的素材
// 对应Python的MaterialNotFound异常
type MaterialNotFoundError struct {
	Condition string
}

func (e *MaterialNotFoundError) Error() string {
	return fmt.Sprintf("material not found: %s", e.Condition)
}

// NewMaterialNotFoundError 创建素材未找到错误
func NewMaterialNotFoundError(condition string) *MaterialNotFoundError {
	return &MaterialNotFoundError{Condition: condition}
}

// AmbiguousMaterialError 找到多个满足条件的素材
// 对应Python的AmbiguousMaterial异常
type AmbiguousMaterialError struct {
	Condition string
	Count     int
}

func (e *AmbiguousMaterialError) Error() string {
	return fmt.Sprintf("ambiguous material: found %d materials matching condition: %s", e.Count, e.Condition)
}

// NewAmbiguousMaterialError 创建素材模糊错误
func NewAmbiguousMaterialError(condition string, count int) *AmbiguousMaterialError {
	return &AmbiguousMaterialError{Condition: condition, Count: count}
}

// ExtensionFailedError 替换素材时延伸片段失败
// 对应Python的ExtensionFailed异常
type ExtensionFailedError struct {
	Reason     string
	SegmentID  string
	MaterialID string
}

func (e *ExtensionFailedError) Error() string {
	return fmt.Sprintf("extension failed for segment %s with material %s: %s", e.SegmentID, e.MaterialID, e.Reason)
}

// NewExtensionFailedError 创建延伸失败错误
func NewExtensionFailedError(reason, segmentID, materialID string) *ExtensionFailedError {
	return &ExtensionFailedError{
		Reason:     reason,
		SegmentID:  segmentID,
		MaterialID: materialID,
	}
}

// DraftNotFoundError 未找到草稿
// 对应Python的DraftNotFound异常
type DraftNotFoundError struct {
	DraftPath string
	DraftName string
}

func (e *DraftNotFoundError) Error() string {
	if e.DraftName != "" {
		return fmt.Sprintf("draft not found: %s", e.DraftName)
	}
	return fmt.Sprintf("draft not found at path: %s", e.DraftPath)
}

// NewDraftNotFoundError 创建草稿未找到错误
func NewDraftNotFoundError(path string) *DraftNotFoundError {
	return &DraftNotFoundError{DraftPath: path}
}

// NewDraftNotFoundErrorByName 根据名称创建草稿未找到错误
func NewDraftNotFoundErrorByName(name string) *DraftNotFoundError {
	return &DraftNotFoundError{DraftName: name}
}

// AutomationError 自动化操作失败
// 对应Python的AutomationError异常
type AutomationError struct {
	Operation string
	Reason    string
}

func (e *AutomationError) Error() string {
	return fmt.Sprintf("automation error in operation '%s': %s", e.Operation, e.Reason)
}

// NewAutomationError 创建自动化错误
func NewAutomationError(operation, reason string) *AutomationError {
	return &AutomationError{Operation: operation, Reason: reason}
}

// ExportTimeoutError 导出超时
// 对应Python的ExportTimeout异常
type ExportTimeoutError struct {
	Duration int64 // 超时时长，单位秒
	FilePath string
}

func (e *ExportTimeoutError) Error() string {
	return fmt.Sprintf("export timeout after %d seconds for file: %s", e.Duration, e.FilePath)
}

// NewExportTimeoutError 创建导出超时错误
func NewExportTimeoutError(duration int64, filePath string) *ExportTimeoutError {
	return &ExportTimeoutError{Duration: duration, FilePath: filePath}
}

// ValidationError 数据验证错误
type ValidationError struct {
	Field  string
	Value  interface{}
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s' with value '%v': %s", e.Field, e.Value, e.Reason)
}

// NewValidationError 创建验证错误
func NewValidationError(field string, value interface{}, reason string) *ValidationError {
	return &ValidationError{Field: field, Value: value, Reason: reason}
}

// JSONProcessingError JSON处理错误
type JSONProcessingError struct {
	Operation string
	Data      string
	Reason    string
}

func (e *JSONProcessingError) Error() string {
	return fmt.Sprintf("JSON processing error during %s: %s (data: %s)", e.Operation, e.Reason, e.Data)
}

// NewJSONProcessingError 创建JSON处理错误
func NewJSONProcessingError(operation, data, reason string) *JSONProcessingError {
	return &JSONProcessingError{Operation: operation, Data: data, Reason: reason}
}

// TypeConversionError 类型转换错误
type TypeConversionError struct {
	SourceType string
	TargetType string
	Value      interface{}
}

func (e *TypeConversionError) Error() string {
	return fmt.Sprintf("type conversion error: cannot convert %v (type %s) to %s", e.Value, e.SourceType, e.TargetType)
}

// NewTypeConversionError 创建类型转换错误
func NewTypeConversionError(sourceType, targetType string, value interface{}) *TypeConversionError {
	return &TypeConversionError{
		SourceType: sourceType,
		TargetType: targetType,
		Value:      value,
	}
}

// ConfigurationError 配置错误
type ConfigurationError struct {
	Component string
	Setting   string
	Reason    string
}

func (e *ConfigurationError) Error() string {
	return fmt.Sprintf("configuration error in %s for setting '%s': %s", e.Component, e.Setting, e.Reason)
}

// NewConfigurationError 创建配置错误
func NewConfigurationError(component, setting, reason string) *ConfigurationError {
	return &ConfigurationError{Component: component, Setting: setting, Reason: reason}
}

// 错误检查辅助函数

// IsTrackNotFound 检查是否为轨道未找到错误
func IsTrackNotFound(err error) bool {
	_, ok := err.(*TrackNotFoundError)
	return ok
}

// IsAmbiguousTrack 检查是否为轨道模糊错误
func IsAmbiguousTrack(err error) bool {
	_, ok := err.(*AmbiguousTrackError)
	return ok
}

// IsSegmentOverlap 检查是否为片段重叠错误
func IsSegmentOverlap(err error) bool {
	_, ok := err.(*SegmentOverlapError)
	return ok
}

// IsMaterialNotFound 检查是否为素材未找到错误
func IsMaterialNotFound(err error) bool {
	_, ok := err.(*MaterialNotFoundError)
	return ok
}

// IsAmbiguousMaterial 检查是否为素材模糊错误
func IsAmbiguousMaterial(err error) bool {
	_, ok := err.(*AmbiguousMaterialError)
	return ok
}

// IsExtensionFailed 检查是否为延伸失败错误
func IsExtensionFailed(err error) bool {
	_, ok := err.(*ExtensionFailedError)
	return ok
}

// IsDraftNotFound 检查是否为草稿未找到错误
func IsDraftNotFound(err error) bool {
	_, ok := err.(*DraftNotFoundError)
	return ok
}

// IsAutomationError 检查是否为自动化错误
func IsAutomationError(err error) bool {
	_, ok := err.(*AutomationError)
	return ok
}

// IsExportTimeout 检查是否为导出超时错误
func IsExportTimeout(err error) bool {
	_, ok := err.(*ExportTimeoutError)
	return ok
}

// IsValidationError 检查是否为验证错误
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// IsJSONProcessingError 检查是否为JSON处理错误
func IsJSONProcessingError(err error) bool {
	_, ok := err.(*JSONProcessingError)
	return ok
}

// IsTypeConversionError 检查是否为类型转换错误
func IsTypeConversionError(err error) bool {
	_, ok := err.(*TypeConversionError)
	return ok
}

// IsConfigurationError 检查是否为配置错误
func IsConfigurationError(err error) bool {
	_, ok := err.(*ConfigurationError)
	return ok
}
