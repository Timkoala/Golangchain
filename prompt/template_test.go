package prompt

import (
	"testing"
)

// TestNewTemplate 测试模板创建
func TestNewTemplate(t *testing.T) {
	content := "Hello {{.Name}}, welcome to {{.Platform}}"
	tmpl, err := NewTemplate("greeting", content)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if tmpl.GetName() != "greeting" {
		t.Errorf("Expected name to be 'greeting', got %s", tmpl.GetName())
	}
}

// TestNewTemplateWithEmptyName 测试空名称
func TestNewTemplateWithEmptyName(t *testing.T) {
	_, err := NewTemplate("", "content")
	if err == nil {
		t.Error("Expected error for empty name")
	}
}

// TestNewTemplateWithEmptyContent 测试空内容
func TestNewTemplateWithEmptyContent(t *testing.T) {
	_, err := NewTemplate("test", "")
	if err == nil {
		t.Error("Expected error for empty content")
	}
}

// TestSetVar 测试设置变量
func TestSetVar(t *testing.T) {
	tmpl, _ := NewTemplate("test", "Hello {{.Name}}")
	err := tmpl.SetVar("Name", "World")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	val, exists := tmpl.GetVar("Name")
	if !exists {
		t.Error("Expected variable to exist")
	}

	if val != "World" {
		t.Errorf("Expected value to be 'World', got %v", val)
	}
}

// TestSetVars 测试批量设置变量
func TestSetVars(t *testing.T) {
	tmpl, _ := NewTemplate("test", "Hello {{.Name}}")
	vars := map[string]interface{}{
		"Name": "Alice",
		"Age":  30,
	}

	err := tmpl.SetVars(vars)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(tmpl.GetVars()) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(tmpl.GetVars()))
	}
}

// TestRender 测试渲染模板
func TestRender(t *testing.T) {
	content := "Hello {{.Name}}, you are {{.Age}} years old"
	tmpl, _ := NewTemplate("test", content)
	tmpl.SetVar("Name", "Bob")
	tmpl.SetVar("Age", 25)

	result, err := tmpl.Render()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := "Hello Bob, you are 25 years old"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestTemplateRegistry 测试模板注册表
func TestTemplateRegistry(t *testing.T) {
	registry := NewTemplateRegistry()

	err := registry.Register("greeting", "Hello {{.Name}}")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	tmpl, err := registry.Get("greeting")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if tmpl.GetName() != "greeting" {
		t.Errorf("Expected name to be 'greeting', got %s", tmpl.GetName())
	}
}

// TestTemplateRegistryDuplicate 测试重复注册
func TestTemplateRegistryDuplicate(t *testing.T) {
	registry := NewTemplateRegistry()
	registry.Register("greeting", "Hello {{.Name}}")

	err := registry.Register("greeting", "Hi {{.Name}}")
	if err == nil {
		t.Error("Expected error for duplicate template")
	}
}

// TestTemplateRegistryList 测试列出模板
func TestTemplateRegistryList(t *testing.T) {
	registry := NewTemplateRegistry()
	registry.Register("greeting", "Hello {{.Name}}")
	registry.Register("farewell", "Goodbye {{.Name}}")

	names := registry.List()
	if len(names) != 2 {
		t.Errorf("Expected 2 templates, got %d", len(names))
	}
}

// TestTemplateRegistryDelete 测试删除模板
func TestTemplateRegistryDelete(t *testing.T) {
	registry := NewTemplateRegistry()
	registry.Register("greeting", "Hello {{.Name}}")

	err := registry.Delete("greeting")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, err = registry.Get("greeting")
	if err == nil {
		t.Error("Expected error for deleted template")
	}
}

// TestPromptBuilder 测试Prompt构建器
func TestPromptBuilder(t *testing.T) {
	builder := NewPromptBuilder()
	builder.AddText("Introduction").
		AddSection("Details", "Some details here").
		AddList([]string{"Item 1", "Item 2", "Item 3"})

	result := builder.Build()
	if result == "" {
		t.Error("Expected non-empty result")
	}

	if !contains(result, "Introduction") {
		t.Error("Expected 'Introduction' in result")
	}

	if !contains(result, "Details") {
		t.Error("Expected 'Details' in result")
	}

	if !contains(result, "Item 1") {
		t.Error("Expected 'Item 1' in result")
	}
}

// TestValidateTemplate 测试模板验证
func TestValidateTemplate(t *testing.T) {
	validContent := "Hello {{.Name}}"
	err := ValidateTemplate(validContent)
	if err != nil {
		t.Errorf("Expected no error for valid template, got %v", err)
	}

	invalidContent := ""
	err = ValidateTemplate(invalidContent)
	if err == nil {
		t.Error("Expected error for empty template")
	}
}

// TestExtractVariables 测试提取变量
func TestExtractVariables(t *testing.T) {
	content := "Hello {{.Name}}, you are {{.Age}} years old. {{.Name}} is great!"
	vars := ExtractVariables(content)

	if len(vars) != 2 {
		t.Errorf("Expected 2 unique variables, got %d", len(vars))
	}

	if !contains(vars, "Name") {
		t.Error("Expected 'Name' in variables")
	}

	if !contains(vars, "Age") {
		t.Error("Expected 'Age' in variables")
	}
}

// Helper function
func contains(slice interface{}, item interface{}) bool {
	switch s := slice.(type) {
	case []string:
		for _, v := range s {
			if v == item {
				return true
			}
		}
	case string:
		if str, ok := item.(string); ok {
			return len(s) > 0 && len(str) > 0 && (s == str || indexOfString(s, str) >= 0)
		}
	}
	return false
}

// indexOfString 查找子字符串
func indexOfString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// BenchmarkRender 基准测试渲染
func BenchmarkRender(b *testing.B) {
	tmpl, _ := NewTemplate("test", "Hello {{.Name}}, you are {{.Age}} years old")
	tmpl.SetVar("Name", "Bob")
	tmpl.SetVar("Age", 25)

	for i := 0; i < b.N; i++ {
		tmpl.Render()
	}
}
