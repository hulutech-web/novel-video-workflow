package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"novel-video-workflow/pkg/tools/aegisub"
	drawthings "novel-video-workflow/pkg/tools/drawthings"
	"novel-video-workflow/pkg/tools/file"
	"novel-video-workflow/pkg/tools/indextts2"
)

func main() {
	fmt.Println("🚀 开始执行小说视频生成完整工作流...")
	
	// 执行自检
	unavailableServices := runSelfCheck()
	if len(unavailableServices) > 0 {
		fmt.Printf("⚠️  以下服务不可用，相关功能将被跳过: %v\n", unavailableServices)
		fmt.Println("⏳ 继续执行可用的工作流步骤...")
	} else {
		fmt.Println("✅ 所有服务均正常")
	}
	
	// 创建logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("创建logger失败: %v\n", err)
		return
	}
	defer logger.Sync()

	fmt.Println("✅ 自检完成，开始执行完整工作流...")
	
	// 从input目录读取内容
	inputDir := "./input"
	items, err := ioutil.ReadDir(inputDir)
	if err != nil {
		fmt.Printf("❌ 无法读取input目录: %v\n", err)
		return
	}

	if len(items) == 0 {
		fmt.Println("❌ input目录为空，请在input目录下放置小说文本文件")
		return
	}

	processor := &WorkflowProcessor{
		logger:        logger,
		fileManager:   file.NewFileManager(),
		ttsClient:     indextts2.NewIndexTTS2Client(logger, "http://localhost:7860"),
		aegisubGen:    aegisub.NewAegisubGenerator(),
		drawThingsGen: drawthings.NewChapterImageGenerator(logger),
	}

	// 处理input目录下的每个文件或子目录
	for _, item := range items {
		if item.IsDir() {
			// 处理目录中的所有文本文件
			novelDir := filepath.Join(inputDir, item.Name())
			err := processor.processNovelDirectory(novelDir)
			if err != nil {
				fmt.Printf("❌ 处理小说目录 %s 失败: %v\n", novelDir, err)
				continue
			}
		} else if strings.HasSuffix(item.Name(), ".txt") {
			// 处理单独的文本文件
			txtFile := filepath.Join(inputDir, item.Name())
			err := processor.processSingleTextFile(txtFile)
			if err != nil {
				fmt.Printf("❌ 处理文本文件 %s 失败: %v\n", txtFile, err)
				continue
			}
		}
	}

	fmt.Println("✅ 小说视频生成完整工作流执行完成！")
}

// runSelfCheck 执行自检程序
func runSelfCheck() []string {
	fmt.Println("🔍 执行自检程序...")
	
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("创建logger失败: %v\n", err)
		return []string{"logger"}
	}
	defer logger.Sync()

	// 检查各项服务
	serviceChecks := []struct {
		name string
		fn   func() error
	}{
		{"Ollama", checkOllama},
		{"DrawThings", func() error { return checkDrawThings(logger) }},
		{"IndexTTS2", checkIndexTTS2},
		{"Aegisub脚本", checkAegisub},
		{"参考音频文件", checkRefAudio},
	}

	var unavailableServices []string
	for _, check := range serviceChecks {
		fmt.Printf("  📋 检查%s...", check.name)
		if err := check.fn(); err != nil {
			fmt.Printf(" ❌ (%v)\n", err)
			unavailableServices = append(unavailableServices, check.name)
		} else {
			fmt.Printf(" ✅\n")
		}
	}

	if len(unavailableServices) > 0 {
		fmt.Printf("⚠️  以下服务不可用: %v\n", unavailableServices)
	} else {
		fmt.Println("✅ 所有服务均正常")
	}
	
	return unavailableServices
}

// checkOllama 检查Ollama服务
func checkOllama() error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://localhost:11434/api/tags")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("状态码: %d", resp.StatusCode)
	}
	
	return nil
}

// checkDrawThings 检查DrawThings服务
func checkDrawThings(logger *zap.Logger) error {
	client := drawthings.NewDrawThingsClient(logger, "http://localhost:7861")
	if !client.APIAvailable {
		return fmt.Errorf("DrawThings API不可用")
	}
	return nil
}

