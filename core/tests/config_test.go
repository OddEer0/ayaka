package ayaka

import (
	"testing"
	"time"

	ayaka "github.com/OddEer0/ayaka/core"
	"github.com/stretchr/testify/assert"
)

func TestWithConfig(t *testing.T) {
	t.Run("Should correct with config", func(t *testing.T) {
		app := ayaka.NewApp(&ayaka.Options{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
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
		app := ayaka.NewApp(&ayaka.Options{
			Name:        "my-app",
			Description: "my-app description testing",
		}).WithConfig(&ayaka.Config{
			StartTimeout:    time.Second * 2,
			GracefulTimeout: time.Second * 3,
			Info: map[string]any{
				"test": "kek",
			},
		})
		assert.Equal(t, &ayaka.Config{}, app.Config())
	})
}
