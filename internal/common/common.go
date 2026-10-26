package common

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func ValidateString(s string, minLen, maxLen *int) error {
	curMinLen := defaultMinLen
	curMaxLen := defaultMaxLen

	if minLen != nil {
		curMinLen = *minLen
	}
	if maxLen != nil {
		curMaxLen = *maxLen
	}

	trimmedS := strings.TrimSpace(s)
	sLen := utf8.RuneCountInString(trimmedS)

	if sLen < curMinLen || sLen > curMaxLen {
		return fmt.Errorf("string '%s' doesn't fits validate rule", s)
	}

	return nil
}
