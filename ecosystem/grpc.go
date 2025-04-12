package ecosystem

import (
	"context"
	ayaka "github.com/OddEer0/ayaka/core"
	"net"
	"sync"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var _ ayaka.Job[string] = (*GrpcJob[string])(nil)

const (
	sliceCap = 8
)

type (
	GrpcJobBuilder[T any] struct {
		address        string
		requestTimeout time.Duration
		recoverHandle  func()
		interceptors   []grpc.UnaryServerInterceptor
		regs           []GrpcRegister[T]
		serverRegs     []GrpcServerRegister
		options        []grpc.ServerOption
	}

	GrpcJob[T any] struct {
		recoverHandle  func()
		srv            *grpc.Server
		mu             sync.Mutex
		address        string
		requestTimeout time.Duration
		interceptors   []grpc.UnaryServerInterceptor
		regs           []GrpcRegister[T]
		serverRegs     []GrpcServerRegister
		options        []grpc.ServerOption
	}

	GrpcRegister[T any] func(ctx context.Context, di T, srv *grpc.Server) error
	GrpcServerRegister  func(srv *grpc.Server) error
)

func (g *GrpcJob[T]) Address() string {
	return g.address
}

func (g *GrpcJob[T]) RequestTimeout() time.Duration {
	return g.requestTimeout
}

func (g *GrpcJob[T]) Interceptors() []grpc.UnaryServerInterceptor {
	return g.interceptors
}

func (g *GrpcJob[T]) Regs() []GrpcRegister[T] {
	return g.regs
}

func (g *GrpcJob[T]) ServerRegs() []GrpcServerRegister {
	return g.serverRegs
}

func (g *GrpcJob[T]) Options() []grpc.ServerOption {
	return g.options
}

func (g *GrpcJobBuilder[T]) Validate() error {
	return validation.ValidateStruct(g,
		validation.Field(&g.address, validation.Required),
		validation.Field(&g.requestTimeout, validation.Required),
	)
}

func NewGrpcJobBuilder[T any]() *GrpcJobBuilder[T] {
	return &GrpcJobBuilder[T]{
		recoverHandle: func() {},
		regs:          make([]GrpcRegister[T], 0, sliceCap),
		serverRegs:    make([]GrpcServerRegister, 0, sliceCap),
		interceptors:  make([]grpc.UnaryServerInterceptor, 0, sliceCap),
		options:       make([]grpc.ServerOption, 0, sliceCap),
	}
}

func (g *GrpcJob[T]) Init(ctx context.Context, di T) error {
	sliceInterceptors := make([]grpc.UnaryServerInterceptor, 0, len(g.interceptors))
	copy(sliceInterceptors, g.interceptors)

	if g.requestTimeout > 0 {
		sliceInterceptors = append(sliceInterceptors, TimeoutInterceptor(g.requestTimeout, g.recoverHandle))
	}

	grpcOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(sliceInterceptors...),
	}
	grpcOptions = append(grpcOptions, g.options...)

	srv := grpc.NewServer(grpcOptions...)

	errCh := make(chan error, 1)
	go func(errCh chan<- error) {
		for _, reg := range g.regs {
			if err := reg(ctx, di, srv); err != nil {
				errCh <- errors.Wrap(err, "[GrpcJob] grpc register error")
				return
			}
		}

		for _, serverRegister := range g.serverRegs {
			if err := serverRegister(srv); err != nil {
				errCh <- errors.Wrap(err, "[GrpcJob] grpc register error")
				return
			}
		}
		errCh <- nil
	}(errCh)

	g.mu.Lock()
	g.srv = srv
	g.mu.Unlock()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (g *GrpcJob[T]) Run(ctx context.Context, di T) error {
	errCh := make(chan error, 1)
	app, err := ayaka.AppFromContext[T](ctx)
	if err != nil {
		return errors.Wrap(err, "[GrpcJob] ayaka.AppFromContext")
	}

	go func() {
		if g.srv != nil {
			app.Logger().Info(ctx, "grpc server started...", map[string]any{"address": g.address})

			lis, err := net.Listen("tcp", g.address)
			if err != nil {
				errCh <- errors.Wrap(err, "[GrpcJob] net.Listen")
				return
			}

			err = g.srv.Serve(lis)
			if err != nil {
				errCh <- errors.Wrap(err, "[GrpcJob] srv.Serve")
				return
			}
		}

		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		app.Logger().Warn(ctx, "grpc server stopped", map[string]any{"address": g.address})
		g.srv.GracefulStop()
		return nil
	}
}

func (g *GrpcJobBuilder[T]) Address(address string) *GrpcJobBuilder[T] {
	g.address = address
	return g
}

func (g *GrpcJobBuilder[T]) RecoverHandler(recoverHandler func()) *GrpcJobBuilder[T] {
	g.recoverHandle = recoverHandler
	return g
}

func (g *GrpcJobBuilder[T]) RequestTimeout(timeout time.Duration) *GrpcJobBuilder[T] {
	g.requestTimeout = timeout
	return g
}

func (g *GrpcJobBuilder[T]) Interceptors(interceptors ...grpc.UnaryServerInterceptor) *GrpcJobBuilder[T] {
	if len(interceptors) > 0 {
		g.interceptors = append(g.interceptors, interceptors...)
	}
	return g
}

func (g *GrpcJobBuilder[T]) Register(regs ...GrpcRegister[T]) *GrpcJobBuilder[T] {
	if len(regs) > 0 {
		g.regs = append(g.regs, regs...)
	}
	return g
}

func (g *GrpcJobBuilder[T]) RegisterServer(regs ...GrpcServerRegister) *GrpcJobBuilder[T] {
	if len(regs) > 0 {
		g.serverRegs = append(g.serverRegs, regs...)
	}
	return g
}

func (g *GrpcJobBuilder[T]) RegisterOptions(options ...grpc.ServerOption) *GrpcJobBuilder[T] {
	if len(options) > 0 {
		g.options = append(g.options, options...)
	}
	return g
}

func (g *GrpcJobBuilder[T]) Build() (*GrpcJob[T], error) {
	err := g.Validate()
	if err != nil {
		return nil, errors.Wrap(err, "[GrpcJobBuilder] validate error")
	}

	return &GrpcJob[T]{
		address:        g.address,
		requestTimeout: g.requestTimeout,
		interceptors:   g.interceptors,
		regs:           g.regs,
		serverRegs:     g.serverRegs,
		options:        g.options,
		mu:             sync.Mutex{},
		recoverHandle:  g.recoverHandle,
	}, nil
}
