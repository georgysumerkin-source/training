package main

import (
	"bytes"
	"errors"
	"log"
	"net"
	"strings"
	"testing"
)

type listenerStub struct {
	net.Listener
	closeError error
	closeCalls int
}

func (l *listenerStub) Close() error { l.closeCalls++; return l.closeError }

func TestListenerClose(t *testing.T) {
	for _, closeErr := range []error{nil, errors.New("close failed")} {
		t.Run(map[bool]string{true: "error", false: "success"}[closeErr != nil], func(t *testing.T) {
			listener := &listenerStub{closeError: closeErr}
			var output bytes.Buffer
			listenerClose(listener, log.New(&output, "", 0))
			if listener.closeCalls != 1 {
				t.Fatalf("Close called %d times", listener.closeCalls)
			}
			if !strings.Contains(output.String(), "Listener closed") {
				t.Fatalf("missing close log: %q", output.String())
			}
			if closeErr != nil && !strings.Contains(output.String(), "Error closing listener: close failed") {
				t.Fatalf("missing error log: %q", output.String())
			}
			if closeErr == nil && strings.Contains(output.String(), "Error") {
				t.Fatalf("unexpected error log: %q", output.String())
			}
		})
	}
}
