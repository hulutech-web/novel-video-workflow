package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"novel-video-workflow/pkg/tools/indextts2"

	"go.uber.org/zap"
)

// TTSProcessor 文本转语音处理器
type TTSProcessor struct {
	logger *zap.Logger
}

// TTSResult TTS处理结果
type TTSResult struct {
	Success    bool   `json:"success"`
	OutputFile string `json:"output_file"`
	Error      string `json:"error,omitempty"`
}

// NewTTSProcessor 创建TTS处理器
func NewTTSProcessor(logger *zap.Logger) *TTSProcessor {
	return &TTSProcessor{
		logger: logger,
	}
}

// Generate 生成音频文件
func (tp *TTSProcessor) Generate(text, outputFile, referenceAudio string) (*TTSResult, error) {
	// 如果没有指定输出文件，生成默认文件名
	if outputFile == "" {
		outputFile = fmt.Sprintf("output/tts_output_%d.wav", tp.getTimestamp())
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return &TTSResult{
			Success: false,
			Error:   fmt.Sprintf("创建输出目录失败: %v", err),
		}, nil
	}

	// 如果没有参考音频，使用默认值或返回错误
	if referenceAudio == "" {
		// 尝试查找默认参考音频
		defaultAudio := tp.findDefaultReferenceAudio()
		if defaultAudio != "" {
			referenceAudio = defaultAudio
			tp.logger.Info("使用默认参考音频", zap.String("audio", defaultAudio))
		} else {
			return &TTSResult{
				Success: false,
				Error:   "未提供参考音频文件",
			}, nil
		}
	}

	// 检查参考音频是否存在
	if _, err := os.Stat(referenceAudio); os.IsNotExist(err) {
		return &TTSResult{
			Success: false,
			Error:   fmt.Sprintf("参考音频文件不存在: %s", referenceAudio),
		}, nil
	}

	// 使用Indextts2客户端生成音频
	client := indextts2.NewIndexTTS2Client(tp.logger, "http://localhost:7860")
	err := client.GenerateTTSWithAudio(referenceAudio, text, outputFile)
	if err != nil {
		tp.logger.Error("TTS生成失败", zap.Error(err))
		return &TTSResult{
			Success: false,
			Error:   fmt.Sprintf("TTS生成失败: %v", err),
		}, nil
	}

	// 验证输出文件
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return &TTSResult{
			Success: false,
			Error:   "TTS生成完成但输出文件不存在",
		}, nil
	}

	return &TTSResult{
		Success:    true,
		OutputFile: outputFile,
	}, nil
}

// getTimestamp 获取时间戳
func (tp *TTSProcessor) getTimestamp() int64 {
	return time.Now().Unix()
}

// findDefaultReferenceAudio 查找默认参考音频
func (tp *TTSProcessor) findDefaultReferenceAudio() string {
	// 检查常见的参考音频文件位置
	possibilities := []string{
		"./ref.m4a",
		"./音色.m4a",
		"./assets/ref_audio/ref.m4a",
		"./assets/ref_audio/音色.m4a",
	}

	for _, path := range possibilities {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
