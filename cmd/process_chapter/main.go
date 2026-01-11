package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"novel-video-workflow/pkg/tools/indextts2"
)

func main() {
	// 定义路径 - 修改为处理chapter_08
	inputChapterPath := "./input/幽灵客栈/chapter_08/chapter_08.txt"
	referenceAudioPath := "./assets/ref_audio/ref.m4a"
	outputAudioPath := "./output/幽灵客栈/chapter_08/audio.wav"

	// 读取章节内容
	fmt.Println("正在读取章节内容...")
	content, err := ioutil.ReadFile(inputChapterPath)
	if err != nil {
		log.Fatalf("无法读取章节文件: %v", err)
	}

	chapterText := string(content)
	fmt.Printf("成功读取章节内容，长度: %d 字符\n", len(chapterText))

	// 确保输出目录存在
	outputDir := filepath.Dir(outputAudioPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("无法创建输出目录: %v", err)
	}

	// 初始化日志
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("无法初始化日志: %v", err)
	}
	defer logger.Sync()

	// 使用Indextts2客户端直接生成音频
	client := indextts2.NewIndexTTS2Client(logger, "http://localhost:7860")

	fmt.Println("正在调用Indextts2 API生成音频...")
	fmt.Printf("参考音频路径: %s\n", referenceAudioPath)
	fmt.Printf("输出音频路径: %s\n", outputAudioPath)
	
	// 显示进度提示
	fmt.Printf("注意：由于文本较长（%d 字符），TTS生成可能需要较长时间，请耐心等待...\n", len(chapterText))
	
	err = client.GenerateTTSWithAudio(referenceAudioPath, chapterText, outputAudioPath)
	if err != nil {
		log.Fatalf("调用Indextts2 API失败: %v", err)
	}

	fmt.Printf("音频文件已成功生成: %s\n", outputAudioPath)
	
	// 验证文件是否真的存在
	if _, err := os.Stat(outputAudioPath); os.IsNotExist(err) {
		log.Fatalf("生成的音频文件不存在: %s", outputAudioPath)
	}
	
	// 获取文件大小
	if fileInfo, err := os.Stat(outputAudioPath); err == nil {
		fmt.Printf("音频文件大小: %d 字节 (%.2f MB)\n", fileInfo.Size(), float64(fileInfo.Size())/(1024*1024))
	}
	
	fmt.Println("TTS生成完成！")
}