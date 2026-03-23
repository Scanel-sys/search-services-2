package core

import "errors"

var ErrBadArguments = errors.New("arguments are not acceptable")
var ErrTooLongMessage = errors.New("too long message")
