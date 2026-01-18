package database

import (
	"fmt"
)

// InitDatabase 初始化数据库连接和表结构
func InitDatabase() error {
	// 创建数据库管理器，这将自动执行迁移
	_, err := NewGormManager()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	return nil
}