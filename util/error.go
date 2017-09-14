package util

import (
	"errors"
	"fmt"
	"strings"
	"path/filepath"
	"runtime/debug"

	"bloodtales/log"
)

type errorStack struct {
	err    error
	stack  []byte
}

func NewError(value interface{}) error {
	var err error
	if valueError, ok := value.(error); ok {
		err = valueError
	} else {
		err = errors.New(fmt.Sprintf("%v", value))
	}

	return &errorStack {
		err: err,
		stack: debug.Stack(),
	}
}

func (e *errorStack) Error() string {
	return e.err.Error()
}

func LogError(message string, err interface{}) {
	log.Errorf("%s: %v", message, err)

	// get stack
	var stack string = ""
	if errStack, ok := err.(*errorStack); ok {
		stack = string(errStack.stack)
	}

	printStack(stack)
}

func PrintStack() {
	printStack(string(debug.Stack()))
}

func printStack(stack string) {
	// get root dir
	rootPath, _ := filepath.Abs(".")
	rootPath = strings.ToLower(strings.Replace(rootPath, "\\", "/", -1))
	lenRootPath := len(rootPath) + 1

	// process stack lines
	stacks := strings.Split(stack, "\n")
	stack = "[red]...[-]"
	call := ""
	parity := 1
	for _, line := range stacks {
		line = strings.TrimSpace(line)
		if parity == 1 {
			path := strings.ToLower(line)
			if strings.Contains(path, rootPath) && !strings.Contains(path, "error.go") {
				// project path found, add to result
				pidx := strings.LastIndex(path, "+")
				if pidx >= 0 {
					path = path[lenRootPath:pidx - 1]
				} else {
					path = path[lenRootPath:]
				}

				stack += fmt.Sprintf("\n[red]%s[-]  [red!](%s)[-]", call, path)
			}
		} else {
			// remove call arguments
			pidx := strings.LastIndex(line, "(")
			if pidx >= 0 {
				call = line[:pidx]
			} else {
				call = line
			}
		}
		parity = 1 - parity
	}
	stack += "\n[red]...[-]"

	// show stack
	log.RawPrint(stack)
}