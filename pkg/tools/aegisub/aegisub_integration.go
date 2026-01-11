package aegisub

import (
	"fmt"
	"path/filepath"
)

// AegisubIntegration 提供Aegisub与其他工具的集成功能
type AegisubIntegration struct {
	aegisubGen *AegisubGenerator
}

// NewAegisubIntegration 创建Aegisub集成实例
func NewAegisubIntegration() *AegisubIntegration {
	return &AegisubIntegration{
		aegisubGen: NewAegisubGenerator(),
	}
}

// ProcessIndextts2Output 处理Indextts2的输出，自动生成字幕
// 这个函数接收Indextts2生成的音频文件和对应的文本，生成SRT字幕文件
func (ai *AegisubIntegration) ProcessIndextts2Output(audioFile, textContent, outputDir string) (string, error) {
	// 生成输出SRT文件路径
	audioFileName := filepath.Base(audioFile)
	srtFileName := audioFileName[:len(audioFileName)-len(filepath.Ext(audioFileName))] + ".srt"
	outputSrt := filepath.Join(outputDir, srtFileName)

	// 使用AegisubGenerator生成字幕
	err := ai.aegisubGen.GenerateSubtitleFromIndextts2Audio(audioFile, textContent, outputSrt)
	if err != nil {
		return "", fmt.Errorf("生成字幕失败: %v", err)
	}

	return outputSrt, nil
}

// ProcessIndextts2OutputWithCustomName 处理Indextts2的输出，使用自定义的输出文件名
func (ai *AegisubIntegration) ProcessIndextts2OutputWithCustomName(audioFile, textContent, outputSrt string) error {
	// 使用AegisubGenerator生成字幕
	err := ai.aegisubGen.GenerateSubtitleFromIndextts2Audio(audioFile, textContent, outputSrt)
	if err != nil {
		return fmt.Errorf("生成字幕失败: %v", err)
	}

	return nil
}