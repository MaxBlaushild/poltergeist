package gameengine

import "errors"

var (
	ErrInvalidSlot = errors.New("slot must be between 0 and 5")
)