// checkIndexTTS2 检查IndexTTS2服务
func checkIndexTTS2() error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://localhost:7860")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("状态码: %d", resp.StatusCode)
	}
	
	return nil
}

// checkAegisub 检查Aegisub脚本
func checkAegisub() error {
	gen := aegisub.NewAegisubGenerator()
	if _, err := os.Stat(gen.ScriptPath); os.IsNotExist(err) {
		return err
	}
	return nil
}

// checkRefAudio 检查参考音频文件
func checkRefAudio() error {
	paths := []string{
		"./assets/ref_audio/ref.m4a",
		"./assets/ref_audio/音色.m4a",
		"/Users/mac/code/ai/novel-video-workflow/assets/ref_audio/ref.m4a",
	}
	
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			// 检查文件大小确保不是空文件
			info, err := os.Stat(path)
			if err != nil {
				continue
			}
			if info.Size() > 1024 { // 确保文件至少有1KB
				return nil
			}
		}
	}
	
	return fmt.Errorf("未找到有效的参考音频文件")
}

// WorkflowProcessor 工作流处理器
type WorkflowProcessor struct {
	logger        *zap.Logger
	fileManager   *file.FileManager
	ttsClient     *indextts2.IndexTTS2Client
	aegisubGen    *aegisub.AegisubGenerator
	drawThingsGen *drawthings.ChapterImageGenerator
}

// processNovelDirectory 处理小说目录
func (wp *WorkflowProcessor) processNovelDirectory(novelDir string) error {
	fmt.Printf("📖 开始处理小说目录: %s\n", novelDir)

	// 从目录名提取小说名称
	novelName := filepath.Base(novelDir)
	fmt.Printf("小说名称: %s\n", novelName)

	// 查找目录下的所有文本文件
	var files []string
	
	// 首先检查子目录中的章节文件
	entries, err := ioutil.ReadDir(novelDir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "chapter_") {
			// 处理章节子目录
			chapterDir := filepath.Join(novelDir, entry.Name())
			chapterFiles, err := filepath.Glob(filepath.Join(chapterDir, "*.txt"))
			if err != nil {
				continue
			}
			files = append(files, chapterFiles...)
		} else if strings.HasSuffix(entry.Name(), ".txt") {
			// 直接添加根目录下的txt文件
			files = append(files, filepath.Join(novelDir, entry.Name()))
		}
	}

	// 如果没有在子目录中找到文件，则查找根目录下的txt文件
	if len(files) == 0 {
		files, err = filepath.Glob(filepath.Join(novelDir, "*.txt"))
		if err != nil {
			return fmt.Errorf("查找文本文件失败: %v", err)
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("在目录 %s 中未找到任何文本文件", novelDir)
	}

	// 为小说创建输出目录
	outputBaseDir := filepath.Join("./output", novelName)
	if err := os.MkdirAll(outputBaseDir, 0755); err != nil {
		return fmt.Errorf("创建小说输出目录失败: %v", err)
	}

	for _, txtFile := range files {
		fmt.Printf("📄 处理文本文件: %s\n", txtFile)
		err := wp.processSingleTextFileWithNovelName(txtFile, novelName)
		if err != nil {
			wp.logger.Error("处理文本文件失败", zap.String("file", txtFile), zap.Error(err))
			continue
		}
	}

	return nil
}

