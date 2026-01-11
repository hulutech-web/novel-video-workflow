package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitNovelIntoChapters(t *testing.T) {
	fm := NewFileManager()
	
	// 测试小说文本
	testNovel := `这是小说的简介部分
第1章 初入江湖
这里是第一章的内容，讲述了主角初次踏入江湖的故事。
第2章 遇险
第二章中，主角遇到了危险，但幸运地得到了高人的救助。
第3章 修炼
主角开始专心修炼，实力不断提升。
第10章 大结局
经过无数磨难，主角终于成为了武林高手。`

	chapters := fm.SplitNovelIntoChapters(testNovel)
	
	if len(chapters) != 4 {
		t.Errorf("期望4个章节，实际得到 %d 个", len(chapters))
	}
	
	if len(chapters) > 0 && !contains(chapters[0], "简介") {
		t.Error("第一个部分应该包含简介内容")
	}
	
	if len(chapters) > 1 && !contains(chapters[1], "初入江湖") {
		t.Error("第二个部分应该包含第一章内容")
	}
	
	if len(chapters) > 2 && !contains(chapters[2], "遇险") {
		t.Error("第三个部分应该包含第二章内容")
	}
	
	if len(chapters) > 3 && !contains(chapters[3], "大结局") {
		t.Error("第四个部分应该包含第十章内容")
	}
}

func TestCreateNovelInputStructure(t *testing.T) {
	fm := NewFileManager()
	
	testNovel := `这是小说的简介部分
第1章 初入江湖
这里是第一章的内容，讲述了主角初次踏入江湖的故事。
第2章 遇险
第二章中，主角遇到了危险，但幸运地得到了高人的救助。`
	
	novelName := "test_novel"
	
	// 清理测试目录
	testDir := filepath.Join("input", novelName)
	os.RemoveAll(testDir)
	
	err := fm.CreateNovelInputStructure(novelName, testNovel)
	if err != nil {
		t.Fatalf("创建小说输入结构失败: %v", err)
	}
	
	// 检查目录结构是否正确创建
	expectedDirs := []string{
		filepath.Join("input", novelName),
		filepath.Join("input", novelName, "chapter_01"),
		filepath.Join("input", novelName, "chapter_02"),
	}
	
	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("目录未创建: %s", dir)
		}
	}
	
	// 检查文件是否存在
	expectedFiles := []string{
		filepath.Join("input", novelName, "chapter_01", "chapter_01.txt"),
		filepath.Join("input", novelName, "chapter_02", "chapter_02.txt"),
	}
	
	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("文件未创建: %s", file)
		}
	}
	
	// 清理测试目录
	os.RemoveAll(filepath.Join("input", novelName))
}

func TestCreateNovelOutputStructure(t *testing.T) {
	fm := NewFileManager()
	
	novelName := "test_novel"
	
	// 首先创建输入目录结构用于测试
	testNovel := `第1章 初入江湖
这里是第一章的内容。
第2章 遇险
这里是第二章的内容。`
	
	// 清理测试目录
	inputDir := filepath.Join("input", novelName)
	os.RemoveAll(inputDir)
	
	err := fm.CreateNovelInputStructure(novelName, testNovel)
	if err != nil {
		t.Fatalf("创建测试输入结构失败: %v", err)
	}
	
	// 测试输出目录结构创建
	outputDir := filepath.Join("output", novelName)
	os.RemoveAll(outputDir)
	
	err = fm.CreateNovelOutputStructure(novelName)
	if err != nil {
		t.Fatalf("创建小说输出结构失败: %v", err)
	}
	
	// 检查输出目录结构是否正确创建
	expectedDirs := []string{
		filepath.Join("output", novelName),
		filepath.Join("output", novelName, "chapter_01"),
		filepath.Join("output", novelName, "chapter_02"),
	}
	
	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("输出目录未创建: %s", dir)
		}
	}
	
	// 清理测试目录
	os.RemoveAll(inputDir)
	os.RemoveAll(outputDir)
}

func contains(s string, substr string) bool {
	return strings.Contains(s, substr)
}