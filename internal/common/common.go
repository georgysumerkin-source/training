package common

import (
	"context"
	"strings"
	"sync"
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

func WithLock(ctx context.Context, locker sync.Locker, fn func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	locker.Lock()
	defer locker.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	return fn()
}

func WithLockResult[T any](ctx context.Context, locker sync.Locker, operation func() (T, error)) (T, error) {
	var result T
	err := WithLock(ctx, locker, func() error {
		var err error
		result, err = operation()
		return err
	})
	return result, err
}
