package exception

import "errors"

// ErrNotFound custom database error for failure to find record
var ErrRecordNotFound = errors.New("record not found")
