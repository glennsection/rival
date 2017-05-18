package util

import (
	"os"
)

type EnvStreamSource struct {
}

var Env *Stream = &Stream {
	source: EnvStreamSource {},
}

func (source EnvStreamSource) Has(name string) bool {
	_, ok := os.LookupEnv(name)
	return ok
}

func (source EnvStreamSource) Set(name string, value interface{}) {
	if err := os.Setenv(name, value.(string)); err != nil {
		panic(err)
	}
}

func (source EnvStreamSource) Get(name string) interface{} {
	return os.Getenv(name)
}
