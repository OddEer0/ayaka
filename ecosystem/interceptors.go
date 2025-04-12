package ecosystem

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TimeoutInterceptor(timeout time.Duration, recover func()) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (interface{}, error) {
		var err error
		var result interface{}

		childCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		done := make(chan struct{})

		go func() {
			defer recover()
			result, err = handler(childCtx, req)
			close(done)
		}()

		select {
		case <-childCtx.Done():
			if errors.Is(childCtx.Err(), context.DeadlineExceeded) {
				return nil, status.New(codes.DeadlineExceeded, "Server timeout, aborting.").Err()
			}

			return nil, status.New(codes.Canceled, "Client cancelled, abandoning.").Err()
		case <-done:
			return result, err
		}
	}
}

// RetryInterceptorOptions - retry query, if it error request
type RetryInterceptorOptions struct {
	// Retry count
	Count uint
	// Retry worked only pick set code, Omit do not work if it has Pick set
	Pick []codes.Code
	// Retry not worked pick set code, Do not work if it has Pick set
	Omit []codes.Code
	// Retry count by code
	CodeCount map[codes.Code]uint
}

func RetryInterceptor(_ *RetryInterceptorOptions) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		return handler(ctx, req)
	}
}
