package aegisub

import (
	"fmt"
	"log"
	"novel-video-workflow/pkg/tools/indextts2"
	"go.uber.org/zap"
)

// 示例：如何在Indextts2完成后自动调用AegisubGenerator
func ExampleIntegration() {
	// 创建logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 创建Indextts2客户端
	client := indextts2.NewIndexTTS2Client(logger, "http://localhost:7860")

	// 示例文本和参考音频
	text := "这是一个示例文本，用于生成语音和字幕。"
	referenceAudio := "./ref.m4a" // 替换为实际的参考音频路径
	outputAudio := "output/example_audio.wav"                        // 输出音频路径

	// 1. 使用Indextts2生成音频
	fmt.Println("步骤1: 使用Indextts2生成音频...")
	err := client.GenerateTTSWithAudio(referenceAudio, text, outputAudio)
	if err != nil {
		log.Fatalf("生成音频失败: %v", err)
	}
	fmt.Printf("音频生成成功: %s\n", outputAudio)

	// 2. 在Indextts2完成后自动调用AegisubGenerator生成字幕
	fmt.Println("步骤2: 自动生成字幕...")
	integration := NewAegisubIntegration()
	
	outputDir := "output"
	outputSrt, err := integration.ProcessIndextts2Output(outputAudio, text, outputDir)
	if err != nil {
		log.Fatalf("生成字幕失败: %v", err)
	}
	fmt.Printf("字幕生成成功: %s\n", outputSrt)

	fmt.Println("完整流程完成！音频文件:", outputAudio, "字幕文件:", outputSrt)
}