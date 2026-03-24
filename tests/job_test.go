package ayaka

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/OddEer0/ayaka"
	"github.com/OddEer0/eelog"
	"github.com/OddEer0/eelog/logtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	initEnd        = "init end"
	initEndWithCtx = "init end with ctx done"
	runEnd         = "run end"
	runEndWithCtx  = "run end with ctx done"
)

type correctJob struct {
	name                string
	initDuration        time.Duration
	ctxDoneInitDuration time.Duration
	runDuration         time.Duration
	ctxDoneRunDuration  time.Duration
	errInit             error
	errRun              error
	panicInit           string
	panicRun            string
}

func (c correctJob) Name() string {
	return c.name
}

func (c correctJob) Init(ctx context.Context, container *Container) error {
	if c.panicInit != "" {
		panic(c.panicInit)
	}
	if c.errInit != nil {
		return c.errInit
	}

	t := time.NewTimer(c.initDuration)
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		if c.ctxDoneInitDuration > 0 {
			ti := time.NewTimer(c.ctxDoneInitDuration)
			<-ti.C
		}
		return ctx.Err()
	}
}

func (c correctJob) Run(ctx context.Context, container *Container) error {
	if c.panicRun != "" {
		panic(c.panicRun)
	}
	if c.errRun != nil {
		return c.errRun
	}

	t := time.NewTimer(c.runDuration)
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		if c.ctxDoneRunDuration > 0 {
			ti := time.NewTimer(c.ctxDoneRunDuration)
			<-ti.C
		}
		return ctx.Err()
	}
}

func TestWithJobErrorApp(t *testing.T) {
	t.Parallel()

	app := ayaka.NewApp[*Container](ayaka.Options[*Container]{
		Name:        "my-app",
		Description: "my-app description testing",
		Version:     "",
	}).WithJob(
		&correctJob{
			initDuration: time.Second * 1,
			runDuration:  time.Second * 1,
		},
	)

	assert.Error(t, app.Err())
	assert.ErrorIs(t, app.Err(), ayaka.ErrInvalidArgument)
	assert.Error(t, app.Start())

	app = ayaka.NewApp[*Container](ayaka.Options[*Container]{
		Name:        "my-app",
		Description: "my-app description testing",
		Version:     "1.0.0",
	})

	assert.Error(t, app.Start())
	assert.ErrorIs(t, app.Err(), ayaka.ErrNoJobs)
	assert.Error(t, app.Err())
}

