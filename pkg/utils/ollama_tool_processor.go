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
	cmd := exec.Command("go", "run", ".", "mcp")
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
	time.Sleep(2 * time.Second)

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

	_, err = s.stdin.Write(append(jsonData, '\n'))
	if err != nil {
		return nil, err
	}

	// 读取响应
	reader := bufio.NewReader(s.stdout)
	responseStr, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}
	err = json.Unmarshal([]byte(strings.TrimSpace(responseStr)), &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Close 关闭MCP服务器
func (s *MCPServer) Close() error {
	s.stdin.Close()
	s.stdout.Close()
	return s.cmd.Wait()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(`{"error": "Usage: go run ollama_tool_processor.go <tool-call-json>"}`)
		return
	}

	// 解析工具调用参数
	var toolCall ToolCall
	if err := json.Unmarshal([]byte(os.Args[1]), &toolCall); err != nil {
		fmt.Printf(`{"error": "Failed to parse tool call: %s"}`, err.Error())
		return
	}

	// 映射工具名称
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
	case "process_chapter":
		actualToolName = "process_chapter"
	default:
		fmt.Printf(`{"error": "Unknown tool: %s"}`, toolCall.Name)
		return
	}

	// 启动MCP服务器并调用工具
	mcpServer, err := NewMCPServer()
	if err != nil {
		fmt.Printf(`{"error": "Failed to start MCP server: %s"}`, err.Error())
		return
	}
	defer mcpServer.Close()

	result, err := mcpServer.CallTool(actualToolName, toolCall.Arguments)
	if err != nil {
		fmt.Printf(`{"error": "Tool call failed: %s"}`, err.Error())
		return
	}

	// 格式化响应
	response := ToolResponse{
		Success: true,
		Data:    result,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Printf(`{"error": "Failed to format response: %s"}`, err.Error())
		return
	}

	fmt.Println(string(responseBytes))
}
