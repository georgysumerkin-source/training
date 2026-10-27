package repository

import "errors"

var ErrTaskNotFound = errors.New("domain not found")
var ErrTaskAlreadyExists = errors.New("domain already exists")
