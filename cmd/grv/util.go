package main

import (
	"fmt"
	"path/filepath"
	"time"

	rw "github.com/mattn/go-runewidth"
)

// Runnable that can be run
type Runnable func()

// Consumer that consumes values
type Consumer func(interface{})

// MaxUInt returns the largest value of the supplied arguments
func MaxUInt(x, y uint) uint {
	if x > y {
		return x
	}

	return y
}

// MinUInt returns the smallest value of the supplied arguments
func MinUInt(x, y uint) uint {
	if x < y {
		return x
	}

	return y
}

// MaxInt returns the largest value of the supplied arguments
func MaxInt(x, y int) int {
	if x > y {
		return x
	}

	return y
}

// MinInt returns the smallest value of the supplied arguments
func MinInt(x, y int) int {
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

// IsNonPrintableCharacter returns true if the provided character is a non-printable ASCII character
func IsNonPrintableCharacter(codePoint rune) bool {
	return (codePoint >= 0 && codePoint < 32) || codePoint == 127
}

// StringWidth sums the RuneWidth for each rune in the provided string
func StringWidth(str string) (width int) {
	for _, char := range str {
		width += RuneWidth(char)
	}

	return
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

// TimeWithLocation returns the provided time with the provided location set
func TimeWithLocation(oldTime time.Time, location *time.Location) time.Time {
	return time.Date(
		oldTime.Year(),
		oldTime.Month(),
		oldTime.Day(),
		oldTime.Hour(),
		oldTime.Minute(),
		oldTime.Second(),
		oldTime.Nanosecond(),
		location)
}
