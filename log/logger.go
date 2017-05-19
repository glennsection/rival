package log

import (
	golog "log"
	"fmt"
	"regexp"
)

type Color int
const (
	ColorClear Color = 0
	ColorBlack = iota + 29
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

//const escape = "\x1b"
const escape = "\033"

func RawPrint(value string) {
	prefix := golog.Prefix()
	golog.SetPrefix("")
	Print(value)
	golog.SetPrefix(prefix)
}

func Print(value string) {
	golog.Print(Sprintf(value))
}

func Println(value string) {
	golog.Println(Sprintf(value))
}

func Printf(format string, v ...interface{}) {
	golog.Println(Sprintf(format, v...))
}

func Warning(value interface{}) {
	Printf("[yellow]WARNING: %v[-]", value)
}

func Warningf(format string, v ...interface{}) {
	Printf(fmt.Sprintf("[yellow]WARNING: %s[-]", format), v...)
}

func Error(value interface{}) {
	Printf("[red!]ERROR:%v[-]", value)
}

func Errorf(format string, v ...interface{}) {
	Printf(fmt.Sprintf("[red!]ERROR: %s[-]", format), v...)
}

func Sprintf(format string, v ...interface{}) string {
	r, _ := regexp.Compile("\\[[a-z\\-!]*\\]") // TODO - allow "\\[" and "\\]"   (?)
	return fmt.Sprintf(r.ReplaceAllStringFunc(format, func(match string) string {
		switch match {
			case "[black]":
				return formatColor(ColorBlack, false)
			case "[red]":
				return formatColor(ColorRed, false)
			case "[green]":
				return formatColor(ColorGreen, false)
			case "[yellow]":
				return formatColor(ColorYellow, false)
			case "[blue]":
				return formatColor(ColorBlue, false)
			case "[magenta]":
				return formatColor(ColorMagenta, false)
			case "[cyan]":
				return formatColor(ColorCyan, false)
			case "[white]":
				return formatColor(ColorWhite, false)
			case "[black!]":
				return formatColor(ColorBlack, true)
			case "[red!]":
				return formatColor(ColorRed, true)
			case "[green!]":
				return formatColor(ColorGreen, true)
			case "[yellow!]":
				return formatColor(ColorYellow, true)
			case "[blue!]":
				return formatColor(ColorBlue, true)
			case "[magenta!]":
				return formatColor(ColorMagenta, true)
			case "[cyan!]":
				return formatColor(ColorCyan, true)
			case "[white!]":
				return formatColor(ColorWhite, true)
			case "[-]":
				return formatColor(ColorClear, false)
		}
		return match
	}), v...)
}

func formatColor(color Color, bold bool) string {
	format := "%s[%dm"
	if bold {
		format = "%s[%d;1m"
	}

	return fmt.Sprintf(format, escape, int(color))
}