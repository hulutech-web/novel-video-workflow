package util

import (
	"reflect"
	"testing"
)

// TestProvideCtorDefaults 测试构造函数默认值提供
func TestProvideCtorDefaults(t *testing.T) {
	// 定义测试结构体
	type TestStruct struct {
		IntField    int
		StringField string
		FloatField  float64
		BoolField   bool
		SliceField  []string
		MapField    map[string]int
	}

	structType := reflect.TypeOf(TestStruct{})
	defaults, err := ProvideCtorDefaults(structType)
	if err != nil {
		t.Fatalf("ProvideCtorDefaults failed: %v", err)
	}

	// 验证默认值
	expectedDefaults := map[string]interface{}{
		"IntField":    0,
		"StringField": "",
		"FloatField":  0.0,
		"BoolField":   false,
		"SliceField":  []string{},
		"MapField":    map[string]int{},
	}

	if len(defaults) != len(expectedDefaults) {
		t.Errorf("期望%d个默认值，得到%d个", len(expectedDefaults), len(defaults))
	}

	for key, expectedValue := range expectedDefaults {
		actualValue, exists := defaults[key]
		if !exists {
			t.Errorf("缺少字段%s的默认值", key)
			continue
		}

		// 对于slice和map，需要特殊比较
		if reflect.TypeOf(expectedValue).Kind() == reflect.Slice {
			if reflect.ValueOf(actualValue).Len() != 0 {
				t.Errorf("字段%s期望空slice，得到%v", key, actualValue)
			}
		} else if reflect.TypeOf(expectedValue).Kind() == reflect.Map {
			if reflect.ValueOf(actualValue).Len() != 0 {
				t.Errorf("字段%s期望空map，得到%v", key, actualValue)
			}
		} else if !reflect.DeepEqual(actualValue, expectedValue) {
			t.Errorf("字段%s期望%v，得到%v", key, expectedValue, actualValue)
		}
	}
}

// TestProvideCtorDefaultsNonStruct 测试非结构体类型
func TestProvideCtorDefaultsNonStruct(t *testing.T) {
	intType := reflect.TypeOf(42)
	_, err := ProvideCtorDefaults(intType)
	if err == nil {
		t.Error("期望非结构体类型返回错误")
	}
}

// TestAssignAttrWithJSON 测试JSON属性赋值
func TestAssignAttrWithJSON(t *testing.T) {
	// 定义测试结构体
	type TestStruct struct {
		Name  string
		Age   int
		Score float64
		Valid bool
	}

	obj := &TestStruct{}
	attrs := []string{"Name", "Age", "Score", "Valid"}
	jsonData := map[string]interface{}{
		"Name":  "测试用户",
		"Age":   25,
		"Score": 95.5,
		"Valid": true,
	}

	err := AssignAttrWithJSON(obj, attrs, jsonData)
	if err != nil {
		t.Fatalf("AssignAttrWithJSON failed: %v", err)
	}

	// 验证赋值结果
	if obj.Name != "测试用户" {
		t.Errorf("Name字段期望'测试用户'，得到'%s'", obj.Name)
	}
	if obj.Age != 25 {
		t.Errorf("Age字段期望25，得到%d", obj.Age)
	}
	if obj.Score != 95.5 {
		t.Errorf("Score字段期望95.5，得到%.1f", obj.Score)
	}
	if obj.Valid != true {
		t.Errorf("Valid字段期望true，得到%v", obj.Valid)
	}
}

// TestAssignAttrWithJSONTypeConversion 测试类型转换
func TestAssignAttrWithJSONTypeConversion(t *testing.T) {
	type TestStruct struct {
		IntFromFloat   int
		FloatFromInt   float64
		StringFromInt  string
		BoolFromString bool
	}

	obj := &TestStruct{}
	attrs := []string{"IntFromFloat", "FloatFromInt", "StringFromInt", "BoolFromString"}
	jsonData := map[string]interface{}{
		"IntFromFloat":   42.7,   // float64 -> int
		"FloatFromInt":   100,    // int -> float64
		"StringFromInt":  123,    // int -> string
		"BoolFromString": "true", // string -> bool
	}

	err := AssignAttrWithJSON(obj, attrs, jsonData)
	if err != nil {
		t.Fatalf("AssignAttrWithJSON with type conversion failed: %v", err)
	}

	// 验证类型转换结果
	if obj.IntFromFloat != 42 {
		t.Errorf("IntFromFloat期望42，得到%d", obj.IntFromFloat)
	}
	if obj.FloatFromInt != 100.0 {
		t.Errorf("FloatFromInt期望100.0，得到%.1f", obj.FloatFromInt)
	}
	if obj.StringFromInt != "123" {
		t.Errorf("StringFromInt期望'123'，得到'%s'", obj.StringFromInt)
	}
	if obj.BoolFromString != true {
		t.Errorf("BoolFromString期望true，得到%v", obj.BoolFromString)
	}
}

