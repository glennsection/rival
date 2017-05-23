package util

import (
	"strconv"
	"strings"
)

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