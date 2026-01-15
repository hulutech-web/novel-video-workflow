// Package draft 定义草稿信息相关结构
package draft

import (
	"fmt"
	"time"
)

// DraftInfo 草稿基本信息
// 扩展结构，提供草稿的详细信息
type DraftInfo struct {
	Name         string    `json:"name"`           // 草稿名称
	Path         string    `json:"path"`           // 草稿完整路径
	ModTime      time.Time `json:"mod_time"`       // 最后修改时间
	HasDraftInfo bool      `json:"has_draft_info"` // 是否包含draft_info.json文件
}

// IsValid 检查草稿是否有效
// 有效的草稿应该包含draft_info.json文件
func (di *DraftInfo) IsValid() bool {
	return di.HasDraftInfo
}

// Age 返回草稿的年龄（距离最后修改的时间）
func (di *DraftInfo) Age() time.Duration {
	return time.Since(di.ModTime)
}

// String 返回草稿信息的字符串表示
func (di *DraftInfo) String() string {
	validStatus := "无效"
	if di.IsValid() {
		validStatus = "有效"
	}

	return fmt.Sprintf("草稿: %s (%s) - 最后修改: %s",
		di.Name,
		validStatus,
		di.ModTime.Format("2006-01-02 15:04:05"))
}
