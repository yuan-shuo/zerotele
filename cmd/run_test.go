// Package cmd 包含 zerotele 工具的核心逻辑
package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yuan-shuo/zerotele/internal/config"
)

func TestRunMetrics(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "config.yaml")
	outputDir := filepath.Join(tmpDir, "metrics")

	// 创建测试 YAML 文件
	yamlContent := `service: test_service
metrics:
  - name: requests_total
    help: Total requests
    type: counter
    methods:
      - inc
`
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test YAML: %v", err)
	}

	opts := &MetricOptions{
		YamlFile:  yamlFile,
		OutputDir: outputDir,
	}

	err := RunMetrics(opts)
	if err != nil {
		t.Errorf("RunMetrics() error = %v", err)
	}

	// 检查文件是否生成
	outputPath := filepath.Join(outputDir, "metrics_gen.go")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("RunMetrics() did not create output file")
	}
}

func TestRunMetricsNoMetrics(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "config.yaml")
	outputDir := filepath.Join(tmpDir, "metrics")

	// 创建没有 metrics 的 YAML 文件
	yamlContent := `service: test_service
logfields:
  - name: user_id
    type: int64
`
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test YAML: %v", err)
	}

	opts := &MetricOptions{
		YamlFile:  yamlFile,
		OutputDir: outputDir,
	}

	err := RunMetrics(opts)
	if err == nil {
		t.Error("RunMetrics() should return error when no metrics defined")
	}
}

func TestRunMetricsInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "config.yaml")
	outputDir := filepath.Join(tmpDir, "metrics")

	// 创建无效的 YAML 文件
	if err := os.WriteFile(yamlFile, []byte("invalid: yaml: content: ["), 0644); err != nil {
		t.Fatalf("Failed to create test YAML: %v", err)
	}

	opts := &MetricOptions{
		YamlFile:  yamlFile,
		OutputDir: outputDir,
	}

	err := RunMetrics(opts)
	if err == nil {
		t.Error("RunMetrics() should return error for invalid YAML")
	}
}

func TestRunLogFields(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "config.yaml")
	outputDir := filepath.Join(tmpDir, "logger")

	// 创建测试 YAML 文件
	yamlContent := `service: test_service
logfields:
  - name: user_id
    type: int64
    comment: 用户ID
    mask: false
`
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test YAML: %v", err)
	}

	opts := &LogFieldOptions{
		YamlFile:  yamlFile,
		OutputDir: outputDir,
	}

	err := RunLogFields(opts)
	if err != nil {
		t.Errorf("RunLogFields() error = %v", err)
	}

	// 检查文件是否生成
	outputPath := filepath.Join(outputDir, "logfields_gen.go")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("RunLogFields() did not create output file")
	}
}

func TestRunLogFieldsWithMask(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "config.yaml")
	outputDir := filepath.Join(tmpDir, "logger")

	// 创建测试 YAML 文件，包含需要 mask 的字段
	yamlContent := `service: test_service
logfields:
  - name: user_id
    type: int64
    comment: 用户ID
    mask: true
`
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test YAML: %v", err)
	}

	opts := &LogFieldOptions{
		YamlFile:  yamlFile,
		OutputDir: outputDir,
		MaskFile:  "mask.go",
	}

	err := RunLogFields(opts)
	if err != nil {
		t.Errorf("RunLogFields() error = %v", err)
	}

	// 检查日志字段文件是否生成
	outputPath := filepath.Join(outputDir, "logfields_gen.go")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("RunLogFields() did not create logfields_gen.go")
	}

	// 检查 mask 文件是否生成
	maskPath := filepath.Join(outputDir, "mask.go")
	if _, err := os.Stat(maskPath); os.IsNotExist(err) {
		t.Error("RunLogFields() did not create mask.go")
	}
}

func TestRunLogFieldsNoLogFields(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "config.yaml")
	outputDir := filepath.Join(tmpDir, "logger")

	// 创建没有 logfields 的 YAML 文件
	yamlContent := `service: test_service
metrics:
  - name: requests_total
    type: counter
`
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test YAML: %v", err)
	}

	opts := &LogFieldOptions{
		YamlFile:  yamlFile,
		OutputDir: outputDir,
	}

	err := RunLogFields(opts)
	if err == nil {
		t.Error("RunLogFields() should return error when no logfields defined")
	}
}

func TestRunLogFieldsInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "config.yaml")
	outputDir := filepath.Join(tmpDir, "logger")

	// 创建无效的 YAML 文件
	if err := os.WriteFile(yamlFile, []byte("invalid: yaml: content: ["), 0644); err != nil {
		t.Fatalf("Failed to create test YAML: %v", err)
	}

	opts := &LogFieldOptions{
		YamlFile:  yamlFile,
		OutputDir: outputDir,
	}

	err := RunLogFields(opts)
	if err == nil {
		t.Error("RunLogFields() should return error for invalid YAML")
	}
}

func TestGenerateMaskFunctionsNoMaskFields(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "logger")
	maskFile := "mask.go"

	// 测试没有 mask 字段的情况
	fields := []config.LogFieldConfig{
		{Name: "user_id", Type: "int64", Comment: "用户ID", Mask: false},
	}

	err := generateMaskFunctions(outputDir, maskFile, fields)
	if err != nil {
		t.Errorf("generateMaskFunctions() error = %v", err)
	}

	// 检查 mask 文件不应该生成（因为没有需要 mask 的字段）
	maskPath := filepath.Join(outputDir, maskFile)
	if _, err := os.Stat(maskPath); !os.IsNotExist(err) {
		t.Error("mask.go should not be created when no mask fields")
	}
}

func TestGenerateMaskFunctionsWithMaskFields(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "logger")
	maskFile := "mask.go"

	// 确保目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// 测试有 mask 字段的情况
	fields := []config.LogFieldConfig{
		{Name: "user_id", Type: "int64", Comment: "用户ID", Mask: true},
	}

	err := generateMaskFunctions(outputDir, maskFile, fields)
	if err != nil {
		t.Errorf("generateMaskFunctions() error = %v", err)
	}

	// 检查 mask 文件应该生成
	maskPath := filepath.Join(outputDir, maskFile)
	if _, err := os.Stat(maskPath); os.IsNotExist(err) {
		t.Error("mask.go should be created when there are mask fields")
	}

	// 检查文件内容
	content, err := os.ReadFile(maskPath)
	if err != nil {
		t.Fatalf("Failed to read mask file: %v", err)
	}
	if len(content) == 0 {
		t.Error("mask.go should not be empty")
	}
}
