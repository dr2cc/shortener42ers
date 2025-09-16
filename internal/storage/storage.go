package storage

import (
	"errors"
)

// Использовались в ошибках sqlite
// Может пригодиться и для postgresql
var (
	// в redirect.NewDb ?
	ErrURLNotFound = errors.New("url not found")
	//
	ErrURLExists = errors.New("url exists")
)
