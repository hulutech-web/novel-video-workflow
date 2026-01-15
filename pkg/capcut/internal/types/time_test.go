package types

import (
	"testing"
)

func TestTim(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int64
		hasError bool
	}{
		// 数值输入
		{"int input", 5000000, 5000000, false},
		{"int64 input", int64(5000000), 5000000, false},
		{"float64 input", 5.5, 6, false}, // 四舍五入

		// 字符串时间输入
		{"seconds only", "5s", 5 * SEC, false},
		{"minutes only", "2m", 2 * 60 * SEC, false},
		{"hours only", "1h", 1 * 3600 * SEC, false},
		{"decimal seconds", "1.5s", int64(1.5 * SEC), false},
		{"complex time", "1h30m45s", 1*3600*SEC + 30*60*SEC + 45*SEC, false},
		{"negative time", "-30s", -30 * SEC, false},
		{"negative complex", "-1h30m", -(1*3600*SEC + 30*60*SEC), false},

		// 错误情况
		{"invalid type", []int{1, 2, 3}, 0, true},
		{"invalid string", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Tim(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestTimerange(t *testing.T) {
	// 基本创建和属性测试
	tr := NewTimerange(1000000, 2000000)
	if tr.Start != 1000000 {
		t.Errorf("Expected start 1000000, got %d", tr.Start)
	}
	if tr.Duration != 2000000 {
		t.Errorf("Expected duration 2000000, got %d", tr.Duration)
	}
	if tr.End() != 3000000 {
		t.Errorf("Expected end 3000000, got %d", tr.End())
	}
}

func TestTimerangeEquals(t *testing.T) {
	tr1 := NewTimerange(1000000, 2000000)
	tr2 := NewTimerange(1000000, 2000000)
	tr3 := NewTimerange(1000000, 3000000)

	if !tr1.Equals(tr2) {
		t.Error("Expected tr1 equals tr2")
	}
	if tr1.Equals(tr3) {
		t.Error("Expected tr1 not equals tr3")
	}
	if tr1.Equals(nil) {
		t.Error("Expected tr1 not equals nil")
	}
}

func TestTimerangeOverlaps(t *testing.T) {
	tests := []struct {
		name     string
		tr1      *Timerange
		tr2      *Timerange
		expected bool
	}{
		{
			"overlapping ranges",
			NewTimerange(1000000, 2000000), // [1s, 3s]
			NewTimerange(2000000, 2000000), // [2s, 4s]
			true,
		},
		{
			"non-overlapping ranges",
			NewTimerange(1000000, 1000000), // [1s, 2s]
			NewTimerange(2000000, 1000000), // [2s, 3s]
			false,
		},
		{
			"touching ranges",
			NewTimerange(1000000, 1000000), // [1s, 2s]
			NewTimerange(2000000, 1000000), // [2s, 3s]
			false,
		},
		{
			"contained range",
			NewTimerange(1000000, 4000000), // [1s, 5s]
			NewTimerange(2000000, 1000000), // [2s, 3s]
			true,
		},
		{
			"nil range",
			NewTimerange(1000000, 2000000),
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tr1.Overlaps(tt.tr2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTrange(t *testing.T) {
	tests := []struct {
		name     string
		start    interface{}
		duration interface{}
		expected *Timerange
		hasError bool
	}{
		{
			"string inputs",
			"1s", "2s",
			NewTimerange(1*SEC, 2*SEC),
			false,
		},
		{
			"mixed inputs",
			1 * SEC, "2s",
			NewTimerange(1*SEC, 2*SEC),
			false,
		},
		{
			"complex time",
			"1h30m", "45s",
			NewTimerange(1*3600*SEC+30*60*SEC, 45*SEC),
			false,
		},
		{
			"invalid start",
			"abc", "1s",
			nil,
			true,
		},
		{
			"invalid duration",
			"1s", "xyz",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Trange(tt.start, tt.duration)
			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !result.Equals(tt.expected) {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestMustTrange(t *testing.T) {
	// 正常情况
	tr := MustTrange("1s", "2s")
	expected := NewTimerange(1*SEC, 2*SEC)
	if !tr.Equals(expected) {
		t.Errorf("Expected %v, got %v", expected, tr)
	}

	// 测试panic情况
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic but got none")
		}
	}()
	MustTrange("invalid", "1s")
}

func TestSrtTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		hasError bool
	}{
		{
			"normal timestamp",
			"01:23:45,678",
			1*3600*SEC + 23*60*SEC + 45*SEC + 678*1000,
			false,
		},
		{
			"zero timestamp",
			"00:00:00,000",
			0,
			false,
		},
		{
			"single digits",
			"1:2:3,4",
			1*3600*SEC + 2*60*SEC + 3*SEC + 4*1000,
			false,
		},
		{
			"invalid format - no comma",
			"01:23:45.678",
			0,
			true,
		},
		{
			"invalid format - too many parts",
			"01:23:45:67,678",
			0,
			true,
		},
		{
			"invalid hours",
			"aa:23:45,678",
			0,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SrtTimestamp(tt.input)
			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"zero", 0, "0s"},
		{"seconds only", 5 * SEC, "5s"},
		{"decimal seconds", int64(1.5 * SEC), "1.500s"},
		{"minutes and seconds", 65 * SEC, "1m5s"},
		{"hours, minutes, seconds", 3661 * SEC, "1h1m1s"},
		{"negative time", -30 * SEC, "-30s"},
		{"complex negative", -(1*3600*SEC + 30*60*SEC), "-1h30m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestMicrosecondsToSeconds(t *testing.T) {
	tests := []struct {
		input    int64
		expected float64
	}{
		{SEC, 1.0},
		{int64(1.5 * SEC), 1.5},
		{0, 0.0},
		{-SEC, -1.0},
	}

	for _, tt := range tests {
		result := MicrosecondsToSeconds(tt.input)
		if result != tt.expected {
			t.Errorf("Input %d: expected %f, got %f", tt.input, tt.expected, result)
		}
	}
}

func TestSecondsToMicroseconds(t *testing.T) {
	tests := []struct {
		input    float64
		expected int64
	}{
		{1.0, SEC},
		{1.5, int64(1.5 * SEC)},
		{0.0, 0},
		{-1.0, -SEC},
		{1.0000005, int64(1.0000005*SEC - 0.5)}, // 测试四舍五入
	}

	for _, tt := range tests {
		result := SecondsToMicroseconds(tt.input)
		if result != tt.expected {
			t.Errorf("Input %f: expected %d, got %d", tt.input, tt.expected, result)
		}
	}
}

func TestTimerangeJSON(t *testing.T) {
	tr := NewTimerange(1000000, 2000000)

	// 测试导出JSON
	jsonData := tr.ExportJSON()
	expectedStart := int64(1000000)
	expectedDuration := int64(2000000)

	if jsonData["start"] != expectedStart {
		t.Errorf("Expected start %d, got %d", expectedStart, jsonData["start"])
	}
	if jsonData["duration"] != expectedDuration {
		t.Errorf("Expected duration %d, got %d", expectedDuration, jsonData["duration"])
	}

	// 测试从JSON导入
	tr2 := &Timerange{}
	importData := map[string]interface{}{
		"start":    float64(1000000),
		"duration": float64(2000000),
	}

	err := tr2.ImportFromJSON(importData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !tr.Equals(tr2) {
		t.Errorf("Expected %v, got %v", tr, tr2)
	}
}
