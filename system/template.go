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
		return err
	}

	// preload all HTML templates
	application.templates = template.Must(template.New("").Funcs(templateFuncMap).ParseFiles(templates...))
	return nil
}

func templateAdd(a, b int) template.HTML {
    return template.HTML(fmt.Sprintf("%d", a + b))
}

func (context *Context) Pagination() template.HTML {
	if context.Params.Has("pagination") {
		pagination := context.Params.Get("pagination").(*Pagination)
		url := context.Request.URL
		urlPattern := fmt.Sprintf("%s?page=\\%d", url.Path)
		return pagination.Links(20, urlPattern)
	}
	return template.HTML("")
}