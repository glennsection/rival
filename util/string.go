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

func StringToStringArray(s string) ([]string) {
	arr := strings.FieldsFunc(s, func (r rune) bool {
		return r == '[' || r == ',' || r == ']'
	})

	return arr
}

func StringToIntArray(s string) ([]int) {
	stringArr := StringToStringArray(s)

	arr := make([]int, len(stringArr))
	for i, num := range stringArr {
		arr[i], _ = strconv.Atoi(num)
	}

	return arr
}