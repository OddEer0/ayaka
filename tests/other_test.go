package ayaka

import (
	"context"
	"testing"

	"github.com/OddEer0/ayaka"
	"github.com/OddEer0/eelog"
	"github.com/OddEer0/eelog/logtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	app := ayaka.NewApp[*Container](ayaka.Options[*Container]{
		Name:        "my-app",
		Description: "my-app description testing",
		Version:     "1.0.0",
		Container:   &Container{},
	})

	ctx := app.Context()
	appRes, err := ayaka.AppFromContext[*Container](ctx)
	assert.NoError(t, err)
	assert.NotNil(t, appRes)

	assert.Nil(t, appRes.Err())
	assert.Equal(t, ayaka.Info{
		Name:        "my-app",
		Description: "my-app description testing",
		Version:     "1.0.0",
	}, appRes.Info())
	assert.NotNil(t, appRes.Container())
	assert.NotZero(t, appRes.Context())

	appRes, err = ayaka.AppFromContext[*Container](context.Background())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ayaka.ErrAppNotFountInContext))
	assert.Nil(t, appRes)
}

func TestWithJob(t *testing.T) {
	logger := logtest.NewLogTest(eelog.DebugLvl)

	app := ayaka.NewApp[*Container](ayaka.Options[*Container]{
		Name:        "my-app",
		Description: "my-app description testing",
		Version:     "1.0.0",
		Container:   &Container{},
		Logger:      logger,
	}).WithJob(
		NoopJob[*Container]{
			name: "my-job",
		},
		NoopJob[*Container]{
			name: "my-job",
		},
		NoopJob[*Container]{
			name: "my-job",
		},
	)

	err := app.Start()
	assert.NoError(t, err)

	filtered := Filter(logger.Messages(), func(s string) bool {
		if s == ayaka.LogWarnJobAlreadyExists {
			return true
		}
		return false
	})
	assert.Len(t, filtered, 2)
}
