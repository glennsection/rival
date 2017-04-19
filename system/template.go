package system

import (
	"os"
	"strings"
	"fmt"
	"path/filepath"
	"html/template"
)

var (
	templateFuncMap = template.FuncMap {
		"add": templateAdd,
	}
)

func (application *Application) LoadTemplates() error {
	var templates []string

	// gather all HTML templates
	fn := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() != true && strings.HasSuffix(f.Name(), ".html") {
			templates = append(templates, path)
		}
		return nil
	}

	err := filepath.Walk("templates", fn)
	if err != nil {
		return err
	}

	// preload all HTML templates
	application.templates = template.Must(template.New("").Funcs(templateFuncMap).ParseFiles(templates...))
	return nil
}

func templateAdd(a, b int) string {
    return fmt.Sprintf("%d", a + b)
}
