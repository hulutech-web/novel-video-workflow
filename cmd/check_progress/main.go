package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	outputDir := "./output/幽灵客栈/chapter_07"
	
	// 读取输出目录中的音频文件
	files, err := ioutil.ReadDir(outputDir)
	if err != nil {
		fmt.Printf("无法读取输出目录: %v\n", err)
		return
	}
	
	audioFiles := []string{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".wav") {
			audioFiles = append(audioFiles, file.Name())
		}
	}
	
	fmt.Printf("章节 07 处理进度:\n")
	fmt.Printf("- 已生成音频文件数量: %d 个\n", len(audioFiles))
	
	totalSize := int64(0)
	for _, file := range audioFiles {
		filePath := filepath.Join(outputDir, file)
		info, _ := os.Stat(filePath)
		totalSize += info.Size()
		fmt.Printf("  - %s (%.2f KB)\n", file, float64(info.Size())/1024)
	}
	
	fmt.Printf("- 总计大小: %.2f MB\n", float64(totalSize)/(1024*1024))
	
	// 读取原始章节文件统计
	inputPath := "./input/幽灵客栈/chapter_07/chapter_07.txt"
	content, err := ioutil.ReadFile(inputPath)
	if err == nil {
		paragraphs := strings.Split(string(content), "\n\n\n")
		fmt.Printf("- 原始段落数量: %d 个\n", len(paragraphs))
		fmt.Printf("- 已处理段落数量: %d 个\n", len(audioFiles))
		fmt.Printf("- 剩余待处理段落数量: %d 个\n", len(paragraphs)-len(audioFiles))
	}
}