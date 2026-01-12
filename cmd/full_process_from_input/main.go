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

func main() {
	// 创建logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("创建logger失败: %v\n", err)
		return
	}
	defer logger.Sync()

	fmt.Println("🚀 启动小说视频生成完整流程...")
	
	// 1. 从input目录读取内容
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

	processor := &FullProcessor{
		logger: logger,
		fileManager: file.NewFileManager(),
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

	fmt.Println("✅ 完整流程处理完成！")
}

// FullProcessor 完整流程处理器
type FullProcessor struct {
	logger        *zap.Logger
	fileManager   *file.FileManager
	drawThingsGen *drawthings.ChapterImageGenerator
}

// processNovelDirectory 处理小说目录
func (fp *FullProcessor) processNovelDirectory(novelDir string) error {
	fmt.Printf("📖 开始处理小说目录: %s\n", novelDir)
	
	// 查找目录下的所有文本文件
	files, err := filepath.Glob(filepath.Join(novelDir, "*.txt"))
	if err != nil {
		return fmt.Errorf("查找文本文件失败: %v", err)
	}

	if len(files) == 0 {
		// 尝试查找子目录中的文件
		subdirs, err := ioutil.ReadDir(novelDir)
		if err != nil {
			return fmt.Errorf("读取子目录失败: %v", err)
		}

		for _, subdir := range subdirs {
			if subdir.IsDir() {
				subdirPath := filepath.Join(novelDir, subdir.Name())
				subFiles, err := filepath.Glob(filepath.Join(subdirPath, "*.txt"))
				if err != nil {
					continue
				}
				files = append(files, subFiles...)
			}
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("在目录 %s 中未找到任何文本文件", novelDir)
	}

	for _, txtFile := range files {
		fmt.Printf("📄 处理文本文件: %s\n", txtFile)
		err := fp.processSingleTextFile(txtFile)
		if err != nil {
			fp.logger.Error("处理文本文件失败", zap.String("file", txtFile), zap.Error(err))
			continue
		}
	}

	return nil
}

// processSingleTextFile 处理单个文本文件
func (fp *FullProcessor) processSingleTextFile(txtFile string) error {
	fmt.Printf("📝 开始处理文本文件: %s\n", txtFile)

	// 步骤1: 读取输入文本
	content, err := ioutil.ReadFile(txtFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	text := string(content)
	fmt.Printf("✅ 步骤1 - 文本读取完成 (长度: %d 字符)\n", len(text))

	// 步骤2: 智能分章节
	fmt.Println("🔄 步骤2 - 开始智能分章节...")
	chapterFiles, err := fp.fileManager.SplitNovelFileIntoChapters(txtFile)
	if err != nil {
		fmt.Printf("⚠️  智能分章节失败，使用简单分段: %v\n", err)
		// 简单按段落分段
		chapterFiles = []string{txtFile}
	} else {
		fmt.Printf("✅ 智能分章节完成 (共 %d 章节文件)\n", len(chapterFiles))
	}

	// 步骤3-6: 对每个章节执行完整流程
	for i, chapterFile := range chapterFiles {
		fmt.Printf("🎬 处理章节文件 %d/%d: %s\n", i+1, len(chapterFiles), filepath.Base(chapterFile))

		// 读取章节内容
		chapterContent, err := ioutil.ReadFile(chapterFile)
		if err != nil {
			fp.logger.Error("读取章节文件失败", zap.String("file", chapterFile), zap.Error(err))
			continue
		}

		chapterTitle := filepath.Base(chapterFile)
		if strings.Contains(chapterTitle, ".") {
			chapterTitle = strings.Split(chapterTitle, ".")[0]
		}

		// 准备输出目录
		outputDir := filepath.Join("./output", 
			fmt.Sprintf("processed_%d", time.Now().Unix()), 
			strings.ReplaceAll(chapterTitle, " ", "_"))
		
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fp.logger.Error("创建输出目录失败", zap.String("dir", outputDir), zap.Error(err))
			continue
		}

		// 步骤3: 生成音频 (模拟 - 实际需要调用TTS服务)
		fmt.Println("🔊 步骤3 - 音频生成 (模拟)")
		// 在实际实现中，这里会调用TTS服务

		// 步骤4: 生成台词/字幕 (模拟 - 实际需要根据音频生成)
		fmt.Println("📜 步骤4 - 台词/字幕生成 (模拟)")
		// 在实际实现中，这里会根据音频和文本生成字幕

		// 步骤5-6: 使用大模型分析场景并生成图像
		fmt.Println("🎨 步骤5-6 - 大模型分析场景并生成图像")
		
		// 使用AI生成提示词并生成图像
		imageOutputDir := filepath.Join(outputDir, "images")
		if err := os.MkdirAll(imageOutputDir, 0755); err != nil {
			fp.logger.Error("创建图像输出目录失败", zap.String("dir", imageOutputDir), zap.Error(err))
			continue
		}

		// 使用AI生成提示词并生成图像
		imageResults, err := fp.drawThingsGen.GenerateImagesFromChapter(
			string(chapterContent), 
			imageOutputDir, 
			1024, 
			1792, 
			true, // 悬疑风格
		)
		if err != nil {
			fp.logger.Warn("使用AI生成图像失败，尝试简化处理", zap.Error(err))
			
			// 尝试使用简化参数
			err = fp.drawThingsGen.Client.GenerateImageFromText(
				chapterTitle+": "+string(chapterContent)[:min(len(string(chapterContent)), 200)], 
				filepath.Join(imageOutputDir, "chapter_image.png"), 
				512, 
				896, 
				true,
			)
			if err != nil {
				fp.logger.Error("简化处理也失败", zap.Error(err))
				continue
			} else {
				fmt.Println("✅ 简化处理图像生成成功")
			}
		} else {
			fmt.Printf("✅ AI图像生成完成 (生成 %d 张图像)\n", len(imageResults))
		}

		fmt.Printf("✅ 章节文件 %s 处理完成\n", filepath.Base(chapterFile))
	}

	fmt.Printf("✅ 文件 %s 处理完成\n", txtFile)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}