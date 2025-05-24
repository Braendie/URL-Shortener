package storage

import "errors"

var (
	ErrURLNOtFound = errors.New("url not found")
	ErrURLExists   = errors.New("url exists")
)