// processSingleTextFileWithNovelName 处理单个文本文件，带小说名称
func (wp *WorkflowProcessor) processSingleTextFileWithNovelName(txtFile string, novelName string) error {
	fmt.Printf("📝 开始处理文本文件: %s\n", txtFile)

	// 步骤1: 读取输入文本
	content, err := ioutil.ReadFile(txtFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	text := string(content)
	fmt.Printf("✅ 步骤1 - 文本读取完成 (长度: %d 字符)\n", len(text))

	// 确定章节号 - 从文件路径或内容中提取，优先从目录名提取以保持与input一致
	chapterNum := wp.extractChapterNumberFromPath(txtFile)
	
	// 从文件路径中提取真实的章节名（与input保持一致）
	chapterDir := filepath.Dir(txtFile)
	realChapterName := filepath.Base(chapterDir)
	
	// 如果目录名是chapter_xx格式，则使用它，否则使用提取的数字
	if strings.HasPrefix(realChapterName, "chapter_") {
		// 从realChapterName中提取数字
		numStr := strings.TrimPrefix(realChapterName, "chapter_")
		var extractedNum int
		// 处理可能包含前导零的数字，如chapter_07
		if strings.HasPrefix(numStr, "_") {
			numStr = strings.TrimPrefix(numStr, "_")
		}
		// 尝试解析数字，先去掉可能的前导零
		if _, err := fmt.Sscanf(numStr, "%d", &extractedNum); err != nil {
			// 如果解析失败，尝试直接转换
			if n, parseErr := fmt.Sscanf(strings.TrimLeft(numStr, "0"), "%d", &extractedNum); parseErr != nil || n == 0 {
				// 如果仍然失败，尝试使用strconv.Atoi
				if num, convErr := strconv.Atoi(strings.TrimLeft(numStr, "0")); convErr == nil {
					extractedNum = num
				}
			}
		}
		if extractedNum > 0 {
			chapterNum = extractedNum
		}
	}
	
	// 为小说创建输出目录结构
	outputDir := filepath.Join("./output", novelName, fmt.Sprintf("chapter_%02d", chapterNum))
	
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		wp.logger.Error("创建输出目录失败", zap.String("dir", outputDir), zap.Error(err))
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 读取章节内容
	chapterContent, err := ioutil.ReadFile(txtFile)
	if err != nil {
		wp.logger.Error("读取章节文件失败", zap.String("file", txtFile), zap.Error(err))
		return fmt.Errorf("读取章节文件失败: %v", err)
	}

	chapterTitle := filepath.Base(txtFile)
	if strings.Contains(chapterTitle, ".") {
		chapterTitle = strings.Split(chapterTitle, ".")[0]
	}

	// 步骤2: 生成音频
	fmt.Println("🔊 步骤2 - 生成音频...")
	audioFile := filepath.Join(outputDir, fmt.Sprintf("chapter_%02d.wav", chapterNum))
	
	// 使用参考音频文件 - 按照用户提供的路径
	refAudioPath := "./assets/ref_audio/ref.m4a"
	if _, err := os.Stat(refAudioPath); os.IsNotExist(err) {
		// 如果默认路径不存在，尝试其他可能的路径
		possiblePaths := []string{
			"./assets/ref_audio/ref.m4a",
			"./assets/ref_audio/音色.m4a",
			"./ref.m4a",
			"./音色.m4a",
			"/Users/mac/code/ai/novel-video-workflow/assets/ref_audio/ref.m4a",
		}
		
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				refAudioPath = path
				break
			}
		}
	}
	
	if _, err := os.Stat(refAudioPath); os.IsNotExist(err) {
		fmt.Printf("⚠️  未找到参考音频文件，跳过音频生成\n")
	} else {
		err = wp.ttsClient.GenerateTTSWithAudio(refAudioPath, string(chapterContent), audioFile)
		if err != nil {
			wp.logger.Warn("生成音频失败", zap.String("chapter", chapterTitle), zap.Error(err))
			fmt.Printf("⚠️  音频生成失败: %v\n", err)
		} else {
			fmt.Printf("✅ 音频生成完成: %s\n", audioFile)
			// 显式关闭IndexTTS2客户端连接
			if wp.ttsClient.HTTPClient != nil {
				wp.ttsClient.HTTPClient.CloseIdleConnections()
			}
		}
	}

	// 步骤3: 生成台词/字幕
	fmt.Println("📜 步骤3 - 生成台词/字幕...")
	subtitleFile := filepath.Join(outputDir, fmt.Sprintf("chapter_%02d.srt", chapterNum))
	
	if _, err := os.Stat(audioFile); err == nil {
		// 如果音频文件存在，生成字幕
		err = wp.aegisubGen.GenerateSubtitleFromIndextts2Audio(audioFile, string(chapterContent), subtitleFile)
		if err != nil {
			wp.logger.Warn("生成字幕失败", zap.String("chapter", chapterTitle), zap.Error(err))
			fmt.Printf("⚠️  字幕生成失败: %v\n", err)
		} else {
			fmt.Printf("✅ 字幕生成完成: %s\n", subtitleFile)
		}
	} else {
		fmt.Printf("⚠️  由于音频文件不存在，跳过字幕生成\n")
	}

	// 步骤4: 生成图像 (使用缩小的像素和Ollama优化的提示词)
	fmt.Println("🎨 步骤4 - 生成图像...")
	imagesDir := filepath.Join(outputDir, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		wp.logger.Error("创建图像目录失败", zap.String("dir", imagesDir), zap.Error(err))
		return fmt.Errorf("创建图像目录失败: %v", err)
	}
	
	// 估算音频时长用于分镜生成
	estimatedAudioDuration := 0
	if _, statErr := os.Stat(audioFile); statErr == nil {
		// 基于音频文件大小估算时长（这是一个近似值，更准确的方法需要音频处理库）
		// 通常WAV文件: 大约每秒 176,400 字节 (44.1kHz * 16位 * 2声道)
		// 但我们的音频可能有不同的参数，这里使用一个大致的估算
		if fileInfo, err := os.Stat(audioFile); err == nil {
			fileSizeMB := float64(fileInfo.Size()) / (1024 * 1024)
			// 假设平均 1MB ≈ 10秒音频
			estimatedAudioDuration = int(fileSizeMB * 10)
			if estimatedAudioDuration < 30 { // 最少30秒
				estimatedAudioDuration = 30
			}
		}
	} else {
		// 如果没有音频文件，基于文本长度估算
		estimatedAudioDuration = len(string(chapterContent)) * 2 / 10 // 每个字符约0.2秒
		if estimatedAudioDuration < 60 { // 最少1分钟
			estimatedAudioDuration = 60
		}
	}
	
	// 使用Ollama优化的提示词生成图像
	err = wp.generateImagesWithOllamaPrompts(string(chapterContent), imagesDir, chapterNum, estimatedAudioDuration)
	if err != nil {
		wp.logger.Warn("生成图像失败", zap.Error(err))
		fmt.Printf("⚠️  图像生成失败: %v\n", err)
	} else {
		fmt.Printf("✅ 图像生成完成，保存在: %s\n", imagesDir)
	}

	fmt.Printf("✅ 章节文件 %s 处理完成\n", filepath.Base(txtFile))
	return nil
}

