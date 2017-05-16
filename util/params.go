package util

import (
	"sync"
	"net/http"
)


type ParamsStreamSource struct {
	bindings        map[string]interface{}
	mutex           sync.RWMutex
	request         *http.Request
}

func NewParamsStream(request *http.Request) *Stream {
	return &Stream {
		source: ParamsStreamSource {
			bindings: map[string]interface{} {},
			request: request,
		},
	}
}

func (source ParamsStreamSource) Has(name string) bool {
	// check bindings
	source.mutex.RLock()
	defer source.mutex.RUnlock()
	_, ok := source.bindings[name]
	return ok
}

func (source ParamsStreamSource) Set(name string, value interface{}) {
	// set bindings
	source.mutex.Lock()
	defer source.mutex.Unlock()
	source.bindings[name] = value
}

func (source ParamsStreamSource) Get(name string) interface{} {
	// first check bindings
	source.mutex.RLock()
	defer source.mutex.RUnlock()
	if val, ok := source.bindings[name]; ok {
		return val
	}

	// then use request params
	return source.request.FormValue(name)
}
