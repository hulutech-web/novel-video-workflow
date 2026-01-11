package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"novel-video-workflow/internal/tools"
	"novel-video-workflow/internal/tools/indextts2"
)

func main() {
	// 创建logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 确保输出目录存在（使用项目根目录的output）
	outputDir := filepath.Join(".", "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("创建输出目录失败: %v\n", err)
		return
	}

	// 示例文本和参考音频
	text := "你好，这是一个测试文本。我们将使用Indextts2生成音频，然后使用Aegisub生成字幕。欢迎来到AI生成的世界！"
	referenceAudio := "./ref.m4a" // 使用相对路径

	// 检查参考音频是否存在
	if _, err := os.Stat(referenceAudio); os.IsNotExist(err) {
		// 尝试其他可能的参考音频路径
		possibilities := []string{
			"/Users/mac/code/ai/novel-video-workflow/ref.m4a",
			"./音色.m4a",
			"/Users/mac/code/ai/novel-video-workflow/音色.m4a",
		}
		
		found := false
		for _, path := range possibilities {
			if _, err := os.Stat(path); err == nil {
				referenceAudio = path
				found = true
				fmt.Printf("找到参考音频文件: %s\n", path)
				break
			}
		}
		
		if !found {
			fmt.Println("错误: 未找到参考音频文件，程序需要 ref.m4a 或 音色.m4a 文件")
			fmt.Println("请确保参考音频文件存在于项目根目录或其他预期路径中")
			return
		}
	}

	outputAudio := filepath.Join(outputDir, fmt.Sprintf("demo_audio_%d.wav", time.Now().Unix()))
	outputSrt := filepath.Join(outputDir, fmt.Sprintf("demo_subtitle_%d.srt", time.Now().Unix()))

	fmt.Printf("开始MCP工作流演示...\n")
	fmt.Printf("音频输出: %s\n", outputAudio)
	fmt.Printf("字幕输出: %s\n", outputSrt)

	// 1. 使用Indextts2生成音频
	fmt.Println("\n步骤1: 调用indextts2服务生成音频...")
	fmt.Printf("  - 文本: %.30s...\n", text)
	fmt.Printf("  - 参考音频: %s\n", referenceAudio)

	client := indextts2.NewIndexTTS2Client(logger, "http://localhost:7860")
	err := client.GenerateTTSWithAudio(referenceAudio, text, outputAudio)
	if err != nil {
		fmt.Printf("  - 错误: %v\n", err)
		fmt.Println("  - 提示: 请确保IndexTTS2服务正在运行 (python app.py)")
		return
	}

	// 验证音频文件是否生成
	if _, err := os.Stat(outputAudio); os.IsNotExist(err) {
		fmt.Printf("  - 错误: 音频文件未生成: %s\n", outputAudio)
		return
	}
	fmt.Printf("  - 成功: 音频生成完成: %s\n", outputAudio)

	// 2. 使用AegisubGenerator生成字幕
	fmt.Println("\n步骤2: 调用AegisubGenerator服务生成字幕...")
	fmt.Printf("  - 音频文件: %s\n", outputAudio)
	fmt.Printf("  - 文本内容: %.30s...\n", text)

	aegisubIntegration := tools.NewAegisubIntegration()
	err = aegisubIntegration.ProcessIndextts2OutputWithCustomName(outputAudio, text, outputSrt)
	if err != nil {
		fmt.Printf("  - 警告: Aegisub字幕生成失败: %v\n", err)
		fmt.Println("  - 尝试使用备用方案生成字幕...")

		// 使用备用方案生成字幕
		err = createSimpleSubtitleFile(text, outputSrt)
		if err != nil {
			fmt.Printf("  - 错误: 备用方案也失败: %v\n", err)
			return
		}
		fmt.Printf("  - 成功: 使用备用方案生成字幕: %s\n", outputSrt)
	} else {
		fmt.Printf("  - 成功: 字幕生成完成: %s\n", outputSrt)
	}

	// 验证生成的文件
	audioInfo, err := os.Stat(outputAudio)
	if err != nil {
		fmt.Printf("无法获取音频文件信息: %v\n", err)
	} else {
		fmt.Printf("  - 音频文件大小: %d bytes\n", audioInfo.Size())
	}

	subtitleInfo, err := os.Stat(outputSrt)
	if err != nil {
		fmt.Printf("无法获取字幕文件信息: %v\n", err)
	} else {
		fmt.Printf("  - 字幕文件大小: %d bytes\n", subtitleInfo.Size())
	}

	fmt.Println("\nMCP工作流演示完成！")
	fmt.Printf("最终输出:\n  - 音频: %s\n  - 字幕: %s\n", outputAudio, outputSrt)
}

// createSimpleSubtitleFile 创建简单的字幕文件（备用方案）
func createSimpleSubtitleFile(text, outputSrt string) error {
	// 简单的字幕生成逻辑：将文本分成几段，每段分配一个时间段
	lines := []string{}
	// 每50个字符分为一段
	for i := 0; i < len(text); i += 50 {
		end := i + 50
		if end > len(text) {
			end = len(text)
		}
		lines = append(lines, text[i:end])
	}

	srtContent := ""
	for i, line := range lines {
		// 计算时间戳，假设每段持续2秒
		startSec := i * 2
		endSec := (i + 1) * 2
		
		startTime := fmt.Sprintf("00:00:%02d,000", startSec)
		endTime := fmt.Sprintf("00:00:%02d,000", endSec)
		
		srtContent += fmt.Sprintf("%d\n%s --> %s\n%s\n\n", i+1, startTime, endTime, line)
	}

	return os.WriteFile(outputSrt, []byte(srtContent), 0644)
}