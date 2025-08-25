package services

import "errors"

var (
	ErrAccessDenied = errors.New("access denied")
	ErrLabelNotFound = errors.New("label not found")
)
