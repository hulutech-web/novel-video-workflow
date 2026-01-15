package draft

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewDraftFolder 测试创建草稿文件夹管理器
func TestNewDraftFolder(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "draft_folder_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试正常创建
	df, err := NewDraftFolder(tempDir)
	if err != nil {
		t.Fatalf("创建DraftFolder失败: %v", err)
	}

	if df.FolderPath != tempDir {
		t.Errorf("期望FolderPath为 '%s', 得到 '%s'", tempDir, df.FolderPath)
	}

	// 测试不存在的路径
	nonExistentPath := filepath.Join(tempDir, "non_existent")
	_, err = NewDraftFolder(nonExistentPath)
	if err == nil {
		t.Error("期望创建不存在路径的DraftFolder时返回错误")
	}
}

// TestListDrafts 测试列出草稿
func TestListDrafts(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "draft_folder_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试草稿文件夹
	draftNames := []string{"draft1", "draft2", "draft3"}
	for _, name := range draftNames {
		draftPath := filepath.Join(tempDir, name)
		if err := os.Mkdir(draftPath, 0755); err != nil {
			t.Fatalf("创建草稿文件夹失败: %v", err)
		}
	}

	// 创建一个普通文件（不应该被列出）
	regularFile := filepath.Join(tempDir, "regular_file.txt")
	if err := os.WriteFile(regularFile, []byte("test"), 0644); err != nil {
		t.Fatalf("创建普通文件失败: %v", err)
	}

	// 创建DraftFolder并列出草稿
	df, err := NewDraftFolder(tempDir)
	if err != nil {
		t.Fatalf("创建DraftFolder失败: %v", err)
	}

	drafts, err := df.ListDrafts()
	if err != nil {
		t.Fatalf("列出草稿失败: %v", err)
	}

	// 验证草稿数量
	if len(drafts) != 3 {
		t.Errorf("期望草稿数量为3，得到%d", len(drafts))
	}

	// 验证草稿名称
	for _, expectedName := range draftNames {
		found := false
		for _, actualName := range drafts {
			if actualName == expectedName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("未找到期望的草稿: %s", expectedName)
		}
	}
}

