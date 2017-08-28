package util

import (
	"fmt"
	"regexp"
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

func IsAlphaNumeric(s string, allowUnderscores bool) (match bool) {
	if allowUnderscores {
		match, _ = regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_]*$", s)
	} else {
		match, _ = regexp.MatchString("^[a-zA-Z][a-zA-Z0-9]*$", s)
	}
	
	return
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

func StringArrayToString(arr []string) (string) {
	str := "["

	length := len(arr)

	for i,element := range arr {
		var format string
		if i != length - 1 {
			format = "%s%s,"
		} else {
			format = "%s%s]"
		}

		str = fmt.Sprintf(format, str, element)
	}

	return str
}

func IntArrayToString(arr []int) (string) {
	str := "["

	length := len(arr)

	for i,element := range arr {
		var format string
		if i != length - 1 {
			format = "%s%d,"
		} else {
			format = "%s%d]"
		}

		str = fmt.Sprintf(format, str, element)
	}

	return str
} 