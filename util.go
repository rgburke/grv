package main

import (
	"fmt"
)

// Min returns the minimum value of the supplied arguments
func Min(x, y uint) uint {
	if x < y {
		return x
	}

	return y
}

// Abs returns the absolute value of an int as a uint
func Abs(x int) uint {
	if x < 0 {
		x = -x
	}

	return uint(x)
}

// NonPrintableCharString converts a control character into a string representation
func NonPrintableCharString(codePoint rune) string {
	switch {
	case codePoint < 32:
		return fmt.Sprintf("^%c", codePoint+64)
	case codePoint == 127:
		return "^?"
	default:
		return string(codePoint)
	}
}
