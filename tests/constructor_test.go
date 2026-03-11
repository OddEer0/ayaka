package ayaka

import (
	"context"
	"testing"

	ayaka2 "github.com/OddEer0/ayaka"
	"github.com/stretchr/testify/assert"
)

func TestConstructor(t *testing.T) {
	t.Run("Should correct init", func(t *testing.T) {
		app := ayaka2.NewApp[*Container](&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
		})
		assert.NoError(t, app.Err())
		assert.NoError(t, app.Start())
		assert.Equal(t, "my-app", app.Info().Name)
		assert.Equal(t, "my-app description testing", app.Info().Description)
		assert.Equal(t, "1.0.0", app.Info().Version)
		assert.NotNil(t, app.Container())
		assert.NotNil(t, app.Config())
		assert.NotEmpty(t, app.Context())
		appFromCtx, err := ayaka2.AppFromContext[*Container](app.Context())
		assert.NoError(t, err)
		assert.NotNil(t, appFromCtx)

		appFromCtx, err = ayaka2.AppFromContext[*Container](context.Background())
		assert.Nil(t, appFromCtx)
		assert.Equal(t, ayaka2.ErrAppNotFountInContext, err)
	})

	t.Run("Should error with empty required Name, Container, Description and Version fields", func(t *testing.T) {
		container := &Container{}
		app := ayaka2.NewApp[*Container](&ayaka2.Options[*Container]{
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   container,
		})
		assert.Error(t, app.Err())
		assert.Error(t, app.Start())

		app = ayaka2.NewApp[*Container](&ayaka2.Options[*Container]{
			Name:      "my-app",
			Version:   "1.0.0",
			Container: container,
		})
		assert.Error(t, app.Err())
		assert.Error(t, app.Start())

		app = ayaka2.NewApp[*Container](&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Container:   container,
		})
		assert.Error(t, app.Err())
		assert.Error(t, app.Start())

		app = ayaka2.NewApp[*Container](&ayaka2.Options[*Container]{
			Name:        "my-app",
			Version:     "1.0.0",
			Description: "my-app description testing",
		})
		assert.Error(t, app.Err())
		assert.Error(t, app.Start())
	})
}
