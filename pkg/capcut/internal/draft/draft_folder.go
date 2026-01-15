// Package draft 定义草稿文件夹管理系统
// 对应Python的 pyJianYingDraft/draft_folder.py
package draft

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"novel-video-workflow/pkg/capcut/internal/script"
)

// DraftFolder 管理一个文件夹及其内的一系列草稿
// 对应Python的Draft_folder类
type DraftFolder struct {
	FolderPath string `json:"folder_path"` // 根路径
}

// NewDraftFolder 创建新的草稿文件夹管理器
// 对应Python的__init__方法
func NewDraftFolder(folderPath string) (*DraftFolder, error) {
	// 检查路径是否存在
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("根文件夹 %s 不存在", folderPath)
	}

	return &DraftFolder{
		FolderPath: folderPath,
	}, nil
}

// ListDrafts 列出文件夹中所有草稿的名称
// 对应Python的list_drafts方法
// 注意: 本函数只是如实地列出子文件夹的名称，并不检查它们是否符合草稿的格式
func (df *DraftFolder) ListDrafts() ([]string, error) {
	entries, err := os.ReadDir(df.FolderPath)
	if err != nil {
		return nil, fmt.Errorf("读取文件夹失败: %v", err)
	}

	var drafts []string
	for _, entry := range entries {
		if entry.IsDir() {
			drafts = append(drafts, entry.Name())
		}
	}

	return drafts, nil
}

// Remove 删除指定名称的草稿
// 对应Python的remove方法
func (df *DraftFolder) Remove(draftName string) error {
	draftPath := filepath.Join(df.FolderPath, draftName)

	// 检查草稿文件夹是否存在
	if _, err := os.Stat(draftPath); os.IsNotExist(err) {
		return fmt.Errorf("草稿文件夹 %s 不存在", draftName)
	}

	// 删除整个草稿文件夹
	if err := os.RemoveAll(draftPath); err != nil {
		return fmt.Errorf("删除草稿文件夹失败: %v", err)
	}

	return nil
}

// InspectMaterial 输出指定名称草稿中的贴纸素材元数据
// 对应Python的inspect_material方法
func (df *DraftFolder) InspectMaterial(draftName string) error {
	draftPath := filepath.Join(df.FolderPath, draftName)

	// 检查草稿文件夹是否存在
	if _, err := os.Stat(draftPath); os.IsNotExist(err) {
		return fmt.Errorf("草稿文件夹 %s 不存在", draftName)
	}

	// 加载草稿文件
	scriptFile, err := df.LoadTemplate(draftName)
	if err != nil {
		return fmt.Errorf("加载草稿失败: %v", err)
	}

	// 检查素材
	scriptFile.InspectMaterial()

	return nil
}

// LoadTemplate 在文件夹中打开一个草稿作为模板，并在其上进行编辑
// 对应Python的load_template方法
func (df *DraftFolder) LoadTemplate(draftName string) (*script.ScriptFile, error) {
	draftPath := filepath.Join(df.FolderPath, draftName)

	// 检查草稿文件夹是否存在
	if _, err := os.Stat(draftPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("草稿文件夹 %s 不存在", draftName)
	}

	// 加载草稿信息文件
	draftInfoPath := filepath.Join(draftPath, "draft_info.json")
	return script.LoadTemplate(draftInfoPath)
}

// DuplicateAsTemplate 复制一份给定的草稿，并在复制出的新草稿上进行编辑
// 对应Python的duplicate_as_template方法
func (df *DraftFolder) DuplicateAsTemplate(templateName, newDraftName string, allowReplace bool) (*script.ScriptFile, error) {
	templatePath := filepath.Join(df.FolderPath, templateName)
	newDraftPath := filepath.Join(df.FolderPath, newDraftName)

	// 检查原始草稿是否存在
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("模板草稿 %s 不存在", templateName)
	}

	// 检查新草稿是否已存在
	if _, err := os.Stat(newDraftPath); err == nil && !allowReplace {
		return nil, fmt.Errorf("新草稿 %s 已存在且不允许覆盖", newDraftName)
	}

	// 如果允许替换且目标已存在，先删除
	if allowReplace {
		if _, err := os.Stat(newDraftPath); err == nil {
			if err := os.RemoveAll(newDraftPath); err != nil {
				return nil, fmt.Errorf("删除已存在的草稿失败: %v", err)
			}
		}
	}

	// 复制草稿文件夹
	if err := copyDir(templatePath, newDraftPath); err != nil {
		return nil, fmt.Errorf("复制草稿文件夹失败: %v", err)
	}

	// 打开复制后的草稿
	return df.LoadTemplate(newDraftName)
}

// copyDir 递归复制目录
// Go标准库没有内置的目录复制功能，需要自己实现
func copyDir(src, dst string) error {
	// 获取源目录信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 创建目标目录
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// 读取源目录内容
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// 递归复制每个条目
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// 递归复制子目录
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// 复制文件
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile 复制单个文件
func copyFile(src, dst string) error {
	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 获取源文件信息
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// 创建目标文件
	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 复制文件内容
	_, err = io.Copy(dstFile, srcFile)
	return err
}

// GetDraftPath 获取指定草稿的完整路径
// 辅助方法，便于其他操作
func (df *DraftFolder) GetDraftPath(draftName string) string {
	return filepath.Join(df.FolderPath, draftName)
}

// DraftExists 检查指定草稿是否存在
// 辅助方法，便于其他操作
func (df *DraftFolder) DraftExists(draftName string) bool {
	draftPath := filepath.Join(df.FolderPath, draftName)
	_, err := os.Stat(draftPath)
	return err == nil
}

// GetDraftInfo 获取草稿的基本信息
// 扩展方法，提供草稿的详细信息
func (df *DraftFolder) GetDraftInfo(draftName string) (*DraftInfo, error) {
	draftPath := filepath.Join(df.FolderPath, draftName)

	// 检查草稿文件夹是否存在
	if _, err := os.Stat(draftPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("草稿文件夹 %s 不存在", draftName)
	}

	// 获取文件夹信息
	info, err := os.Stat(draftPath)
	if err != nil {
		return nil, fmt.Errorf("获取草稿信息失败: %v", err)
	}

	// 检查是否有draft_info.json文件
	draftInfoPath := filepath.Join(draftPath, "draft_info.json")
	hasDraftInfo := true
	if _, err := os.Stat(draftInfoPath); os.IsNotExist(err) {
		hasDraftInfo = false
	}

	return &DraftInfo{
		Name:         draftName,
		Path:         draftPath,
		ModTime:      info.ModTime(),
		HasDraftInfo: hasDraftInfo,
	}, nil
}

// ListDraftsWithInfo 列出文件夹中所有草稿及其详细信息
// 扩展方法，提供更丰富的草稿列表信息
func (df *DraftFolder) ListDraftsWithInfo() ([]*DraftInfo, error) {
	drafts, err := df.ListDrafts()
	if err != nil {
		return nil, err
	}

	var draftInfos []*DraftInfo
	for _, draftName := range drafts {
		info, err := df.GetDraftInfo(draftName)
		if err != nil {
			// 跳过无法获取信息的草稿，但不中断整个过程
			continue
		}
		draftInfos = append(draftInfos, info)
	}

	return draftInfos, nil
}
