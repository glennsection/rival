package data

import (
	"time"
)

func TicksToTime(ticks int64) time.Time {
	// 1 tick = 100 nanoseconds
	nticks := ticks % 10000000
	sec := (ticks - nticks) / 10000000
	return time.Unix(sec, nticks * 100)
}

func TimeToTicks(time time.Time) int64 {
	// 1 tick = 100 nanoseconds
	nsec := time.UnixNano()
	return nsec / 100
}