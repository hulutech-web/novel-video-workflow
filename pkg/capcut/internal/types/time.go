// Package types 通用类型定义
// 定义时间范围类以及与时间相关的辅助函数
// 对应Python的 time_util.py
package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SEC 一秒=1e6微秒
const SEC = 1000000

// Tim 将输入的字符串转换为微秒，也可直接输入微秒数
// 支持类似 "1h52m3s" 或 "0.15s" 这样的格式，可包含负号以表示负偏移
// 对应Python的tim函数
func Tim(inp interface{}) (int64, error) {
	switch v := inp.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v + 0.5), nil // 四舍五入
	case string:
		return parseTimeString(v)
	default:
		return 0, fmt.Errorf("unsupported type: %T", inp)
	}
}

// parseTimeString 解析时间字符串
func parseTimeString(inp string) (int64, error) {
	sign := int64(1)
	inp = strings.TrimSpace(strings.ToLower(inp))

	if strings.HasPrefix(inp, "-") {
		sign = -1
		inp = inp[1:]
	}

	// 检查是否包含任何有效的时间单位
	if !strings.Contains(inp, "h") && !strings.Contains(inp, "m") && !strings.Contains(inp, "s") {
		return 0, fmt.Errorf("invalid time format: %s", inp)
	}

	var totalTime float64 = 0

	// 定义时间单位和对应的微秒数
	units := []struct {
		unit   string
		factor float64
	}{
		{"h", 3600 * SEC},
		{"m", 60 * SEC},
		{"s", SEC},
	}

	lastIndex := 0
	for _, u := range units {
		unitIndex := strings.Index(inp, u.unit)
		if unitIndex == -1 {
			continue
		}

		valueStr := inp[lastIndex:unitIndex]
		if valueStr == "" {
			continue
		}

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid time value: %s", valueStr)
		}

		totalTime += value * u.factor
		lastIndex = unitIndex + 1
	}

	return int64(totalTime+0.5) * sign, nil // 四舍五入
}

// Timerange 记录了起始时间及持续长度的时间范围
// 对应Python的Timerange类
type Timerange struct {
	Start    int64 `json:"start"`    // 起始时间，单位为微秒
	Duration int64 `json:"duration"` // 持续长度，单位为微秒
}

// NewTimerange 构造一个时间范围
func NewTimerange(start, duration int64) *Timerange {
	return &Timerange{
		Start:    start,
		Duration: duration,
	}
}

// ImportFromJSON 从JSON对象中恢复Timerange
func (tr *Timerange) ImportFromJSON(jsonData map[string]interface{}) error {
	if start, ok := jsonData["start"]; ok {
		switch v := start.(type) {
		case float64:
			tr.Start = int64(v)
		case string:
			s, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid start value: %s", v)
			}
			tr.Start = s
		default:
			return fmt.Errorf("invalid start type: %T", v)
		}
	}

	if duration, ok := jsonData["duration"]; ok {
		switch v := duration.(type) {
		case float64:
			tr.Duration = int64(v)
		case string:
			d, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid duration value: %s", v)
			}
			tr.Duration = d
		default:
			return fmt.Errorf("invalid duration type: %T", v)
		}
	}

	return nil
}

// End 结束时间，单位为微秒
func (tr *Timerange) End() int64 {
	return tr.Start + tr.Duration
}

// Equals 判断两个时间范围是否相等
func (tr *Timerange) Equals(other *Timerange) bool {
	if other == nil {
		return false
	}
	return tr.Start == other.Start && tr.Duration == other.Duration
}

// Overlaps 判断两个时间范围是否有重叠
func (tr *Timerange) Overlaps(other *Timerange) bool {
	if other == nil {
		return false
	}
	return !(tr.End() <= other.Start || other.End() <= tr.Start)
}

// String 返回字符串表示
func (tr *Timerange) String() string {
	return fmt.Sprintf("[start=%d, end=%d]", tr.Start, tr.End())
}

