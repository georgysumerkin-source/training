package internal_errors

import "errors"

var ErrInvalidTask = errors.New("invalid task params")
var ErrInvalidTransition = errors.New("invalid status transition")
