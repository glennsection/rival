package util

import (
	"strconv"
	"strings"
	"encoding/json"
)

func init() {
	AddTemplateFunc("jsonEncode", t_JsonEncode)
	AddTemplateFunc("jsonEncodeIndent", t_JsonEncodeIndent)
	AddTemplateFunc("truncate", t_Truncate)
}

func t_JsonEncode(value interface{}) string {
	var encoded string
	raw, err := json.Marshal(value)
	if err == nil {
		encoded = string(raw)
	} else {
		encoded = "INVALID"
	}

	return encoded
}

func t_JsonEncodeIndent(value interface{}) string {
	var encoded string
	raw, err := json.MarshalIndent(value, "", "\t")
	if err == nil {
		encoded = string(raw)
	} else {
		encoded = "INVALID"
	}

	return encoded
}

func t_Truncate(value string, maxLength int) string {
	if len(value) > maxLength {
		value = value[:maxLength] + "..."
	}
	return value
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