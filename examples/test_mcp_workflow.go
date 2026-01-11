package main

import (
	"context"
	"fmt"
	"novel-video-workflow/internal/mcp"
	"novel-video-workflow/internal/workflow"
	"go.uber.org/zap"
	"novel-video-workflow/internal/tools/indextts2"
	"novel-video-workflow/internal/tools"
	"path/filepath"
)

// TestMCPWorkflow 测试MCP工作流，依次调用indextts2和AegisubGenerator
func TestMCPWorkflow() {
	// 创建logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 创建工作流处理器
	processor, err := workflow.NewProcessor(logger)
	if err != nil {
		fmt.Printf("创建处理器失败: %v\n", err)
		return
	}

	// 创建MCP服务器
	mcpServer, err := mcp.NewServer(processor, logger)
	if err != nil {
		fmt.Printf("创建MCP服务器失败: %v\n", err)
		return
	}

	fmt.Println("开始测试MCP工作流...")

	// 示例文本和参考音频
	text := "这是一个示例文本，用于测试MCP工作流。我们将使用Indextts2生成音频，然后使用Aegisub生成字幕。"
	referenceAudio := "/Users/mac/code/ai/novel-video-workflow/ref.m4a" // 替换为实际的参考音频路径
	outputAudio := "output/test_audio.wav"
	outputSrt := "output/test_subtitle.srt"

	// 1. 调用indextts2服务生成音频
	fmt.Println("步骤1: 调用indextts2服务生成音频...")
	ctx := context.Background()
	
	// 直接使用indextts2客户端进行音频生成
	client := indextts2.NewIndexTTS2Client(logger, "http://localhost:7860")
	err = client.GenerateTTSWithAudio(referenceAudio, text, outputAudio)
	if err != nil {
		fmt.Printf("Indextts2音频生成失败: %v\n", err)
		return
	}
	fmt.Printf("Indextts2音频生成成功: %s\n", outputAudio)

	// 2. 调用AegisubGenerator服务生成字幕
	fmt.Println("步骤2: 调用AegisubGenerator服务生成字幕...")
	
	// 使用AegisubIntegration处理音频和文本，生成字幕
	aegisubIntegration := tools.NewAegisubIntegration()
	err = aegisubIntegration.ProcessIndextts2OutputWithCustomName(outputAudio, text, outputSrt)
	if err != nil {
		fmt.Printf("Aegisub字幕生成失败: %v\n", err)
		return
	}
	fmt.Printf("Aegisub字幕生成成功: %s\n", outputSrt)

	fmt.Println("MCP工作流测试完成！")
	fmt.Printf("音频文件: %s\n", outputAudio)
	fmt.Printf("字幕文件: %s\n", outputSrt)
}

func main() {
	TestMCPWorkflow()
}