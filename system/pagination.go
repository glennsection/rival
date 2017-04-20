package system

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
	Number      int        `json:"number"`
	Active      bool       `json:"active"`
	Link        string     `json:"link"`
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

func (pagination *Pagination) All(result interface{}) error {
	return pagination.Query.All(result)
}

func (pagination *Pagination) Links(numLinks int, urlPattern string) template.HTML {
	var pages []*Page

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

	// render template
	var out bytes.Buffer
	tmpl := template.Must(template.New("pagination").Parse(paginationTemplate))
	ctx := map[string]interface{}{
		"links": pages,
	}
	tmpl.Execute(&out, ctx)
	return template.HTML(out.String())
}

const (
	paginationTemplate string = `
{{ if .links }}
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
