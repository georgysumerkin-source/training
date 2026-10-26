package common

import (
	"strings"
	"unicode/utf8"

	"github.com/georgysumerkin-source/training/internal/internal_errors"
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
		return internal_errors.ErrInvalidTask
	}

	return nil
}
