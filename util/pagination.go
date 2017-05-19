package util

import (
	"bytes"
	"fmt"
	"math"
	"html/template"

	"gopkg.in/mgo.v2"
)

type Pagination struct {
	Query       *mgo.Query

	// internal
	total       int
	limit       int
	page        int
	pageTotal   int
}

type Page struct {
	Number      int
	Active      bool
	Link        string
}

func Paginate(query *mgo.Query, limit int, page int) (pagination *Pagination, err error) {
	pagination = new(Pagination)
	pagination.page = page
	pagination.limit = limit

	pagination.total, err = query.Count()
	if err != nil {
		return
	}

	pagination.pageTotal = int(math.Ceil(float64(pagination.total) / float64(limit)))

	if limit > 0 {
		if limit > 1000 { // to avoid memory leak
			limit = 999
		}
		query = query.Limit(limit)
	}

	if page >= 1 {
		query = query.Skip((page - 1) * limit)
	}
	
	pagination.Query = query

	return pagination, err
}

func (pagination *Pagination) GetTotal() int {
	return pagination.total
}

func (pagination *Pagination) GetLimit() int {
	return pagination.limit
}

func (pagination *Pagination) GetPage() int {
	return pagination.page
}

func (pagination *Pagination) GetPageTotal() int {
	return pagination.pageTotal
}

func (pagination *Pagination) All(result interface{}) error {
	return pagination.Query.All(result)
}

func (pagination *Pagination) Links(numLinks int, urlPattern string) template.HTML {
	var pages []*Page

	// get first/last visible page links
	if pagination.pageTotal > 1 {
		pageStart := 1
		pageEnd := 1

		if pagination.pageTotal < numLinks {
			pageStart = 1
			pageEnd = pagination.pageTotal
		} else {
			pageStart = pagination.page - int(math.Floor(float64(numLinks)/float64(2)))
			pageEnd = pagination.page + int(math.Floor(float64(numLinks)/float64(2)))

			if pageStart < 1 {
				pageEnd += int(math.Abs(float64(pageStart))) + 1
				pageStart = 1
			}

			if pageEnd > pagination.pageTotal {
				pageStart -= (pageEnd - pagination.pageTotal) - 1
				pageEnd = pagination.pageTotal
			}
		}

		for i := pageStart; i <= pageEnd; i++ { 
			page := new(Page)

			page.Number = i
			page.Link = fmt.Sprintf(urlPattern, page.Number)
			page.Active = (page.Number == pagination.page)

			pages = append(pages, page)
		}
	}

	// first visible element on page
	first := 0
	if pagination.total > 0 {
		first = ((pagination.page - 1) * pagination.limit) + 1
	}

	// last visible element on page
	last := first - 1 + pagination.limit
	if last > pagination.total {
		last = pagination.total
	}

	// render template
	var out bytes.Buffer
	tmpl := template.Must(template.New("pagination").Parse(paginationTemplate))
	ctx := map[string]interface{}{
		"first": first,
		"last": last,
		"total": pagination.total,
		"links": pages,
	}
	tmpl.Execute(&out, ctx)
	return template.HTML(out.String())
}

func (context *Context) Paginate(query *mgo.Query, limit int) (pagination *Pagination, err error) {
	// parse parameters
	page := context.Params.GetInt("page", 1)

	pagination, err = Paginate(query, limit, page)

	context.Params.Set("pagination", pagination)
	return
}

func (context *Context) GetPagination() *Pagination {
	pagination, _ := context.Params.Get("pagination").(*Pagination)
	return pagination
}

func (context *Context) RenderPagination() template.HTML {
	if context.Params.Has("pagination") {
		pagination := context.Params.Get("pagination").(*Pagination)
		url := context.Request.URL
		urlPattern := fmt.Sprintf("%s?page=%%d", url.Path)
		return pagination.Links(20, urlPattern)
	}
	return template.HTML("")
}

const (
	paginationTemplate string = `
{{ if .links }}
	<span class="pagination-summary">{{ .first }} to {{ .last }} of {{ .total }}</span>
	<ul class="pagination">
    {{ range .links }}
      	{{ if .Active }}
        	<li class="active"><a href="#">{{ .Number }}</a></li>
      	{{ else }}
        	<li><a href="{{ .Link }}">{{ .Number }}</a></li>
      	{{ end }}
    {{ end }}
  	</ul>
{{ end }}
`
)
