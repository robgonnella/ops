package exception

import "errors"

// ErrRecordNotFound custom database error for failure to find record
var ErrRecordNotFound = errors.New("record not found")
