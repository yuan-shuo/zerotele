// Package maskgen 处理 mask 函数的生成和补全
package maskgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuan-shuo/zerotele/internal/config"
)

func TestNew(t *testing.T) {
	g := New("logger")
	if g == nil {
		t.Fatal("New() returned nil")
	}
	if g.packageName != "logger" {
		t.Errorf("New() packageName = %v, want %v", g.packageName, "logger")
	}
}

func TestHasMaskFields(t *testing.T) {
	tests := []struct {
		name     string
		fields   []config.LogFieldConfig
		expected bool
	}{
		{
			"has mask",
			[]config.LogFieldConfig{
				{Name: "user_id", Type: "int64", Mask: true},
				{Name: "user_name", Type: "string", Mask: false},
			},
			true,
		},
		{
			"no mask",
			[]config.LogFieldConfig{
				{Name: "user_id", Type: "int64", Mask: false},
				{Name: "user_name", Type: "string", Mask: false},
			},
			false,
		},
		{
			"empty fields",
			[]config.LogFieldConfig{},
			false,
		},
		{
			"nil fields",
			nil,
			false,
		},
		{
			"all masked",
			[]config.LogFieldConfig{
				{Name: "password", Type: "string", Mask: true},
				{Name: "email", Type: "string", Mask: true},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasMaskFields(tt.fields)
			if result != tt.expected {
				t.Errorf("HasMaskFields() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFilterMaskFields(t *testing.T) {
	fields := []config.LogFieldConfig{
		{Name: "user_id", Type: "int64", Mask: true},
		{Name: "user_name", Type: "string", Mask: false},
		{Name: "email", Type: "string", Mask: true},
	}

	result := filterMaskFields(fields)
	if len(result) != 2 {
		t.Errorf("filterMaskFields() returned %d fields, want 2", len(result))
	}

	// 检查返回的字段是否正确
	names := make(map[string]bool)
	for _, f := range result {
		names[f.Name] = true
	}
	if !names["user_id"] || !names["email"] {
		t.Error("filterMaskFields() returned wrong fields")
	}
	if names["user_name"] {
		t.Error("filterMaskFields() should not include non-mask fields")
	}
}

func TestGenerateNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	maskFile := filepath.Join(tmpDir, "mask.go")

	g := New("logger")
	fields := []config.LogFieldConfig{
		{Name: "user_id", Type: "int64", Comment: "用户ID", Mask: true},
		{Name: "email", Type: "string", Comment: "邮箱", Mask: true},
	}

	err := g.Generate(maskFile, fields)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 检查文件是否生成
	content, err := os.ReadFile(maskFile)
	if err != nil {
		t.Fatalf("Failed to read mask file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package logger") {
		t.Error("Generated file should contain package declaration")
	}
	if !strings.Contains(contentStr, "func maskUserId") {
		t.Error("Generated file should contain maskUserId function")
	}
	if !strings.Contains(contentStr, "func maskEmail") {
		t.Error("Generated file should contain maskEmail function")
	}
	if !strings.Contains(contentStr, "TODO: 实现脱敏逻辑") {
		t.Error("Generated file should contain TODO comment")
	}
}

func TestGenerateAppendToExisting(t *testing.T) {
	tmpDir := t.TempDir()
	maskFile := filepath.Join(tmpDir, "mask.go")

	// 先创建现有文件
	existingContent := `package logger

// maskUserId 对用户ID进行脱敏
// 请在此实现具体的脱敏逻辑
func maskUserId(userid int64) any {
	// 已实现脱敏逻辑
	return "***"
}
`
	if err := os.WriteFile(maskFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	g := New("logger")
	// 包含一个已存在的字段和一个新字段
	fields := []config.LogFieldConfig{
		{Name: "user_id", Type: "int64", Comment: "用户ID", Mask: true},
		{Name: "email", Type: "string", Comment: "邮箱", Mask: true},
	}

	err := g.Generate(maskFile, fields)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 检查文件内容
	content, err := os.ReadFile(maskFile)
	if err != nil {
		t.Fatalf("Failed to read mask file: %v", err)
	}

	contentStr := string(content)
	// 应该保留原有内容
	if !strings.Contains(contentStr, "已实现脱敏逻辑") {
		t.Error("Existing content should be preserved")
	}
	// 应该添加新函数
	if !strings.Contains(contentStr, "func maskEmail") {
		t.Error("New maskEmail function should be added")
	}
	// 不应该重复添加已有函数
	if strings.Count(contentStr, "func maskUserId") != 1 {
		t.Error("Existing function should not be duplicated")
	}
}

func TestGenerateNoMaskFields(t *testing.T) {
	tmpDir := t.TempDir()
	maskFile := filepath.Join(tmpDir, "mask.go")

	g := New("logger")
	fields := []config.LogFieldConfig{
		{Name: "user_name", Type: "string", Comment: "用户名", Mask: false},
	}

	err := g.Generate(maskFile, fields)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 不应该创建文件
	if _, err := os.Stat(maskFile); !os.IsNotExist(err) {
		t.Error("File should not be created when no mask fields")
	}
}

func TestGenerateEmptyFields(t *testing.T) {
	tmpDir := t.TempDir()
	maskFile := filepath.Join(tmpDir, "mask.go")

	g := New("logger")
	err := g.Generate(maskFile, []config.LogFieldConfig{})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 不应该创建文件
	if _, err := os.Stat(maskFile); !os.IsNotExist(err) {
		t.Error("File should not be created when fields are empty")
	}
}

func TestParseExistingMaskFile(t *testing.T) {
	tmpDir := t.TempDir()
	maskFile := filepath.Join(tmpDir, "mask.go")

	content := `package logger

func maskUserId(userid int64) any {
	return userid
}

func maskEmail(email string) any {
	return email
}
`
	if err := os.WriteFile(maskFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	g := New("logger")
	existingContent, funcs, err := g.parseExistingMaskFile(maskFile)
	if err != nil {
		t.Fatalf("parseExistingMaskFile() error = %v", err)
	}

	if existingContent != content {
		t.Error("parseExistingMaskFile() returned wrong content")
	}

	if len(funcs) != 2 {
		t.Errorf("parseExistingMaskFile() returned %d functions, want 2", len(funcs))
	}

	if !funcs["maskUserId"] || !funcs["maskEmail"] {
		t.Error("parseExistingMaskFile() returned wrong function names")
	}
}

func TestParseExistingMaskFileInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	maskFile := filepath.Join(tmpDir, "mask.go")

	// 写入无效的 Go 代码
	invalidContent := `package logger

func maskUserId(userid int64) any {
	return userid
// 缺少右括号
`
	if err := os.WriteFile(maskFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	g := New("logger")
	_, _, err := g.parseExistingMaskFile(maskFile)
	if err == nil {
		t.Error("parseExistingMaskFile() should return error for invalid Go code")
	}
}

func TestGenerateMaskCode(t *testing.T) {
	g := New("logger")
	fields := []config.LogFieldConfig{
		{Name: "user_id", Type: "int64", Comment: "用户ID"},
		{Name: "email", Type: "string", Comment: "邮箱"},
	}

	// 测试新文件
	code := g.generateMaskCode(fields, true)
	if !strings.Contains(code, "package logger") {
		t.Error("New file code should contain package declaration")
	}
	if !strings.Contains(code, "func maskUserId") {
		t.Error("Code should contain maskUserId function")
	}
	if !strings.Contains(code, "func maskEmail") {
		t.Error("Code should contain maskEmail function")
	}
	if !strings.Contains(code, "对用户ID进行脱敏") {
		t.Error("Code should contain comment with field comment")
	}

	// 测试追加到现有文件
	code = g.generateMaskCode(fields, false)
	if strings.Contains(code, "package logger") {
		t.Error("Append code should not contain package declaration")
	}
	if !strings.Contains(code, "func maskUserId") {
		t.Error("Append code should contain maskUserId function")
	}
}

func TestGenerateMaskCodeEmptyFields(t *testing.T) {
	g := New("logger")
	code := g.generateMaskCode([]config.LogFieldConfig{}, true)
	if !strings.Contains(code, "package logger") {
		t.Error("Code should contain package even with empty fields")
	}
	// 不应该包含任何函数
	if strings.Contains(code, "func mask") {
		t.Error("Code should not contain any mask functions when fields are empty")
	}
}
