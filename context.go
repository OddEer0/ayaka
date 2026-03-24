package ayaka

import (
	"context"
)

type appKey[T any] struct{}

func AppWithContext[T any](ctx context.Context, app *App[T]) context.Context {
	return context.WithValue(ctx, appKey[T]{}, app)
}

func AppFromContext[T any](ctx context.Context) (*App[T], error) {
	val := ctx.Value(appKey[T]{})
	if val == nil {
		return nil, ErrAppNotFountInContext
	}
	result, ok := val.(*App[T])
	if !ok {
		return nil, ErrAppNotFountInContext
	}
	return result, nil
}
