package common_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/georgysumerkin-source/training/internal/common"
	"github.com/georgysumerkin-source/training/internal/internal_errors"
)

func TestValidateString(t *testing.T) {
	min, max := 2, 4
	for _, tt := range []struct {
		name, value string
		min, max    *int
		invalid     bool
	}{
		{"default empty", "", nil, nil, false},
		{"default maximum", strings.Repeat("a", 64), nil, nil, false},
		{"default too long", strings.Repeat("a", 65), nil, nil, true},
		{"minimum", "ab", &min, &max, false},
		{"maximum", "abcd", &min, &max, false},
		{"too short", "a", &min, &max, true},
		{"too long", "abcde", &min, &max, true},
		{"trim whitespace", " \t ab \n", &min, &max, false},
		{"only whitespace", " \t\n", &min, &max, true},
		{"unicode counts runes", "ёж🙂界", &min, &max, false},
		{"only minimum", "a", &min, nil, true},
		{"only maximum", "", nil, &max, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := common.ValidateString(tt.value, tt.min, tt.max)
			if tt.invalid {
				if !errors.Is(err, internal_errors.ErrInvalidTask) {
					t.Fatalf("error = %v, want ErrInvalidTask", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}
