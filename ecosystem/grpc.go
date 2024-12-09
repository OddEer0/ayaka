package ecosystem

import (
	"context"
	"net"
	"sync"
	"time"

	ayaka "github.com/OddEer0/ayaka/core"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpcvalidator "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator"
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"go.uber.org/dig"
	"google.golang.org/grpc"
)

type (
	GrpcJobBuilder struct {
		address        string
		requestTimeout time.Duration
		maxRetry       int
		interceptors   []grpc.UnaryServerInterceptor
		regs           []GrpcRegister
		serverRegs     []GrpcServerRegister
		tracer         opentracing.Tracer
	}

	GrpcJob struct {
		srv            *grpc.Server
		mu             sync.Mutex
		address        string
		requestTimeout time.Duration
		maxRetry       int
		interceptors   []grpc.UnaryServerInterceptor
		regs           []GrpcRegister
		serverRegs     []GrpcServerRegister
		tracer         opentracing.Tracer
	}

	GrpcRegister       func(ctx context.Context, di *dig.Container, srv *grpc.Server) error
	GrpcServerRegister func(srv *grpc.Server) error
)

func (g *GrpcJobBuilder) Validate() error {
	return validation.ValidateStruct(g,
		validation.Field(&g.address, validation.Required),
		validation.Field(&g.requestTimeout, validation.Required),
		validation.Field(&g.maxRetry, validation.Required),
		validation.Field(&g.tracer, validation.Required),
	)
}

func NewGrpcJobBuilder() *GrpcJobBuilder {
	return &GrpcJobBuilder{
		regs:         make([]GrpcRegister, 0, 8),
		interceptors: make([]grpc.UnaryServerInterceptor, 0, 8),
	}
}

func (g *GrpcJob) Init(ctx context.Context, di *dig.Container) error {
	sliceInterceptors := append(g.interceptors,
		grpcprometheus.UnaryServerInterceptor,
		otgrpc.OpenTracingServerInterceptor(g.tracer),
		grpcvalidator.UnaryServerInterceptor(),
	)

	if g.requestTimeout != 0 {
		sliceInterceptors = append(sliceInterceptors, TimeoutInterceptor(g.requestTimeout))
	}

	grpcOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(sliceInterceptors...),
		grpc.ChainStreamInterceptor(
			otgrpc.OpenTracingStreamServerInterceptor(g.tracer),
			recovery.StreamServerInterceptor(),
			grpcprometheus.StreamServerInterceptor,
			grpcvalidator.StreamServerInterceptor(),
		),
	}

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

func (g *GrpcJob) Run(ctx context.Context, di *dig.Container) error {
	errCh := make(chan error, 1)
	var log ayaka.Logger
	err := di.Invoke(func(logger ayaka.Logger) {
		log = logger
	})
	if err != nil {
		return errors.Wrap(err, "[GrpcJob] di.Invoke")
	}

	go func() {
		if g.srv != nil {
			log.Info(ctx, "grpc server started...", map[string]any{"address": g.address})

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
		log.Warn(ctx, "grpc server stopped", map[string]any{"address": g.address})
		g.srv.GracefulStop()
		return nil
	}
}

func (g *GrpcJobBuilder) Address(address string) *GrpcJobBuilder {
	g.address = address
	return g
}

func (g *GrpcJobBuilder) RequestTimeout(timeout time.Duration) *GrpcJobBuilder {
	g.requestTimeout = timeout
	return g
}

func (g *GrpcJobBuilder) MaxRetry(max int) *GrpcJobBuilder {
	g.maxRetry = max
	return g
}

func (g *GrpcJobBuilder) Interceptors(interceptors ...grpc.UnaryServerInterceptor) *GrpcJobBuilder {
	g.interceptors = interceptors
	g.interceptors = append(g.interceptors, interceptors...)
	return g
}

func (g *GrpcJobBuilder) Tracer(tracer opentracing.Tracer) *GrpcJobBuilder {
	g.tracer = tracer
	return g
}

func (g *GrpcJobBuilder) Register(regs ...GrpcRegister) *GrpcJobBuilder {
	for _, reg := range regs {
		g.regs = append(g.regs, reg)
	}
	return g
}

func (g *GrpcJobBuilder) RegisterServer(regs ...GrpcServerRegister) *GrpcJobBuilder {
	for _, reg := range regs {
		g.serverRegs = append(g.serverRegs, reg)
	}
	return g
}

func (g *GrpcJobBuilder) Build() (*GrpcJob, error) {
	err := g.Validate()
	if err != nil {
		return nil, errors.Wrap(err, "[GrpcJobBuilder] validate error")
	}

	return &GrpcJob{
		address:        g.address,
		requestTimeout: g.requestTimeout,
		maxRetry:       g.maxRetry,
		interceptors:   g.interceptors,
		regs:           g.regs,
		tracer:         g.tracer,
		mu:             sync.Mutex{},
	}, nil
}
