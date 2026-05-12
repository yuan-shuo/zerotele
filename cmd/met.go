// Package cmd 包含 zerotele 工具的核心逻辑
package cmd

import (
	"fmt"

	"github.com/yuan-shuo/zerotele/internal/config"
	"github.com/yuan-shuo/zerotele/internal/generator"
)

// MetricOptions 包含指标命令行选项
type MetricOptions struct {
	YamlFile  string
	OutputDir string
}

// ParseMetricFlags 解析指标子命令参数
func ParseMetricFlags(args []string) (*MetricOptions, error) {
	// 先手动解析 -d 参数
	var opts MetricOptions
	var newArgs []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-d":
			if i+1 < len(args) {
				opts.OutputDir = args[i+1]
				i++
			}
		default:
			newArgs = append(newArgs, args[i])
		}
	}

	// 验证必选参数
	if opts.OutputDir == "" {
		return nil, fmt.Errorf("-d is required")
	}

	// 获取 YAML 文件路径
	if len(newArgs) < 1 {
		return nil, fmt.Errorf("YAML file path is required")
	}
	opts.YamlFile = newArgs[0]

	return &opts, nil
}

// RunMetrics 执行指标代码生成
func RunMetrics(opts *MetricOptions) error {
	// 加载配置
	cfg, err := config.Load(opts.YamlFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// 检查是否有指标配置
	if len(cfg.Metrics) == 0 {
		return fmt.Errorf("no metrics defined in config")
	}

	// 创建生成器
	gen, err := generator.New()
	if err != nil {
		return fmt.Errorf("creating generator: %w", err)
	}

	// 生成指标代码
	genOpts := generator.MetricsOptions{
		OutputDir: opts.OutputDir,
	}
	if err := gen.GenerateMetrics(cfg, genOpts); err != nil {
		return fmt.Errorf("generating metrics code: %w", err)
	}

	fmt.Printf("Generated %s/metrics_gen.go\n", opts.OutputDir)
	return nil
}
