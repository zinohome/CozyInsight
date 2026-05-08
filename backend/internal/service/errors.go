package service

import "errors"

var (
	ErrNotOwner = errors.New("permission denied: not owner")
)
