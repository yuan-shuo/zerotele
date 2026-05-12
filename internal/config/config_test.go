// Package config 处理 YAML 配置文件的解析
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestToPascal(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user_id", "UserId"},
		{"request_duration_ms", "RequestDurationMs"},
		{"simple", "Simple"},
		{"", ""},
		{"a", "A"},
		{"alreadyPascal", "AlreadyPascal"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToPascal(tt.input)
			if result != tt.expected {
				t.Errorf("ToPascal(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToCamel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user_id", "userId"},
		{"request_duration_ms", "requestDurationMs"},
		{"simple", "simple"},
		{"", ""},
		{"a", "a"},
		{"AlreadyPascal", "alreadyPascal"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToCamel(tt.input)
			if result != tt.expected {
				t.Errorf("ToCamel(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLabelIsEnum(t *testing.T) {
	tests := []struct {
		name     string
		label    Label
		expected bool
	}{
		{"enum values", Label{Name: "source", Vals: []string{"app", "web"}}, true},
		{"wildcard only", Label{Name: "error", Vals: []string{"*"}}, false},
		{"empty vals", Label{Name: "empty", Vals: []string{}}, false},
		{"mixed with wildcard", Label{Name: "mixed", Vals: []string{"app", "*"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.label.IsEnum()
			if result != tt.expected {
				t.Errorf("IsEnum() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLogFieldConfigGetFieldPascalName(t *testing.T) {
	f := LogFieldConfig{Name: "user_id"}
	if got := f.GetFieldPascalName(); got != "UserId" {
		t.Errorf("GetFieldPascalName() = %v, want %v", got, "UserId")
	}
}

func TestMetricConfigGetMetricPascalName(t *testing.T) {
	m := MetricConfig{Name: "request_duration_ms"}
	if got := m.GetMetricPascalName(); got != "RequestDurationMs" {
		t.Errorf("GetMetricPascalName() = %v, want %v", got, "RequestDurationMs")
	}
}

func TestMetricConfigGetLabelNames(t *testing.T) {
	m := MetricConfig{
		Labels: []Label{
			{Name: "method"},
			{Name: "path"},
		},
	}
	names := m.GetLabelNames()
	if len(names) != 2 || names[0] != "method" || names[1] != "path" {
		t.Errorf("GetLabelNames() = %v, want [method path]", names)
	}
}

func TestLogFieldConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		field   LogFieldConfig
		wantErr bool
	}{
		{
			name:    "valid field",
			field:   LogFieldConfig{Name: "user_id", Type: "int64"},
			wantErr: false,
		},
		{
			name:    "empty name",
			field:   LogFieldConfig{Name: "", Type: "int64"},
			wantErr: true,
		},
		{
			name:    "invalid snake_case",
			field:   LogFieldConfig{Name: "UserId", Type: "int64"},
			wantErr: true,
		},
		{
			name:    "empty type",
			field:   LogFieldConfig{Name: "user_id", Type: ""},
			wantErr: true,
		},
		{
			name:    "invalid type",
			field:   LogFieldConfig{Name: "user_id", Type: "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.field.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLabelValidate(t *testing.T) {
	tests := []struct {
		name    string
		label   Label
		wantErr bool
	}{
		{
			name:    "valid label",
			label:   Label{Name: "source", Vals: []string{"app", "web"}},
			wantErr: false,
		},
		{
			name:    "empty name",
			label:   Label{Name: "", Vals: []string{"app"}},
			wantErr: true,
		},
		{
			name:    "invalid snake_case",
			label:   Label{Name: "Source", Vals: []string{"app"}},
			wantErr: true,
		},
		{
			name:    "empty vals",
			label:   Label{Name: "source", Vals: []string{}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.label.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		metric  MetricConfig
		wantErr bool
	}{
		{
			name: "valid counter",
			metric: MetricConfig{
				Name:    "requests_total",
				Help:    "Total requests",
				Type:    "counter",
				Methods: []string{"inc"},
			},
			wantErr: false,
		},
		{
			name: "valid histogram",
			metric: MetricConfig{
				Name:    "request_duration_ms",
				Help:    "Request duration",
				Type:    "histogram",
				Methods: []string{"observe"},
				Buckets: []float64{5, 10, 25},
			},
			wantErr: false,
		},
		{
			name:    "empty name",
			metric:  MetricConfig{Name: "", Help: "test", Type: "counter", Methods: []string{"inc"}},
			wantErr: true,
		},
		{
			name:    "empty help",
			metric:  MetricConfig{Name: "test", Help: "", Type: "counter", Methods: []string{"inc"}},
			wantErr: true,
		},
		{
			name:    "invalid type",
			metric:  MetricConfig{Name: "test", Help: "test", Type: "invalid", Methods: []string{"inc"}},
			wantErr: true,
		},
		{
			name:    "histogram without buckets",
			metric:  MetricConfig{Name: "test", Help: "test", Type: "histogram", Methods: []string{"observe"}},
			wantErr: true,
		},
		{
			name:    "empty methods",
			metric:  MetricConfig{Name: "test", Help: "test", Type: "counter", Methods: []string{}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metric.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTeleConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     TeleConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: TeleConfig{
				Service: "user_service",
				LogFields: []LogFieldConfig{
					{Name: "user_id", Type: "int64"},
				},
				Metrics: []MetricConfig{
					{Name: "requests_total", Help: "Total requests", Type: "counter", Methods: []string{"inc"}},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty service",
			cfg:     TeleConfig{Service: ""},
			wantErr: true,
		},
		{
			name:    "invalid service name",
			cfg:     TeleConfig{Service: "UserService"},
			wantErr: true,
		},
		{
			name: "invalid logfield",
			cfg: TeleConfig{
				Service:   "test",
				LogFields: []LogFieldConfig{{Name: "Invalid", Type: "int64"}},
			},
			wantErr: true,
		},
		{
			name: "invalid metric",
			cfg: TeleConfig{
				Service: "test",
				Metrics: []MetricConfig{{Name: "Invalid", Help: "test", Type: "counter", Methods: []string{"inc"}}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 测试有效的 YAML 文件
	validYAML := `
service: user_service
logfields:
- name: user_id
  type: int64
  mask: true
  comment: 用户ID
metrics:
- name: requests_total
  help: Total requests
  type: counter
  labels:
  - name: method
    vals: [GET, POST]
  methods: [inc]
`
	validFile := filepath.Join(tmpDir, "valid.yaml")
	if err := os.WriteFile(validFile, []byte(validYAML), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg, err := Load(validFile)
	if err != nil {
		t.Errorf("Load() error = %v", err)
	}
	if cfg.Service != "user_service" {
		t.Errorf("Load() Service = %v, want %v", cfg.Service, "user_service")
	}
	if len(cfg.LogFields) != 1 || cfg.LogFields[0].Name != "user_id" {
		t.Errorf("Load() LogFields = %v", cfg.LogFields)
	}
	if len(cfg.Metrics) != 1 || cfg.Metrics[0].Name != "requests_total" {
		t.Errorf("Load() Metrics = %v", cfg.Metrics)
	}

	// 测试不存在的文件
	_, err = Load(filepath.Join(tmpDir, "notexist.yaml"))
	if err == nil {
		t.Error("Load() should return error for non-existent file")
	}

	// 测试无效的 YAML
	invalidYAML := `
service: InvalidService
`
	invalidFile := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(invalidFile, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = Load(invalidFile)
	if err == nil {
		t.Error("Load() should return error for invalid config")
	}
}
