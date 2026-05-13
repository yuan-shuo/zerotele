// Package cmd 包含 zerotele 工具的核心逻辑
package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCheckFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		wantYaml string
	}{
		{
			name:     "valid args",
			args:     []string{"config.yaml"},
			wantErr:  false,
			wantYaml: "config.yaml",
		},
		{
			name:     "valid args with path",
			args:     []string{"./testoutput/zerotele.yaml"},
			wantErr:  false,
			wantYaml: "./testoutput/zerotele.yaml",
		},
		{
			name:    "missing yaml file",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "empty args",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseCheckFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCheckFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if opts.YamlFile != tt.wantYaml {
					t.Errorf("ParseCheckFlags() YamlFile = %v, want %v", opts.YamlFile, tt.wantYaml)
				}
			}
		})
	}
}

func TestRunCheck(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()

	// 创建有效的 YAML 文件
	validYaml := `service: test_service
logfields:
  - name: user_id
    type: int64
    comment: 用户ID
metrics:
  - name: request_total
    help: Total requests
    type: counter
    labels:
      - name: method
        vals: [GET, POST]
    methods: [inc]
`
	validFile := filepath.Join(tempDir, "valid.yaml")
	if err := os.WriteFile(validFile, []byte(validYaml), 0644); err != nil {
		t.Fatalf("Failed to create valid yaml file: %v", err)
	}

	// 创建包含多个错误的 YAML 文件
	invalidYaml := `service: 1invalid
logfields:
  - name: UserID
    type: invalid_type
    comment: 用户ID
metrics:
  - name: 2invalid_metric
    help: ""
    type: unknown_type
    labels:
      - name: 3bad_label
        vals: []
    methods: []
`
	invalidFile := filepath.Join(tempDir, "invalid.yaml")
	if err := os.WriteFile(invalidFile, []byte(invalidYaml), 0644); err != nil {
		t.Fatalf("Failed to create invalid yaml file: %v", err)
	}

	// 创建格式错误的 YAML 文件
	badFormatYaml := `service: test
logfields:
  - name: user_id
    type: int64
  - invalid yaml content: [
`
	badFormatFile := filepath.Join(tempDir, "bad_format.yaml")
	if err := os.WriteFile(badFormatFile, []byte(badFormatYaml), 0644); err != nil {
		t.Fatalf("Failed to create bad format yaml file: %v", err)
	}

	tests := []struct {
		name    string
		opts    *CheckOptions
		wantErr bool
	}{
		{
			name: "valid yaml file",
			opts: &CheckOptions{
				YamlFile: validFile,
			},
			wantErr: false,
		},
		{
			name: "invalid yaml file with multiple errors",
			opts: &CheckOptions{
				YamlFile: invalidFile,
			},
			wantErr: true,
		},
		{
			name: "bad format yaml file",
			opts: &CheckOptions{
				YamlFile: badFormatFile,
			},
			wantErr: true,
		},
		{
			name: "non-existent file",
			opts: &CheckOptions{
				YamlFile: filepath.Join(tempDir, "not_exist.yaml"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunCheck(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunCheck_MultipleErrors(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()

	// 创建包含多个错误的 YAML 文件
	invalidYaml := `service: 1invalid
logfields:
  - name: UserID
    type: invalid_type
  - name: another_bad_name
    type: bad_type
metrics:
  - name: bad_metric_name
    help: ""
    type: counter
    labels:
      - name: bad_label
        vals: [value1]
    methods: []
`
	invalidFile := filepath.Join(tempDir, "multiple_errors.yaml")
	if err := os.WriteFile(invalidFile, []byte(invalidYaml), 0644); err != nil {
		t.Fatalf("Failed to create invalid yaml file: %v", err)
	}

	opts := &CheckOptions{
		YamlFile: invalidFile,
	}

	err := RunCheck(opts)
	if err == nil {
		t.Error("RunCheck() expected error for invalid yaml, got nil")
	}
}
