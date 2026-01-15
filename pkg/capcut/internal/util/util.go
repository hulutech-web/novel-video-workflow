// Package util 提供各种辅助工具函数
// 对应Python的 pyJianYingDraft/util.py 和根目录util.py
package util

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// JsonExportable 定义可导出为JSON的类型
// 对应Python的JsonExportable类型别名
type JsonExportable interface{}

// JSONExportable 接口定义可导出JSON的对象
type JSONExportable interface {
	ExportJSON() map[string]interface{}
}

// JSONImportable 接口定义可从JSON导入的对象
type JSONImportable interface {
	ImportJSON(data map[string]interface{}) error
}

// ProvideCtorDefaults 为结构体类型提供默认值
// 对应Python的provide_ctor_defaults函数
func ProvideCtorDefaults(t reflect.Type) (map[string]interface{}, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type must be a struct, got %s", t.Kind())
	}

	defaults := make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 跳过不可导出的字段
		if !field.IsExported() {
			continue
		}

		fieldName := field.Name
		fieldType := field.Type

		// 根据字段类型提供默认值
		switch fieldType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			defaults[fieldName] = 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			defaults[fieldName] = uint64(0)
		case reflect.Float32, reflect.Float64:
			defaults[fieldName] = 0.0
		case reflect.String:
			defaults[fieldName] = ""
		case reflect.Bool:
			defaults[fieldName] = false
		case reflect.Slice:
			defaults[fieldName] = reflect.MakeSlice(fieldType, 0, 0).Interface()
		case reflect.Map:
			defaults[fieldName] = reflect.MakeMap(fieldType).Interface()
		case reflect.Ptr:
			defaults[fieldName] = nil
		default:
			// 对于复杂类型，尝试创建零值
			if fieldType.Kind() == reflect.Struct {
				defaults[fieldName] = reflect.Zero(fieldType).Interface()
			} else {
				defaults[fieldName] = nil
			}
		}
	}

	return defaults, nil
}

// AssignAttrWithJSON 根据JSON数据为对象属性赋值
// 对应Python的assign_attr_with_json函数
func AssignAttrWithJSON(obj interface{}, attrs []string, jsonData map[string]interface{}) error {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr || objValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("obj must be a pointer to struct")
	}

	objValue = objValue.Elem()

	for _, attr := range attrs {
		// 查找字段
		fieldValue := objValue.FieldByName(attr)
		if !fieldValue.IsValid() {
			continue // 跳过不存在的字段
		}

		if !fieldValue.CanSet() {
			return fmt.Errorf("field %s cannot be set", attr)
		}

		// 获取JSON数据中的值
		jsonValue, exists := jsonData[attr]
		if !exists {
			continue // 跳过JSON中不存在的字段
		}

		// 获取字段类型
		fieldType := fieldValue.Type()

		// 尝试设置值
		if err := setFieldValue(fieldValue, fieldType, jsonValue); err != nil {
			return fmt.Errorf("failed to set field %s: %v", attr, err)
		}
	}

	return nil
}

// setFieldValue 设置字段值，处理类型转换
func setFieldValue(fieldValue reflect.Value, fieldType reflect.Type, jsonValue interface{}) error {
	// 如果字段实现了JSONImportable接口
	if fieldValue.CanInterface() {
		if importable, ok := fieldValue.Interface().(JSONImportable); ok {
			if jsonMap, ok := jsonValue.(map[string]interface{}); ok {
				return importable.ImportJSON(jsonMap)
			}
		}
	}

	// 跳过Go的直接类型转换，使用我们自己的转换逻辑
	// 这样可以避免int到string的直接转换产生字符而不是字符串

	// 处理特殊情况
	switch fieldType.Kind() {
	case reflect.String:
		if val, ok := convertToString(jsonValue); ok {
			fieldValue.SetString(val)
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, ok := convertToInt(jsonValue); ok {
			fieldValue.SetInt(val)
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val, ok := convertToUint(jsonValue); ok {
			fieldValue.SetUint(val)
			return nil
		}
	case reflect.Float32, reflect.Float64:
		if val, ok := convertToFloat(jsonValue); ok {
			fieldValue.SetFloat(val)
			return nil
		}
	case reflect.Bool:
		if val, ok := convertToBool(jsonValue); ok {
			fieldValue.SetBool(val)
			return nil
		}
	case reflect.Map:
		// 处理map类型
		jsonValueReflect := reflect.ValueOf(jsonValue)
		if jsonValueReflect.Type().AssignableTo(fieldType) {
			fieldValue.Set(jsonValueReflect)
			return nil
		}
	case reflect.Slice:
		// 处理slice类型
		jsonValueReflect := reflect.ValueOf(jsonValue)
		if jsonValueReflect.Type().AssignableTo(fieldType) {
			fieldValue.Set(jsonValueReflect)
			return nil
		}
	}

	return fmt.Errorf("cannot convert %T to %s", jsonValue, fieldType)
}

// ExportAttrToJSON 将对象属性导出为JSON数据
// 对应Python的export_attr_to_json函数
func ExportAttrToJSON(obj interface{}, attrs []string) (map[string]interface{}, error) {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}

	if objValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("obj must be a struct or pointer to struct")
	}

	jsonData := make(map[string]interface{})

	for _, attr := range attrs {
		fieldValue := objValue.FieldByName(attr)
		if !fieldValue.IsValid() {
			continue // 跳过不存在的字段
		}

		if !fieldValue.CanInterface() {
			continue // 跳过不可访问的字段
		}

		fieldInterface := fieldValue.Interface()

		// 如果字段实现了JSONExportable接口
		if exportable, ok := fieldInterface.(JSONExportable); ok {
			jsonData[attr] = exportable.ExportJSON()
		} else {
			// 直接使用字段值
			jsonData[attr] = fieldInterface
		}
	}

	return jsonData, nil
}

