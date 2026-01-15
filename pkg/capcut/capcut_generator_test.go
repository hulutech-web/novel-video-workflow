package capcut

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCapcutGeneratorRealData 测试 CapcutGenerator 使用真实数据
func TestCapcutGeneratorRealData(t *testing.T) {
	// 检查测试数据是否存在
	testDir := "/Users/mac/code/ai/novel-video-workflow/output/幽灵客栈/chapter_07"
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skipf("跳过测试: 目录 %s 不存在", testDir)
	}

	// 验证必要文件是否存在
	audioFile := filepath.Join(testDir, "chapter_07.wav")
	imageFile := filepath.Join(testDir, "scene_01.png")
	srtFile := filepath.Join(testDir, "chapter_07.srt")

	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		t.Skipf("跳过测试: 音频文件 %s 不存在", audioFile)
	}

	if _, err := os.Stat(imageFile); os.IsNotExist(err) {
		t.Skipf("跳过测试: 图片文件 %s 不存在", imageFile)
	}

	if _, err := os.Stat(srtFile); os.IsNotExist(err) {
		t.Logf("警告: 字幕文件 %s 不存在，将不测试字幕功能", srtFile)
	}

	// 创建 CapcutGenerator 实例
	generator := NewCapcutGenerator(nil)

	// 测试生成项目功能 - 直接导入到剪映
	err := generator.GenerateProject(testDir)
	if err != nil {
		t.Errorf("使用真实数据生成剪映项目失败: %v", err)
	}
}

// TestCapcutGeneratorBasic 测试 CapcutGenerator 的基本功能
func TestCapcutGeneratorBasic(t *testing.T) {
	// 创建临时测试目录
	tempDir := t.TempDir()
	
	// 创建模拟输入文件
	inputDir := filepath.Join(tempDir, "input")
	err := os.MkdirAll(inputDir, 0755)
	if err != nil {
		t.Fatalf("创建输入目录失败: %v", err)
	}

	// 创建模拟音频文件 (实际上只是一个空文件用于测试)
	audioFile := filepath.Join(inputDir, "test.mp3")
	err = os.WriteFile(audioFile, []byte("dummy audio content"), 0644)
	if err != nil {
		t.Fatalf("创建音频文件失败: %v", err)
	}

	// 创建模拟图片文件
	imageFile := filepath.Join(inputDir, "test.jpg")
	err = os.WriteFile(imageFile, []byte("dummy image content"), 0644)
	if err != nil {
		t.Fatalf("创建图片文件失败: %v", err)
	}

	// 创建模拟字幕文件
	srtFile := filepath.Join(inputDir, "test.srt")
	srtContent := `1
00:00:01,000 --> 00:00:03,000
这是第一行字幕

2
00:00:04,000 --> 00:00:06,000
这是第二行字幕`
	err = os.WriteFile(srtFile, []byte(srtContent), 0644)
	if err != nil {
		t.Fatalf("创建字幕文件失败: %v", err)
	}

	// 创建 CapcutGenerator 实例
	generator := NewCapcutGenerator(nil)

	// 测试生成项目功能 - 直接导入到剪映
	err = generator.GenerateProject(inputDir)
	if err != nil {
		t.Errorf("生成剪映项目失败: %v", err)
	}
}

// TestCapcutGeneratorWithOutputDir 测试带输出目录的生成功能
func TestCapcutGeneratorWithOutputDir(t *testing.T) {
	// 创建临时测试目录
	tempDir := t.TempDir()
	
	// 创建模拟输入文件
	inputDir := filepath.Join(tempDir, "input")
	err := os.MkdirAll(inputDir, 0755)
	if err != nil {
		t.Fatalf("创建输入目录失败: %v", err)
	}

	// 创建模拟音频文件 (实际上只是一个空文件用于测试)
	audioFile := filepath.Join(inputDir, "test.mp3")
	err = os.WriteFile(audioFile, []byte("dummy audio content"), 0644)
	if err != nil {
		t.Fatalf("创建音频文件失败: %v", err)
	}

	// 创建模拟图片文件
	imageFile := filepath.Join(inputDir, "test.jpg")
	err = os.WriteFile(imageFile, []byte("dummy image content"), 0644)
	if err != nil {
		t.Fatalf("创建图片文件失败: %v", err)
	}

	// 创建模拟字幕文件
	srtFile := filepath.Join(inputDir, "test.srt")
	srtContent := `1
00:00:01,000 --> 00:00:03,000
这是第一行字幕

2
00:00:04,000 --> 00:00:06,000
这是第二行字幕`
	err = os.WriteFile(srtFile, []byte(srtContent), 0644)
	if err != nil {
		t.Fatalf("创建字幕文件失败: %v", err)
	}

	// 创建输出目录
	outputDir := filepath.Join(tempDir, "output")
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("创建输出目录失败: %v", err)
	}

	// 创建 CapcutGenerator 实例
	generator := NewCapcutGenerator(nil)

	// 测试生成项目功能
	err = generator.GenerateProjectWithOutputDir(inputDir, outputDir)
	if err != nil {
		t.Errorf("生成剪映项目失败: %v", err)
	}

	// 验证输出目录是否创建了项目文件
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("读取输出目录失败: %v", err)
	}

	foundOutputFile := false
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			foundOutputFile = true
			break
		}
	}

	if !foundOutputFile {
		t.Error("未找到生成的项目文件")
	}

	// 清理测试数据
	os.RemoveAll(tempDir)
}

// TestNewCapcutGenerator 测试 NewCapcutGenerator 函数
func TestNewCapcutGenerator(t *testing.T) {
	logger := struct{}{} // 使用空结构体模拟日志记录器
	generator := NewCapcutGenerator(logger)

	if generator == nil {
		t.Error("NewCapcutGenerator 返回了 nil")
	}

	if generator.Logger == nil {
		t.Error("NewCapcutGenerator 未正确设置 Logger")
	}
}