package database

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// ProcessStatus 表示处理状态
type ProcessStatus string

const (
	StatusPending    ProcessStatus = "pending"    // 待处理
	StatusProcessing ProcessStatus = "processing" // 处理中
	StatusCompleted  ProcessStatus = "completed"  // 已完成
	StatusFailed     ProcessStatus = "failed"     // 失败
	StatusSkipped    ProcessStatus = "skipped"    // 跳过
)

// MyTime 自定义时间类型，用于处理时间戳
type MyTime struct {
	time.Time
}

// GormDataType GORM数据类型
func (MyTime) GormDataType() string {
	return "timestamp"
}

// Scan 实现scanner接口
func (t *MyTime) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	switch v := value.(type) {
	case time.Time:
		t.Time = v
	case string:
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			t.Time = parsedTime
		} else {
			return fmt.Errorf("can't parse %s to MyTime", v)
		}
	case []byte:
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", string(v)); err == nil {
			t.Time = parsedTime
		} else {
			return fmt.Errorf("can't parse %s to MyTime", string(v))
		}
	default:
		return fmt.Errorf("can't parse %v to MyTime", value)
	}
	return nil
}

// Value 实现valuer接口
func (t MyTime) Value() (driver.Value, error) {
	return t.Time, nil
}

// MarshalJSON 实现json序列化
func (t MyTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, t.Time.Format("2006-01-02 15:04:05"))), nil
}

// UnmarshalJSON 实现json反序列化
func (t *MyTime) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")
	if str == "null" {
		return nil
	}
	
	parsedTime, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		return err
	}
	t.Time = parsedTime
	return nil
}