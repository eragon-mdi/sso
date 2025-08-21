package domain

import "errors"

var (
	ErrNotFound   = errors.New("no content found")
	ErrValidation = errors.New("bad expertion")
	ErrDuplicate  = errors.New("duplicate")
)
