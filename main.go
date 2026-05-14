// zerotele 是一个代码生成工具，根据 YAML 配置生成结构化日志字段代码和 Prometheus 指标代码
package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/yuan-shuo/zerotele/cmd"
)

//go:embed version.json
var versionJSON []byte

var version string

func init() {
	var v struct {
		Version string `json:"version"`
		Suffix  string `json:"suffix"`
	}
	if err := json.Unmarshal(versionJSON, &v); err != nil {
		version = "unknown"
	} else {
		version = v.Version
		if v.Suffix != "" {
			version = version + "-" + v.Suffix
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "lf":
		if err := runLogFields(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "met":
		if err := runMetrics(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "check":
		if err := runCheck(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version", "-v", "--version":
		fmt.Printf("zerotele version %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func runLogFields(args []string) error {
	opts, err := cmd.ParseLogFieldFlags(args)
	if err != nil {
		return err
	}
	opts.Version = version
	return cmd.RunLogFields(opts)
}

func runMetrics(args []string) error {
	opts, err := cmd.ParseMetricFlags(args)
	if err != nil {
		return err
	}
	opts.Version = version
	return cmd.RunMetrics(opts)
}

func runCheck(args []string) error {
	opts, err := cmd.ParseCheckFlags(args)
	if err != nil {
		return err
	}
	return cmd.RunCheck(opts)
}

func printUsage() {
	fmt.Println(`zerotele - 统一的日志字段和指标代码生成工具

用法:
  zerotele <command> [options] <yaml-file>

命令:
  lf     生成日志字段代码
  met    生成指标代码
  check  校验 YAML 配置文件

日志字段命令 (lf):
  zerotele lf [options] <yaml-file>
  
  选项:
    -d string    输出目录 (必填)
    -m string    生成/追加 mask 函数到指定文件 (可选)
  
  示例:
    zerotele lf zerotele.yaml -d ./internal/logger
    zerotele lf zerotele.yaml -d ./internal/logger -m mask.go

指标命令 (met):
  zerotele met [options] <yaml-file>
  
  选项:
    -d string    输出目录 (必填)
  
  示例:
    zerotele met zerotele.yaml -d ./internal/metrics

校验命令 (check):
  zerotele check <yaml-file>

  示例:
    zerotele check zerotele.yaml

其他命令:
  version, -v, --version   显示版本信息
  help, -h, --help          显示帮助信息

配置文件格式参考 zerotele.example.yaml`)
}
