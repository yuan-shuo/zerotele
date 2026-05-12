// Package cmd 包含 zerotele 工具的核心逻辑
package cmd

import (
	"testing"
)

func TestParseLogFieldFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		wantYaml string
		wantDir  string
		wantMask string
	}{
		{
			name:     "valid args with mask",
			args:     []string{"config.yaml", "-d", "./output", "-m", "mask.go"},
			wantErr:  false,
			wantYaml: "config.yaml",
			wantDir:  "./output",
			wantMask: "mask.go",
		},
		{
			name:     "valid args without mask",
			args:     []string{"config.yaml", "-d", "./output"},
			wantErr:  false,
			wantYaml: "config.yaml",
			wantDir:  "./output",
			wantMask: "",
		},
		{
			name:     "args in different order",
			args:     []string{"-d", "./logger", "-m", "mask.go", "config.yaml"},
			wantErr:  false,
			wantYaml: "config.yaml",
			wantDir:  "./logger",
			wantMask: "mask.go",
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
			opts, err := ParseLogFieldFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLogFieldFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if opts.YamlFile != tt.wantYaml {
					t.Errorf("ParseLogFieldFlags() YamlFile = %v, want %v", opts.YamlFile, tt.wantYaml)
				}
				if opts.OutputDir != tt.wantDir {
					t.Errorf("ParseLogFieldFlags() OutputDir = %v, want %v", opts.OutputDir, tt.wantDir)
				}
				if opts.MaskFile != tt.wantMask {
					t.Errorf("ParseLogFieldFlags() MaskFile = %v, want %v", opts.MaskFile, tt.wantMask)
				}
			}
		})
	}
}
