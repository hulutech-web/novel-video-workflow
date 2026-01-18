package database

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// GetAppDataPath 获取应用数据存储路径
func GetAppDataPath(appName string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %v", err)
	}

	var appDataPath string
	switch runtime.GOOS {
	case "windows":
		appDataPath = filepath.Join(usr.HomeDir, "AppData", "Local", appName)
	case "darwin": // macOS
		appDataPath = filepath.Join(usr.HomeDir, "Library", "Application Support", appName)
	case "linux":
		appDataPath = filepath.Join(usr.HomeDir, ".local", "share", appName)
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// 确保目录存在
	if err := os.MkdirAll(appDataPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create app data directory: %v", err)
	}

	return appDataPath, nil
}

// GetDatabasePath 获取数据库文件路径
func GetDatabasePath() (string, error) {
	appDataPath, err := GetAppDataPath("novel-video-workflow")
	if err != nil {
		return "", err
	}

	dbPath := filepath.Join(appDataPath, "database.sqlite")
	return dbPath, nil
}