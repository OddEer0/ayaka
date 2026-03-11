package ayaka

import (
	"context"
	"testing"
	"time"

	"errors"

	ayaka "github.com/OddEer0/ayaka/core"
	"github.com/stretchr/testify/assert"
)

type Container struct{}

func TestWithConfig(t *testing.T) {
	t.Run("Should correct with config", func(t *testing.T) {
		app := ayaka.NewApp[*Container](&ayaka.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
		}).WithConfig(&ayaka.Config{
			StartTimeout:    time.Second * 2,
			GracefulTimeout: time.Second * 3,
			Info: map[string]any{
				"test": "kek",
			},
		})
		assert.NoError(t, app.Err())
		assert.NoError(t, app.Start())
		assert.Equal(t, &ayaka.Config{
			StartTimeout:    time.Second * 2,
			GracefulTimeout: time.Second * 3,
			Info: map[string]any{
				"test": "kek",
			},
		}, app.Config())
	})

	t.Run("Should not worked with error app", func(t *testing.T) {
		app := ayaka.NewApp[*Container](&ayaka.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Container:   &Container{},
		}).WithConfig(&ayaka.Config{
			StartTimeout:    time.Second * 2,
			GracefulTimeout: time.Second * 3,
			Info: map[string]any{
				"test": "kek",
			},
		})
		assert.Equal(t, &ayaka.Config{}, app.Config())
	})

	t.Run("Should correct with interceptor", func(t *testing.T) {
		app := ayaka.NewApp[*Container](&ayaka.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			ConfigInterceptor: func(ctx context.Context, conf *ayaka.Config) (*ayaka.Config, error) {
				conf.StartTimeout = time.Second * 2
				conf.GracefulTimeout = time.Second * 3
				conf.Info = map[string]any{
					"test": "kek",
				}
				return conf, nil
			},
		}).WithConfig(&ayaka.Config{})
		assert.NoError(t, app.Err())
		assert.NoError(t, app.Start())
		assert.Equal(t, &ayaka.Config{
			StartTimeout:    time.Second * 2,
			GracefulTimeout: time.Second * 3,
			Info: map[string]any{
				"test": "kek",
			},
		}, app.Config())
	})

	t.Run("Should correct error with error interceptor", func(t *testing.T) {
		app := ayaka.NewApp[*Container](&ayaka.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			ConfigInterceptor: func(ctx context.Context, conf *ayaka.Config) (*ayaka.Config, error) {
				return conf, errors.New("some error")
			},
		}).WithConfig(&ayaka.Config{})
		assert.Error(t, app.Err())
		assert.Error(t, app.Start())
		assert.Equal(t, &ayaka.Config{}, app.Config())
	})
}
