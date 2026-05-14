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
