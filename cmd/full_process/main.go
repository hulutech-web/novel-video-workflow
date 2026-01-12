package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	drawthings "novel-video-workflow/pkg/tools/drawthings"
	"novel-video-workflow/pkg/tools/file"
)

// NovelVideoProcessor 小说视频处理器，实现完整的处理流程
type NovelVideoProcessor struct {
	logger        *zap.Logger
	fileManager   *file.FileManager
	drawThingsGen *drawthings.ChapterImageGenerator
}

// NewNovelVideoProcessor 创建新的小说视频处理器
func NewNovelVideoProcessor(logger *zap.Logger) *NovelVideoProcessor {
	return &NovelVideoProcessor{
		logger:        logger,
		fileManager:   file.NewFileManager(),
		drawThingsGen: drawthings.NewChapterImageGenerator(logger),
	}
}

// ProcessNovel 完整处理小说的函数 - 从输入文本到最终图像
func (nvp *NovelVideoProcessor) ProcessNovel(novelPath string) error {
	nvp.logger.Info("开始处理小说", zap.String("novel_path", novelPath))

	// 步骤1: 读取输入小说文件
	content, err := ioutil.ReadFile(novelPath)
	if err != nil {
		return fmt.Errorf("读取小说文件失败: %v", err)
	}

	novelText := string(content)
	nvp.logger.Info("小说文本读取成功", zap.Int("字符数", len(novelText)))

	// 步骤2: 智能分章节
	nvp.logger.Info("开始智能分章节...")
	chapterTexts, err := nvp.fileManager.SplitNovelFileIntoChapters(novelPath)
	if err != nil {
		return fmt.Errorf("智能分章节失败: %v", err)
	}

	nvp.logger.Info("章节分割完成", zap.Int("章节数量", len(chapterTexts)))

	// 步骤3-6: 对每个章节执行完整处理流程
	for i, chapterContent := range chapterTexts {
		// 从章节内容中提取标题（取第一行或前几个字符作为标题）
		title := nvp.extractChapterTitle(chapterContent)
		if title == "" {
			title = fmt.Sprintf("章节_%02d", i+1)
		}

		nvp.logger.Info("处理章节", zap.Int("章节号", i+1), zap.String("章节标题", title))

		// 准备输出目录
		outputDir := filepath.Join("./output", "processed_novel", title)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			nvp.logger.Error("创建输出目录失败", zap.String("dir", outputDir), zap.Error(err))
			continue
		}

		// 步骤3: 生成音频 (此处为示意，实际需要调用TTS服务)
		nvp.logger.Info("章节音频生成 (待实现)", zap.String("chapter", title))
		// 这一步在实际环境中会调用TTS服务

		// 步骤4: 生成台词/字幕 (待实现)
		nvp.logger.Info("章节台词生成 (待实现)", zap.String("chapter", title))
		// 这一步在实际环境中会根据音频和文本生成字幕

		// 步骤5-6: 使用大模型分析场景并生成图像
		nvp.logger.Info("使用大模型分析场景并生成图像", zap.String("chapter", title))
		
		// 使用AI生成提示词并生成图像
		imageOutputDir := filepath.Join(outputDir, "images")
		if err := os.MkdirAll(imageOutputDir, 0755); err != nil {
			nvp.logger.Error("创建图像输出目录失败", zap.String("dir", imageOutputDir), zap.Error(err))
			continue
		}

		// 生成图像 - 使用AI分析场景生成提示词
		imageResults, err := nvp.drawThingsGen.GenerateImagesFromChapter(
			chapterContent, 
			imageOutputDir, 
			1024, 
			1792, 
			true, // 悬疑风格
		)
		if err != nil {
			nvp.logger.Error("生成章节图像失败", zap.String("chapter", title), zap.Error(err))
			// 尝试使用简化参数
			imageResults, err = nvp.drawThingsGen.GenerateImagesFromChapter(
				title+": "+chapterContent, 
				imageOutputDir, 
				512, 
				896, 
				true,
			)
			if err != nil {
				nvp.logger.Error("使用简化参数生成图像也失败", zap.String("chapter", title), zap.Error(err))
				continue
			}
		}

		nvp.logger.Info("章节图像生成完成", 
			zap.String("chapter", title), 
			zap.Int("生成图像数量", len(imageResults)))

		// 输出结果摘要
		for j, result := range imageResults {
			if j < 3 { // 只显示前3个结果
				nvp.logger.Info("生成的图像", 
					zap.String("image_file", result.ImageFile),
					zap.String("prompt_used", truncateString(result.ImagePrompt, 60)))
			}
		}

		// 添加延迟以避免API过载
		time.Sleep(1 * time.Second)
	}

	nvp.logger.Info("小说处理完成", zap.String("input", novelPath))
	return nil
}

// extractChapterTitle 从章节内容中提取标题
func (nvp *NovelVideoProcessor) extractChapterTitle(content string) string {
	lines := strings.SplitN(content, "\n", 3) // 只取前几行
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			// 如果行长度不太长，认为是标题
			if len(trimmed) < 50 && (strings.Contains(trimmed, "第") && (strings.Contains(trimmed, "章") || strings.Contains(trimmed, "节"))) {
				return trimmed
			}
		}
	}
	// 如果没找到合适的标题，返回空字符串
	return ""
}

