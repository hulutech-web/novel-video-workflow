package aegisub

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// AegisubGenerator 是使用Aegisub软件生成字幕的工具
type AegisubGenerator struct {
	LuaScriptPath string
	ScriptPath    string
}

// NewAegisubGenerator 创建一个新的Aegisub字幕生成器实例
func NewAegisubGenerator() *AegisubGenerator {
	return &AegisubGenerator{
		LuaScriptPath: filepath.Join("pkg", "tools", "aegisub", "aegisub_generator.lua"),
		ScriptPath:    filepath.Join("pkg", "tools", "aegisub", "aegisub_subtitle_gen.sh"),
	}
}

// GenerateSubtitle 生成字幕文件
// 参数: audioFile - 音频文件路径, textFile - 文本文件路径, outputSrt - 输出SRT文件路径
func (ag *AegisubGenerator) GenerateSubtitle(audioFile, textFile, outputSrt string) error {
	// 检查输入文件是否存在
	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		return fmt.Errorf("音频文件不存在: %s", audioFile)
	}

	if _, err := os.Stat(textFile); os.IsNotExist(err) {
		return fmt.Errorf("文本文件不存在: %s", textFile)
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputSrt)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 执行shell脚本来调用Aegisub生成字幕
	cmd := exec.Command(ag.ScriptPath, audioFile, textFile, outputSrt)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行Aegisub字幕生成脚本失败: %v, 输出: %s", err, string(output))
	}

	fmt.Printf("Aegisub字幕生成成功: %s\n", string(output))
	return nil
}

// GenerateSubtitleWithDefaults 使用默认配置生成字幕
func (ag *AegisubGenerator) GenerateSubtitleWithDefaults() error {
	// 使用Lua脚本中的默认配置
	defaultConfig := map[string]string{
		"text_file":  "/Users/mac/Documents/ai/chapter6/novel06.txt",
		"audio_file": "/Users/mac/Documents/ai/chapter6/spk_1767952937.wav",
		"output_srt": "/Users/mac/Documents/ai/chapter6/novel_word_ratio_final.srt",
	}

	return ag.GenerateSubtitle(
		defaultConfig["audio_file"],
		defaultConfig["text_file"],
		defaultConfig["output_srt"],
	)
}

// GenerateSubtitleFromText 使用指定的音频文件和文本内容生成字幕
// 此方法将文本内容写入临时文件，然后调用Aegisub生成字幕
func (ag *AegisubGenerator) GenerateSubtitleFromText(audioFile, textContent, outputSrt string) error {
	// 创建临时文本文件
	tempTextFile, err := createTempTextFile(textContent)
	if err != nil {
		return fmt.Errorf("创建临时文本文件失败: %v", err)
	}
	defer os.Remove(tempTextFile) // 清理临时文件

	// 调用字幕生成方法
	return ag.GenerateSubtitle(audioFile, tempTextFile, outputSrt)
}

// createTempTextFile 创建包含文本内容的临时文件
func createTempTextFile(textContent string) (string, error) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "subtitle_text_*.txt")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// 写入文本内容，确保使用UTF-8编码
	_, err = tempFile.WriteString(textContent)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

// GenerateSubtitleFromIndextts2Audio 使用Indextts2生成的音频和提供的文本生成字幕
func (ag *AegisubGenerator) GenerateSubtitleFromIndextts2Audio(indextts2AudioFile, textContent, outputSrt string) error {
	// 验证Indextts2音频文件是否存在
	if _, err := os.Stat(indextts2AudioFile); os.IsNotExist(err) {
		return fmt.Errorf("Indextts2音频文件不存在: %s", indextts2AudioFile)
	}

	// 使用提供的文本内容生成字幕
	return ag.GenerateSubtitleFromText(indextts2AudioFile, textContent, outputSrt)
}
