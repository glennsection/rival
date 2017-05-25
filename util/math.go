package util

import (
	"fmt"
	"html/template"
)

func init() {
	AddTemplateFunc("add", t_Add)
	AddTemplateFunc("fmt", t_Fmt)
}

func t_Add(a, b int) template.HTML {
	return template.HTML(fmt.Sprintf("%d", a + b))
}

func t_Fmt(format string, args ...interface{}) template.HTML {
	return template.HTML(fmt.Sprintf(format, args...))
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func RoundToInt(f float64) int {
	if f < -0.5 {
		return int(f - 0.5)
	}
	if f > 0.5 {
		return int(f + 0.5)
	}
	return 0
}

type Bits int

func SetMask(value Bits, mask Bits) Bits {
    return (value | mask)
}

func CheckMask(value Bits, mask Bits) bool {
	return (value & mask) == mask
}
