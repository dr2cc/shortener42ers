package storage

import (
	"errors"
)

// Использовались в ошибках sqlite
// Может пригодиться и для postgresql
var (
	// в redirect.NewDB ?
	ErrURLNotFound = errors.New("url not found")
	//
	ErrURLExists = errors.New("url exists")
)
