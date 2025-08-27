package services

import (
	"code.vikunja.io/api/pkg/models"
	"errors"
)

var (
	ErrAccessDenied  = models.ErrGenericForbidden{}
	ErrLabelNotFound = errors.New("label not found")
)
