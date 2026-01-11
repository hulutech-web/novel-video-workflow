package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type FileManager struct{}

func NewFileManager() *FileManager {
	return &FileManager{}
}

type ChapterStructure struct {
	ChapterDir  string
	TextFile    string
	AudioDir    string
	SubtitleDir string
	ImageDir    string
	SceneDir    string
}

func (fm *FileManager) CreateChapterStructure(chapterNum int, text string, baseDir string) (*ChapterStructure, error) {
	// 创建章节目录
	chapterDir := filepath.Join(baseDir, fmt.Sprintf("chapter_%02d", chapterNum))
	if err := os.MkdirAll(chapterDir, 0755); err != nil {
		return nil, fmt.Errorf("创建章节目录失败: %w", err)
	}

	// 创建子目录
	subdirs := []string{"audio", "subtitles", "images", "scenes"}
	dirPaths := make(map[string]string)

	for _, subdir := range subdirs {
		dirPath := filepath.Join(chapterDir, subdir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return nil, fmt.Errorf("创建子目录 %s 失败: %w", subdir, err)
		}
		dirPaths[subdir] = dirPath
	}

	// 保存文本文件
	textFile := filepath.Join(chapterDir, fmt.Sprintf("chapter_%d.txt", chapterNum))
	if err := os.WriteFile(textFile, []byte(text), 0644); err != nil {
		return nil, fmt.Errorf("保存文本文件失败: %w", err)
	}

	return &ChapterStructure{
		ChapterDir:  chapterDir,
		TextFile:    textFile,
		AudioDir:    dirPaths["audio"],
		SubtitleDir: dirPaths["subtitles"],
		ImageDir:    dirPaths["images"],
		SceneDir:    dirPaths["scenes"],
	}, nil
}

func (fm *FileManager) SaveJSON(filepath string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化JSON失败: %w", err)
	}

	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return fmt.Errorf("写入JSON文件失败: %w", err)
	}

	return nil
}

// CreateNovelInputStructure 创建小说输入目录结构
func (fm *FileManager) CreateNovelInputStructure(novelName, novelText string) error {
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	// 创建小说主目录
	novelDir := filepath.Join(wd, "input", novelName)
	if err := os.MkdirAll(novelDir, 0755); err != nil {
		return fmt.Errorf("创建小说目录失败: %w", err)
	}

	// 拆分章节
	chapters := fm.SplitNovelIntoChapters(novelText)

	// 为每个章节创建目录和文件
	for i, chapterText := range chapters {
		// 从章节文本中提取章节号
		chapterNum := fm.extractChapterNumber(chapterText)
		if chapterNum == 0 { // 如果无法提取章节号，使用顺序号
			chapterNum = i + 1
		}
		chapterDir := filepath.Join(novelDir, fmt.Sprintf("chapter_%02d", chapterNum))
		if err := os.MkdirAll(chapterDir, 0755); err != nil {
			return fmt.Errorf("创建章节目录失败: %w", err)
		}

		// 创建章节文本文件
		chapterFile := filepath.Join(chapterDir, fmt.Sprintf("chapter_%02d.txt", chapterNum))
		if err := os.WriteFile(chapterFile, []byte(chapterText), 0644); err != nil {
			return fmt.Errorf("保存章节文件失败: %w", err)
		}
	}

	return nil
}

// SplitNovelIntoChapters 将小说文本按章节拆分
func (fm *FileManager) SplitNovelIntoChapters(novelText string) []string {
	// 使用正则表达式匹配章节标记
	// 匹配以“第x章”、“第xx章”、“第xxx章”等开头的行
	re := regexp.MustCompile(`(?m)^\s*第[\p{N}\p{L}]+[章节][^\r\n]*$`)
	matches := re.FindAllStringIndex(novelText, -1)

	var chapters []string

	// 如果没有找到章节标记，则将整个文本作为一个章节
	if len(matches) == 0 {
		trimmed := strings.TrimSpace(novelText)
		if trimmed != "" {
			chapters = append(chapters, trimmed)
		}
		return chapters
	}

	for i, match := range matches {
		var chapterStart, chapterEnd int
		
		// 当前章节的开始位置是章节标记的开始
		chapterStart = match[0]
		
		// 当前章节的结束位置是下一个章节标记的开始，或者是文本的结尾
		if i+1 < len(matches) {
			chapterEnd = matches[i+1][0]
		} else {
			chapterEnd = len(novelText)
		}
		
		// 提取包含章节标题的完整章节内容
		chapterContent := strings.TrimSpace(novelText[chapterStart:chapterEnd])
		if chapterContent != "" {
			chapters = append(chapters, chapterContent)
		}
	}

	return chapters
}

