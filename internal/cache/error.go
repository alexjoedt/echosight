package cache

import "errors"

var (
	ErrAlreadyExists error = errors.New("key already exists")
	ErrNotExists     error = errors.New("key does not exist")
)