// GoString 返回Go语法的字符串表示
func (tr *Timerange) GoString() string {
	return fmt.Sprintf("Timerange{Start: %d, Duration: %d}", tr.Start, tr.Duration)
}

// ExportJSON 导出为JSON格式
func (tr *Timerange) ExportJSON() map[string]int64 {
	return map[string]int64{
		"start":    tr.Start,
		"duration": tr.Duration,
	}
}

// Trange Timerange的简便构造函数，接受字符串或微秒数作为参数
// 支持类似 "1h52m3s" 或 "0.15s" 这样的格式
// 对应Python的trange函数
func Trange(start, duration interface{}) (*Timerange, error) {
	startMicros, err := Tim(start)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %w", err)
	}

	durationMicros, err := Tim(duration)
	if err != nil {
		return nil, fmt.Errorf("invalid duration: %w", err)
	}

	return NewTimerange(startMicros, durationMicros), nil
}

// MustTrange Trange的不返回错误版本，遇到错误会panic
// 主要用于测试和已知正确的场景
func MustTrange(start, duration interface{}) *Timerange {
	tr, err := Trange(start, duration)
	if err != nil {
		panic(err)
	}
	return tr
}

// SrtTimestamp 解析SRT中的时间戳字符串，返回微秒数
// 格式: "01:23:45,678"
// 对应Python的srt_tstamp函数
func SrtTimestamp(srtTimestamp string) (int64, error) {
	// 使用正则表达式解析SRT时间戳格式: HH:MM:SS,mmm
	re := regexp.MustCompile(`^(\d{1,2}):(\d{1,2}):(\d{1,2}),(\d{1,3})$`)
	matches := re.FindStringSubmatch(srtTimestamp)

	if len(matches) != 5 {
		return 0, fmt.Errorf("invalid SRT timestamp format: %s", srtTimestamp)
	}

	hours, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid hours: %s", matches[1])
	}

	minutes, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes: %s", matches[2])
	}

	seconds, err := strconv.Atoi(matches[3])
	if err != nil {
		return 0, fmt.Errorf("invalid seconds: %s", matches[3])
	}

	milliseconds, err := strconv.Atoi(matches[4])
	if err != nil {
		return 0, fmt.Errorf("invalid milliseconds: %s", matches[4])
	}

	// 转换为微秒
	totalMicros := int64(hours)*3600*SEC + int64(minutes)*60*SEC + int64(seconds)*SEC + int64(milliseconds)*1000

	return totalMicros, nil
}

// FormatDuration 将微秒格式化为可读的时间字符串
// 例如: 3661000000 -> "1h1m1s"
func FormatDuration(micros int64) string {
	if micros == 0 {
		return "0s"
	}

	sign := ""
	if micros < 0 {
		sign = "-"
		micros = -micros
	}

	hours := micros / (3600 * SEC)
	micros %= (3600 * SEC)

	minutes := micros / (60 * SEC)
	micros %= (60 * SEC)

	seconds := float64(micros) / SEC

	var parts []string
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 {
		if seconds == float64(int64(seconds)) {
			parts = append(parts, fmt.Sprintf("%.0fs", seconds))
		} else {
			parts = append(parts, fmt.Sprintf("%.3fs", seconds))
		}
	}

	if len(parts) == 0 {
		return "0s"
	}

	return sign + strings.Join(parts, "")
}

// MicrosecondsToSeconds 将微秒转换为秒（浮点数）
func MicrosecondsToSeconds(micros int64) float64 {
	return float64(micros) / SEC
}

// SecondsToMicroseconds 将秒（浮点数）转换为微秒
func SecondsToMicroseconds(seconds float64) int64 {
	if seconds >= 0 {
		return int64(seconds*SEC + 0.5) // 四舍五入
	} else {
		return int64(seconds*SEC - 0.5) // 负数的四舍五入
	}
}
