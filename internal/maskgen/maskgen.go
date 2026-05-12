// Package maskgen 处理 mask 函数的生成和补全
package maskgen

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/yuan-shuo/zerotele/internal/config"
)

// Generator 负责生成和补全 mask 函数
type Generator struct {
	packageName string
}

// New 创建一个新的 mask 生成器
func New(packageName string) *Generator {
	return &Generator{
		packageName: packageName,
	}
}

// Generate 生成或补全 mask 函数
func (g *Generator) Generate(maskFilePath string, fields []config.LogFieldConfig) error {
	// 过滤出需要 mask 的字段
	maskFields := filterMaskFields(fields)
	if len(maskFields) == 0 {
		return nil
	}

	// 检查 mask 文件是否已存在
	isNewFile := false
	existingFuncs := make(map[string]bool)
	var existingContent string

	if _, err := os.Stat(maskFilePath); err != nil {
		// 文件不存在，标记为新文件
		isNewFile = true
	} else {
		content, funcs, err := g.parseExistingMaskFile(maskFilePath)
		if err != nil {
			return fmt.Errorf("parsing existing mask file: %w", err)
		}
		existingContent = content
		existingFuncs = funcs
	}

	// 找出需要新增的 mask 函数
	var newFields []config.LogFieldConfig
	for _, field := range maskFields {
		funcName := "mask" + field.GetFieldPascalName()
		if !existingFuncs[funcName] {
			newFields = append(newFields, field)
		}
	}

	if len(newFields) == 0 {
		return nil
	}

	// 生成新的 mask 函数代码
	newCode := g.generateMaskCode(newFields, isNewFile)

	// 写入文件
	if isNewFile {
		// 新文件：生成完整代码并进行 gofmt 格式化
		formatted, err := format.Source([]byte(newCode))
		if err != nil {
			// 格式化失败，使用原始代码
			formatted = []byte(newCode)
		}
		if err := os.WriteFile(maskFilePath, formatted, 0644); err != nil {
			return fmt.Errorf("writing mask file: %w", err)
		}
	} else {
		// 已有文件：直接追加新代码，不进行格式化
		existingContent = strings.TrimRight(existingContent, "\n")
		fullContent := existingContent + "\n\n" + newCode + "\n"
		if err := os.WriteFile(maskFilePath, []byte(fullContent), 0644); err != nil {
			return fmt.Errorf("writing mask file: %w", err)
		}
	}

	return nil
}

// filterMaskFields 过滤出需要 mask 的字段
func filterMaskFields(fields []config.LogFieldConfig) []config.LogFieldConfig {
	var result []config.LogFieldConfig
	for _, f := range fields {
		if f.Mask {
			result = append(result, f)
		}
	}
	return result
}

// parseExistingMaskFile 解析已存在的 mask 文件
func (g *Generator) parseExistingMaskFile(path string) (string, map[string]bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		return "", nil, fmt.Errorf("parsing Go file: %w", err)
	}

	funcs := make(map[string]bool)
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			funcs[fn.Name.Name] = true
		}
	}

	return string(content), funcs, nil
}

// generateMaskCode 生成 mask 函数代码
func (g *Generator) generateMaskCode(fields []config.LogFieldConfig, isNewFile bool) string {
	var sb strings.Builder

	if isNewFile {
		sb.WriteString(fmt.Sprintf("package %s\n", g.packageName))
	}

	for i, field := range fields {
		if i > 0 || isNewFile {
			sb.WriteString("\n")
		}

		funcName := "mask" + field.GetFieldPascalName()
		paramName := strings.ToLower(field.GetFieldPascalName())

		sb.WriteString(fmt.Sprintf("// %s 对%s进行脱敏\n", funcName, field.Comment))
		sb.WriteString("// 请在此实现具体的脱敏逻辑\n")
		sb.WriteString(fmt.Sprintf("func %s(%s %s) any {\n", funcName, paramName, field.Type))
		sb.WriteString("\t// TODO: 实现脱敏逻辑\n")
		sb.WriteString(fmt.Sprintf("\treturn %s\n", paramName))
		sb.WriteString("}")
	}

	return sb.String()
}

// HasMaskFields 检查是否有需要 mask 的字段
func HasMaskFields(fields []config.LogFieldConfig) bool {
	for _, f := range fields {
		if f.Mask {
			return true
		}
	}
	return false
}
