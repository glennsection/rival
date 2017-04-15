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

func Print(line string) {
	golog.Print(Sprintf(line))
}

func Println(line string) {
	golog.Println(Sprintf(line))
}

func Printf(format string, v ...interface{}) {
	golog.Printf(Sprintf(format, v...))
}

func Warning(line string) {
	golog.Println(Sprintf("[yellow]WARNING: " + line + "[-]"))
}

func Warningf(format string, v ...interface{}) {
	golog.Printf(Sprintf("[yellow]WARNING: " + format + "[-]", v...))
}

func Error(line string) {
	golog.Println(Sprintf("[red!]ERROR: " + line + "[-]"))
}

func Errorf(format string, v ...interface{}) {
	golog.Printf(Sprintf("[red!]ERROR: " + format + "[-]", v...))
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