package ayaka

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	ayaka2 "github.com/OddEer0/ayaka"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestMarshalConfig(t *testing.T) {
	cfg := ayaka2.Config{
		StartTimeout:    time.Second * 3,
		GracefulTimeout: time.Second * 2,
		Info: map[string]interface{}{
			"test": "kek",
		},
	}

	expected := map[string]interface{}{
		"start_timeout":    3,
		"graceful_timeout": 2,
		"info": map[string]interface{}{
			"test": "kek",
		},
	}

	result, err := json.Marshal(cfg)
	assert.NoError(t, err)
	extectedResult, err := json.Marshal(expected)
	assert.NoError(t, err)
	assert.Equal(t, string(extectedResult), string(result))
}

func TestContext(t *testing.T) {
	app := ayaka2.NewApp[*Container](&ayaka2.Options[*Container]{
		Name:        "my-app",
		Description: "my-app description testing",
		Version:     "1.0.0",
		Container:   &Container{},
	})

	ctx := app.Context()
	appRes, err := ayaka2.AppFromContext[*Container](ctx)
	assert.NoError(t, err)
	assert.NotNil(t, appRes)

	assert.Nil(t, appRes.Err())
	assert.Equal(t, &ayaka2.Config{}, appRes.Config())
	assert.Equal(t, ayaka2.Info{
		Name:        "my-app",
		Description: "my-app description testing",
		Version:     "1.0.0",
	}, appRes.Info())
	assert.NotNil(t, appRes.Container())
	assert.NotZero(t, appRes.Context())

	appRes, err = ayaka2.AppFromContext[*Container](context.Background())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ayaka2.ErrAppNotFountInContext))
	assert.Nil(t, appRes)
}

func TestNoopLogger(t *testing.T) {
	logger := ayaka2.NoopLogger{}
	ctx := context.Background()
	message := "message string"
	info := map[string]any{}

	logger.Debug(ctx, message, info)
	logger.Info(ctx, message, info)
	logger.Warn(ctx, message, info)
	logger.Error(ctx, message, info)
}
