package main

import (
	"fmt"
	"log"
	"novel-video-workflow/pkg/capcut"
)

func main() {
	// 创建 CapcutGenerator 实例
	generator := capcut.NewCapcutGenerator(nil)

	// 使用真实的项目数据
	inputDir := "/Users/mac/code/ai/novel-video-workflow/output/幽灵客栈/chapter_07"
	
	fmt.Printf("开始生成剪映项目，输入目录: %s\n", inputDir)
	
	// 使用原始实现生成项目
	err := generator.GenerateProject(inputDir)
	if err != nil {
		log.Printf("生成项目失败: %v", err)
		return
	}

	fmt.Println("剪映项目已成功生成并导入!")
}