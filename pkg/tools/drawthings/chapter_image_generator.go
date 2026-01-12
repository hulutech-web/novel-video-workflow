package drawthings

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ChapterImageGenerator 章节图像生成器
type ChapterImageGenerator struct {
	Client       *DrawThingsClient
	OllamaClient *OllamaClient
	Logger       *zap.Logger
}

// NewChapterImageGenerator 创建章节图像生成器
func NewChapterImageGenerator(logger *zap.Logger) *ChapterImageGenerator {
	client := NewDrawThingsClient(logger, "http://localhost:7861")
	ollamaClient := NewOllamaClient(logger, "http://localhost:11434", "qwen3:4b") // 使用默认Ollama配置
	return &ChapterImageGenerator{
		Client:       client,
		OllamaClient: ollamaClient,
		Logger:       logger,
	}
}

// ParagraphImage 生成的段落图像信息
type ParagraphImage struct {
	ParagraphText string `json:"paragraph_text"`
	ImageFile     string `json:"image_file"`
	ImagePrompt   string `json:"image_prompt"` // 新增：图像提示词
	Index         int    `json:"index"`
}

// GenerateImagesFromChapter 根据章节文本生成图像
func (c *ChapterImageGenerator) GenerateImagesFromChapter(chapterText, outputDir string, width, height int, isSuspense bool) ([]ParagraphImage, error) {
	// 将相对路径转换为绝对路径
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		c.Logger.Warn("无法解析输出目录路径", zap.String("output_dir", outputDir), zap.Error(err))
		absOutputDir = outputDir // 如果无法解析，继续使用原路径
	}

	// 按段落分割章节文本
	paragraphs := c.splitChapterIntoParagraphs(chapterText)

	var results []ParagraphImage

	c.Logger.Info("开始生成章节图像",
		zap.String("output_dir", absOutputDir),
		zap.Int("paragraph_count", len(paragraphs)))

	// 检查DrawThings API可用性
	if !c.Client.APIAvailable {
		c.Logger.Warn("DrawThings API不可用，将跳过图像生成步骤", zap.String("api_url", c.Client.BaseURL))
		return results, fmt.Errorf("DrawThings API不可用，请确保Stable Diffusion WebUI正在运行在 %s", c.Client.BaseURL)
	}

	for i, paragraph := range paragraphs {
		// 跳过空白段落
		trimmedPara := strings.TrimSpace(paragraph)
		if trimmedPara == "" {
			continue
		}

		// 生成图像文件名
		imageFile := filepath.Join(absOutputDir, fmt.Sprintf("paragraph_%03d.png", i+1))

		// 确保输出目录存在
		if err := os.MkdirAll(absOutputDir, 0755); err != nil {
			c.Logger.Error("创建输出目录失败", zap.String("dir", absOutputDir), zap.Error(err))
			continue
		}

		// 使用Ollama生成更精确的图像提示词
		styleDesc := "悬疑风格，氛围紧张，暗淡光线，神秘感"
		if isSuspense {
			styleDesc = "悬疑惊悚风格，周围环境模糊成黑影, 空气凝滞,浅景深, 胶片颗粒感, 低饱和度，极致悬疑氛围, 阴沉窒息感, 夏季，环境阴霾，其他部分模糊不可见"
		}

		imagePrompt, err := c.OllamaClient.GenerateImagePrompt(trimmedPara, styleDesc)
		if err != nil {
			c.Logger.Warn("使用Ollama生成图像提示词失败，使用原始文本",
				zap.Int("paragraph_index", i),
				zap.String("paragraph", trimmedPara),
				zap.Error(err))
			// 如果Ollama失败，使用原始文本加上悬疑风格
			if isSuspense {
				imagePrompt = trimmedPara + ", 周围环境模糊成黑影, 空气凝滞,浅景深, 胶片颗粒感, 低饱和度，极致悬疑氛围, 阴沉窒息感, 夏季，环境阴霾，其他部分模糊不可见"
			} else {
				imagePrompt = trimmedPara
			}
		}

		// 使用生成的提示词调用DrawThings API生成图像
		err = c.Client.GenerateImageFromText(imagePrompt, imageFile, width, height, false) // isSuspense已经在提示词中处理
		if err != nil {
			c.Logger.Warn("生成段落图像失败",
				zap.Int("paragraph_index", i),
				zap.String("paragraph", trimmedPara),
				zap.String("prompt", imagePrompt),
				zap.Error(err))
			continue
		}

		// 添加到结果
		results = append(results, ParagraphImage{
			ParagraphText: trimmedPara,
			ImageFile:     imageFile,
			ImagePrompt:   imagePrompt, // 记录使用的提示词
			Index:         i,
		})

		c.Logger.Info("段落图像生成成功",
			zap.Int("index", i),
			zap.String("image_file", imageFile),
			zap.String("paragraph_preview", c.truncateString(trimmedPara, 50)),
			zap.String("prompt_preview", c.truncateString(imagePrompt, 80)))

		// 添加小延迟避免API过载
		time.Sleep(100 * time.Millisecond)
	}

	c.Logger.Info("章节图像生成完成",
		zap.Int("generated_count", len(results)),
		zap.Int("total_paragraphs", len(paragraphs)))

	return results, nil
}

