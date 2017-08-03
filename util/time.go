package util

import (
	"time"
	"html/template"
)

func init() {
	AddTemplateFunc("shortTime", t_ShortTime)
}

func t_ShortTime(t time.Time) template.HTML {
	return template.HTML(t.Format("02/01/2006 03:04 PM"))
}

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

func DurationToTicks(duration time.Duration) int64 {
	// 1 tick = 100 nanoseconds
	nsec := duration.Nanoseconds()
	return nsec / 100
}

func GetCurrentDate() time.Time {
	year, month, day := time.Now().UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func GetTomorrowDate() time.Time {
	return GetCurrentDate().AddDate(0, 0, 1)
}