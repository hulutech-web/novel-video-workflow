package main

import (
	"context"
	"fmt"
	"novel-video-workflow/pkg/mcp"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"novel-video-workflow/pkg/workflow"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("启动小说视频工作流系统...")

	// 启动 MCP 服务器
	go runMCPModeBackground()

	// 启动 Web 服务器
	go runWebModeBackground()

	fmt.Println("MCP 服务器和 Web 服务器正在后台运行...")
	fmt.Println("- MCP 服务器: 供 AI 代理和其他客户端调用")
	fmt.Println("- Web 服务器: http://localhost:8080 供用户界面操作")
	fmt.Println("按 Ctrl+C 停止服务")

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n正在关闭服务器...")
}

func runMCPModeBackground() {
	fmt.Println("启动 MCP 服务器模式...")
	
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

	if err := mcpServer.Start(ctx); err != nil {
		logger.Fatal("MCP服务器启动失败", zap.Error(err))
	}
}

func runWebModeBackground() {
	fmt.Println("启动 Web 服务器模式...")
	
	// 等待片刻，确保 MCP 服务器先启动
	time.Sleep(2 * time.Second)
	
	// 直接运行web服务器
	cmd := exec.Command("go", "run", "cmd/web_server/web_server.go")
	cmd.Dir = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("Web服务器运行出错: %v\n", err)
	}
}

// 旧的函数保留作为备用
func runMCPMode() {
	fmt.Println("启动MCP服务器模式...")
	
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

func runWebMode() {
	fmt.Println("启动Web服务器模式...")
	
	// 直接运行web服务器
	cmd := exec.Command("go", "run", "cmd/web_server/web_server.go")
	cmd.Dir = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("启动Web服务器失败: %v\n", err)
		os.Exit(1)
	}
}

func runBatchMode() {
	fmt.Println("启动批处理模式...")
	
	// 运行完整的批处理工作流
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 加载配置
	var configPath string
	wd, _ := os.Getwd()
	configPath = filepath.Join(wd, "config.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		exe, exeErr := os.Executable()
		if exeErr != nil {
			logger.Fatal("无法获取可执行文件路径", zap.Error(exeErr))
		}
		exeDir := filepath.Dir(exe)
		configPath = filepath.Join(exeDir, "config.yaml")
	}

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		logger.Fatal("读取配置文件失败",
			zap.String("configPath", configPath),
			zap.Error(err),
		)
	}

	// 创建工作流处理器
	processor, err := workflow.NewProcessor(logger)
	if err != nil {
		logger.Fatal("创建工作流处理器失败", zap.Error(err))
	}

	// 执行完整的批处理工作流
	logger.Info("开始执行批处理工作流...")

	// 这里可以根据配置执行完整的工作流
	// 示例：执行所有MCP工具
	mcpServer, err := mcp.NewServer(processor, logger)
	if err != nil {
		logger.Fatal("创建MCP服务器失败", zap.Error(err))
	}

	availableTools := mcpServer.GetHandler().GetToolNames()
	logger.Info("可用工具数量", zap.Int("count", len(availableTools)))

	// 执行所有工具或根据配置执行特定工具
	for _, toolName := range availableTools {
		logger.Info("执行工具", zap.String("tool", toolName))
		// 这里可以根据具体需求执行工具
		// 实际实现可能会更复杂，例如从输入目录读取小说文件并处理
	}

	logger.Info("批处理工作流完成")
}


