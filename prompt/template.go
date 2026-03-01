// Package prompt 提供Prompt模板系统
package prompt

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

// Template 定义Prompt模板
type Template struct {
	name     string
	template *template.Template
	vars     map[string]interface{}
}

// TemplateConfig 定义模板配置
type TemplateConfig struct {
	Name    string
	Content string
	Vars    map[string]interface{}
}

// NewTemplate 创建新的模板
func NewTemplate(name, content string) (*Template, error) {
	if name == "" {
		return nil, errors.New("template name cannot be empty")
	}
	if content == "" {
		return nil, errors.New("template content cannot be empty")
	}

	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &Template{
		name:     name,
		template: tmpl,
		vars:     make(map[string]interface{}),
	}, nil
}

// SetVar 设置模板变量
func (t *Template) SetVar(key string, value interface{}) error {
	if key == "" {
		return errors.New("variable key cannot be empty")
	}
	t.vars[key] = value
	return nil
}

// SetVars 批量设置模板变量
func (t *Template) SetVars(vars map[string]interface{}) error {
	if vars == nil {
		return errors.New("variables map cannot be nil")
	}
	for key, value := range vars {
		if err := t.SetVar(key, value); err != nil {
			return err
		}
	}
	return nil
}

// Render 渲染模板
func (t *Template) Render() (string, error) {
	var buf bytes.Buffer
	if err := t.template.Execute(&buf, t.vars); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}
	return buf.String(), nil
}

// GetVar 获取模板变量
func (t *Template) GetVar(key string) (interface{}, bool) {
	val, exists := t.vars[key]
	return val, exists
}

// GetVars 获取所有模板变量
func (t *Template) GetVars() map[string]interface{} {
	return t.vars
}

// GetName 获取模板名称
func (t *Template) GetName() string {
	return t.name
}

// TemplateRegistry 定义模板注册表
type TemplateRegistry struct {
	templates map[string]*Template
}

// NewTemplateRegistry 创建新的模板注册表
func NewTemplateRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		templates: make(map[string]*Template),
	}
}

// Register 注册模板
func (r *TemplateRegistry) Register(name, content string) error {
	if name == "" {
		return errors.New("template name cannot be empty")
	}
	if _, exists := r.templates[name]; exists {
		return fmt.Errorf("template %s already exists", name)
	}

	tmpl, err := NewTemplate(name, content)
	if err != nil {
		return err
	}

	r.templates[name] = tmpl
	return nil
}

// Get 获取模板
func (r *TemplateRegistry) Get(name string) (*Template, error) {
	tmpl, exists := r.templates[name]
	if !exists {
		return nil, fmt.Errorf("template %s not found", name)
	}
	return tmpl, nil
}

// List 列出所有模板名称
func (r *TemplateRegistry) List() []string {
	names := make([]string, 0, len(r.templates))
	for name := range r.templates {
		names = append(names, name)
	}
	return names
}

// Delete 删除模板
func (r *TemplateRegistry) Delete(name string) error {
	if _, exists := r.templates[name]; !exists {
		return fmt.Errorf("template %s not found", name)
	}
	delete(r.templates, name)
	return nil
}

// PromptBuilder 定义Prompt构建器
type PromptBuilder struct {
	parts []string
}

// NewPromptBuilder 创建新的Prompt构建器
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		parts: make([]string, 0),
	}
}

// AddText 添加文本
func (pb *PromptBuilder) AddText(text string) *PromptBuilder {
	if text != "" {
		pb.parts = append(pb.parts, text)
	}
	return pb
}

// AddSection 添加分段
func (pb *PromptBuilder) AddSection(title, content string) *PromptBuilder {
	if title != "" && content != "" {
		pb.parts = append(pb.parts, fmt.Sprintf("## %s\n%s", title, content))
	}
	return pb
}

// AddList 添加列表
func (pb *PromptBuilder) AddList(items []string) *PromptBuilder {
	if len(items) > 0 {
		for _, item := range items {
			pb.parts = append(pb.parts, fmt.Sprintf("- %s", item))
		}
	}
	return pb
}

// Build 构建Prompt
func (pb *PromptBuilder) Build() string {
	return strings.Join(pb.parts, "\n")
}

// ValidateTemplate 验证模板内容
func ValidateTemplate(content string) error {
	if content == "" {
		return errors.New("template content cannot be empty")
	}

	// 检查是否有有效的模板变量
	varPattern := regexp.MustCompile(`\{\{\.(\w+)\}\}`)
	if !varPattern.MatchString(content) && !strings.Contains(content, "{{") {
		// 如果没有模板变量，至少应该有一些内容
		if len(strings.TrimSpace(content)) == 0 {
			return errors.New("template content is empty after trimming")
		}
	}

	// 尝试解析模板
	_, err := template.New("validate").Parse(content)
	return err
}

// ExtractVariables 提取模板中的变量
func ExtractVariables(content string) []string {
	varPattern := regexp.MustCompile(`\{\{\.(\w+)\}\}`)
	matches := varPattern.FindAllStringSubmatch(content, -1)

	vars := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			vars = append(vars, match[1])
			seen[match[1]] = true
		}
	}

	return vars
}
