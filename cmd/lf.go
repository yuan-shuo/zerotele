// Package cmd 包含 zerotele 工具的核心逻辑
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/yuan-shuo/zerotele/internal/config"
	"github.com/yuan-shuo/zerotele/internal/generator"
	"github.com/yuan-shuo/zerotele/internal/maskgen"
)

// LogFieldOptions 包含日志字段命令行选项
type LogFieldOptions struct {
	YamlFile  string
	OutputDir string
	MaskFile  string
}

// ParseLogFieldFlags 解析日志字段子命令参数
func ParseLogFieldFlags(args []string) (*LogFieldOptions, error) {
	// 先手动解析 -d 和 -m 参数
	var opts LogFieldOptions
	var newArgs []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-d":
			if i+1 < len(args) {
				opts.OutputDir = args[i+1]
				i++
			}
		case "-m":
			if i+1 < len(args) {
				opts.MaskFile = args[i+1]
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

// RunLogFields 执行日志字段代码生成
func RunLogFields(opts *LogFieldOptions) error {
	// 加载配置
	cfg, err := config.Load(opts.YamlFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// 检查是否有日志字段配置
	if len(cfg.LogFields) == 0 {
		return fmt.Errorf("no logfields defined in config")
	}

	// 创建生成器
	gen, err := generator.New()
	if err != nil {
		return fmt.Errorf("creating generator: %w", err)
	}

	// 生成日志字段代码
	genOpts := generator.Options{
		OutputDir: opts.OutputDir,
	}
	if err := gen.GenerateLogFields(cfg.LogFields, genOpts); err != nil {
		return fmt.Errorf("generating logfields code: %w", err)
	}

	fmt.Printf("Generated %s/logfields_gen.go\n", opts.OutputDir)

	// 生成 mask 函数（如果指定了 -m 参数）
	if opts.MaskFile != "" {
		if err := generateMaskFunctions(opts.OutputDir, opts.MaskFile, cfg.LogFields); err != nil {
			return fmt.Errorf("generating mask functions: %w", err)
		}
	}

	return nil
}

// generateMaskFunctions 生成 mask 函数
func generateMaskFunctions(outputDir, maskFile string, fields []config.LogFieldConfig) error {
	// 如果没有需要 mask 的字段，跳过
	if !maskgen.HasMaskFields(fields) {
		fmt.Println("No mask fields found, skipping mask generation")
		return nil
	}

	maskPath := filepath.Join(outputDir, maskFile)

	// 确定包名
	packageName := filepath.Base(outputDir)
	if packageName == "." || packageName == "/" || packageName == "" {
		packageName = "logger"
	}

	maskGen := maskgen.New(packageName)
	if err := maskGen.Generate(maskPath, fields); err != nil {
		return err
	}

	fmt.Printf("Generated/Updated %s\n", maskPath)
	return nil
}
