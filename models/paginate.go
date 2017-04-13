package models

import (
	"math"

	"gopkg.in/mgo.v2"
)

func Paginate(query *mgo.Query, limit int, page int) (*mgo.Query, int, error) {
	total, err := query.Count()
	pages := int(math.Ceil(float64(total) / float64(limit)))

	if limit > 0 {
		if limit > 1000 { // to avoid memory leak
			limit = 999
		}
		query = query.Limit(limit)
	}

	if page >= 1 {
		query = query.Skip((page - 1) * limit)
	}
	
	return query, pages, err
}
