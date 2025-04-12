package ecosystem

import (
	"context"
	ayaka "github.com/OddEer0/ayaka/core"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
)

const (
	DefaultHttpIdleTimeout         = 60 * time.Second
	DefaultHttpHeaderMaxBytesLimit = 4 * 1024 * 1024
)

type (
	HttpJobBuilder[T any] struct {
		address        string
		requestTimeout time.Duration
		idleTimeout    time.Duration
		maxHeaderBytes int

		regs        []HttpRegister[T]
		middlewares []func(http.Handler) http.Handler
	}

	HttpJob[T any] struct {
		address        string
		requestTimeout time.Duration
		idleTimeout    time.Duration
		maxHeaderBytes int

		handler     *chi.Mux
		regs        []HttpRegister[T]
		middlewares []func(http.Handler) http.Handler
	}

	HttpRegister[T any] func(ctx context.Context, di T, handler *chi.Mux) (*chi.Mux, error)
)

func (h *HttpJob[T]) Address() string {
	return h.address
}

func (h *HttpJob[T]) RequestTimeout() time.Duration {
	return h.requestTimeout
}

func (h *HttpJob[T]) IdleTimeout() time.Duration {
	return h.idleTimeout
}

func (h *HttpJob[T]) MaxHeaderBytes() int {
	return h.maxHeaderBytes
}

func (h *HttpJob[T]) Handler() http.Handler {
	return h.handler
}

func (h *HttpJob[T]) Regs() []HttpRegister[T] {
	return h.regs
}

func (h *HttpJob[T]) Middlewares() []func(http.Handler) http.Handler {
	return h.middlewares
}

func (h *HttpJob[T]) Init(ctx context.Context, container T) error {
	errCh := make(chan error, 1)
	go func(errCh chan<- error) {
		var err error
		h.handler.Use(h.middlewares...)
		for _, reg := range h.regs {
			if h.handler, err = reg(ctx, container, h.handler); err != nil {
				errCh <- errors.Wrap(err, "[HttpJob] http register failed")
				return
			}
		}
		errCh <- nil
	}(errCh)

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (h *HttpJob[T]) Run(ctx context.Context, container T) error {
	errCh := make(chan error, 1)
	app, err := ayaka.AppFromContext[T](ctx)
	if err != nil {
		return errors.Wrap(err, "[HttpJob] ayaka.AppFromContext")
	}

	srv := http.Server{
		Addr:           h.address,
		Handler:        h.handler,
		WriteTimeout:   h.requestTimeout,
		ReadTimeout:    h.requestTimeout,
		IdleTimeout:    h.idleTimeout,
		MaxHeaderBytes: h.maxHeaderBytes,
	}

	go func() {
		app.Logger().Info(ctx, "http server started...", map[string]any{"address": h.address})
		err = srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		err := srv.Shutdown(ctx)
		if err != nil {
			app.Logger().Warn(ctx, "http server failed graceful stopped", map[string]any{"address": h.address})
			return errors.Wrap(err, "[HttpJob] failed to shutdown http server")
		}
		app.Logger().Warn(ctx, "http server stopped", map[string]any{"address": h.address})
		return nil
	}
}

func NewHttpJobBuilder[T any]() *HttpJobBuilder[T] {
	return &HttpJobBuilder[T]{
		regs:        make([]HttpRegister[T], 0, sliceCap),
		middlewares: make([]func(http.Handler) http.Handler, 0, sliceCap),
	}
}

func (b *HttpJobBuilder[T]) Address(address string) *HttpJobBuilder[T] {
	b.address = address
	return b
}

func (b *HttpJobBuilder[T]) RequestTimeout(requestTimeout time.Duration) *HttpJobBuilder[T] {
	b.requestTimeout = requestTimeout
	return b
}

func (b *HttpJobBuilder[T]) IdleTimeout(idleTimeout time.Duration) *HttpJobBuilder[T] {
	b.idleTimeout = idleTimeout
	return b
}

func (b *HttpJobBuilder[T]) MaxHeaderBytes(maxHeaderBytes int) *HttpJobBuilder[T] {
	b.maxHeaderBytes = maxHeaderBytes
	return b
}

func (b *HttpJobBuilder[T]) Middleware(middlewares ...func(http.Handler) http.Handler) *HttpJobBuilder[T] {
	if len(middlewares) > 0 {
		b.middlewares = append(b.middlewares, middlewares...)
	}
	return b
}

func (b *HttpJobBuilder[T]) Register(regs ...HttpRegister[T]) *HttpJobBuilder[T] {
	if len(regs) > 0 {
		b.regs = append(b.regs, regs...)
	}
	return b
}

func (b *HttpJobBuilder[T]) Validate() error {
	return validation.ValidateStruct(b,
		validation.Field(&b.address, validation.Required),
		validation.Field(&b.requestTimeout, validation.Required),
	)
}

func (b *HttpJobBuilder[T]) Build() (*HttpJob[T], error) {
	if err := b.Validate(); err != nil {
		return nil, errors.Wrap(err, "[HttpJob] validation failed")
	}
	if b.idleTimeout == 0 {
		b.idleTimeout = DefaultHttpIdleTimeout
	}
	if b.maxHeaderBytes == 0 {
		b.maxHeaderBytes = DefaultHttpHeaderMaxBytesLimit
	}

	return &HttpJob[T]{
		address:        b.address,
		requestTimeout: b.requestTimeout,
		idleTimeout:    b.idleTimeout,
		maxHeaderBytes: b.maxHeaderBytes,
		handler:        chi.NewRouter(),
		regs:           b.regs,
		middlewares:    b.middlewares,
	}, nil
}
