package ayaka

import (
	"github.com/pkg/errors"
	"sync"
)

var (
	ErrGracefulTimeout         = errors.New("graceful timeout error")
	ErrAppNotFountInContext    = errors.New("app not found in context")
	ErrIncorrectValueInContext = errors.New("incorrect value in context")
)

type singleError struct {
	mu       sync.Mutex
	err      error
	callback func()
}

func (s *singleError) add(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err == nil {
		if s.callback != nil {
			s.callback()
		}
		s.err = err
	}
}

func (s *singleError) get() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

func newSingleError(callback func()) *singleError {
	return &singleError{
		err:      nil,
		mu:       sync.Mutex{},
		callback: callback,
	}
}
