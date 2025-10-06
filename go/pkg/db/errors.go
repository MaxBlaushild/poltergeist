package db

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrMaxPartySizeReached = errors.New("max party size reached")