// processSingleTextFile 处理单个文本文件（保持原有方法用于兼容性）
func (wp *WorkflowProcessor) processSingleTextFile(txtFile string) error {
	// 从文件路径推断小说名称
	dir := filepath.Dir(txtFile)
	novelName := filepath.Base(dir)
	
	// 如果文件直接在input目录下，则从小说文件名推断
	if novelName == "input" {
		novelName = strings.TrimSuffix(filepath.Base(txtFile), ".txt")
	}
	
	return wp.processSingleTextFileWithNovelName(txtFile, novelName)
}

// extractChapterNumberFromPath 从文件路径中提取章节号
func (wp *WorkflowProcessor) extractChapterNumberFromPath(filePath string) int {
	// 先尝试从父级目录名提取章节号
	dir := filepath.Dir(filePath)
	baseDir := filepath.Base(dir)
	
	if strings.HasPrefix(baseDir, "chapter_") {
		numStr := strings.TrimPrefix(baseDir, "chapter_")
		// 去除可能的前导下划线
		numStr = strings.TrimPrefix(numStr, "_")
		
		var num int
		// 处理可能包含前导零的数字，如chapter_07
		if strings.HasPrefix(numStr, "_") {
			numStr = strings.TrimPrefix(numStr, "_")
		}
		// 尝试解析数字，先去掉可能的前导零
		if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil {
			// 如果解析失败，尝试直接转换
			if n, parseErr := fmt.Sscanf(strings.TrimLeft(numStr, "0"), "%d", &num); parseErr != nil || n == 0 {
				// 如果仍然失败，尝试使用strconv.Atoi
				if num2, convErr := strconv.Atoi(strings.TrimLeft(numStr, "0")); convErr == nil {
					num = num2
				} else {
					// 如果数字解析失败，尝试中文数字转换
					num = wp.fileManager.ConvertChineseNumberToArabic(numStr)
				}
			}
		}
		return num
	}
	
	// 如果目录名不是chapter_x格式，尝试从文件名提取
	baseFile := filepath.Base(filePath)
	nameWithoutExt := strings.TrimSuffix(baseFile, ".txt")
	
	if strings.HasPrefix(nameWithoutExt, "chapter_") {
		numStr := strings.TrimPrefix(nameWithoutExt, "chapter_")
		numStr = strings.TrimPrefix(numStr, "_")
		
		var num int
		// 处理可能包含前导零的数字，如chapter_07
		if strings.HasPrefix(numStr, "_") {
			numStr = strings.TrimPrefix(numStr, "_")
		}
		// 尝试解析数字，先去掉可能的前导零
		if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil {
			// 如果解析失败，尝试直接转换
			if n, parseErr := fmt.Sscanf(strings.TrimLeft(numStr, "0"), "%d", &num); parseErr != nil || n == 0 {
				// 如果仍然失败，尝试使用strconv.Atoi
				if num2, convErr := strconv.Atoi(strings.TrimLeft(numStr, "0")); convErr == nil {
					num = num2
				} else {
					// 如果数字解析失败，尝试中文数字转换
					num = wp.fileManager.ConvertChineseNumberToArabic(numStr)
				}
			}
		}
		return num
	}
	
	// 如果都没找到，返回默认值1
	return 1
}

