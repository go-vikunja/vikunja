package errs

import (
	"errors"
)

var (
	ResourceNotFoundError      = errors.New("caldav: resource not found")
	ResourceAlreadyExistsError = errors.New("caldav: resource already exists")
	UnauthorizedError          = errors.New("caldav: unauthorized. credentials needed.")
	ForbiddenError             = errors.New("caldav: forbidden operation.")
)
