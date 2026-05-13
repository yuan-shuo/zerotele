// Package cmd 包含 zerotele 工具的核心逻辑
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/yuan-shuo/zerotele/internal/config"
)

// CheckOptions 包含 check 子命令的选项
type CheckOptions struct {
	YamlFile string
}

// ParseCheckFlags 解析 check 子命令的参数
func ParseCheckFlags(args []string) (*CheckOptions, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("YAML file path is required")
	}

	return &CheckOptions{
		YamlFile: args[0],
	}, nil
}

// RunCheck 执行 YAML 文件校验
func RunCheck(opts *CheckOptions) error {
	// 检查文件是否存在
	if _, err := os.Stat(opts.YamlFile); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", opts.YamlFile)
	}

	fmt.Printf("Checking %s...\n", opts.YamlFile)

	// 尝试加载并校验配置
	_, err := config.Load(opts.YamlFile)
	if err != nil {
		fmt.Println("Validation failed:")
		// 处理 errors.Join 合并的多个错误
		// 尝试将错误按换行分割并逐行打印
		errStr := err.Error()
		// errors.Join 使用 "\n" 连接错误，我们将其分割并格式化输出
		lines := strings.Split(errStr, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				fmt.Printf("  - %s\n", line)
			}
		}
		return fmt.Errorf("validation failed with %d error(s)", len(lines))
	}

	// 输出校验结果
	fmt.Println("YAML file is valid!")
	return nil
}