// generateImagesWithOllamaPrompts 使用Ollama优化的提示词生成图像
func (wp *WorkflowProcessor) generateImagesWithOllamaPrompts(content, imagesDir string, chapterNum int, audioDurationSecs int) error {
	// 使用Ollama分析整个章节内容并生成分镜提示词
	styleDesc := "悬疑惊悚风格，周围环境模糊成黑影, 空气凝滞,浅景深, 胶片颗粒感, 低饱和度，极致悬疑氛围, 阴沉窒息感, 夏季，环境阴霾，其他部分模糊不可见"
	
	// 使用实际音频时长，如果未提供则估算
	estimatedDurationSecs := audioDurationSecs
	if estimatedDurationSecs <= 0 {
		// 估算音频时长（假设每分钟300字，即每个字符约0.2秒）
		estimatedDurationSecs = len(content) * 2 / 10 // 简化估算，大约每个字符0.2秒
		if estimatedDurationSecs < 60 { // 最少1分钟
			estimatedDurationSecs = 60
		}
	}
	
	// 让Ollama分析整个章节并生成分镜
	wp.logger.Info("开始Ollama分镜分析", zap.Int("chapter_num", chapterNum), zap.Int("content_length", len(content)), zap.Int("estimated_duration_secs", estimatedDurationSecs))
	sceneDescriptions, err := wp.drawThingsGen.OllamaClient.AnalyzeScenesAndGeneratePrompts(content, styleDesc, estimatedDurationSecs)
	if err != nil {
		wp.logger.Warn("使用Ollama分析场景并生成分镜提示词失败",
			zap.Error(err))
		
		// 如果Ollama场景分析失败，回退到原来的段落处理方式
		wp.logger.Info("Ollama分镜分析失败，回退到段落处理方式")
		paragraphs := wp.splitChapterIntoParagraphsWithMerge(content)
		
		for idx, paragraph := range paragraphs {
			if strings.TrimSpace(paragraph) == "" {
				continue
			}
			
			optimizedPrompt, err := wp.drawThingsGen.OllamaClient.GenerateImagePrompt(paragraph, styleDesc)
			if err != nil {
				wp.logger.Warn("使用Ollama生成图像提示词失败，使用原始文本",
					zap.Int("paragraph_index", idx),
					zap.String("paragraph", paragraph),
					zap.Error(err))
				optimizedPrompt = paragraph + ", 周围环境模糊成黑影, 空气凝滞,浅景深, 胶片颗粒感, 低饱和度，极致悬疑氛围, 阴沉窒息感, 夏季，环境阴霾，其他部分模糊不可见"
			}
			
			imageFile := filepath.Join(imagesDir, fmt.Sprintf("paragraph_%02d.png", idx+1))
			
			err = wp.drawThingsGen.Client.GenerateImageFromText(
				optimizedPrompt,
				imageFile, 
				512, // 缩小宽度
				896, // 缩小高度
				false, // 风格已在提示词中处理
			)
			if err != nil {
				wp.logger.Warn("生成图像失败", zap.String("paragraph", paragraph[:min(len(paragraph), 50)]), zap.Error(err))
				fmt.Printf("⚠️  段落图像生成失败: %v\n", err)
			} else {
				fmt.Printf("✅ 段落图像生成完成: %s\n", imageFile)
			}
		}
		
		return nil
	}
	
	// 如果Ollama分镜分析成功，使用生成的分镜描述生成图像
	wp.logger.Info("Ollama分镜分析成功", zap.Int("scene_count", len(sceneDescriptions)))
	for idx, sceneDesc := range sceneDescriptions {
		imageFile := filepath.Join(imagesDir, fmt.Sprintf("scene_%02d.png", idx+1))
		
		// 使用分镜描述生成图像
		err = wp.drawThingsGen.Client.GenerateImageFromText(
			sceneDesc,
			imageFile, 
			512, // 缩小宽度
			896, // 缩小高度
			false, // 风格已在提示词中处理
		)
		if err != nil {
			wp.logger.Warn("生成分镜图像失败", zap.String("scene", sceneDesc[:min(len(sceneDesc), 50)]), zap.Error(err))
			fmt.Printf("⚠️  分镜图像生成失败: %v\n", err)
		} else {
			fmt.Printf("✅ 分镜图像生成完成: %s\n", imageFile)
		}
	}
	
	return nil
}

