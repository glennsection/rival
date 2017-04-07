package system

import (
	"time"
	"log"
)

type Profiler struct {
	Handler func(string, time.Duration)
}

var profiler *Profiler = nil

func HandleProfiling(handler func(string, time.Duration)) {
	if profiler == nil {
		// create singleton
		profiler = &Profiler{}
	}
	profiler.Handler = handler
}

func Profile(name string, startTime time.Time) {
	elapsedTime := time.Since(startTime)

	// notify handler or log or results
	if profiler == nil {
		log.Printf("Profiled function: %s took %v", name, elapsedTime)
	} else {
		profiler.Handler(name, elapsedTime)
	}
}