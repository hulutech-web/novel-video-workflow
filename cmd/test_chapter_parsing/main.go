package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.uber.org/zap"
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

	fmt.Println("🧪 开始测试章节编号解析功能...")

	// 创建FileManager实例
	fm := file.NewFileManager()

	// 读取输入目录中的小说
	inputDir := "./input"
	items, err := os.ReadDir(inputDir)
	if err != nil {
		fmt.Printf("❌ 无法读取input目录: %v\n", err)
		return
	}

	if len(items) == 0 {
		fmt.Println("❌ input目录为空，请在input目录下放置小说文本文件")
		return
	}

	// 处理input目录下的每个文件或子目录
	for _, item := range items {
		if item.IsDir() {
			// 处理目录中的所有文本文件
			novelDir := filepath.Join(inputDir, item.Name())
			err := processNovelDirectory(fm, novelDir, item.Name())
			if err != nil {
				fmt.Printf("❌ 处理小说目录 %s 失败: %v\n", novelDir, err)
				continue
			}
		} else if strings.HasSuffix(item.Name(), ".txt") {
			// 处理单独的文本文件
			txtFile := filepath.Join(inputDir, item.Name())
			err := processSingleTextFile(fm, txtFile)
			if err != nil {
				fmt.Printf("❌ 处理文本文件 %s 失败: %v\n", txtFile, err)
				continue
			}
		}
	}

	fmt.Println("✅ 章节编号解析功能测试完成！")
}

// processNovelDirectory 处理小说目录
func processNovelDirectory(fm *file.FileManager, novelDir string, novelName string) error {
	fmt.Printf("📖 开始处理小说目录: %s\n", novelDir)

	// 查找目录下的所有文本文件
	var files []string

	// 首先检查子目录中的章节文件
	entries, err := os.ReadDir(novelDir)
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

	for _, txtFile := range files {
		fmt.Printf("📄 处理文本文件: %s\n", txtFile)
		err := processSingleTextFile(fm, txtFile)
		if err != nil {
			return fmt.Errorf("处理文本文件失败: %v", err)
		}
	}

	return nil
}

// processSingleTextFile 处理单个文本文件
func processSingleTextFile(fm *file.FileManager, txtFile string) error {
	fmt.Printf("📝 开始处理文本文件: %s\n", txtFile)

	// 步骤1: 读取输入文本
	content, err := os.ReadFile(txtFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	text := string(content)
	fmt.Printf("✅ 步骤1 - 文本读取完成 (长度: %d 字符)\n", len(text))

	// 步骤2: 尝试从文本中提取章节号
	chapterNum := fm.ExtractChapterNumber(text)
	fmt.Printf("🔢 提取到的章节号: %d\n", chapterNum)

	// 验证章节号是否正确
	if chapterNum == 0 {
		fmt.Println("⚠️  未能从文本中提取到有效的章节号")
	} else {
		fmt.Printf("✅ 章节号解析成功: %d -> 格式化为: chapter_%02d\n", chapterNum, chapterNum)
	}

	// 步骤3: 演示如何从文件路径中提取章节号
	chapterNumFromPath := extractChapterNumberFromPath(txtFile)
	fmt.Printf("🔢 从文件路径 %s 提取到的章节号: %d\n", txtFile, chapterNumFromPath)

	// 验证路径提取的章节号是否与内容提取的一致
	if chapterNum == chapterNumFromPath {
		fmt.Println("✅ 文件内容和路径章节号匹配")
	} else {
		fmt.Println("⚠️  文件内容和路径章节号不匹配")
	}

	// 步骤4: 展示使用两位数格式的输出目录
	outputDir := filepath.Join("./output", "test", fmt.Sprintf("chapter_%02d", chapterNum))
	fmt.Printf("📁 输出目录格式化为: %s\n", outputDir)

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	fmt.Printf("✅ 章节 %d 处理完成\n", chapterNum)
	return nil
}

// extractChapterNumberFromPath 从文件路径中提取章节号
func extractChapterNumberFromPath(filePath string) int {
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
				}
			}
		}
		return num
	}

	// 如果都没找到，返回默认值1
	return 1
}