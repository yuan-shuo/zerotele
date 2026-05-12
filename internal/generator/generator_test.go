// Package generator 处理代码生成
package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yuan-shuo/zerotele/internal/config"
)

func TestNew(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if gen == nil {
		t.Fatal("New() returned nil")
	}
	if gen.metricsTmpl == nil {
		t.Error("New() metricsTmpl should not be nil")
	}
	if gen.logTmpl == nil {
		t.Error("New() logTmpl should not be nil")
	}
}

func TestGetPackageName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"metrics", "metrics"},
		{"./metrics", "metrics"},
		{"aaa/bbb/ccc", "ccc"},
		{"aaa\\bbb\\ccc", "ccc"},
		{".", "metrics"},
		{"/", "metrics"},
		{"", "metrics"},
		{"./", "metrics"},
		{"../metrics", "metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := getPackageName(tt.input)
			if result != tt.expected {
				t.Errorf("getPackageName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateMetrics(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tmpDir := t.TempDir()
	// 使用子目录确保包名有效
	outputDir := filepath.Join(tmpDir, "metrics")

	cfg := &config.TeleConfig{
		Service: "user",
		Metrics: []config.MetricConfig{
			{
				Name:    "requests_total",
				Help:    "Total requests",
				Type:    "counter",
				Methods: []string{"inc"},
				Labels: []config.Label{
					{Name: "method", Vals: []string{"GET", "POST"}},
				},
			},
		},
	}

	opts := MetricsOptions{OutputDir: outputDir}
	err = gen.GenerateMetrics(cfg, opts)
	if err != nil {
		t.Errorf("GenerateMetrics() error = %v", err)
	}

	// 检查文件是否生成
	outputPath := filepath.Join(outputDir, "metrics_gen.go")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("GenerateMetrics() did not create output file")
	}

	// 检查文件内容
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if len(content) == 0 {
		t.Error("GenerateMetrics() created empty file")
	}
}

func TestGenerateMetricsWithHistogram(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "metrics")

	cfg := &config.TeleConfig{
		Service: "user",
		Metrics: []config.MetricConfig{
			{
				Name:    "request_duration_ms",
				Help:    "Request duration",
				Type:    "histogram",
				Methods: []string{"observe"},
				Buckets: []float64{5, 10, 25, 50, 100},
				Labels: []config.Label{
					{Name: "method", Vals: []string{"*"}},
				},
			},
		},
	}

	opts := MetricsOptions{OutputDir: outputDir}
	err = gen.GenerateMetrics(cfg, opts)
	if err != nil {
		t.Errorf("GenerateMetrics() error = %v", err)
	}

	// 检查文件是否生成
	outputPath := filepath.Join(outputDir, "metrics_gen.go")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// 检查是否包含 histogram 相关内容
	contentStr := string(content)
	if !contains(contentStr, "HistogramVec") {
		t.Error("Generated code should contain HistogramVec")
	}
	if !contains(contentStr, "Observe") {
		t.Error("Generated code should contain Observe method")
	}
}

func TestGenerateMetricsWithGauge(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "metrics")

	cfg := &config.TeleConfig{
		Service: "user",
		Metrics: []config.MetricConfig{
			{
				Name:    "active_connections",
				Help:    "Active connections",
				Type:    "gauge",
				Methods: []string{"set", "inc", "dec"},
				Labels:  []config.Label{},
			},
		},
	}

	opts := MetricsOptions{OutputDir: outputDir}
	err = gen.GenerateMetrics(cfg, opts)
	if err != nil {
		t.Errorf("GenerateMetrics() error = %v", err)
	}

	outputPath := filepath.Join(outputDir, "metrics_gen.go")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "GaugeVec") {
		t.Error("Generated code should contain GaugeVec")
	}
	if !contains(contentStr, "Set") {
		t.Error("Generated code should contain Set method")
	}
	if !contains(contentStr, "Inc") {
		t.Error("Generated code should contain Inc method")
	}
	if !contains(contentStr, "Dec") {
		t.Error("Generated code should contain Dec method")
	}
}

func TestGenerateLogFields(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "logger")

	cfg := &config.TeleConfig{
		Service: "user",
		LogFields: []config.LogFieldConfig{
			{Name: "user_id", Type: "int64", Comment: "用户ID", Mask: true},
			{Name: "user_name", Type: "string", Comment: "用户名", Mask: false},
		},
	}

	opts := LogOptions{OutputDir: outputDir}
	err = gen.GenerateLogFields(cfg, opts)
	if err != nil {
		t.Errorf("GenerateLogFields() error = %v", err)
	}

	// 检查文件是否生成
	outputPath := filepath.Join(outputDir, "logfields_gen.go")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("GenerateLogFields() did not create output file")
	}

	// 检查文件内容
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if len(content) == 0 {
		t.Error("GenerateLogFields() created empty file")
	}

	contentStr := string(content)
	if !contains(contentStr, "UserId") {
		t.Error("Generated code should contain UserId type")
	}
	if !contains(contentStr, "UserName") {
		t.Error("Generated code should contain UserName type")
	}
	if !contains(contentStr, "MaskSensitive") {
		t.Error("Generated code should contain MaskSensitive for masked field")
	}
}

func TestGenerateLogFieldsEmpty(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "logger")

	cfg := &config.TeleConfig{
		Service:   "user",
		LogFields: []config.LogFieldConfig{},
	}

	opts := LogOptions{OutputDir: outputDir}
	err = gen.GenerateLogFields(cfg, opts)
	if err != nil {
		t.Errorf("GenerateLogFields() error = %v", err)
	}

	outputPath := filepath.Join(outputDir, "logfields_gen.go")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("GenerateLogFields() should create file even with empty fields")
	}
}

func TestGenerateMetricsNestedDir(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "aaa", "bbb", "metrics")

	cfg := &config.TeleConfig{
		Service: "user",
		Metrics: []config.MetricConfig{
			{
				Name:    "test_metric",
				Help:    "Test metric",
				Type:    "counter",
				Methods: []string{"inc"},
			},
		},
	}

	opts := MetricsOptions{OutputDir: nestedDir}
	err = gen.GenerateMetrics(cfg, opts)
	if err != nil {
		t.Errorf("GenerateMetrics() error = %v", err)
	}

	outputPath := filepath.Join(nestedDir, "metrics_gen.go")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("GenerateMetrics() should create nested directories")
	}

	// 检查包名是否正确
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if !contains(string(content), "package metrics") {
		t.Error("Generated code should have correct package name")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
