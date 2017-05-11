package system

import (
	"os"
	"time"
	"strings"
	"fmt"
	"path/filepath"
	"html/template"
)

var (
	// default template functions
	templateFuncMap = template.FuncMap {
		"add": templateAdd,
		"shortTime": templateShortTime,
	}
)

func AddTemplateFunc(name string, f interface{}) {
	templateFuncMap[name] = f
}

func (application *Application) loadTemplates() {
	var templates []string

	// filter to gather all HTML templates
	fn := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() != true && strings.HasSuffix(f.Name(), ".html") {
			templates = append(templates, path)
		}
		return nil
	}

	// gather all HTML templates
	err := filepath.Walk("templates", fn)
	if err != nil {
		panic(err)
	}

	// preload all HTML templates
	application.templates = template.Must(template.New("").Funcs(templateFuncMap).ParseFiles(templates...))
}

func templateAdd(a, b int) template.HTML {
	return template.HTML(fmt.Sprintf("%d", a + b))
}

func templateShortTime(t time.Time) template.HTML {
	return template.HTML(t.Format("02/01/2006 03:04 PM"))
}