func TestSingleJob(t *testing.T) {
	t.Run("Should correct init and run job", func(t *testing.T) {
		t.Parallel()
		logger := logtest.NewLogTest(eelog.DebugLvl)

		app := ayaka.NewApp[*Container](ayaka.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithJob(correctJob{
			name:         "my-test-job-1",
			initDuration: time.Second * 1,
			runDuration:  time.Second * 1,
		})

		err := app.Start()
		assert.NoError(t, err)
		assert.NoError(t, app.Err())
	})

	t.Run("Should correct error handle init job", func(t *testing.T) {
		t.Parallel()
		logger := logtest.NewLogTest(eelog.DebugLvl)
		myErr := errors.New("my error")

		app := ayaka.NewApp[*Container](ayaka.Options[*Container]{
			Name:            "my-app",
			Description:     "my-app description testing",
			Version:         "1.0.0",
			Container:       &Container{},
			Logger:          logger,
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}).WithJob(
			correctJob{
				name:    "my-test-job-1",
				errInit: myErr,
			},
		)

		err := app.Start()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ayaka.ErrJobInitFailed)

		assert.Equal(t, []string{
			ayaka.LogInfoInitAllJobsStarted,
			ayaka.LogInfoJobInitStarted,
			ayaka.LogErrorJobInitFailure,
		}, logger.Messages())
		assert.Equal(t, []eelog.Level{
			eelog.InfoLvl, eelog.InfoLvl, eelog.ErrorLvl,
		}, logger.Levels())
	})

	t.Run("Should correct panic handle init job", func(t *testing.T) {
		t.Parallel()
		logger := logtest.NewLogTest(eelog.DebugLvl)
		panicMessage := "panic init haha!!!"

		app := ayaka.NewApp(ayaka.Options[*Container]{
			Name:            "my-app",
			Description:     "my-app description testing",
			Version:         "1.0.0",
			Container:       &Container{},
			Logger:          logger,
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}).WithJob(
			correctJob{
				name:      "my-test-job-1",
				panicInit: panicMessage,
			},
		)

		err := app.Start()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ayaka.ErrJobInitPanic)

		assert.Equal(t, []string{
			ayaka.LogInfoInitAllJobsStarted,
			ayaka.LogInfoJobInitStarted,
			ayaka.LogErrorJobInitPanic,
		}, logger.Messages())
		assert.Equal(t, []eelog.Level{
			eelog.InfoLvl, eelog.InfoLvl, eelog.ErrorLvl,
		}, logger.Levels())
	})

	t.Run("Should correct error handle run job", func(t *testing.T) {
		t.Parallel()
		logger := logtest.NewLogTest(eelog.DebugLvl)
		myErr := errors.New("my error")

		app := ayaka.NewApp(ayaka.Options[*Container]{
			Name:            "my-app",
			Description:     "my-app description testing",
			Version:         "1.0.0",
			Container:       &Container{},
			Logger:          logger,
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}).WithJob(correctJob{
			errRun: myErr,
		})

		err := app.Start()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ayaka.ErrJobRunFailed)

		assert.Equal(t, []string{
			ayaka.LogInfoInitAllJobsStarted,
			ayaka.LogInfoJobInitStarted,
			ayaka.LogInfoJobInitFinished,
			ayaka.LogInfoInitAllJobsFinished,
			ayaka.LogInfoRunAllJobsStarted,
			ayaka.LogInfoJobRunStarted,
			ayaka.LogErrorJobRunFailure,
		}, logger.Messages())
		assert.Equal(t, []eelog.Level{
			eelog.InfoLvl, eelog.InfoLvl, eelog.InfoLvl, eelog.InfoLvl,
			eelog.InfoLvl, eelog.InfoLvl, eelog.ErrorLvl,
		}, logger.Levels())
	})

	t.Run("Should correct panic handle run job", func(t *testing.T) {
		t.Parallel()
		logger := logtest.NewLogTest(eelog.DebugLvl)
		panicMessage := "panic run haha!!!"

		app := ayaka.NewApp(ayaka.Options[*Container]{
			Name:            "my-app",
			Description:     "my-app description testing",
			Version:         "1.0.0",
			Container:       &Container{},
			Logger:          logger,
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}).WithJob(correctJob{
			panicRun: panicMessage,
		})

		err := app.Start()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ayaka.ErrJobRunPanic)

		assert.Equal(t, []string{
			ayaka.LogInfoInitAllJobsStarted,
			ayaka.LogInfoJobInitStarted,
			ayaka.LogInfoJobInitFinished,
			ayaka.LogInfoInitAllJobsFinished,
			ayaka.LogInfoRunAllJobsStarted,
			ayaka.LogInfoJobRunStarted,
			ayaka.LogErrorJobRunPanic,
		}, logger.Messages())
		assert.Equal(t, []eelog.Level{
			eelog.InfoLvl, eelog.InfoLvl, eelog.InfoLvl, eelog.InfoLvl,
			eelog.InfoLvl, eelog.InfoLvl, eelog.ErrorLvl,
		}, logger.Levels())
	})
}