// splitChapterIntoParagraphsWithMerge 将章节文本分割为段落，并对短段落进行合并
func (wp *WorkflowProcessor) splitChapterIntoParagraphsWithMerge(text string) []string {
	// 按换行符分割文本
	lines := strings.Split(text, "\n")

	var rawParagraphs []string
	var currentParagraph strings.Builder

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			// 遇到空行，结束当前段落
			if currentParagraph.Len() > 0 {
				rawParagraphs = append(rawParagraphs, strings.TrimSpace(currentParagraph.String()))
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
		rawParagraphs = append(rawParagraphs, strings.TrimSpace(currentParagraph.String()))
	}

	// 合并短段落
	var mergedParagraphs []string
	minLength := 50 // 设定最小长度阈值，低于此值的段落将与相邻段落合并
	
	for i := 0; i < len(rawParagraphs); i++ {
		currentPara := rawParagraphs[i]
		
		// 如果当前段落太短，考虑与下一个段落合并
		if len(currentPara) < minLength && i < len(rawParagraphs)-1 {
			// 与下一个段落合并
			merged := currentPara + " " + rawParagraphs[i+1]
			mergedParagraphs = append(mergedParagraphs, merged)
			i++ // 跳过下一个段落，因为它已经被合并了
		} else {
			// 检查是否当前段落太短但已经是最后一段
			if len(currentPara) < minLength && len(mergedParagraphs) > 0 {
				// 将其与前一段落合并
				lastIdx := len(mergedParagraphs) - 1
				mergedParagraphs[lastIdx] = mergedParagraphs[lastIdx] + " " + currentPara
			} else {
				// 添加正常段落
				mergedParagraphs = append(mergedParagraphs, currentPara)
			}
		}
	}

	// 过滤掉过短的段落（比如只有标点符号）
	var filtered []string
	for _, para := range mergedParagraphs {
		// 只保留非空且有一定长度的段落
		if len(strings.TrimSpace(para)) > 3 { // 至少3个字符
			filtered = append(filtered, para)
		}
	}

	return filtered
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}