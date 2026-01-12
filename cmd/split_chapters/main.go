package main

import (
	"fmt"
	"os"
	"path/filepath"

	"novel-video-workflow/pkg/tools/file"
)

func main() {
	fmt.Println("🔄 开始拆分小说章节...")

	// 创建FileManager实例
	fm := file.NewFileManager()

	// 指定要拆分的小说文件路径
	novelFilePath := "./input/幽灵客栈/幽灵客栈.txt"

	// 检查文件是否存在
	if _, err := os.Stat(novelFilePath); os.IsNotExist(err) {
		fmt.Printf("❌ 小说文件不存在: %s\n", novelFilePath)
		return
	}

	// 调用章节拆分功能
	createdFiles, err := fm.SplitNovelFileIntoChapters(novelFilePath)
	if err != nil {
		fmt.Printf("❌ 拆分小说章节失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 章节拆分完成！创建了 %d 个章节文件\n", len(createdFiles))
	
	for _, file := range createdFiles {
		fmt.Printf("📄 %s\n", file)
	}

	// 验证拆分结果
	fmt.Println("\n📋 验证拆分结果:")
	inputDir := "./input/幽灵客栈"
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		fmt.Printf("❌ 读取输入目录失败: %v\n", err)
		return
	}

	chapterCount := 0
	for _, entry := range entries {
		if entry.IsDir() && filepath.HasPrefix(entry.Name(), "chapter_") {
			chapterCount++
			chapterDir := filepath.Join(inputDir, entry.Name())
			chapterTxt := filepath.Join(chapterDir, entry.Name()+".txt")
			
			if _, err := os.Stat(chapterTxt); err == nil {
				fmt.Printf("✅ 章节目录: %s (包含 %s)\n", chapterDir, filepath.Base(chapterTxt))
			} else {
				fmt.Printf("⚠️  章节目录: %s (缺少 %s)\n", chapterDir, filepath.Base(chapterTxt))
			}
		}
	}

	fmt.Printf("\n📊 总计: %d 个章节目录\n", chapterCount)
}