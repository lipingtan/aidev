package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

//go:embed templates/*
var templateFS embed.FS

// TemplateData 模板渲染数据
type TemplateData struct {
	Name       string // 插件名（原始，如 "myplugin"）
	NamePascal string // PascalCase（如 "Myplugin"）
	ModuleName string // Go module 名（如 "game-server/plugins/myplugin"）
}

// fileMapping 模板文件到输出路径的映射
type fileMapping struct {
	tmpl   string // 模板路径（相对于 templates/）
	output string // 输出路径（相对于 outputDir）
}

func main() {
	if len(os.Args) < 3 || os.Args[1] != "new-plugin" {
		fmt.Println("用法: go run scaffold/main.go new-plugin <plugin-name>")
		os.Exit(1)
	}
	pluginName := os.Args[2]

	// 生成插件项目到 ./plugins/{pluginName}/
	outputDir := filepath.Join(".", "plugins", pluginName)
	if err := generatePlugin(pluginName, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "生成失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("插件 %s 已生成到 %s\n", pluginName, outputDir)
}

// toPascalCase 将插件名转为 PascalCase（首字母大写）
func toPascalCase(name string) string {
	if len(name) == 0 {
		return name
	}
	runes := []rune(name)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// generatePlugin 生成插件项目骨架
func generatePlugin(pluginName, outputDir string) error {
	data := TemplateData{
		Name:       pluginName,
		NamePascal: toPascalCase(pluginName),
		ModuleName: "game-server/plugins/" + pluginName,
	}

	// 模板文件到输出路径的映射
	mappings := []fileMapping{
		{tmpl: "go.mod.tmpl", output: "go.mod"},
		{tmpl: "main.go.tmpl", output: "main.go"},
		{tmpl: "handler.go.tmpl", output: "handler.go"},
		{tmpl: "plugin.json.tmpl", output: "plugin.json"},
		{tmpl: "Makefile.tmpl", output: "Makefile"},
		{tmpl: "README.md.tmpl", output: "README.md"},
		{tmpl: "frontend/index.ts.tmpl", output: filepath.Join("frontend", "src", "index.ts")},
		{tmpl: "frontend/package.json.tmpl", output: filepath.Join("frontend", "package.json")},
	}

	for _, m := range mappings {
		if err := renderTemplate(data, m.tmpl, filepath.Join(outputDir, m.output)); err != nil {
			return fmt.Errorf("渲染 %s 失败: %w", m.tmpl, err)
		}
	}

	return nil
}

// renderTemplate 读取嵌入的模板文件并渲染到目标路径
func renderTemplate(data TemplateData, tmplPath, outputPath string) error {
	// 读取模板内容
	content, err := fs.ReadFile(templateFS, filepath.ToSlash(filepath.Join("templates", tmplPath)))
	if err != nil {
		return fmt.Errorf("读取模板 %s: %w", tmplPath, err)
	}

	// 解析模板
	tmpl, err := template.New(tmplPath).Parse(string(content))
	if err != nil {
		return fmt.Errorf("解析模板 %s: %w", tmplPath, err)
	}

	// 创建输出目录
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录 %s: %w", dir, err)
	}

	// 渲染到文件
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("执行模板 %s: %w", tmplPath, err)
	}

	if err := os.WriteFile(outputPath, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("写入文件 %s: %w", outputPath, err)
	}

	return nil
}
