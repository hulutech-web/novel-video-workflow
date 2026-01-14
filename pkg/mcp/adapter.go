package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"go.uber.org/zap"
)

// MCPAdapter 是一个适配器，允许外部工具与MCP服务器交互
type MCPAdapter struct {
	logger *zap.Logger
}

// NewMCPAdapter 创建新的MCP适配器
func NewMCPAdapter(logger *zap.Logger) *MCPAdapter {
	return &MCPAdapter{
		logger: logger,
	}
}

// CallTool 调用MCP工具
func (a *MCPAdapter) CallTool(toolName string, arguments map[string]interface{}) (map[string]interface{}, error) {
	// 创建MCP工具调用消息
	callToolMsg := map[string]interface{}{
		"method":  "tools/call",
		"id":      fmt.Sprintf("tool-call-%d", time.Now().UnixNano()),
		"jsonrpc": "2.0",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	jsonData, err := json.Marshal(callToolMsg)
	if err != nil {
		a.logger.Error("Failed to marshal tool call", zap.Error(err))
		return nil, err
	}

	// 通过标准输入输出与MCP服务器通信
	// 我们创建一个子进程来运行MCP服务器
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = "." // 设置工作目录为项目根目录

	stdin, err := cmd.StdinPipe()
	if err != nil {
		a.logger.Error("Failed to create stdin pipe", zap.Error(err))
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		a.logger.Error("Failed to create stdout pipe", zap.Error(err))
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		a.logger.Error("Failed to start MCP server process", zap.Error(err))
		return nil, err
	}

	// 发送工具调用请求
	_, err = stdin.Write(append(jsonData, '\n'))
	if err != nil {
		a.logger.Error("Failed to write to stdin", zap.Error(err))
		_ = cmd.Wait()
		return nil, err
	}

	// 等待响应
	reader := bufio.NewReader(stdout)
	responseStr, err := reader.ReadString('\n')
	if err != nil {
		a.logger.Error("Failed to read response", zap.Error(err))
		_ = cmd.Wait()
		return nil, err
	}

	// 关闭stdin并等待进程结束
	stdin.Close()
	_ = cmd.Wait()

	var response map[string]interface{}
	err = json.Unmarshal([]byte(responseStr), &response)
	if err != nil {
		a.logger.Error("Failed to unmarshal response", zap.Error(err))
		return nil, err
	}

	return response, nil
}

// RunAsMCPService 运行MCP服务模式
func (a *MCPAdapter) RunAsMCPService(ctx context.Context) error {
	// 设置环境变量以表明这是MCP服务模式
	os.Setenv("MCP_STDIO_MODE", "true")

	// 从main.go导入处理器和日志
	// 注意：这里我们不能直接导入main包，所以我们需要重构代码结构
	// 或者通过命令行方式启动MCP服务

	a.logger.Info("Starting MCP service adapter")

	// 实际上，我们会通过命令行启动主程序
	cmd := exec.CommandContext(ctx, "go", "run", "main.go")
	cmd.Env = append(os.Environ(), "MCP_STDIO_MODE=true")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP service: %w", err)
	}

	// 启动goroutine来处理stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			a.logger.Info("MCP Service Error", zap.String("msg", scanner.Text()))
		}
	}()

	// 启动goroutine来处理stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			a.logger.Info("MCP Service Output", zap.String("msg", line))
		}
	}()

	// 等待命令完成或上下文取消
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("Context cancelled, stopping MCP service")
		_ = cmd.Process.Kill()
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// GetAvailableTools 获取可用工具列表
func (a *MCPAdapter) GetAvailableTools() []string {
	// 这里应该连接到实际的MCP服务器并获取工具列表
	// 由于我们无法直接连接，我们可以返回一个预定义的列表
	// 或者通过运行一个特殊的命令来获取工具列表
	tools := []string{
		"generate_indextts2_audio",
		"generate_subtitles_from_indextts2",
		"file_split_novel_into_chapters",
		"generate_image_from_text",
		"generate_image_from_image",
		"generate_images_from_chapter",
		"generate_images_from_chapter_with_ai_prompt",
	}

	return tools
}

// ProcessWithOllamaDesktop 为Ollama Desktop提供简化的工具调用接口
func (a *MCPAdapter) ProcessWithOllamaDesktop(toolName string, params map[string]interface{}) (map[string]interface{}, error) {
	a.logger.Info("Processing tool call for Ollama Desktop",
		zap.String("tool", toolName),
		zap.Any("params", params))

	// 验证工具名称
	availableTools := a.GetAvailableTools()
	isValidTool := false
	for _, tool := range availableTools {
		if tool == toolName {
			isValidTool = true
			break
		}
	}

	if !isValidTool {
		return nil, fmt.Errorf("invalid tool: %s, available tools: %v", toolName, availableTools)
	}

	// 调用实际的MCP工具
	result, err := a.CallTool(toolName, params)
	if err != nil {
		a.logger.Error("Tool call failed",
			zap.String("tool", toolName),
			zap.Error(err))
		return nil, err
	}

	return result, nil
}
