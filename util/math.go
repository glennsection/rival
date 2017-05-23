package util

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func RoundToInt(f float64) int {
	if f < -0.5 {
		return int(f - 0.5)
	}
	if f > 0.5 {
		return int(f + 0.5)
	}
	return 0
}

type Bits int

func SetMask(value Bits, mask Bits) Bits {
    return (value | mask)
}

func CheckMask(value Bits, mask Bits) bool {
	return (value & mask) == mask
}
