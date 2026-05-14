// Package template 提供日志字段代码生成的模板定义
package template

import (
	"strings"
	"testing"
	"text/template"

	"github.com/yuan-shuo/zerotele/internal/config"
)

func TestLogFuncMap(t *testing.T) {
	fm := LogFuncMap()
	if fm == nil {
		t.Error("LogFuncMap() should not return nil")
	}

	expectedFuncs := []string{"toPascal", "toLower", "toCamel"}
	for _, fn := range expectedFuncs {
		if _, ok := fm[fn]; !ok {
			t.Errorf("LogFuncMap() missing function %q", fn)
		}
	}
}

func TestLogFieldsTemplate(t *testing.T) {
	// 测试模板是否可以正常解析
	tmpl, err := template.New("test").Funcs(LogFuncMap()).Parse(LogFieldsTemplate)
	if err != nil {
		t.Fatalf("Failed to parse LogFieldsTemplate: %v", err)
	}

	// 准备测试数据
	data := struct {
		Service     string
		LogFields   []config.LogFieldConfig
		PackageName string
		Version     string
	}{
		Service: "user",
		LogFields: []config.LogFieldConfig{
			{Name: "user_id", Type: "int64", Comment: "用户ID", Mask: true},
			{Name: "user_name", Type: "string", Comment: "用户名", Mask: false},
		},
		PackageName: "logger",
		Version:     "0.1.0",
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	result := buf.String()
	if result == "" {
		t.Error("Template output should not be empty")
	}

	// 检查生成的代码包含关键内容
	if !strings.Contains(result, "package logger") {
		t.Error("Generated code should contain package declaration")
	}
	if !strings.Contains(result, "UserId") {
		t.Error("Generated code should contain field type name")
	}
	if !strings.Contains(result, "WUserId") {
		t.Error("Generated code should contain wrapper function")
	}
	if !strings.Contains(result, "MaskSensitive") {
		t.Error("Generated code should contain mask function for masked field")
	}
	if !strings.Contains(result, "user_id") {
		t.Error("Generated code should contain original field name")
	}
}

func TestLogFieldsTemplateWithoutMask(t *testing.T) {
	tmpl, err := template.New("test").Funcs(LogFuncMap()).Parse(LogFieldsTemplate)
	if err != nil {
		t.Fatalf("Failed to parse LogFieldsTemplate: %v", err)
	}

	// 准备没有 mask 的测试数据
	data := struct {
		Service     string
		LogFields   []config.LogFieldConfig
		PackageName string
		Version     string
	}{
		Service: "user",
		LogFields: []config.LogFieldConfig{
			{Name: "request_id", Type: "string", Comment: "请求ID", Mask: false},
		},
		PackageName: "logger",
		Version:     "0.1.0",
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	result := buf.String()
	// 不应该包含 MaskSensitive
	if strings.Contains(result, "MaskSensitive") {
		t.Error("Generated code should not contain MaskSensitive for non-masked field")
	}
}

func TestLogFieldsTemplateEmptyFields(t *testing.T) {
	tmpl, err := template.New("test").Funcs(LogFuncMap()).Parse(LogFieldsTemplate)
	if err != nil {
		t.Fatalf("Failed to parse LogFieldsTemplate: %v", err)
	}

	// 准备空的测试数据
	data := struct {
		Service     string
		LogFields   []config.LogFieldConfig
		PackageName string
		Version     string
	}{
		Service:     "user",
		LogFields:   []config.LogFieldConfig{},
		PackageName: "logger",
		Version:     "0.1.0",
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	result := buf.String()
	if result == "" {
		t.Error("Template output should not be empty even with no fields")
	}

	// 应该包含基本的包声明和 fieldKeys 结构体
	if !strings.Contains(result, "package logger") {
		t.Error("Generated code should contain package declaration")
	}
	if !strings.Contains(result, "fieldKeys") {
		t.Error("Generated code should contain fieldKeys struct")
	}
}

func TestLogFieldsTemplateLogBuilder(t *testing.T) {
	tmpl, err := template.New("test").Funcs(LogFuncMap()).Parse(LogFieldsTemplate)
	if err != nil {
		t.Fatalf("Failed to parse LogFieldsTemplate: %v", err)
	}

	data := struct {
		Service     string
		LogFields   []config.LogFieldConfig
		PackageName string
		Version     string
	}{
		Service: "user",
		LogFields: []config.LogFieldConfig{
			{Name: "user_id", Type: "int64", Comment: "用户ID", Mask: false},
			{Name: "user_name", Type: "string", Comment: "用户名", Mask: false},
		},
		PackageName: "logger",
		Version:     "0.1.0",
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	result := buf.String()

	// 检查 LogBuilder 结构体
	if !strings.Contains(result, "type LogBuilder struct") {
		t.Error("Generated code should contain LogBuilder struct")
	}

	// 检查 L 函数
	if !strings.Contains(result, "func L(ctx context.Context, content string) *LogBuilder") {
		t.Error("Generated code should contain L function")
	}

	// 检查 LogBuilder 的字段方法（以 W 开头）
	if !strings.Contains(result, "func (b *LogBuilder) WUserId(v int64) *LogBuilder") {
		t.Error("Generated code should contain LogBuilder.WUserId method")
	}
	if !strings.Contains(result, "func (b *LogBuilder) WUserName(v string) *LogBuilder") {
		t.Error("Generated code should contain LogBuilder.WUserName method")
	}
}

func TestLogFieldsTemplateSortFields(t *testing.T) {
	tmpl, err := template.New("test").Funcs(LogFuncMap()).Parse(LogFieldsTemplate)
	if err != nil {
		t.Fatalf("Failed to parse LogFieldsTemplate: %v", err)
	}

	data := struct {
		Service     string
		LogFields   []config.LogFieldConfig
		PackageName string
		Version     string
	}{
		Service: "user",
		LogFields: []config.LogFieldConfig{
			{Name: "user_id", Type: "int64", Comment: "用户ID", Mask: false},
			{Name: "user_name", Type: "string", Comment: "用户名", Mask: false},
		},
		PackageName: "logger",
		Version:     "0.1.0",
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	result := buf.String()

	// 检查 SortableField 结构体
	if !strings.Contains(result, "type SortableField struct") {
		t.Error("Generated code should contain SortableField struct")
	}

	// 检查 sortFields 函数
	if !strings.Contains(result, "func sortFields(fields []SortableField) []logx.LogField") {
		t.Error("Generated code should contain sortFields function")
	}

	// 检查 fieldOrder 变量
	if !strings.Contains(result, "var fieldOrder = map[string]int") {
		t.Error("Generated code should contain fieldOrder map")
	}
}

func TestLogFieldsTemplateLogLevelMethods(t *testing.T) {
	tmpl, err := template.New("test").Funcs(LogFuncMap()).Parse(LogFieldsTemplate)
	if err != nil {
		t.Fatalf("Failed to parse LogFieldsTemplate: %v", err)
	}

	data := struct {
		Service     string
		LogFields   []config.LogFieldConfig
		PackageName string
		Version     string
	}{
		Service: "user",
		LogFields: []config.LogFieldConfig{
			{Name: "user_id", Type: "int64", Comment: "用户ID", Mask: false},
		},
		PackageName: "logger",
		Version:     "0.1.0",
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	result := buf.String()

	// 检查不带 s 后缀的方法（不排序）
	if !strings.Contains(result, "func (b *LogBuilder) Debug()") {
		t.Error("Generated code should contain Debug method")
	}
	if !strings.Contains(result, "func (b *LogBuilder) Info()") {
		t.Error("Generated code should contain Info method")
	}
	if !strings.Contains(result, "func (b *LogBuilder) Error()") {
		t.Error("Generated code should contain Error method")
	}
	if !strings.Contains(result, "func (b *LogBuilder) Slow()") {
		t.Error("Generated code should contain Slow method")
	}

	// 检查带 s 后缀的方法（排序）
	if !strings.Contains(result, "func (b *LogBuilder) Debugs()") {
		t.Error("Generated code should contain Debugs method")
	}
	if !strings.Contains(result, "func (b *LogBuilder) Infos()") {
		t.Error("Generated code should contain Infos method")
	}
	if !strings.Contains(result, "func (b *LogBuilder) Errors()") {
		t.Error("Generated code should contain Errors method")
	}
	if !strings.Contains(result, "func (b *LogBuilder) Slows()") {
		t.Error("Generated code should contain Slows method")
	}
}

func TestLogFieldsTemplateFieldOrder(t *testing.T) {
	tmpl, err := template.New("test").Funcs(LogFuncMap()).Parse(LogFieldsTemplate)
	if err != nil {
		t.Fatalf("Failed to parse LogFieldsTemplate: %v", err)
	}

	// 测试多个字段的顺序
	data := struct {
		Service     string
		LogFields   []config.LogFieldConfig
		PackageName string
		Version     string
	}{
		Service: "user",
		LogFields: []config.LogFieldConfig{
			{Name: "first_field", Type: "string", Comment: "第一个字段", Mask: false},
			{Name: "second_field", Type: "int64", Comment: "第二个字段", Mask: false},
			{Name: "third_field", Type: "bool", Comment: "第三个字段", Mask: false},
		},
		PackageName: "logger",
		Version:     "0.1.0",
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	result := buf.String()

	// 检查 fieldOrder 中字段的顺序是否正确（格式：fieldKeys.FirstField: 0）
	if !strings.Contains(result, "fieldKeys.FirstField: 0") {
		t.Error("Generated code should contain FirstField with order 0")
	}
	if !strings.Contains(result, "fieldKeys.SecondField: 1") {
		t.Error("Generated code should contain SecondField with order 1")
	}
	if !strings.Contains(result, "fieldKeys.ThirdField: 2") {
		t.Error("Generated code should contain ThirdField with order 2")
	}
}
