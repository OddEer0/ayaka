package ayaka

import (
	"context"
)

type appKey struct{}

func AppWithContext[T any](ctx context.Context, app *App[T]) context.Context {
	return context.WithValue(ctx, appKey{}, app)
}

func AppFromContext[T any](ctx context.Context) (*ReadonlyApp[T], error) {
	val := ctx.Value(appKey{})
	if val == nil {
		return nil, ErrAppNotFountInContext
	}
	result := val.(*App[T])
	return &ReadonlyApp[T]{
		app: result,
	}, nil
}
