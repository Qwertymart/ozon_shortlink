package entity

import "errors"

var (
	ErrNotFound      = errors.New("url not found")
	ErrConflict      = errors.New("url already exists")
	ErrInvalidFormat = errors.New("invalid url format")
)