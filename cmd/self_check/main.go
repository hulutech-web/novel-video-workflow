package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	drawthings "novel-video-workflow/pkg/tools/drawthings"
	"novel-video-workflow/pkg/tools/aegisub"
)

func main() {
	fmt.Println("🔍 开始执行自检程序...")
	
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("❌ 创建logger失败: %v\n", err)
		os.Exit(1)
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

	allPassed := true
	for _, check := range serviceChecks {
		fmt.Printf("  📋 检查%s...", check.name)
		if err := check.fn(); err != nil {
			fmt.Printf(" ❌ (%v)\n", err)
			allPassed = false
		} else {
			fmt.Printf(" ✅\n")
		}
	}

	if !allPassed {
		fmt.Println("❌ 自检失败，存在服务不可用的情况")
		os.Exit(1)
	}

	fmt.Println("✅ 所有服务均正常，可以开始执行完整工作流")
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