// splitChapterIntoParagraphs 将章节文本分割为段落
func (c *ChapterImageGenerator) splitChapterIntoParagraphs(text string) []string {
	// 按换行符分割文本
	lines := strings.Split(text, "\n")

	var paragraphs []string
	var currentParagraph strings.Builder

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			// 遇到空行，结束当前段落
			if currentParagraph.Len() > 0 {
				paragraphs = append(paragraphs, strings.TrimSpace(currentParagraph.String()))
				currentParagraph.Reset()
			}
		} else {
			// 添加到当前段落
			if currentParagraph.Len() > 0 {
				currentParagraph.WriteString(" ")
			}
			currentParagraph.WriteString(trimmedLine)
		}
	}

	// 处理最后一个段落
	if currentParagraph.Len() > 0 {
		paragraphs = append(paragraphs, strings.TrimSpace(currentParagraph.String()))
	}

	// 过滤掉过短的段落（比如只有标点符号）
	var filtered []string
	for _, para := range paragraphs {
		// 只保留非空且有一定长度的段落
		if len(strings.TrimSpace(para)) > 3 { // 至少3个字符
			filtered = append(filtered, para)
		}
	}

	return filtered
}

// truncateString 截断字符串
func (c *ChapterImageGenerator) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GenerateImageSequenceFromText 根据文本生成一系列图像，用于视频制作
func (c *ChapterImageGenerator) GenerateImageSequenceFromText(text, outputDir, baseFilename string, width, height int, isSuspense bool) ([]string, error) {
	// 分割文本为句子或有意义的片段
	segments := c.segmentText(text)

	var imageFiles []string

	c.Logger.Info("开始生成文本序列图像",
		zap.String("output_dir", outputDir),
		zap.Int("segment_count", len(segments)))

	for i, segment := range segments {
		// 跳过空白段
		trimmedSeg := strings.TrimSpace(segment)
		if trimmedSeg == "" {
			continue
		}

		// 生成图像文件名
		imageFile := filepath.Join(outputDir, fmt.Sprintf("%s_%03d.png", baseFilename, i+1))

		// 确保输出目录存在
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			c.Logger.Error("创建输出目录失败", zap.String("dir", outputDir), zap.Error(err))
			continue
		}

		// 生成图像
		err := c.Client.GenerateImageFromText(trimmedSeg, imageFile, width, height, isSuspense)
		if err != nil {
			c.Logger.Warn("生成文本片段图像失败",
				zap.Int("segment_index", i),
				zap.String("segment", trimmedSeg),
				zap.Error(err))
			continue
		}

		imageFiles = append(imageFiles, imageFile)

		c.Logger.Info("文本片段图像生成成功",
			zap.Int("index", i),
			zap.String("image_file", imageFile),
			zap.String("segment_preview", c.truncateString(trimmedSeg, 50)))

		// 添加小延迟避免API过载
		time.Sleep(100 * time.Millisecond)
	}

	c.Logger.Info("文本序列图像生成完成",
		zap.Int("generated_count", len(imageFiles)),
		zap.Int("total_segments", len(segments)))

	return imageFiles, nil
}

// segmentText 将文本分割为适合图像生成的片段
func (c *ChapterImageGenerator) segmentText(text string) []string {
	// 首先按句子分割
	// 注意：这只是一个简单的实现，实际应用中可能需要更复杂的NLP处理
	var segments []string

	// 按换行符分割
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			// 如果行太长，进一步分割
			if len(trimmedLine) > 100 { // 如果超过100个字符，尝试按句子分割
				subSegments := c.splitLongLine(trimmedLine)
				segments = append(segments, subSegments...)
			} else {
				segments = append(segments, trimmedLine)
			}
		}
	}

	return segments
}

// splitLongLine 将长行按句子分割
func (c *ChapterImageGenerator) splitLongLine(line string) []string {
	// 按常见的句子结束标点符号分割
	sentenceEndings := []string{".", "。", "!", "！", "?", "？", "；", ";", "……"}

	// 这里可以实现更复杂的句子分割逻辑
	// 当前返回原行，但在实际应用中应该按句子分割
	var result []string
	currentSegment := ""

	// 简单按标点符号分割，确保每段不超过一定长度
	for _, char := range line {
		currentSegment += string(char)

		// 检查是否是句子结束符
		for _, ending := range sentenceEndings {
			if strings.HasSuffix(currentSegment, ending) {
				// 检查段长度，如果不太长则分割
				if len(currentSegment) > 50 { // 如果超过50个字符
					result = append(result, strings.TrimSpace(currentSegment))
					currentSegment = ""
				}
				break
			}
		}

		// 如果段落太长也强制分割
		if len(currentSegment) >= 100 {
			result = append(result, strings.TrimSpace(currentSegment))
			currentSegment = ""
		}
	}

	// 添加剩余部分
	if currentSegment != "" {
		result = append(result, strings.TrimSpace(currentSegment))
	}

	return result
}

// ProcessChapterTextFile 处理章节文本文件
func (c *ChapterImageGenerator) ProcessChapterTextFile(textFilePath, outputDir string, width, height int, isSuspense bool) error {
	// 读取文本文件
	file, err := os.Open(textFilePath)
	if err != nil {
		return fmt.Errorf("打开文本文件失败: %v", err)
	}
	defer file.Close()

	// 读取全部内容
	scanner := bufio.NewScanner(file)
	var content strings.Builder
	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文本文件失败: %v", err)
	}

	text := content.String()

	// 生成图像
	_, err = c.GenerateImagesFromChapter(text, outputDir, width, height, isSuspense)
	if err != nil {
		return fmt.Errorf("生成章节图像失败: %v", err)
	}

	return nil
}
