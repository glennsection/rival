package util

import (
	"os"
	"time"
	"strings"
	"fmt"
	"path/filepath"
	"html/template"
)

var (
	// internal
	templateFuncMap = template.FuncMap {
		// defaults
		"add": templateAdd,
		"shortTime": templateShortTime,
	}
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

func templateAdd(a, b int) template.HTML {
	return template.HTML(fmt.Sprintf("%d", a + b))
}

func templateShortTime(t time.Time) template.HTML {
	return template.HTML(t.Format("02/01/2006 03:04 PM"))
}
