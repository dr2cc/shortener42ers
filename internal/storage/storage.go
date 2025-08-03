package storage

import (
	"errors"
)

// Используются в sqlite.go
var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url exists")
)
