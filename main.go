package main

import (
	"context"
	"novel-video-workflow/internal/mcp"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"novel-video-workflow/internal/workflow"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// 1. 初始化日志（第一个操作，用于记录）
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 2. 加载配置文件 - 首先尝试当前工作目录，然后尝试可执行文件目录
	var configPath string
	var err error

	// 尝试在当前工作目录查找配置文件
	wd, _ := os.Getwd()
	configPath = filepath.Join(wd, "config.yaml")

	// 如果当前工作目录没有配置文件，尝试可执行文件所在目录
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		exe, exeErr := os.Executable()
		if exeErr != nil {
			logger.Fatal("无法获取可执行文件路径", zap.Error(exeErr))
		}
		exeDir := filepath.Dir(exe)
		configPath = filepath.Join(exeDir, "config.yaml")
	}

	viper.SetConfigFile(configPath) // 关键：明确指定文件
	if err := viper.ReadInConfig(); err != nil {
		// 使用logger输出到stderr，而不是log.Fatalf或fmt.Printf
		logger.Fatal("读取配置文件失败",
			zap.String("configPath", configPath),
			zap.Error(err),
		)
	}
	// 重要：不要向stdout打印任何内容！使用logger记录到stderr。
	logger.Info("配置文件加载成功", zap.String("path", configPath))

	// 3. 创建工作流处理器和MCP服务器
	processor, err := workflow.NewProcessor(logger)
	if err != nil {
		logger.Fatal("创建工作流处理器失败", zap.Error(err))
	}

	mcpServer, err := mcp.NewServer(processor, logger)
	if err != nil {
		logger.Fatal("创建MCP服务器失败", zap.Error(err))
	}

	// 4. 启动服务器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := mcpServer.Start(ctx); err != nil {
			logger.Fatal("MCP服务器启动失败", zap.Error(err))
		}
	}()

	// 5. 等待退出信号
	logger.Info("小说视频MCP服务器已启动，等待连接...",
		zap.Strings("tools", mcpServer.GetToolNames()))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("正在关闭服务器...")
	cancel()
}
