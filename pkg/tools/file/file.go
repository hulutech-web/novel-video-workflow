package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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