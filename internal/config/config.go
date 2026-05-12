// Package config 处理 YAML 配置文件的解析
package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// snakeCaseRegex 匹配有效的 snake_case 格式（OpenTelemetry 规范）
var snakeCaseRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)

// 支持的 Go 基础类型
var validTypes = map[string]bool{
	"bool":       true,
	"string":     true,
	"int":        true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"uint":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"uintptr":    true,
	"byte":       true,
	"rune":       true,
	"float32":    true,
	"float64":    true,
	"complex64":  true,
	"complex128": true,
}

// TeleConfig 表示统一的 YAML 配置根结构
type TeleConfig struct {
	Service   string           `yaml:"service"`
	LogFields []LogFieldConfig `yaml:"logfields"`
	Metrics   []MetricConfig   `yaml:"metrics"`
}

// LogFieldConfig 表示日志字段配置
type LogFieldConfig struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Mask    bool   `yaml:"mask"`
	Comment string `yaml:"comment"`
}

// MetricConfig 表示指标配置
type MetricConfig struct {
	Name    string    `yaml:"name"`
	Help    string    `yaml:"help"`
	Type    string    `yaml:"type"` // counter, gauge, histogram
	Labels  []Label   `yaml:"labels"`
	Methods []string  `yaml:"methods"`
	Buckets []float64 `yaml:"buckets,omitempty"` // 仅 histogram 使用
}

// Label 表示指标标签
type Label struct {
	Name string   `yaml:"name"`
	Vals []string `yaml:"vals"`
}

// GetLabelNames 返回标签名称列表
func (m *MetricConfig) GetLabelNames() []string {
	names := make([]string, len(m.Labels))
	for i, l := range m.Labels {
		names[i] = l.Name
	}
	return names
}

// Load 从指定路径加载 YAML 配置文件
func Load(path string) (*TeleConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading YAML file %s: %w", path, err)
	}

	var cfg TeleConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	// 校验配置
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate 校验配置是否合法
func (c *TeleConfig) Validate() error {
	// 校验 service 名称
	if strings.TrimSpace(c.Service) == "" {
		return fmt.Errorf("service name is required")
	}
	if !snakeCaseRegex.MatchString(c.Service) {
		return fmt.Errorf("service name '%s' must be valid snake_case format", c.Service)
	}

	// 校验日志字段
	for i, f := range c.LogFields {
		if err := f.Validate(); err != nil {
			return fmt.Errorf("logfield[%d]: %w", i, err)
		}
	}

	// 校验指标
	for i, m := range c.Metrics {
		if err := m.Validate(); err != nil {
			return fmt.Errorf("metric[%d]: %w", i, err)
		}
	}

	return nil
}

// Validate 校验日志字段配置
func (f LogFieldConfig) Validate() error {
	if strings.TrimSpace(f.Name) == "" {
		return fmt.Errorf("field name is required")
	}
	if !snakeCaseRegex.MatchString(f.Name) {
		return fmt.Errorf("field name '%s' must be valid snake_case format", f.Name)
	}
	if strings.TrimSpace(f.Type) == "" {
		return fmt.Errorf("field '%s': type is required", f.Name)
	}
	if !validTypes[f.Type] {
		return fmt.Errorf("field '%s': invalid type '%s'", f.Name, f.Type)
	}
	return nil
}

// Validate 校验指标配置
func (m MetricConfig) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("metric name is required")
	}
	if !snakeCaseRegex.MatchString(m.Name) {
		return fmt.Errorf("metric name '%s' must be valid snake_case format", m.Name)
	}
	if strings.TrimSpace(m.Help) == "" {
		return fmt.Errorf("metric '%s': help is required", m.Name)
	}
	if m.Type != "counter" && m.Type != "gauge" && m.Type != "histogram" {
		return fmt.Errorf("metric '%s': invalid type '%s', must be counter/gauge/histogram", m.Name, m.Type)
	}
	if m.Type == "histogram" && len(m.Buckets) == 0 {
		return fmt.Errorf("metric '%s': histogram requires buckets", m.Name)
	}
	if len(m.Methods) == 0 {
		return fmt.Errorf("metric '%s': methods are required", m.Name)
	}

	// 校验标签
	for i, l := range m.Labels {
		if err := l.Validate(); err != nil {
			return fmt.Errorf("label[%d]: %w", i, err)
		}
	}

	return nil
}

// Validate 校验标签配置
func (l Label) Validate() error {
	if strings.TrimSpace(l.Name) == "" {
		return fmt.Errorf("label name is required")
	}
	if !snakeCaseRegex.MatchString(l.Name) {
		return fmt.Errorf("label name '%s' must be valid snake_case format", l.Name)
	}
	if len(l.Vals) == 0 {
		return fmt.Errorf("label '%s': vals are required", l.Name)
	}
	return nil
}

// ToPascal 将 snake_case 转换为 PascalCase
func ToPascal(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// ToCamel 将 snake_case 转换为 camelCase（首字母小写）
func ToCamel(s string) string {
	pascal := ToPascal(s)
	if len(pascal) > 0 {
		return strings.ToLower(pascal[:1]) + pascal[1:]
	}
	return pascal
}

// GetFieldPascalName 获取日志字段的 PascalCase 名称
func (f LogFieldConfig) GetFieldPascalName() string {
	return ToPascal(f.Name)
}

// GetMetricPascalName 获取指标的 PascalCase 名称
func (m MetricConfig) GetMetricPascalName() string {
	return ToPascal(m.Name)
}

// IsEnum 检查标签值是否是枚举（不是通配符 *）
func (l Label) IsEnum() bool {
	if len(l.Vals) == 0 {
		return false
	}
	for _, v := range l.Vals {
		if v == "*" {
			return false
		}
	}
	return true
}
