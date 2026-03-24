package ayaka

import (
	"context"
	"errors"
	"testing"

	"github.com/OddEer0/ayaka"
	"github.com/OddEer0/eelog"
	"github.com/OddEer0/eelog/logtest"
	"github.com/stretchr/testify/assert"
)

type NoopJob[T any] struct {
	name string
}

func (n NoopJob[T]) Name() string {
	return n.name
}

func (n NoopJob[T]) Init(_ context.Context, _ T) error {
	return nil
}

func (n NoopJob[T]) Run(_ context.Context, _ T) error {
	return nil
}

func NewNoopJob[T any](name string) NoopJob[T] {
	return NoopJob[T]{name: name}
}

type Container struct {
	value string
}

func TestNewApp_init(t *testing.T) {
	app := ayaka.NewApp(ayaka.Options[Container]{
		Name:        "test-app",
		Version:     "1.0.0",
		Description: "testing application",
	}).WithJob(
		NewNoopJob[Container]("job-1"),
	)

	err := app.Start()
	assert.Nil(t, err)

	assert.Equal(t, ayaka.DefaultStartTimeout, app.StartTimeout())
	assert.Equal(t, ayaka.DefaultGracefulTimeout, app.GracefulTimeout())
}

func TestNewApp_validate_option(t *testing.T) {
	t.Run("Should require name", func(t *testing.T) {
		app := ayaka.NewApp(ayaka.Options[Container]{
			Name:        "",
			Version:     "1.0.0",
			Description: "testing application",
		}).WithJob(
			NewNoopJob[Container]("job-1"),
		)

		err := app.Start()
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ayaka.ErrInvalidArgument))
	})

	t.Run("Should require version", func(t *testing.T) {
		app := ayaka.NewApp(ayaka.Options[Container]{
			Name:        "test-app",
			Version:     "",
			Description: "testing application",
		}).WithJob(
			NewNoopJob[Container]("job-1"),
		)

		err := app.Start()
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ayaka.ErrInvalidArgument))
	})

	t.Run("Should require description", func(t *testing.T) {
		app := ayaka.NewApp(ayaka.Options[Container]{
			Name:        "test-app",
			Version:     "1.0.0",
			Description: "",
		}).WithJob(
			NewNoopJob[Container]("job-1"),
		)

		err := app.Start()
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ayaka.ErrInvalidArgument))
	})
}

func TestNewAppOptionalOption(t *testing.T) {
	startTimeout := ayaka.DefaultStartTimeout * 2
	gracefulTimeout := ayaka.DefaultGracefulTimeout * 2

	beforeInitCalled := false
	afterInitCalled := false
	beforeRunCalled := false
	afterRunCalled := false

	log := logtest.NewLogTest(eelog.DebugLvl)

	app := ayaka.NewApp(ayaka.Options[Container]{
		Name:        "test-app",
		Version:     "1.0.0",
		Description: "testing application",
		Container: Container{
			value: "value",
		},
		StartTimeout:    startTimeout,
		GracefulTimeout: gracefulTimeout,
		Logger:          log,
		Hooks: ayaka.Hooks[Container]{
			BeforeInit: func(ctx context.Context, job ayaka.Job[Container]) {
				beforeInitCalled = true
			},
			AfterInit: func(ctx context.Context, job ayaka.Job[Container], err error) {
				afterInitCalled = true
			},
			BeforeRun: func(ctx context.Context, job ayaka.Job[Container]) {
				beforeRunCalled = true
			},
			AfterRun: func(ctx context.Context, job ayaka.Job[Container], err error) {
				afterRunCalled = true
			},
		},
	}).WithJob(
		NewNoopJob[Container]("job-1"),
	)
	err := app.Start()
	assert.Nil(t, err)

	assert.NotNil(t, app.Logger())
	assert.True(t, beforeInitCalled)
	assert.True(t, afterInitCalled)
	assert.True(t, beforeRunCalled)
	assert.True(t, afterRunCalled)
	assert.Equal(t, startTimeout, app.StartTimeout())
	assert.Equal(t, gracefulTimeout, app.GracefulTimeout())

	assert.Equal(t, []string{
		ayaka.LogInfoInitAllJobsStarted,
		ayaka.LogInfoInitAllJobsFinished,
		ayaka.LogInfoRunAllJobsStarted,
		ayaka.LogInfoRunAllJobsFinished,
	}, Filter(log.Messages(), filterPickGlobalLogsMessages))
}
