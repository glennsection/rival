package util

import (
	"errors"
	"fmt"
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

func PrintError(message string, err interface{}) {
	log.Errorf("%s: %v", message, err)

	// show stack (TODO - strip non-local file traces)
	if errStack, ok := err.(*errorStack); ok {
		log.Printf("[red]%v[-]", string(errStack.stack))
	} else {
		log.Printf("[red]%v[-]", string(debug.Stack()))
	}
}