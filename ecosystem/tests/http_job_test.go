package grpc_job

import (
	"context"
	ayaka "github.com/OddEer0/ayaka/core"
	"github.com/OddEer0/ayaka/ecosystem"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func noopHttpRegister[T any](ctx context.Context, di T, handler *chi.Mux) (*chi.Mux, error) {
	return handler, nil
}

func noopMiddleware(handler http.Handler) http.Handler {
	return handler
}

func TestHttpJobBuilder(t *testing.T) {
	t.Run("Should correct build http job", func(t *testing.T) {
		address := "localhost:10101"
		requestTimeout := time.Second * 5

		job, err := ecosystem.NewHttpJobBuilder[*container]().
			Address(address).
			RequestTimeout(requestTimeout).
			Register().
			Middleware().
			Build()

		assert.NoError(t, err)
		assert.Equal(t, address, job.Address())
		assert.Equal(t, requestTimeout, job.RequestTimeout())
		assert.Equal(t, ecosystem.DefaultHttpIdleTimeout, job.IdleTimeout())
		assert.Equal(t, ecosystem.DefaultHttpHeaderMaxBytesLimit, job.MaxHeaderBytes())

		assert.Equal(t, 0, len(job.Middlewares()))
		assert.Equal(t, 0, len(job.Regs()))

		idleTimeout := time.Second * 42
		maxBytes := 69

		job, err = ecosystem.NewHttpJobBuilder[*container]().
			Address(address).
			RequestTimeout(requestTimeout).
			IdleTimeout(idleTimeout).
			MaxHeaderBytes(maxBytes).
			Register().
			Middleware().
			Build()

		assert.NoError(t, err)
		assert.Equal(t, address, job.Address())
		assert.Equal(t, requestTimeout, job.RequestTimeout())
		assert.Equal(t, idleTimeout, job.IdleTimeout())
		assert.Equal(t, maxBytes, job.MaxHeaderBytes())

		assert.Equal(t, 0, len(job.Middlewares()))
		assert.Equal(t, 0, len(job.Regs()))
	})

	t.Run("Should correct error building http without address, request-timeout", func(t *testing.T) {
		address := "localhost:10101"
		requestTimeout := time.Second * 5

		job, err := ecosystem.NewHttpJobBuilder[*container]().
			Address(address).
			Build()

		assert.Error(t, err)
		assert.Nil(t, job)

		job, err = ecosystem.NewHttpJobBuilder[*container]().
			RequestTimeout(requestTimeout).
			Build()

		assert.Error(t, err)
		assert.Nil(t, job)
	})

	t.Run("Should correct work middleware, register", func(t *testing.T) {
		address := "localhost:10101"
		requestTimeout := time.Second * 5
		job, err := ecosystem.NewHttpJobBuilder[*container]().
			Address(address).
			RequestTimeout(requestTimeout).
			Register(noopHttpRegister[*container], noopHttpRegister[*container], noopHttpRegister[*container]).
			Middleware(noopMiddleware, noopMiddleware).
			Build()

		assert.NoError(t, err)
		assert.Equal(t, address, job.Address())
		assert.Equal(t, requestTimeout, job.RequestTimeout())

		assert.Equal(t, 3, len(job.Regs()))
		assert.Equal(t, 2, len(job.Middlewares()))
	})
}

func TestHttpJobSignature(t *testing.T) {
	address := "localhost:10101"
	requestTimeout := time.Second * 5

	job, err := ecosystem.NewHttpJobBuilder[*container]().
		Address(address).
		RequestTimeout(requestTimeout).
		Register().
		Middleware().
		Build()

	assert.NoError(t, err)

	ayaka.NewApp(&ayaka.Options[*container]{
		Name:        "aya",
		Description: "kekw",
		Version:     "0.0.1",
		Container:   &container{},
	}).WithJob(ayaka.JobEntry[*container]{
		Key: "xd",
		Job: job,
	})
}
