package main

import (
	"fmt"
)

func Min(x, y uint) uint {
	if x < y {
		return x
	}

	return y
}

func nonPrintableCharString(codePoint rune) string {
	switch {
	case codePoint < 32:
		return fmt.Sprintf("^%c", codePoint+64)
	case codePoint == 127:
		return "^?"
	default:
		return string(codePoint)
	}
}
