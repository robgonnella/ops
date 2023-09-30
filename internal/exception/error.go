package exception

import "errors"

// ErrRecordNotFound custom error for failure to find record
var ErrRecordNotFound = errors.New("record not found")
