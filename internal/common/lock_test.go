package common_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/georgysumerkin-source/training/internal/common"
)

func TestWithLockCallback(t *testing.T) {
	failure := errors.New("operation failed")
	for _, tt := range []struct {
		name string
		err  error
	}{{"success", nil}, {"error", failure}} {
		t.Run(tt.name, func(t *testing.T) {
			var mu sync.Mutex
			calls := 0
			err := common.WithLock(context.Background(), &mu, func() error {
				calls++
				if mu.TryLock() {
					mu.Unlock()
					t.Fatal("callback ran without lock")
				}
				return tt.err
			})
			if calls != 1 || !errors.Is(err, tt.err) {
				t.Fatalf("calls = %d, error = %v; want 1, %v", calls, err, tt.err)
			}
			if !mu.TryLock() {
				t.Fatal("lock not released")
			}
			mu.Unlock()
		})
	}
}

func TestWithLockPanicReleasesLock(t *testing.T) {
	var mu sync.Mutex
	func() {
		defer func() {
			if got := recover(); got != "failure" {
				t.Errorf("panic = %v", got)
			}
		}()
		_ = common.WithLock(context.Background(), &mu, func() error { panic("failure") })
	}()
	if !mu.TryLock() {
		t.Fatal("lock not released after panic")
	}
	mu.Unlock()
}

type cancelingLocker struct {
	sync.Mutex
	cancel context.CancelFunc
}

func (l *cancelingLocker) Lock() { l.Mutex.Lock(); l.cancel() }

func TestWithLockCanceledContext(t *testing.T) {
	t.Run("before lock", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		var mu sync.Mutex
		mu.Lock()
		defer mu.Unlock()
		// An already canceled call must return without trying to acquire the held lock.
		err := common.WithLock(ctx, &mu, func() error { t.Fatal("operation called"); return nil })
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error: %v", err)
		}
	})
	t.Run("during acquisition", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		mu := &cancelingLocker{cancel: cancel}
		err := common.WithLock(ctx, mu, func() error { t.Fatal("operation called"); return nil })
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error: %v", err)
		}
		if !mu.TryLock() {
			t.Fatal("lock not released")
		}
		mu.Unlock()
	})
}

func TestWithLockResult(t *testing.T) {
	var mu sync.RWMutex
	failure := errors.New("read failed")
	for _, wantErr := range []error{nil, failure} {
		got, err := common.WithLockResult(context.Background(), mu.RLocker(), func() (int, error) {
			if mu.TryLock() {
				mu.Unlock()
				t.Fatal("read lock not held")
			}
			if !mu.TryRLock() {
				t.Fatal("read lock should allow other readers")
			}
			mu.RUnlock()
			return 42, wantErr
		})
		if got != 42 || !errors.Is(err, wantErr) {
			t.Fatalf("result: %d, %v", got, err)
		}
		if !mu.TryLock() {
			t.Fatal("read lock not released")
		}
		mu.Unlock()
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, err := common.WithLockResult(ctx, &mu, func() (int, error) { t.Fatal("operation called"); return 42, nil })
	if got != 0 || !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled result: %d, %v", got, err)
	}
}