// ProcessTextDirectly 直接处理输入文本的函数
func (nvp *NovelVideoProcessor) ProcessTextDirectly(text string, outputDir string) error {
	nvp.logger.Info("开始处理直接输入的文本", zap.Int("字符数", len(text)))

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	nvp.logger.Info("步骤1: 文本已接收", zap.Int("字符数", len(text)))

	// 智能分段（模拟章节分割）
	paragraphs := nvp.splitIntoScenes(text)
	nvp.logger.Info("步骤2: 文本已分段", zap.Int("段落数", len(paragraphs)))

	// 步骤3: 音频生成 (模拟)
	nvp.logger.Info("步骤3: 音频生成 (模拟)")

	// 步骤4: 台词生成 (模拟)
	nvp.logger.Info("步骤4: 台词生成 (模拟)")

	// 步骤5-6: 大模型分析场景并生成图像
	nvp.logger.Info("步骤5-6: 大模型分析场景并生成图像")
	
	imageOutputDir := filepath.Join(outputDir, "images")
	if err := os.MkdirAll(imageOutputDir, 0755); err != nil {
		return fmt.Errorf("创建图像输出目录失败: %v", err)
	}

	// 使用AI生成提示词并生成图像
	fullContent := strings.Join(paragraphs, "\n\n")
	imageResults, err := nvp.drawThingsGen.GenerateImagesFromChapter(
		fullContent,
		imageOutputDir,
		1024,
		1792,
		true, // 悬疑风格
	)
	if err != nil {
		nvp.logger.Warn("使用完整内容生成图像失败，尝试简化处理", zap.Error(err))
		
		// 尝试逐段处理
		for i, para := range paragraphs {
			if len(strings.TrimSpace(para)) < 10 { // 跳过太短的段落
				continue
			}
			
			imageFile := filepath.Join(imageOutputDir, fmt.Sprintf("scene_%03d.png", i+1))
			err := nvp.drawThingsGen.Client.GenerateImageFromText(
				para,
				imageFile,
				512,
				896,
				true,
			)
			if err != nil {
				nvp.logger.Warn("生成单个场景图像失败", zap.Int("scene", i+1), zap.Error(err))
				continue
			}
			nvp.logger.Info("单个场景图像生成成功", zap.String("file", imageFile))
		}
	} else {
		nvp.logger.Info("图像批量生成完成", zap.Int("生成数量", len(imageResults)))
	}

	nvp.logger.Info("完整处理流程完成", zap.String("output_dir", outputDir))
	return nil
}

// splitIntoScenes 将文本分割成适合生成图像的场景
func (nvp *NovelVideoProcessor) splitIntoScenes(text string) []string {
	// 按段落分割
	paragraphs := strings.Split(text, "\n\n")
	
	var scenes []string
	for _, para := range paragraphs {
		trimmed := strings.TrimSpace(para)
		if len(trimmed) > 0 {
			scenes = append(scenes, trimmed)
		}
	}
	
	// 如果段落太少或太长，进一步分割
	var finalScenes []string
	for _, scene := range scenes {
		if len(scene) > 500 { // 如果单个场景太长，按句子分割
			subScenes := nvp.splitLongScene(scene)
			finalScenes = append(finalScenes, subScenes...)
		} else {
			finalScenes = append(finalScenes, scene)
		}
	}
	
	return finalScenes
}

// splitLongScene 分割长场景
func (nvp *NovelVideoProcessor) splitLongScene(scene string) []string {
	var subScenes []string
	
	// 按句子分割（中文句号、英文句号、感叹号、问号）
	sentences := strings.FieldsFunc(scene, func(r rune) bool {
		return r == '。' || r == '.' || r == '!' || r == '！' || r == '?' || r == '？'
	})
	
	currentScene := ""
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}
		
		if len(currentScene)+len(sentence) < 300 { // 控制每个子场景的长度
			if currentScene != "" {
				currentScene += "。" + sentence
			} else {
				currentScene = sentence
			}
		} else {
			if currentScene != "" {
				subScenes = append(subScenes, currentScene)
			}
			currentScene = sentence
		}
	}
	
	if currentScene != "" {
		subScenes = append(subScenes, currentScene)
	}
	
	return subScenes
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func main() {
	// 创建logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("创建logger失败: %v\n", err)
		return
	}
	defer logger.Sync()

	// 创建处理器
	processor := NewNovelVideoProcessor(logger)

	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("使用方法:")
		fmt.Println("  go run main.go <input_text_file>                    # 处理小说文件")
		fmt.Println("  go run main.go \"直接输入的文本内容\"              # 直接处理文本")
		return
	}

	input := os.Args[1]

	// 判断是文件还是直接文本
	if _, err := os.Stat(input); err == nil {
		// 是文件
		fmt.Printf("开始处理小说文件: %s\n", input)
		if err := processor.ProcessNovel(input); err != nil {
			fmt.Printf("处理小说失败: %v\n", err)
			return
		}
	} else {
		// 是直接输入的文本
		fmt.Println("开始处理直接输入的文本...")
		outputDir := "./output/direct_processing_" + fmt.Sprintf("%d", time.Now().Unix())
		if err := processor.ProcessTextDirectly(input, outputDir); err != nil {
			fmt.Printf("处理文本失败: %v\n", err)
			return
		}
		fmt.Printf("处理完成，输出目录: %s\n", outputDir)
	}

	fmt.Println("所有处理完成！")
}