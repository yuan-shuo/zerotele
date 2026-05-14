// Package template 提供代码生成的模板定义和辅助函数
package template

import (
	"strings"
	"testing"
	"text/template"

	"github.com/yuan-shuo/zerotele/internal/config"
)

func TestHasMethod(t *testing.T) {
	tests := []struct {
		name     string
		methods  []string
		target   string
		expected bool
	}{
		{"has inc", []string{"inc", "add"}, "inc", true},
		{"has add", []string{"inc", "add"}, "add", true},
		{"not has", []string{"inc", "add"}, "set", false},
		{"empty methods", []string{}, "inc", false},
		{"nil methods", nil, "inc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasMethod(tt.methods, tt.target)
			if result != tt.expected {
				t.Errorf("HasMethod(%v, %q) = %v, want %v", tt.methods, tt.target, result, tt.expected)
			}
		})
	}
}

func TestIsEnum(t *testing.T) {
	tests := []struct {
		name     string
		vals     []string
		expected bool
	}{
		{"enum values", []string{"app", "web"}, true},
		{"single enum", []string{"app"}, true},
		{"wildcard only", []string{"*"}, false},
		{"empty vals", []string{}, false},
		{"mixed with wildcard", []string{"app", "*"}, false},
		{"nil vals", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEnum(tt.vals)
			if result != tt.expected {
				t.Errorf("IsEnum(%v) = %v, want %v", tt.vals, result, tt.expected)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		strs     []string
		sep      string
		expected string
	}{
		{[]string{"a", "b", "c"}, ",", "a,b,c"},
		{[]string{"a", "b", "c"}, " | ", "a | b | c"},
		{[]string{}, ",", ""},
		{[]string{"single"}, ",", "single"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := Join(tt.strs, tt.sep)
			if result != tt.expected {
				t.Errorf("Join(%v, %q) = %q, want %q", tt.strs, tt.sep, result, tt.expected)
			}
		})
	}
}

func TestEnumVars(t *testing.T) {
	tests := []struct {
		metricName string
		labelName  string
		vals       []string
		expected   string
	}{
		{
			"registrations_total", "source",
			[]string{"app", "web"},
			"RegistrationsTotalSourceApp, RegistrationsTotalSourceWeb",
		},
		{
			"request_duration", "method",
			[]string{"GET", "POST"},
			"RequestDurationMethodGET, RequestDurationMethodPOST",
		},
		{
			"test", "status",
			[]string{"ok"},
			"TestStatusOk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.metricName+"_"+tt.labelName, func(t *testing.T) {
			result := EnumVars(tt.metricName, tt.labelName, tt.vals)
			if result != tt.expected {
				t.Errorf("EnumVars() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestLabelParams(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		labels     []config.Label
		expected   string
	}{
		{
			"empty labels",
			"test", []config.Label{},
			"",
		},
		{
			"single enum label",
			"registrations_total",
			[]config.Label{{Name: "source", Vals: []string{"app", "web"}}},
			"source RegistrationsTotalSource",
		},
		{
			"single wildcard label",
			"registrations_total",
			[]config.Label{{Name: "error_type", Vals: []string{"*"}}},
			"error_type string",
		},
		{
			"mixed labels",
			"registrations_total",
			[]config.Label{
				{Name: "source", Vals: []string{"app", "web"}},
				{Name: "error_type", Vals: []string{"*"}},
			},
			"source RegistrationsTotalSource, error_type string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LabelParams(tt.metricName, tt.labels)
			if result != tt.expected {
				t.Errorf("LabelParams() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestLabelArg(t *testing.T) {
	tests := []struct {
		name       string
		label      config.Label
		metricName string
		expected   string
	}{
		{
			"enum label",
			config.Label{Name: "source", Vals: []string{"app", "web"}},
			"registrations_total",
			"source.registrationsTotalsourceValue()",
		},
		{
			"wildcard label",
			config.Label{Name: "error_type", Vals: []string{"*"}},
			"registrations_total",
			"error_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LabelArg(tt.label, tt.metricName)
			if result != tt.expected {
				t.Errorf("LabelArg() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFuncMap(t *testing.T) {
	fm := FuncMap()
	if fm == nil {
		t.Error("FuncMap() should not return nil")
	}

	expectedFuncs := []string{"toPascal", "camelCase", "hasMethod", "isEnum", "join", "enumVars", "labelParams", "labelArg"}
	for _, fn := range expectedFuncs {
		if _, ok := fm[fn]; !ok {
			t.Errorf("FuncMap() missing function %q", fn)
		}
	}
}

func TestMetricsTemplate(t *testing.T) {
	// 测试模板是否可以正常解析
	tmpl, err := template.New("test").Funcs(FuncMap()).Parse(MetricsTemplate)
	if err != nil {
		t.Fatalf("Failed to parse MetricsTemplate: %v", err)
	}

	// 准备测试数据
	data := struct {
		Service     string
		Metrics     []config.MetricConfig
		PackageName string
		Version     string
	}{
		Service: "user",
		Metrics: []config.MetricConfig{
			{
				Name:    "requests_total",
				Help:    "Total requests",
				Type:    "counter",
				Methods: []string{"inc", "add"},
				Labels: []config.Label{
					{Name: "method", Vals: []string{"GET", "POST"}},
					{Name: "path", Vals: []string{"*"}},
				},
			},
		},
		PackageName: "metrics",
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
	if !strings.Contains(result, "package metrics") {
		t.Error("Generated code should contain package declaration")
	}
	if !strings.Contains(result, "RequestsTotal") {
		t.Error("Generated code should contain metric type name")
	}
	if !strings.Contains(result, "requests_total") {
		t.Error("Generated code should contain original metric name")
	}
}