// TestRemove 测试删除草稿
func TestRemove(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "draft_folder_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试草稿文件夹
	draftName := "test_draft"
	draftPath := filepath.Join(tempDir, draftName)
	if err := os.Mkdir(draftPath, 0755); err != nil {
		t.Fatalf("创建草稿文件夹失败: %v", err)
	}

	// 在草稿文件夹中创建一些文件
	testFile := filepath.Join(draftPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 创建DraftFolder
	df, err := NewDraftFolder(tempDir)
	if err != nil {
		t.Fatalf("创建DraftFolder失败: %v", err)
	}

	// 验证草稿存在
	if !df.DraftExists(draftName) {
		t.Error("期望草稿存在")
	}

	// 删除草稿
	if err := df.Remove(draftName); err != nil {
		t.Fatalf("删除草稿失败: %v", err)
	}

	// 验证草稿已删除
	if df.DraftExists(draftName) {
		t.Error("期望草稿已被删除")
	}

	// 测试删除不存在的草稿
	err = df.Remove("non_existent_draft")
	if err == nil {
		t.Error("期望删除不存在的草稿时返回错误")
	}
}

// TestLoadTemplate 测试加载模板
func TestLoadTemplate(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "draft_folder_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试草稿文件夹
	draftName := "test_draft"
	draftPath := filepath.Join(tempDir, draftName)
	if err := os.Mkdir(draftPath, 0755); err != nil {
		t.Fatalf("创建草稿文件夹失败: %v", err)
	}

	// 创建draft_info.json文件
	draftInfo := map[string]interface{}{
		"fps":      float64(30),
		"duration": float64(5000000),
		"canvas_config": map[string]interface{}{
			"width":  float64(1920),
			"height": float64(1080),
		},
		"materials": map[string]interface{}{
			"videos": []interface{}{},
			"audios": []interface{}{},
		},
		"tracks": []interface{}{},
	}

	draftInfoPath := filepath.Join(draftPath, "draft_info.json")
	jsonBytes, err := json.Marshal(draftInfo)
	if err != nil {
		t.Fatalf("序列化draft_info失败: %v", err)
	}

	if err := os.WriteFile(draftInfoPath, jsonBytes, 0644); err != nil {
		t.Fatalf("写入draft_info.json失败: %v", err)
	}

	// 创建DraftFolder
	df, err := NewDraftFolder(tempDir)
	if err != nil {
		t.Fatalf("创建DraftFolder失败: %v", err)
	}

	// 测试加载模板
	scriptFile, err := df.LoadTemplate(draftName)
	if err != nil {
		t.Fatalf("加载模板失败: %v", err)
	}

	if scriptFile.FPS != 30 {
		t.Errorf("期望FPS为30，得到%d", scriptFile.FPS)
	}

	if scriptFile.Width != 1920 {
		t.Errorf("期望Width为1920，得到%d", scriptFile.Width)
	}

	if scriptFile.Height != 1080 {
		t.Errorf("期望Height为1080，得到%d", scriptFile.Height)
	}

	// 测试加载不存在的草稿
	_, err = df.LoadTemplate("non_existent_draft")
	if err == nil {
		t.Error("期望加载不存在的草稿时返回错误")
	}
}

// TestDuplicateAsTemplate 测试复制草稿作为模板
func TestDuplicateAsTemplate(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "draft_folder_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建原始草稿文件夹
	templateName := "template_draft"
	templatePath := filepath.Join(tempDir, templateName)
	if err := os.Mkdir(templatePath, 0755); err != nil {
		t.Fatalf("创建模板草稿文件夹失败: %v", err)
	}

	// 创建draft_info.json文件
	draftInfo := map[string]interface{}{
		"fps":      float64(25),
		"duration": float64(8000000),
		"canvas_config": map[string]interface{}{
			"width":  float64(1280),
			"height": float64(720),
		},
		"materials": map[string]interface{}{
			"videos": []interface{}{},
			"audios": []interface{}{},
		},
		"tracks": []interface{}{},
	}

	draftInfoPath := filepath.Join(templatePath, "draft_info.json")
	jsonBytes, err := json.Marshal(draftInfo)
	if err != nil {
		t.Fatalf("序列化draft_info失败: %v", err)
	}

	if err := os.WriteFile(draftInfoPath, jsonBytes, 0644); err != nil {
		t.Fatalf("写入draft_info.json失败: %v", err)
	}

	// 创建一些额外文件
	extraFile := filepath.Join(templatePath, "extra.txt")
	if err := os.WriteFile(extraFile, []byte("extra content"), 0644); err != nil {
		t.Fatalf("创建额外文件失败: %v", err)
	}

	// 创建DraftFolder
	df, err := NewDraftFolder(tempDir)
	if err != nil {
		t.Fatalf("创建DraftFolder失败: %v", err)
	}

	// 测试复制草稿
	newDraftName := "new_draft"
	scriptFile, err := df.DuplicateAsTemplate(templateName, newDraftName, false)
	if err != nil {
		t.Fatalf("复制草稿失败: %v", err)
	}

	// 验证复制的草稿属性
	if scriptFile.FPS != 25 {
		t.Errorf("期望FPS为25，得到%d", scriptFile.FPS)
	}

	if scriptFile.Width != 1280 {
		t.Errorf("期望Width为1280，得到%d", scriptFile.Width)
	}

	// 验证新草稿文件夹存在
	if !df.DraftExists(newDraftName) {
		t.Error("期望新草稿文件夹存在")
	}

	// 验证额外文件也被复制
	newExtraFile := filepath.Join(df.GetDraftPath(newDraftName), "extra.txt")
	if _, err := os.Stat(newExtraFile); os.IsNotExist(err) {
		t.Error("期望额外文件也被复制")
	}

	// 测试重复复制（不允许覆盖）
	_, err = df.DuplicateAsTemplate(templateName, newDraftName, false)
	if err == nil {
		t.Error("期望不允许覆盖时返回错误")
	}

	// 测试重复复制（允许覆盖）
	_, err = df.DuplicateAsTemplate(templateName, newDraftName, true)
	if err != nil {
		t.Fatalf("允许覆盖时复制失败: %v", err)
	}

	// 测试复制不存在的模板
	_, err = df.DuplicateAsTemplate("non_existent_template", "another_draft", false)
	if err == nil {
		t.Error("期望复制不存在的模板时返回错误")
	}
}

// TestInspectMaterial 测试素材检查
func TestInspectMaterial(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "draft_folder_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试草稿文件夹
	draftName := "test_draft"
	draftPath := filepath.Join(tempDir, draftName)
	if err := os.Mkdir(draftPath, 0755); err != nil {
		t.Fatalf("创建草稿文件夹失败: %v", err)
	}

	// 创建包含素材的draft_info.json文件
	draftInfo := map[string]interface{}{
		"fps":      float64(30),
		"duration": float64(5000000),
		"canvas_config": map[string]interface{}{
			"width":  float64(1920),
			"height": float64(1080),
		},
		"materials": map[string]interface{}{
			"stickers": []interface{}{
				map[string]interface{}{
					"resource_id": "sticker_123",
					"name":        "测试贴纸",
				},
			},
			"effects": []interface{}{
				map[string]interface{}{
					"type":        "text_shape",
					"effect_id":   "effect_456",
					"resource_id": "bubble_789",
					"name":        "文字气泡",
				},
			},
		},
		"tracks": []interface{}{},
	}

	draftInfoPath := filepath.Join(draftPath, "draft_info.json")
	jsonBytes, err := json.Marshal(draftInfo)
	if err != nil {
		t.Fatalf("序列化draft_info失败: %v", err)
	}

	if err := os.WriteFile(draftInfoPath, jsonBytes, 0644); err != nil {
		t.Fatalf("写入draft_info.json失败: %v", err)
	}

	// 创建DraftFolder
	df, err := NewDraftFolder(tempDir)
	if err != nil {
		t.Fatalf("创建DraftFolder失败: %v", err)
	}

	// 测试素材检查（这个方法主要是打印，我们只验证它不会出错）
	err = df.InspectMaterial(draftName)
	if err != nil {
		t.Fatalf("素材检查失败: %v", err)
	}

	// 测试检查不存在的草稿
	err = df.InspectMaterial("non_existent_draft")
	if err == nil {
		t.Error("期望检查不存在的草稿时返回错误")
	}
}

// TestDraftExists 测试草稿存在检查
func TestDraftExists(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "draft_folder_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试草稿文件夹
	draftName := "test_draft"
	draftPath := filepath.Join(tempDir, draftName)
	if err := os.Mkdir(draftPath, 0755); err != nil {
		t.Fatalf("创建草稿文件夹失败: %v", err)
	}

	// 创建DraftFolder
	df, err := NewDraftFolder(tempDir)
	if err != nil {
		t.Fatalf("创建DraftFolder失败: %v", err)
	}

	// 测试存在的草稿
	if !df.DraftExists(draftName) {
		t.Error("期望草稿存在")
	}

	// 测试不存在的草稿
	if df.DraftExists("non_existent_draft") {
		t.Error("期望草稿不存在")
	}
}

// TestGetDraftPath 测试获取草稿路径
func TestGetDraftPath(t *testing.T) {
	tempDir := "/test/folder"
	df := &DraftFolder{FolderPath: tempDir}

	draftName := "test_draft"
	expectedPath := filepath.Join(tempDir, draftName)
	actualPath := df.GetDraftPath(draftName)

	if actualPath != expectedPath {
		t.Errorf("期望路径为 '%s', 得到 '%s'", expectedPath, actualPath)
	}
}

// TestGetDraftInfo 测试获取草稿信息
func TestGetDraftInfo(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "draft_folder_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试草稿文件夹
	draftName := "test_draft"
	draftPath := filepath.Join(tempDir, draftName)
	if err := os.Mkdir(draftPath, 0755); err != nil {
		t.Fatalf("创建草稿文件夹失败: %v", err)
	}

	// 创建draft_info.json文件
	draftInfoPath := filepath.Join(draftPath, "draft_info.json")
	if err := os.WriteFile(draftInfoPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("创建draft_info.json失败: %v", err)
	}

	// 创建DraftFolder
	df, err := NewDraftFolder(tempDir)
	if err != nil {
		t.Fatalf("创建DraftFolder失败: %v", err)
	}

	// 获取草稿信息
	info, err := df.GetDraftInfo(draftName)
	if err != nil {
		t.Fatalf("获取草稿信息失败: %v", err)
	}

	if info.Name != draftName {
		t.Errorf("期望草稿名称为 '%s', 得到 '%s'", draftName, info.Name)
	}

	if info.Path != draftPath {
		t.Errorf("期望草稿路径为 '%s', 得到 '%s'", draftPath, info.Path)
	}

	if !info.HasDraftInfo {
		t.Error("期望草稿包含draft_info.json")
	}

	if !info.IsValid() {
		t.Error("期望草稿为有效状态")
	}

	// 测试获取不存在的草稿信息
	_, err = df.GetDraftInfo("non_existent_draft")
	if err == nil {
		t.Error("期望获取不存在的草稿信息时返回错误")
	}
}

// TestListDraftsWithInfo 测试列出草稿及其信息
func TestListDraftsWithInfo(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "draft_folder_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试草稿文件夹
	draftNames := []string{"draft1", "draft2", "draft3"}
	for i, name := range draftNames {
		draftPath := filepath.Join(tempDir, name)
		if err := os.Mkdir(draftPath, 0755); err != nil {
			t.Fatalf("创建草稿文件夹失败: %v", err)
		}

		// 只为部分草稿创建draft_info.json
		if i < 2 {
			draftInfoPath := filepath.Join(draftPath, "draft_info.json")
			if err := os.WriteFile(draftInfoPath, []byte("{}"), 0644); err != nil {
				t.Fatalf("创建draft_info.json失败: %v", err)
			}
		}
	}

	// 创建DraftFolder
	df, err := NewDraftFolder(tempDir)
	if err != nil {
		t.Fatalf("创建DraftFolder失败: %v", err)
	}

	// 获取草稿信息列表
	infos, err := df.ListDraftsWithInfo()
	if err != nil {
		t.Fatalf("获取草稿信息列表失败: %v", err)
	}

	if len(infos) != 3 {
		t.Errorf("期望草稿信息数量为3，得到%d", len(infos))
	}

	// 验证有效草稿数量
	validCount := 0
	for _, info := range infos {
		if info.IsValid() {
			validCount++
		}
	}

	if validCount != 2 {
		t.Errorf("期望有效草稿数量为2，得到%d", validCount)
	}
}

// TestDraftInfoAge 测试草稿年龄计算
func TestDraftInfoAge(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-time.Hour) // 1小时前

	info := &DraftInfo{
		Name:         "test_draft",
		Path:         "/test/path",
		ModTime:      pastTime,
		HasDraftInfo: true,
	}

	age := info.Age()
	if age < time.Hour-time.Second || age > time.Hour+time.Second {
		t.Errorf("期望年龄约为1小时，得到%v", age)
	}
}

// TestDraftInfoString 测试草稿信息字符串表示
func TestDraftInfoString(t *testing.T) {
	now := time.Now()

	info := &DraftInfo{
		Name:         "test_draft",
		Path:         "/test/path",
		ModTime:      now,
		HasDraftInfo: true,
	}

	str := info.String()
	if str == "" {
		t.Error("期望非空字符串")
	}

	// 检查是否包含关键信息
	if !containsString(str, "test_draft") {
		t.Error("字符串应包含草稿名称")
	}

	if !containsString(str, "有效") {
		t.Error("字符串应包含有效状态")
	}

	// 测试无效草稿
	info.HasDraftInfo = false
	str = info.String()
	if !containsString(str, "无效") {
		t.Error("字符串应包含无效状态")
	}
}

// containsString 检查字符串是否包含子字符串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

// containsSubstring 简单的子字符串检查
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
