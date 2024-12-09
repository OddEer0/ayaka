package ecosystem

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func TimeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (interface{}, error) {
		var err error
		var result interface{}

		childCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		done := make(chan struct{})

		go func() {
			result, err = handler(childCtx, req)
			close(done)
		}()

		select {
		case <-childCtx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return nil, status.New(codes.DeadlineExceeded, "Server timeout, aborting.").Err()
			}

			return nil, status.New(codes.Canceled, "Client cancelled, abandoning.").Err()
		case <-done:
			return result, err
		}
	}
}
