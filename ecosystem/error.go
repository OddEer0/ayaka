package ecosystem

import "github.com/pkg/errors"

var (
	ErrTracerNotFoundInContext = errors.New("tracer not found in context")
	ErrTypeAssertionFailed     = errors.New("type assertion failed")
)
