package ayaka

import (
	"sync"

	"errors"
)

var (
	ErrGracefulTimeout      = errors.New("graceful timeout error")
	ErrAppNotFountInContext = errors.New("app not found in context")
	ErrInvalidArgument      = errors.New("invalid argument")
	ErrNoJobs               = errors.New("no jobs registered")
	ErrJobInitFailed        = errors.New("job init failed")
	ErrJobRunFailed         = errors.New("job run failed")
	ErrJobInitPanic         = errors.New("job init panic")
	ErrJobRunPanic          = errors.New("job run panic")
)

type singleError struct {
	once sync.Once
	err  error
}

func (s *singleError) add(err error, cancel func()) {
	s.once.Do(func() {
		cancel()
		s.err = err
	})
}

func (s *singleError) get() error {
	return s.err
}

func newSingleError() *singleError {
	return &singleError{}
}
