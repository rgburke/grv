package main

import (
	"testing"
)

func TestMin(t *testing.T) {
	var minTests = []struct {
		arg1           uint
		arg2           uint
		expectedResult uint
	}{
		{
			arg1:           1,
			arg2:           2,
			expectedResult: 1,
		},
		{
			arg1:           5,
			arg2:           4,
			expectedResult: 4,
		},
		{
			arg1:           5,
			arg2:           5,
			expectedResult: 5,
		},
	}

	for _, minTest := range minTests {
		actualResult := Min(minTest.arg1, minTest.arg2)

		if actualResult != minTest.expectedResult {
			t.Errorf("Min return arg does not match expected arg. Expected: %v, Actual: %v", minTest.expectedResult, actualResult)
		}
	}
}

func TestAbs(t *testing.T) {
	var absTests = []struct {
		arg            int
		expectedResult uint
	}{
		{
			arg:            1,
			expectedResult: 1,
		},
		{
			arg:            -1,
			expectedResult: 1,
		},
		{
			arg:            0,
			expectedResult: 0,
		},
	}

	for _, absTest := range absTests {
		actualResult := Abs(absTest.arg)

		if actualResult != absTest.expectedResult {
			t.Errorf("Abs return arg does not match expected arg. Expected: %v, Actual: %v", absTest.expectedResult, actualResult)
		}
	}
}

func TestIsNonPrintableCharacter(t *testing.T) {
	var isPrintableTests = []struct {
		arg            rune
		expectedResult bool
	}{
		{
			arg:            0,
			expectedResult: true,
		},
		{
			arg:            31,
			expectedResult: true,
		},
		{
			arg:            32,
			expectedResult: false,
		},
	}

	for _, isPrintableCharTest := range isPrintableTests {
		actualResult := IsNonPrintableCharacter(isPrintableCharTest.arg)

		if actualResult != isPrintableCharTest.expectedResult {
			t.Errorf("IsNonPrintableCharacter return value does not match expected value. Expected: %v, Actual: %v", isPrintableCharTest.expectedResult, actualResult)
		}
	}
}

func TestRuneWidth(t *testing.T) {
	var runeWidthTests = []struct {
		arg            rune
		expectedResult int
	}{
		{
			arg:            0,
			expectedResult: 2,
		},
		{
			arg:            'a',
			expectedResult: 1,
		},
		{
			arg:            'ü',
			expectedResult: 1,
		},
		{
			arg:            '世',
			expectedResult: 2,
		},
	}

	for _, runeWidthCharTest := range runeWidthTests {
		actualResult := RuneWidth(runeWidthCharTest.arg)

		if actualResult != runeWidthCharTest.expectedResult {
			t.Errorf("RuneWidth return value does not match expected value. Expected: %v, Actual: %v", runeWidthCharTest.expectedResult, actualResult)
		}
	}
}

func TestNonPrintableCharString(t *testing.T) {
	var printableCharTests = []struct {
		arg            rune
		expectedResult string
	}{
		{
			arg:            0,
			expectedResult: "^@",
		},
		{
			arg:            31,
			expectedResult: "^_",
		},
		{
			arg:            32,
			expectedResult: " ",
		},
		{
			arg:            65,
			expectedResult: "A",
		},
		{
			arg:            127,
			expectedResult: "^?",
		},
	}

	for _, printableCharTest := range printableCharTests {
		actualResult := NonPrintableCharString(printableCharTest.arg)

		if actualResult != printableCharTest.expectedResult {
			t.Errorf("Abs return arg does not match expected arg. Expected: %v, Actual: %v", printableCharTest.expectedResult, actualResult)
		}
	}
}
