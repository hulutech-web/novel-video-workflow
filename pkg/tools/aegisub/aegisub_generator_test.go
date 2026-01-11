package aegisub

import (
	"os"
	"testing"
)

func TestAegisubGenerator(t *testing.T) {
	// 创建AegisubGenerator实例
	gen := NewAegisubGenerator()
	
	// 测试实例是否正确创建
	if gen == nil {
		t.Fatal("Failed to create AegisubGenerator instance")
	}
	
	// 验证路径设置
	if gen.LuaScriptPath == "" {
		t.Error("LuaScriptPath should not be empty")
	}
	
	if gen.ScriptPath == "" {
		t.Error("ScriptPath should not be empty")
	}
	
	// 注意：实际的GenerateSubtitle调用需要Aegisub软件安装在系统上
	// 因此我们只测试结构和路径，而不执行实际的字幕生成
	t.Logf("AegisubGenerator created successfully with LuaScriptPath: %s, ScriptPath: %s", 
		   gen.LuaScriptPath, gen.ScriptPath)
}

func TestCreateTempTextFile(t *testing.T) {
	textContent := "This is a test subtitle content.\nIt has multiple lines.\nEach line represents a subtitle segment."
	
	tempFile, err := createTempTextFile(textContent)
	if err != nil {
		t.Fatalf("Failed to create temp text file: %v", err)
	}
	defer os.Remove(tempFile) // 清理临时文件
	
	// 检查文件是否存在
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Fatalf("Temp file was not created: %v", err)
	}
	
	// 读取文件内容验证
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}
	
	if string(content) != textContent {
		t.Errorf("File content mismatch. Expected: %s, Got: %s", textContent, string(content))
	}
	
	t.Logf("Temp file created successfully: %s", tempFile)
}

func TestAegisubIntegration(t *testing.T) {
	// 创建AegisubIntegration实例
	integration := NewAegisubIntegration()
	
	if integration == nil {
		t.Fatal("Failed to create AegisubIntegration instance")
	}
	
	if integration.aegisubGen == nil {
		t.Error("AegisubGenerator not initialized in integration")
	}
	
	t.Log("AegisubIntegration created successfully")
}