// Package cmd 包含 zerotele 工具的核心逻辑
package cmd

import (
	"testing"
)

func TestParseMetricFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		wantYaml string
		wantDir  string
	}{
		{
			name:     "valid args",
			args:     []string{"config.yaml", "-d", "./output"},
			wantErr:  false,
			wantYaml: "config.yaml",
			wantDir:  "./output",
		},
		{
			name:     "args in different order",
			args:     []string{"-d", "./metrics", "config.yaml"},
			wantErr:  false,
			wantYaml: "config.yaml",
			wantDir:  "./metrics",
		},
		{
			name:    "missing -d flag",
			args:    []string{"config.yaml"},
			wantErr: true,
		},
		{
			name:    "missing yaml file",
			args:    []string{"-d", "./output"},
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
			opts, err := ParseMetricFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMetricFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if opts.YamlFile != tt.wantYaml {
					t.Errorf("ParseMetricFlags() YamlFile = %v, want %v", opts.YamlFile, tt.wantYaml)
				}
				if opts.OutputDir != tt.wantDir {
					t.Errorf("ParseMetricFlags() OutputDir = %v, want %v", opts.OutputDir, tt.wantDir)
				}
			}
		})
	}
}
