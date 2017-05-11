package system

import (
	"fmt"
	"strings"
	"html/template"

	"gopkg.in/mgo.v2"
)

func (application *Application) initializeSort() {
	AddTemplateFunc("sortHeader", templateSortHeader)
}

func templateSortHeader(context *Context, name string, field string) template.HTML {
	sorted := (context.Params.GetString("sort-field", "") == field)
	desc := context.Params.GetBool("sort-desc", false)

	// icon and query URL
	icon := "sort"
	url := *context.Request.URL
	query := url.Query()

	if sorted {
		if desc {
			icon = "sort-desc"
			query.Del("sort")
		} else {
			icon = "sort-asc"
			query.Set("sort", fmt.Sprintf("%s-desc", field))
		}
	} else {
		query.Set("sort", fmt.Sprintf("%s-asc", field))
	}

	// construct new URL
	url.RawQuery = query.Encode()

	return template.HTML(fmt.Sprintf("<a href=\"%s\">%s <i class=\"fa fa-%s text-info\"></i></a>", url.String(), name, icon))
}

func (context *Context) Sort(query *mgo.Query) (*mgo.Query) {
	// parse parameters
	sort := context.Params.GetString("sort", "")
	desc := false

	if sort != "" {
		// determine direction
		var sortQuery string
		if strings.HasSuffix(sort, "-desc") {
			desc = true
			sort = strings.TrimSuffix(sort, "-desc")
			sortQuery = fmt.Sprintf("-%s", sort)
		} else {
			sort = strings.TrimSuffix(sort, "-asc")
			sortQuery = sort
		}

		// update query
		query = query.Sort(sortQuery)

		// update params
		context.Params.Set("sort-field", sort)
		context.Params.Set("sort-desc", desc)
	}

	return query;
}