func TestMultipleJobs(t *testing.T) {
	t.Run("Should correct init and run jobs", func(t *testing.T) {
		t.Parallel()
		logger := logtest.NewLogTest(eelog.DebugLvl)

		jobCount := 4
		j := 1
		multiJ := 300
		jobEntries := make([]ayaka.Job[*Container], 0, jobCount)
		for i := 0; i < jobCount; i++ {
			jobEntries = append(jobEntries,
				correctJob{
					name:         fmt.Sprintf("my-test-job-%d", i+1),
					initDuration: time.Millisecond * time.Duration(j*multiJ),
				},
			)
			j++
		}

		app := ayaka.NewApp(ayaka.Options[*Container]{
			Name:            "my-app",
			Description:     "my-app description testing",
			Version:         "1.0.0",
			Container:       &Container{},
			Logger:          logger,
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}).WithJob(jobEntries...)

		ti := time.Now()

		err := app.Start()

		duration := time.Since(ti)
		assert.True(t, duration > time.Millisecond*time.Duration((j-1)*multiJ))
		assert.NoError(t, err)
		assert.NoError(t, app.Err())

		assert.Equal(t, []string{
			ayaka.LogInfoInitAllJobsStarted,
			ayaka.LogInfoInitAllJobsFinished,
			ayaka.LogInfoRunAllJobsStarted,
			ayaka.LogInfoRunAllJobsFinished,
		}, Filter(logger.Messages(), filterPickGlobalLogsMessages))
	})

	t.Run("Should correct error handle init jobs", func(t *testing.T) {
		t.Parallel()
		logger := logtest.NewLogTest(eelog.DebugLvl)
		myError := errors.New("my error")

		jobCount := 2
		j := 1
		multiJ := 300
		jobEntries := make([]ayaka.Job[*Container], 0, jobCount)

		for i := 0; i < jobCount; i++ {
			jobEntries = append(jobEntries,
				correctJob{
					name:         fmt.Sprintf("my-test-job-%d", i+1),
					initDuration: time.Millisecond * time.Duration(j*multiJ),
				},
			)
			j++
		}

		// error
		jobEntries = append(jobEntries,
			correctJob{
				name:    fmt.Sprintf("my-test-job-%d", j),
				errInit: myError,
			},
		)

		app := ayaka.NewApp(ayaka.Options[*Container]{
			Name:            "my-app",
			Description:     "my-app description testing",
			Version:         "1.0.0",
			Container:       &Container{},
			Logger:          logger,
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}).WithJob(jobEntries...)

		ti := time.Now()

		err := app.Start()

		duration := time.Since(ti)
		assert.True(t, duration < time.Millisecond*time.Duration(j*multiJ))
		assert.Error(t, err)
		assert.ErrorIs(t, err, myError)
		assert.ErrorIs(t, err, ayaka.ErrJobInitFailed)
		assert.NoError(t, app.Err())
	})

	t.Run("Should correct error panic init jobs", func(t *testing.T) {
		t.Parallel()
		logger := logtest.NewLogTest(eelog.DebugLvl)

		jobCount := 2
		j := 1
		multiJ := 300
		jobEntries := make([]ayaka.Job[*Container], 0, jobCount)
		for i := 0; i < jobCount; i++ {
			jobEntries = append(jobEntries,
				correctJob{
					name:         fmt.Sprintf("my-test-job-%d", i+1),
					initDuration: time.Millisecond * time.Duration(j*multiJ),
				},
			)
			j++
		}

		// error
		jobEntries = append(jobEntries,
			correctJob{
				name:      fmt.Sprintf("my-test-job-%d", j),
				panicInit: "panic xd",
			},
		)

		app := ayaka.NewApp(ayaka.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithJob(jobEntries...)

		ti := time.Now()

		err := app.Start()

		duration := time.Since(ti)
		assert.True(t, duration < time.Millisecond*time.Duration(j*multiJ))
		assert.Error(t, err)
		assert.ErrorIs(t, err, ayaka.ErrJobInitPanic)
		assert.NoError(t, app.Err())
	})

	t.Run("Should correct error handle run jobs", func(t *testing.T) {
		t.Parallel()
		logger := logtest.NewLogTest(eelog.DebugLvl)
		myError := errors.New("my error")

		jobCount := 2
		j := 1
		multiJ := 300
		jobEntries := make([]ayaka.Job[*Container], 0, jobCount)
		for i := 0; i < jobCount; i++ {
			jobEntries = append(jobEntries,
				correctJob{
					name:         fmt.Sprintf("my-test-job-%d", i+1),
					initDuration: time.Millisecond * time.Duration((jobCount-1)*multiJ),
					runDuration:  time.Second * 5,
				},
			)
			j++
		}

		// error
		jobEntries = append(jobEntries,
			correctJob{
				name:         fmt.Sprintf("my-test-job-%d", j),
				initDuration: time.Millisecond * time.Duration((j-1)*multiJ),
				errRun:       myError,
			},
		)

		app := ayaka.NewApp(ayaka.Options[*Container]{
			Name:            "my-app",
			Description:     "my-app description testing",
			Version:         "1.0.0",
			Container:       &Container{},
			Logger:          logger,
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}).WithJob(jobEntries...)

		ti := time.Now()

		err := app.Start()

		duration := time.Since(ti)
		assert.True(t, duration < time.Millisecond*time.Duration(j*multiJ))
		assert.Error(t, err)
		assert.ErrorIs(t, err, ayaka.ErrJobRunFailed)
		assert.ErrorIs(t, err, myError)
		assert.NoError(t, app.Err())

	})

	t.Run("Should correct panic handler run jobs", func(t *testing.T) {
		t.Parallel()
		logger := logtest.NewLogTest(eelog.DebugLvl)

		jobCount := 2
		j := 1
		multiJ := 300
		jobEntries := make([]ayaka.Job[*Container], 0, jobCount)
		for i := 0; i < jobCount; i++ {
			jobEntries = append(jobEntries,
				correctJob{
					name:         fmt.Sprintf("my-test-job-%d", i+1),
					initDuration: time.Millisecond * time.Duration((jobCount-1)*multiJ),
					runDuration:  time.Second * 5,
				},
			)
			j++
		}

		// error
		jobEntries = append(jobEntries,
			correctJob{
				name:         fmt.Sprintf("my-test-job-%d", j),
				initDuration: time.Millisecond * time.Duration((j-1)*multiJ),
				panicRun:     "panic xd",
			},
		)

		app := ayaka.NewApp(ayaka.Options[*Container]{
			Name:            "my-app",
			Description:     "my-app description testing",
			Version:         "1.0.0",
			Container:       &Container{},
			Logger:          logger,
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}).WithJob(jobEntries...)

		ti := time.Now()

		err := app.Start()

		duration := time.Since(ti)
		assert.True(t, duration < time.Millisecond*time.Duration(j*multiJ))
		assert.Error(t, err)
		assert.ErrorIs(t, err, ayaka.ErrJobRunPanic)
		assert.NoError(t, app.Err())
	})
}

func TestJobsTimeout(t *testing.T) {
	t.Run("Should correct stop init with start timeout 1", func(t *testing.T) {
		t.Parallel()
		logger := logtest.NewLogTest(eelog.DebugLvl)

		app := ayaka.NewApp(ayaka.Options[*Container]{
			Name:            "my-app",
			Description:     "my-app description testing",
			Version:         "1.0.0",
			Container:       &Container{},
			Logger:          logger,
			StartTimeout:    time.Second * 1,
			GracefulTimeout: time.Second * 2,
		}).WithJob(
			correctJob{
				name:         "my-test-job",
				initDuration: time.Second * 2,
				runDuration:  time.Second * 1,
			},
		)

		err := app.Start()
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.NoError(t, app.Err())
	})

	t.Run("Should correct stop init with start timeout 2", func(t *testing.T) {
		t.Parallel()
		logger := logtest.NewLogTest(eelog.DebugLvl)

		app := ayaka.NewApp(ayaka.Options[*Container]{
			Name:            "my-app",
			Description:     "my-app description testing",
			Version:         "1.0.0",
			Container:       &Container{},
			Logger:          logger,
			StartTimeout:    time.Second * 1,
			GracefulTimeout: time.Second * 2,
		}).WithJob(
			correctJob{
				name:         "my-test-job-1",
				initDuration: time.Second * 2,
				runDuration:  time.Second * 1,
			},
			correctJob{
				name:         "my-test-job-2",
				initDuration: time.Second * 0,
				runDuration:  time.Second * 5,
			},
		)

		err := app.Start()
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.NoError(t, app.Err())
	})

	t.Run("Should correct graceful timeout init job", func(t *testing.T) {
		t.Parallel()
		logger := logtest.NewLogTest(eelog.DebugLvl)

		myErr := errors.New("my error")

		app := ayaka.NewApp(ayaka.Options[*Container]{
			Name:            "my-app",
			Description:     "my-app description testing",
			Version:         "1.0.0",
			Container:       &Container{},
			Logger:          logger,
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 1,
		}).WithJob(
			correctJob{
				name:                "my-test-job",
				initDuration:        time.Second * 1,
				runDuration:         time.Second * 1,
				ctxDoneInitDuration: time.Second * 2,
			},
			correctJob{
				name:    "my-test-job-2",
				errInit: myErr,
			},
		)

		err := app.Start()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ayaka.ErrGracefulTimeout)
		assert.NoError(t, app.Err())
	})

	t.Run("Should correct graceful timeout run job", func(t *testing.T) {
		t.Parallel()
		logger := logtest.NewLogTest(eelog.DebugLvl)

		myErr := errors.New("my error")

		app := ayaka.NewApp(ayaka.Options[*Container]{
			Name:            "my-app",
			Description:     "my-app description testing",
			Version:         "1.0.0",
			Container:       &Container{},
			Logger:          logger,
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 1,
		}).WithJob(
			correctJob{
				name:               "my-test-job",
				initDuration:       time.Second * 1,
				runDuration:        time.Second * 2,
				ctxDoneRunDuration: time.Second * 2,
			},
			correctJob{
				name:   "my-test-job-2",
				errRun: myErr,
			},
		)

		err := app.Start()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ayaka.ErrGracefulTimeout)
		assert.NoError(t, app.Err())
	})
}
