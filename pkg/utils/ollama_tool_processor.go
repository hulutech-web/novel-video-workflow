package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ToolCall 表示工具调用
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResponse 工具调用响应
type ToolResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
	Error   string                 `json:"error,omitempty"`
}

// MCPServer 用于管理MCP服务器进程
type MCPServer struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

// NewMCPServer 创建新的MCP服务器实例
func NewMCPServer() (*MCPServer, error) {
	cmd := exec.Command("go", "run", "main.go")
	cmd.Dir = "."

	// 设置环境变量，让主程序知道它是作为MCP服务器运行
	cmd.Env = append(os.Environ(), "MCP_STDIO_MODE=true")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// 等待服务器初始化
	time.Sleep(3 * time.Second)

	return &MCPServer{
		cmd:    cmd,
		stdin:  stdin.(io.WriteCloser),
		stdout: stdout.(io.ReadCloser),
	}, nil
}

// CallTool 调用MCP工具
func (s *MCPServer) CallTool(toolName string, arguments interface{}) (map[string]interface{}, error) {
	callToolMsg := map[string]interface{}{
		"method":  "tools/call",
		"id":      fmt.Sprintf("tool-call-%d", time.Now().Unix()),
		"jsonrpc": "2.0",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	jsonData, err := json.Marshal(callToolMsg)
	if err != nil {
		return nil, err
	}

	// 发送请求
	_, err = s.stdin.Write(append(jsonData, '\n'))
	if err != nil {
		return nil, err
	}

	// 设置读取超时
	done := make(chan struct{})
	responseChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(done)
		reader := bufio.NewReader(s.stdout)
		responseStr, err := reader.ReadString('\n')
		if err != nil {
			errChan <- err
			return
		}
		responseChan <- responseStr
	}()

	select {
	case responseStr := <-responseChan:
		var response map[string]interface{}
		err = json.Unmarshal([]byte(strings.TrimSpace(responseStr)), &response)
		if err != nil {
			return nil, err
		}
		return response, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(30 * time.Second): // 30秒超时
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// Close 关闭MCP服务器
func (s *MCPServer) Close() error {
	if s.stdin != nil {
		s.stdin.Close()
	}
	if s.stdout != nil {
		s.stdout.Close()
	}

	// 等待命令完成，但设置超时
	done := make(chan error, 1)
	go func() {
		done <- s.cmd.Wait()
	}()

	select {
	case <-done:
		// 正常完成
	case <-time.After(5 * time.Second):
		// 超时，强制终止进程
		if s.cmd.Process != nil {
			s.cmd.Process.Kill()
		}
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		response := map[string]interface{}{
			"error": "Usage: go run ollama_tool_processor.go <tool-call-json>",
		}
		responseBytes, _ := json.Marshal(response)
		fmt.Println(string(responseBytes))
		return
	}

	// 解析工具调用参数
	var toolCall ToolCall
	if err := json.Unmarshal([]byte(os.Args[1]), &toolCall); err != nil {
		response := map[string]interface{}{
			"error": fmt.Sprintf("Failed to parse tool call: %s", err.Error()),
		}
		responseBytes, _ := json.Marshal(response)
		fmt.Println(string(responseBytes))
		return
	}

	// 映射工具名称 - 为Ollama Desktop提供友好的工具名
	actualToolName := toolCall.Name
	switch toolCall.Name {
	case "novel_video_workflow_generate_audio":
		actualToolName = "generate_indextts2_audio"
		// 设置默认值
		if _, exists := toolCall.Arguments["reference_audio"]; !exists {
			toolCall.Arguments["reference_audio"] = "./assets/ref_audio/ref.m4a"
		}
		if _, exists := toolCall.Arguments["output_file"]; !exists {
			toolCall.Arguments["output_file"] = fmt.Sprintf("./output/ollama_output_%d.wav", time.Now().Unix())
		}
	case "novel_video_workflow_process_chapter":
		actualToolName = "process_chapter"
	case "novel_video_workflow_generate_image":
		actualToolName = "generate_image_from_text"
		if _, exists := toolCall.Arguments["output_file"]; !exists {
			toolCall.Arguments["output_file"] = fmt.Sprintf("./output/image_%d.png", time.Now().Unix())
		}
	case "novel_video_workflow_generate_chapter_images":
		actualToolName = "generate_images_from_chapter"
		if _, exists := toolCall.Arguments["width"]; !exists {
			toolCall.Arguments["width"] = 512
		}
		if _, exists := toolCall.Arguments["height"]; !exists {
			toolCall.Arguments["height"] = 896
		}
	case "novel_video_workflow_split_novel":
		actualToolName = "file_split_novel_into_chapters"
	default:
		// 如果工具名不匹配映射，使用原始工具名
		actualToolName = toolCall.Name
	}

	// 启动MCP服务器并调用工具
	mcpServer, err := NewMCPServer()
	if err != nil {
		response := map[string]interface{}{
			"error": fmt.Sprintf("Failed to start MCP server: %s", err.Error()),
		}
		responseBytes, _ := json.Marshal(response)
		fmt.Println(string(responseBytes))
		return
	}
	defer mcpServer.Close()

	result, err := mcpServer.CallTool(actualToolName, toolCall.Arguments)
	if err != nil {
		response := map[string]interface{}{
			"error": fmt.Sprintf("Tool call failed: %s", err.Error()),
		}
		responseBytes, _ := json.Marshal(response)
		fmt.Println(string(responseBytes))
		return
	}

	// 格式化响应
	response := ToolResponse{
		Success: true,
		Data:    result,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		response := map[string]interface{}{
			"error": fmt.Sprintf("Failed to format response: %s", err.Error()),
		}
		responseBytes, _ := json.Marshal(response)
		fmt.Println(string(responseBytes))
		return
	}

	fmt.Println(string(responseBytes))
}
