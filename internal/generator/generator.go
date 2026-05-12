// Package generator 处理代码生成
package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/yuan-shuo/zerotele/internal/config"
	tmplpkg "github.com/yuan-shuo/zerotele/internal/template"
)

// Options 包含代码生成选项
type Options struct {
	OutputDir string
}

// Generator 负责生成代码
type Generator struct {
	logTmpl *template.Template
	metTmpl *template.Template
}

// New 创建一个新的生成器
func New() (*Generator, error) {
	logTmpl, err := template.New("logfields").Funcs(tmplpkg.FuncMap()).Parse(tmplpkg.LogFieldsTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing logfields template: %w", err)
	}

	metTmpl, err := template.New("metrics").Funcs(tmplpkg.FuncMap()).Parse(tmplpkg.MetricsTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing metrics template: %w", err)
	}

	return &Generator{
		logTmpl: logTmpl,
		metTmpl: metTmpl,
	}, nil
}

// GenerateLogFields 生成日志字段代码
func (g *Generator) GenerateLogFields(fields []config.LogFieldConfig, opts Options) error {
	data := struct {
		Fields      []config.LogFieldConfig
		PackageName string
	}{
		Fields:      fields,
		PackageName: getPackageName(opts.OutputDir),
	}

	var buf bytes.Buffer
	if err := g.logTmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing logfields template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting logfields code: %w\nRaw output:\n%s", err, buf.String())
	}

	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory %s: %w", opts.OutputDir, err)
	}

	outputPath := filepath.Join(opts.OutputDir, "logfields_gen.go")
	if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
		return fmt.Errorf("writing logfields file %s: %w", outputPath, err)
	}

	return nil
}

// GenerateMetrics 生成指标代码
func (g *Generator) GenerateMetrics(service string, metrics []config.MetricConfig, opts Options) error {
	data := struct {
		Service     string
		Metrics     []config.MetricConfig
		PackageName string
	}{
		Service:     service,
		Metrics:     metrics,
		PackageName: getPackageName(opts.OutputDir),
	}

	var buf bytes.Buffer
	if err := g.metTmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing metrics template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting metrics code: %w\nRaw output:\n%s", err, buf.String())
	}

	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory %s: %w", opts.OutputDir, err)
	}

	outputPath := filepath.Join(opts.OutputDir, "metrics_gen.go")
	if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
		return fmt.Errorf("writing metrics file %s: %w", outputPath, err)
	}

	return nil
}

// getPackageName 从输出目录路径提取包名
func getPackageName(outputDir string) string {
	normalized := strings.ReplaceAll(outputDir, "\\", "/")
	cleanPath := path.Clean(normalized)
	base := path.Base(cleanPath)

	if base == "." || base == "/" || base == "" {
		return "main"
	}

	// 检查是否是有效的 Go 标识符
	if !isValidGoIdentifier(base) {
		return "main"
	}

	return base
}

// isValidGoIdentifier 检查字符串是否是有效的 Go 标识符
func isValidGoIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// 第一个字符必须是字母或下划线
	c := s[0]
	if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_') {
		return false
	}

	// 其余字符可以是字母、数字或下划线
	for i := 1; i < len(s); i++ {
		c = s[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}

	return true
}
