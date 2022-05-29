package core

import (
	"fmt"
	"strconv"
)

// TODO: Make this more efficient and improve unit-test coverage
const (
	nineDigits        = 1000000000.0
	digitsOfPrecision = 18
)

func canonicalize(s []byte) (string, error) {
	var err error
	if len(s) > digitsOfPrecision {
		return "", fmt.Errorf("number has %d digits, exceeds max of %d", len(s), digitsOfPrecision)
	}
	var f float64
	f, err = strconv.ParseFloat(string(s), 64)
	if err != nil {
		return "", err
	}
	if f >= nineDigits || f <= -nineDigits {
		return "", fmt.Errorf("number is outside of range [%f, %f]", -nineDigits, nineDigits)
	}
	return fmt.Sprintf("%019.0f", (f+nineDigits)*nineDigits), nil
}
