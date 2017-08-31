package util

import (
	"time"
	"math/rand"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomIntn(n int) int {
	return rand.Intn(n)
}

func RandomRange(min int, max int) int {
	return rand.Intn(max - min + 1) + min
}