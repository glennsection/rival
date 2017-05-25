package util

import (
	"os"
	"strings"
	"path/filepath"
	"html/template"
)

var (
	// internal
	templateFuncMap = template.FuncMap {}
	templates *template.Template
)

func AddTemplateFunc(name string, f interface{}) {
	templateFuncMap[name] = f
}

func LoadTemplates() {
	var templatePaths []string

	// filter to gather all HTML templates
	fn := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() != true && strings.HasSuffix(f.Name(), ".html") {
			templatePaths = append(templatePaths, path)
		}
		return nil
	}

	// gather all HTML templates
	Must(filepath.Walk("templates", fn))

	// preload all HTML templates
	templates = template.Must(template.New("").Funcs(templateFuncMap).ParseFiles(templatePaths...))
}

func GetTemplates() *template.Template {
	return templates
}
