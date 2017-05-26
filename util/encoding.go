package util

import (
	"strconv"
	"strings"
	"html/template"
	"encoding/json"
)

func init() {
	AddTemplateFunc("jsonEncode", t_JsonEncode)
}

func t_JsonEncode(value interface{}) template.HTML {
	var encoded string
	raw, err := json.Marshal(value)
	if err == nil {
		encoded = string(raw)
	} else {
		encoded = "INVALID"
	}

	return template.HTML(encoded)
}


func StringToIntArray(s string) (arr []int) {
	nums := strings.FieldsFunc(s, func (r rune) bool {
		return r == '[' || r == ',' || r == ']'
	})

	arr = make([]int, len(nums))
	for i, num := range nums {
		arr[i], _ = strconv.Atoi(num)
	}

	return
}