// TestExportAttrToJSON 测试属性JSON导出
func TestExportAttrToJSON(t *testing.T) {
	type TestStruct struct {
		Name  string
		Age   int
		Score float64
		Valid bool
	}

	obj := TestStruct{
		Name:  "导出测试",
		Age:   30,
		Score: 88.5,
		Valid: false,
	}

	attrs := []string{"Name", "Age", "Score", "Valid"}
	jsonData, err := ExportAttrToJSON(obj, attrs)
	if err != nil {
		t.Fatalf("ExportAttrToJSON failed: %v", err)
	}

	// 验证导出结果
	expectedData := map[string]interface{}{
		"Name":  "导出测试",
		"Age":   30,
		"Score": 88.5,
		"Valid": false,
	}

	if len(jsonData) != len(expectedData) {
		t.Errorf("期望%d个字段，得到%d个", len(expectedData), len(jsonData))
	}

	for key, expectedValue := range expectedData {
		actualValue, exists := jsonData[key]
		if !exists {
			t.Errorf("缺少字段%s", key)
			continue
		}

		if !reflect.DeepEqual(actualValue, expectedValue) {
			t.Errorf("字段%s期望%v，得到%v", key, expectedValue, actualValue)
		}
	}
}

// TestHexToRGB 测试十六进制颜色转RGB
func TestHexToRGB(t *testing.T) {
	testCases := []struct {
		hex      string
		r, g, b  float64
		hasError bool
	}{
		{"#FF0000", 1.0, 0.0, 0.0, false}, // 红色
		{"#00FF00", 0.0, 1.0, 0.0, false}, // 绿色
		{"#0000FF", 0.0, 0.0, 1.0, false}, // 蓝色
		{"#FFFFFF", 1.0, 1.0, 1.0, false}, // 白色
		{"#000000", 0.0, 0.0, 0.0, false}, // 黑色
		{"FF0000", 1.0, 0.0, 0.0, false},  // 无#前缀
		{"#F00", 1.0, 0.0, 0.0, false},    // 简写形式
		{"#80", 0.0, 0.0, 0.0, true},      // 无效长度
		{"#GGGGGG", 0.0, 0.0, 0.0, true},  // 无效字符
	}

	for _, tc := range testCases {
		r, g, b, err := HexToRGB(tc.hex)

		if tc.hasError {
			if err == nil {
				t.Errorf("颜色%s期望返回错误，但没有错误", tc.hex)
			}
			continue
		}

		if err != nil {
			t.Errorf("颜色%s不期望错误，但得到错误: %v", tc.hex, err)
			continue
		}

		// 使用容差比较浮点数
		tolerance := 0.001
		if abs(r-tc.r) > tolerance || abs(g-tc.g) > tolerance || abs(b-tc.b) > tolerance {
			t.Errorf("颜色%s期望RGB(%.3f, %.3f, %.3f)，得到RGB(%.3f, %.3f, %.3f)",
				tc.hex, tc.r, tc.g, tc.b, r, g, b)
		}
	}
}

// abs 计算浮点数绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// TestIsWindowsPath 测试Windows路径检测
func TestIsWindowsPath(t *testing.T) {
	testCases := []struct {
		path      string
		isWindows bool
	}{
		{"C:\\Users\\test", true},
		{"D:\\Program Files", true},
		{"\\\\server\\share", true},
		{"/usr/local/bin", false},
		{"./relative/path", false},
		{"../parent/path", false},
		{"file.txt", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := IsWindowsPath(tc.path)
		if result != tc.isWindows {
			t.Errorf("路径'%s'期望Windows检测结果为%v，得到%v", tc.path, tc.isWindows, result)
		}
	}
}

// TestURLToHash 测试URL哈希转换
func TestURLToHash(t *testing.T) {
	testCases := []struct {
		url    string
		length int
		name   string
	}{
		{"https://example.com", 16, "默认长度"},
		{"https://example.com", 8, "短哈希"},
		{"https://example.com", 32, "长哈希"},
		{"https://example.com", 0, "无效长度"},
		{"https://example.com", -1, "负长度"},
		{"https://example.com", 100, "超长长度"},
	}

	for _, tc := range testCases {
		hash := URLToHash(tc.url, tc.length)

		expectedLength := tc.length
		if expectedLength <= 0 || expectedLength > 64 {
			expectedLength = 16 // 默认长度
		}

		if len(hash) != expectedLength {
			t.Errorf("%s: 期望哈希长度为%d，得到%d (哈希: %s)", tc.name, expectedLength, len(hash), hash)
		}

		// 验证哈希只包含十六进制字符
		for _, char := range hash {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
				t.Errorf("%s: 哈希包含无效字符'%c'", tc.name, char)
				break
			}
		}
	}

	// 测试相同URL产生相同哈希
	url := "https://test.example.com"
	hash1 := URLToHash(url, 16)
	hash2 := URLToHash(url, 16)
	if hash1 != hash2 {
		t.Errorf("相同URL应该产生相同哈希，得到'%s'和'%s'", hash1, hash2)
	}

	// 测试不同URL产生不同哈希
	url1 := "https://example1.com"
	url2 := "https://example2.com"
	hash1 = URLToHash(url1, 16)
	hash2 = URLToHash(url2, 16)
	if hash1 == hash2 {
		t.Errorf("不同URL不应该产生相同哈希，都得到'%s'", hash1)
	}
}

