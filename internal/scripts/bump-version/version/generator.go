package version

import (
	"html/template"
	"os"
	"path/filepath"
)

// TemplateGenerator implements the Generator interface using templates
type TemplateGenerator struct {
	outFile      string
	outDir       string
	templatePath string
	templateName string
}

// NewTemplateGenerator returns a new instance of TemplateGenerator
func NewTemplateGenerator(outFile, templatePath string) *TemplateGenerator {
	return &TemplateGenerator{
		outFile:      outFile,
		outDir:       filepath.Dir(outFile),
		templatePath: templatePath,
		templateName: filepath.Base(templatePath),
	}
}

// Generate implements the Generate method using templates
func (t *TemplateGenerator) Generate(data VersionData) error {
	if err := os.MkdirAll(t.outDir, 0751); err != nil {
		return err
	}

	file, err := os.Create(t.outFile)

	if err != nil {
		return err
	}

	defer file.Close()

	template, err := template.New(t.templateName).ParseFiles(t.templatePath)

	if err != nil {
		return err
	}

	return template.Execute(file, data)
}
