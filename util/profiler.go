package util

import (
	"time"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	_ "net/http/pprof"
)

type Profiler struct {
	Handler func(string, time.Duration)
}

var profiler *Profiler = nil

func HandleProfiling(handler func(string, time.Duration)) {
	if profiler == nil {
		// create singleton
		profiler = &Profiler {}
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

func StartCPUProfile() {
	path := Env.GetString("PROFILE_CPU_PATH", "cpu.prf")

	f, err := os.Create(path)
	Must(err)

	err = pprof.StartCPUProfile(f)
	Must(err)
}

func StopCPUProfile() {
	pprof.StopCPUProfile()
}

func WriteHeapProfile() {
	path := Env.GetString("PROFILE_MEM_PATH", "mem.prf")

	f, err := os.Create(path)
	Must(err)
	defer f.Close()

	runtime.GC()
	err = pprof.WriteHeapProfile(f)
	Must(err)
}