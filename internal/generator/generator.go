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

// MetricsOptions 包含指标代码生成选项
type MetricsOptions struct {
	OutputDir string
	Version   string
}

// Generator 负责生成代码
type Generator struct {
	metricsTmpl *template.Template
	logTmpl     *template.Template
}

// New 创建一个新的生成器
func New() (*Generator, error) {
	// 解析指标模板
	metricsTmpl, err := template.New("metrics").Funcs(tmplpkg.FuncMap()).Parse(tmplpkg.MetricsTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing metrics template: %w", err)
	}

	// 解析日志模板
	logTmpl, err := template.New("logfields").Funcs(tmplpkg.LogFuncMap()).Parse(tmplpkg.LogFieldsTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing logfields template: %w", err)
	}

	return &Generator{
		metricsTmpl: metricsTmpl,
		logTmpl:     logTmpl,
	}, nil
}

// GenerateMetrics 根据配置生成指标代码文件
func (g *Generator) GenerateMetrics(cfg *config.TeleConfig, opts MetricsOptions) error {
	// 准备模板数据
	data := struct {
		Service     string
		Metrics     []config.MetricConfig
		PackageName string
		Version     string
	}{
		Service:     cfg.Service,
		Metrics:     cfg.Metrics,
		PackageName: getPackageName(opts.OutputDir),
		Version:     opts.Version,
	}

	// 执行模板
	var buf bytes.Buffer
	if err := g.metricsTmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing metrics template: %w", err)
	}

	// 格式化代码
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting code: %w\nRaw output:\n%s", err, buf.String())
	}

	// 确保输出目录存在
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory %s: %w", opts.OutputDir, err)
	}

	// 写入文件
	outputPath := filepath.Join(opts.OutputDir, "metrics_gen.go")
	if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
		return fmt.Errorf("writing file %s: %w", outputPath, err)
	}

	return nil
}

// LogOptions 包含日志字段代码生成选项
type LogOptions struct {
	OutputDir string
	Version   string
}

// GenerateLogFields 根据配置生成日志字段代码文件
func (g *Generator) GenerateLogFields(cfg *config.TeleConfig, opts LogOptions) error {
	// 准备模板数据
	data := struct {
		Service     string
		LogFields   []config.LogFieldConfig
		PackageName string
		Version     string
	}{
		Service:     cfg.Service,
		LogFields:   cfg.LogFields,
		PackageName: getPackageName(opts.OutputDir),
		Version:     opts.Version,
	}

	// 执行模板
	var buf bytes.Buffer
	if err := g.logTmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing logfields template: %w", err)
	}

	// 格式化代码
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting code: %w\nRaw output:\n%s", err, buf.String())
	}

	// 确保输出目录存在
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory %s: %w", opts.OutputDir, err)
	}

	// 写入文件
	outputPath := filepath.Join(opts.OutputDir, "logfields_gen.go")
	if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
		return fmt.Errorf("writing file %s: %w", outputPath, err)
	}

	return nil
}

// getPackageName 从输出目录路径提取包名
// 例如: "aaa/aaa/bbb" -> "bbb", "." -> "metrics"
func getPackageName(outputDir string) string {
	// 统一使用 / 作为分隔符处理
	normalized := strings.ReplaceAll(outputDir, "\\", "/")
	cleanPath := path.Clean(normalized)
	base := path.Base(cleanPath)

	// 如果路径是 "." 或 "/" 等，使用默认包名
	if base == "." || base == "/" || base == "" {
		return "metrics"
	}

	return base
}