// TestTypeConversionHelpers 测试类型转换辅助函数
func TestTypeConversionHelpers(t *testing.T) {
	// 测试convertToInt
	intTests := []struct {
		input    interface{}
		expected int64
		success  bool
	}{
		{42, 42, true},
		{int8(42), 42, true},
		{int16(42), 42, true},
		{int32(42), 42, true},
		{int64(42), 42, true},
		{uint(42), 42, true},
		{float64(42.7), 42, true},
		{"42", 42, true},
		{"invalid", 0, false},
		{true, 0, false},
	}

	for _, test := range intTests {
		result, success := convertToInt(test.input)
		if success != test.success {
			t.Errorf("convertToInt(%v): 期望成功=%v，得到成功=%v", test.input, test.success, success)
		}
		if success && result != test.expected {
			t.Errorf("convertToInt(%v): 期望%d，得到%d", test.input, test.expected, result)
		}
	}

	// 测试convertToFloat
	floatTests := []struct {
		input    interface{}
		expected float64
		success  bool
	}{
		{42, 42.0, true},
		{float32(42.5), 42.5, true},
		{float64(42.7), 42.7, true},
		{"42.5", 42.5, true},
		{"invalid", 0.0, false},
		{true, 0.0, false},
	}

	for _, test := range floatTests {
		result, success := convertToFloat(test.input)
		if success != test.success {
			t.Errorf("convertToFloat(%v): 期望成功=%v，得到成功=%v", test.input, test.success, success)
		}
		if success && abs(result-test.expected) > 0.001 {
			t.Errorf("convertToFloat(%v): 期望%.3f，得到%.3f", test.input, test.expected, result)
		}
	}

	// 测试convertToString
	stringTests := []struct {
		input    interface{}
		expected string
		success  bool
	}{
		{"hello", "hello", true},
		{[]byte("hello"), "hello", true},
		{42, "42", true},
		{42.5, "42.5", true},
		{true, "true", true},
	}

	for _, test := range stringTests {
		result, success := convertToString(test.input)
		if success != test.success {
			t.Errorf("convertToString(%v): 期望成功=%v，得到成功=%v", test.input, test.success, success)
		}
		if success && result != test.expected {
			t.Errorf("convertToString(%v): 期望'%s'，得到'%s'", test.input, test.expected, result)
		}
	}

	// 测试convertToBool
	boolTests := []struct {
		input    interface{}
		expected bool
		success  bool
	}{
		{true, true, true},
		{false, false, true},
		{"true", true, true},
		{"false", false, true},
		{"1", true, true},
		{"0", false, true},
		{1, true, true},
		{0, false, true},
		{42, true, true},
		{0.0, false, true},
		{1.5, true, true},
		{"invalid", false, false},
	}

	for _, test := range boolTests {
		result, success := convertToBool(test.input)
		if success != test.success {
			t.Errorf("convertToBool(%v): 期望成功=%v，得到成功=%v", test.input, test.success, success)
		}
		if success && result != test.expected {
			t.Errorf("convertToBool(%v): 期望%v，得到%v", test.input, test.expected, result)
		}
	}
}

// TestExportable 实现JSONExportable接口的测试结构体
type TestExportable struct {
	Name  string
	Value int
}

// ExportJSON 实现JSONExportable接口
func (te *TestExportable) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"name":  te.Name,
		"value": te.Value,
	}
}

// TestJSONExportableInterface 测试JSONExportable接口
func TestJSONExportableInterface(t *testing.T) {
	// 定义包含JSONExportable字段的结构体
	type ContainerStruct struct {
		ID         string
		Exportable *TestExportable
	}

	obj := ContainerStruct{
		ID: "container123",
		Exportable: &TestExportable{
			Name:  "测试对象",
			Value: 42,
		},
	}

	attrs := []string{"ID", "Exportable"}
	jsonData, err := ExportAttrToJSON(obj, attrs)
	if err != nil {
		t.Fatalf("ExportAttrToJSON with JSONExportable failed: %v", err)
	}

	// 验证ID字段
	if jsonData["ID"] != "container123" {
		t.Errorf("ID字段期望'container123'，得到'%v'", jsonData["ID"])
	}

	// 验证Exportable字段
	exportableData, ok := jsonData["Exportable"].(map[string]interface{})
	if !ok {
		t.Fatalf("Exportable字段期望为map[string]interface{}，得到%T", jsonData["Exportable"])
	}

	if exportableData["name"] != "测试对象" {
		t.Errorf("Exportable.name期望'测试对象'，得到'%v'", exportableData["name"])
	}

	if exportableData["value"] != 42 {
		t.Errorf("Exportable.value期望42，得到%v", exportableData["value"])
	}
}
