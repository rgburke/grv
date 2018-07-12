package main

import (
	"testing"
	"time"
)

func TestMinUInt(t *testing.T) {
	var tests = []struct {
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

	for _, test := range tests {
		actualResult := MinUInt(test.arg1, test.arg2)

		if actualResult != test.expectedResult {
			t.Errorf("MinUInt return arg does not match expected value. Expected: %v, Actual: %v", test.expectedResult, actualResult)
		}
	}
}

func TestMaxUInt(t *testing.T) {
	var tests = []struct {
		arg1           uint
		arg2           uint
		expectedResult uint
	}{
		{
			arg1:           1,
			arg2:           2,
			expectedResult: 2,
		},
		{
			arg1:           5,
			arg2:           4,
			expectedResult: 5,
		},
		{
			arg1:           5,
			arg2:           5,
			expectedResult: 5,
		},
	}

	for _, test := range tests {
		actualResult := MaxUInt(test.arg1, test.arg2)

		if actualResult != test.expectedResult {
			t.Errorf("MaxUInt return arg does not match expected value. Expected: %v, Actual: %v", test.expectedResult, actualResult)
		}
	}
}

func TestMinInt(t *testing.T) {
	var tests = []struct {
		arg1           int
		arg2           int
		expectedResult int
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
		{
			arg1:           -1,
			arg2:           -2,
			expectedResult: -2,
		},
	}

	for _, test := range tests {
		actualResult := MinInt(test.arg1, test.arg2)

		if actualResult != test.expectedResult {
			t.Errorf("MinInt return arg does not match expected value. Expected: %v, Actual: %v", test.expectedResult, actualResult)
		}
	}
}

func TestMaxInt(t *testing.T) {
	var tests = []struct {
		arg1           int
		arg2           int
		expectedResult int
	}{
		{
			arg1:           1,
			arg2:           2,
			expectedResult: 2,
		},
		{
			arg1:           5,
			arg2:           4,
			expectedResult: 5,
		},
		{
			arg1:           5,
			arg2:           5,
			expectedResult: 5,
		},
		{
			arg1:           -1,
			arg2:           -2,
			expectedResult: -1,
		},
	}

	for _, test := range tests {
		actualResult := MaxInt(test.arg1, test.arg2)

		if actualResult != test.expectedResult {
			t.Errorf("MaxInt return arg does not match expected value. Expected: %v, Actual: %v", test.expectedResult, actualResult)
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
			t.Errorf("Abs return arg does not match expected value. Expected: %v, Actual: %v", absTest.expectedResult, actualResult)
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
			t.Errorf("Abs return arg does not match expected value. Expected: %v, Actual: %v", printableCharTest.expectedResult, actualResult)
		}
	}
}

func TestTimeWithLocation(t *testing.T) {
	var tests = []struct {
		arg1           time.Time
		arg2           *time.Location
		expectedResult time.Time
	}{
		{
			arg1:           time.Date(2018, 07, 12, 20, 25, 45, 0, time.UTC),
			arg2:           time.FixedZone("TestZone", 3600),
			expectedResult: time.Date(2018, 07, 12, 20, 25, 45, 0, time.FixedZone("TestZone", 3600)),
		},
		{
			arg1:           time.Date(2018, 07, 12, 20, 25, 45, 0, time.UTC),
			arg2:           time.UTC,
			expectedResult: time.Date(2018, 07, 12, 20, 25, 45, 0, time.UTC),
		},
	}

	for _, test := range tests {
		actualResult := TimeWithLocation(test.arg1, test.arg2)

		if !actualResult.Equal(test.expectedResult) {
			t.Errorf("TimeWithLocation return arg does not match expected value. Expected: %v, Actual: %v", test.expectedResult, actualResult)
		}
	}
}
