package ayaka_test

import (
	"context"
	"testing"

	ayaka "github.com/OddEer0/ayaka/core"
	"github.com/stretchr/testify/assert"
)

func TestConstructor(t *testing.T) {
	t.Run("Should correct init", func(t *testing.T) {
		app := ayaka.NewApp(&ayaka.Options{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
		})
		assert.NoError(t, app.Err())
		assert.NoError(t, app.Start())
		assert.Equal(t, "my-app", app.Info().Name)
		assert.Equal(t, "my-app description testing", app.Info().Description)
		assert.Equal(t, "1.0.0", app.Info().Version)
		assert.NotNil(t, app.Dependency())
		assert.NotNil(t, app.Config())
		assert.NotEmpty(t, app.Context())
		appFromCtx, err := ayaka.AppFromContext(app.Context())
		assert.NoError(t, err)
		assert.NotNil(t, appFromCtx)

		appFromCtx, err = ayaka.AppFromContext(context.Background())
		assert.Nil(t, appFromCtx)
		assert.Equal(t, ayaka.ErrAppNotFountInContext, err)
	})

	t.Run("Should error with empty required Name, Description and Version fields", func(t *testing.T) {
		app := ayaka.NewApp(&ayaka.Options{
			Description: "my-app description testing",
			Version:     "1.0.0",
		})
		assert.Error(t, app.Err())
		assert.Error(t, app.Start())

		app = ayaka.NewApp(&ayaka.Options{
			Name:    "my-app",
			Version: "1.0.0",
		})
		assert.Error(t, app.Err())
		assert.Error(t, app.Start())

		app = ayaka.NewApp(&ayaka.Options{
			Name:        "my-app",
			Description: "my-app description testing",
		})
		assert.Error(t, app.Err())
		assert.Error(t, app.Start())
	})
}