// HexToRGB 将十六进制颜色代码转换为RGB元组 (范围0.0-1.0)
// 对应根目录util.py的hex_to_rgb函数
func HexToRGB(hexColor string) (r, g, b float64, err error) {
	// 移除#前缀
	hexColor = strings.TrimPrefix(hexColor, "#")

	// 处理简写形式 (例如 #fff)
	if len(hexColor) == 3 {
		hexColor = string(hexColor[0]) + string(hexColor[0]) +
			string(hexColor[1]) + string(hexColor[1]) +
			string(hexColor[2]) + string(hexColor[2])
	}

	if len(hexColor) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hexadecimal color code: %s", hexColor)
	}

	// 解析RGB分量
	rInt, err := strconv.ParseInt(hexColor[0:2], 16, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hexadecimal color code: %s", hexColor)
	}

	gInt, err := strconv.ParseInt(hexColor[2:4], 16, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hexadecimal color code: %s", hexColor)
	}

	bInt, err := strconv.ParseInt(hexColor[4:6], 16, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hexadecimal color code: %s", hexColor)
	}

	// 转换为0.0-1.0范围
	r = float64(rInt) / 255.0
	g = float64(gInt) / 255.0
	b = float64(bInt) / 255.0

	return r, g, b, nil
}

// IsWindowsPath 检测路径是否为Windows风格
// 对应根目录util.py的is_windows_path函数
func IsWindowsPath(path string) bool {
	// 检查是否以驱动器字母开头 (例如 C:\) 或包含Windows风格分隔符
	matched, _ := regexp.MatchString(`^[a-zA-Z]:\\|\\\\`, path)
	return matched
}

// URLToHash 将URL转换为固定长度的哈希字符串
// 对应根目录util.py的url_to_hash函数
func URLToHash(url string, length int) string {
	if length <= 0 || length > 64 {
		length = 16 // 默认长度
	}

	// 使用SHA-256生成哈希
	hash := sha256.Sum256([]byte(url))
	hashStr := fmt.Sprintf("%x", hash)

	// 截取到指定长度
	if length > len(hashStr) {
		return hashStr
	}
	return hashStr[:length]
}

// 类型转换辅助函数

func convertToInt(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i, true
		}
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i, true
		}
	}
	return 0, false
}

func convertToUint(value interface{}) (uint64, bool) {
	switch v := value.(type) {
	case uint:
		return uint64(v), true
	case uint8:
		return uint64(v), true
	case uint16:
		return uint64(v), true
	case uint32:
		return uint64(v), true
	case uint64:
		return v, true
	case int:
		if v >= 0 {
			return uint64(v), true
		}
	case int8:
		if v >= 0 {
			return uint64(v), true
		}
	case int16:
		if v >= 0 {
			return uint64(v), true
		}
	case int32:
		if v >= 0 {
			return uint64(v), true
		}
	case int64:
		if v >= 0 {
			return uint64(v), true
		}
	case float32:
		if v >= 0 {
			return uint64(v), true
		}
	case float64:
		if v >= 0 {
			return uint64(v), true
		}
	case string:
		if u, err := strconv.ParseUint(v, 10, 64); err == nil {
			return u, true
		}
	case json.Number:
		if f, err := v.Float64(); err == nil && f >= 0 {
			return uint64(f), true
		}
	}
	return 0, false
}

func convertToFloat(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f, true
		}
	}
	return 0, false
}

func convertToString(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v), true
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), true
	case float32, float64:
		return fmt.Sprintf("%g", v), true
	case bool:
		return fmt.Sprintf("%t", v), true
	default:
		return fmt.Sprintf("%v", v), true
	}
}

func convertToBool(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		if b, err := strconv.ParseBool(v); err == nil {
			return b, true
		}
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0, true
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() != 0, true
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0, true
	}
	return false, false
}