// extractChapterNumber 从章节文本中提取章节号
func (fm *FileManager) extractChapterNumber(chapterText string) int {
	// 查找章节标题行
	lines := strings.Split(chapterText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 使用正则表达式匹配章节标记
		re := regexp.MustCompile(`^\s*第([\p{N}\p{L}]+)[章节]`)
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			chapterNumStr := matches[1]
			// 尝试解析数字
			if num, err := strconv.Atoi(chapterNumStr); err == nil {
				return num
			}
			// 如果是汉字数字，转换为阿拉伯数字
			return fm.convertChineseNumberToArabic(chapterNumStr)
		}
	}
	return 0 // 无法提取章节号时返回0
}

// convertChineseNumberToArabic 将中文数字转换为阿拉伯数字
func (fm *FileManager) convertChineseNumberToArabic(chineseNum string) int {
	chineseToArabic := map[string]int{
		"一": 1, "二": 2, "三": 3, "四": 4, "五": 5,
		"六": 6, "七": 7, "八": 8, "九": 9, "十": 10,
		"十一": 11, "十二": 12, "十三": 13, "十四": 14, "十五": 15,
		"十六": 16, "十七": 17, "十八": 18, "十九": 19, "二十": 20,
		"零": 0, "两": 2,
	}
	if num, exists := chineseToArabic[chineseNum]; exists {
		return num
	}
	return 0
}

// CreateNovelOutputStructure 创建小说输出目录结构
func (fm *FileManager) CreateNovelOutputStructure(novelName string) error {
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	// 创建小说输出主目录
	novelOutputDir := filepath.Join(wd, "output", novelName)
	if err := os.MkdirAll(novelOutputDir, 0755); err != nil {
		return fmt.Errorf("创建小说输出目录失败: %w", err)
	}

	// 获取输入目录中的章节数量
	inputDir := filepath.Join(wd, "input", novelName)
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("读取输入目录失败: %w", err)
	}

	// 为每个章节创建输出目录
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "chapter_") {
			chapterOutputDir := filepath.Join(novelOutputDir, entry.Name())
			if err := os.MkdirAll(chapterOutputDir, 0755); err != nil {
				return fmt.Errorf("创建章节输出目录失败: %w", err)
			}
		}
	}

	return nil
}

// GetNovelChaptersFromInput 获取输入目录中的所有章节文件
func (fm *FileManager) GetNovelChaptersFromInput(novelName string) ([]string, error) {
	inputDir := filepath.Join("input", novelName)
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return nil, fmt.Errorf("读取输入目录失败: %w", err)
	}

	var chapterFiles []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "chapter_") {
			chapterFile := filepath.Join(inputDir, entry.Name(), entry.Name()+".txt")
			if _, err := os.Stat(chapterFile); err == nil { // 检查章节文件是否存在
				chapterFiles = append(chapterFiles, chapterFile)
			}
		}
	}

	// 按名称排序以保证章节顺序
	for i := 0; i < len(chapterFiles)-1; i++ {
		for j := i + 1; j < len(chapterFiles); j++ {
			if filepath.Base(chapterFiles[i]) > filepath.Base(chapterFiles[j]) {
				chapterFiles[i], chapterFiles[j] = chapterFiles[j], chapterFiles[i]
			}
		}
	}

	return chapterFiles, nil
}

// SplitNovelFileIntoChapters 从文件读取小说并将其拆分为多个章节文件
func (fm *FileManager) SplitNovelFileIntoChapters(novelFilePath string) ([]string, error) {
	// 读取小说文件
	content, err := os.ReadFile(novelFilePath)
	if err != nil {
		return nil, fmt.Errorf("无法读取小说文件: %v", err)
	}

	novelText := string(content)

	// 拆分章节
	chapters := fm.SplitNovelIntoChapters(novelText)

	if len(chapters) == 0 {
		return nil, fmt.Errorf("在文件中未找到任何章节")
	}

	// 确定输出目录
	dir := filepath.Dir(novelFilePath)

	// 为每个章节创建目录和文件
	var createdFiles []string
	for i, chapterText := range chapters {
		// 从章节文本中提取章节号
		chapterNum := fm.extractChapterNumber(chapterText)
		if chapterNum == 0 { // 如果无法提取章节号，使用顺序号
			chapterNum = i + 1
		}
		chapterDir := filepath.Join(dir, fmt.Sprintf("chapter_%02d", chapterNum))
		if err := os.MkdirAll(chapterDir, 0755); err != nil {
			return nil, fmt.Errorf("创建章节目录失败: %w", err)
		}

		// 创建章节文本文件
		chapterFile := filepath.Join(chapterDir, fmt.Sprintf("chapter_%02d.txt", chapterNum))
		if err := os.WriteFile(chapterFile, []byte(chapterText), 0644); err != nil {
			return nil, fmt.Errorf("保存章节文件失败: %w", err)
		}
		createdFiles = append(createdFiles, chapterFile)
	}

	return createdFiles, nil
}