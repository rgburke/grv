package main

import (
	"fmt"
	"path/filepath"

	rw "github.com/mattn/go-runewidth"
)

// MinUint returns the minimum value of the supplied arguments
func MinUint(x, y uint) uint {
	if x < y {
		return x
	}

	return y
}

// MaxInt returns the largest values of the supplied arguments
func MaxInt(x, y int) int {
	if x > y {
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

// IsNonPrintableCharacter returns true if the provided character is a non-printable ASCII character
func IsNonPrintableCharacter(codePoint rune) bool {
	return (codePoint >= 0 && codePoint < 32) || codePoint == 127
}

// RuneWidth is a wrapper around go-runewidth.RuneWidth and
// only differs from the original for ASCII non-printable characters
func RuneWidth(codePoint rune) int {
	if IsNonPrintableCharacter(codePoint) {
		return 2
	}

	return rw.RuneWidth(codePoint)
}

// NonPrintableCharString converts a control character into a string representation
func NonPrintableCharString(codePoint rune) string {
	if IsNonPrintableCharacter(codePoint) {
		if codePoint == 127 {
			return "^?"
		}

		return fmt.Sprintf("^%c", codePoint+64)
	}

	return string(codePoint)
}

// CanonicalPath returns the canonical version of the provided path
func CanonicalPath(path string) (canonicalPath string, err error) {
	canonicalPath, err = filepath.EvalSymlinks(path)
	if err != nil {
		return
	}

	return filepath.Abs(canonicalPath)
}
