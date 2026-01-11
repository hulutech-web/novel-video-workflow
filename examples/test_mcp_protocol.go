package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"novel-video-workflow/internal/mcp"
	"novel-video-workflow/internal/workflow"
	"os"
	"time"

	"go.uber.org/zap"
)

// MCPCallRequest MCP调用请求
type MCPCallRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

// MCPCallResponse MCP调用响应
type MCPCallResponse struct {
	Result interface{} `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// CallMCPTool 调用MCP工具
func CallMCPTool(method string, params interface{}) (*MCPCallResponse, error) {
	req := MCPCallRequest{
		Method: method,
		Params: params,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建管道用于模拟stdio通信
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	// 启动MCP服务器
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	processor, err := workflow.NewProcessor(logger)
	if err != nil {
		return nil, fmt.Errorf("创建处理器失败: %w", err)
	}

	mcpServer, err := mcp.NewServer(processor, logger)
	if err != nil {
		return nil, fmt.Errorf("创建MCP服务器失败: %w", err)
	}

	// 将请求写入stdin
	go func() {
		defer stdinWriter.Close()
		stdinWriter.Write(reqBytes)
	}()

	// 启动MCP服务器处理
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		mcpServer.Start(ctx)
	}()

	// 读取响应
	buf := make([]byte, 4096)
	n, err := stdoutReader.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	response := &MCPCallResponse{}
	err = json.Unmarshal(buf[:n], response)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return response, nil
}

// TestMCPWorkflowWithProtocol 使用MCP协议测试工作流
func TestMCPWorkflowWithProtocol() {
	fmt.Println("开始使用MCP协议测试工作流...")

	// 示例文本和参考音频
	text := "这是一个示例文本，用于测试MCP工作流。我们将使用Indextts2生成音频，然后使用Aegisub生成字幕。"
	referenceAudio := "/Users/mac/code/ai/novel-video-workflow/ref.m4a" // 替换为实际的参考音频路径
	outputAudio := "output/test_audio.wav"
	outputSrt := "output/test_subtitle.srt"

	// 1. 调用indextts2服务生成音频
	fmt.Println("步骤1: 通过MCP调用indextts2服务...")
	
	indextts2Params := map[string]interface{}{
		"text":            text,
		"reference_audio": referenceAudio,
		"output_file":     outputAudio,
	}

	indextts2Resp, err := CallMCPTool("generate_indextts2_audio", indextts2Params)
	if err != nil {
		fmt.Printf("调用indextts2服务失败: %v\n", err)
		return
	}

	if indextts2Resp.Error != nil {
		fmt.Printf("indextts2服务返回错误: %d - %s\n", indextts2Resp.Error.Code, indextts2Resp.Error.Message)
		return
	}

	fmt.Printf("indextts2服务调用成功: %+v\n", indextts2Resp.Result)

	// 2. 调用AegisubGenerator服务生成字幕
	fmt.Println("步骤2: 通过MCP调用AegisubGenerator服务...")
	
	aegisubParams := map[string]interface{}{
		"audio_file":    outputAudio,
		"text_content":  text,
		"output_file":   outputSrt,
	}

	aegisubResp, err := CallMCPTool("generate_subtitles_from_indextts2", aegisubParams)
	if err != nil {
		fmt.Printf("调用AegisubGenerator服务失败: %v\n", err)
		return
	}

	if aegisubResp.Error != nil {
		fmt.Printf("AegisubGenerator服务返回错误: %d - %s\n", aegisubResp.Error.Code, aegisubResp.Error.Message)
		return
	}

	fmt.Printf("AegisubGenerator服务调用成功: %+v\n", aegisubResp.Result)

	fmt.Println("MCP工作流协议测试完成！")
	fmt.Printf("音频文件: %s\n", outputAudio)
	fmt.Printf("字幕文件: %s\n", outputSrt)
}

func main() {
	// 确保输出目录存在
	os.MkdirAll("output", 0755)

	// 运行测试
	TestMCPWorkflowWithProtocol()
}