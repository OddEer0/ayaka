package ayaka

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	ayaka2 "github.com/OddEer0/ayaka"
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
	initDuration        time.Duration
	ctxDoneInitDuration time.Duration
	runDuration         time.Duration
	ctxDoneRunDuration  time.Duration
	errInit             error
	errRun              error
	panicInit           string
	panicRun            string
}

func (c correctJob) Init(ctx context.Context, container *Container) error {
	app, err := ayaka2.AppFromContext[*Container](ctx)
	if err != nil {
		return err
	}
	logger := app.Logger()

	if c.panicInit != "" {
		panic(c.panicInit)
	}
	if c.errInit != nil {
		return c.errInit
	}

	t := time.NewTimer(c.initDuration)
	select {
	case <-t.C:
		logger.Debug(ctx, initEnd, nil)
		return nil
	case <-ctx.Done():
		if c.ctxDoneInitDuration > 0 {
			ti := time.NewTimer(c.ctxDoneInitDuration)
			<-ti.C
		}
		logger.Debug(ctx, initEndWithCtx, nil)
		return ctx.Err()
	}
}

func (c correctJob) Run(ctx context.Context, container *Container) error {
	app, err := ayaka2.AppFromContext[*Container](ctx)
	if err != nil {
		return err
	}
	logger := app.Logger()

	if c.panicRun != "" {
		panic(c.panicRun)
	}
	if c.errRun != nil {
		return c.errRun
	}

	t := time.NewTimer(c.runDuration)
	select {
	case <-t.C:
		logger.Debug(ctx, runEnd, nil)
		return nil
	case <-ctx.Done():
		if c.ctxDoneRunDuration > 0 {
			ti := time.NewTimer(c.ctxDoneRunDuration)
			<-ti.C
		}
		logger.Debug(ctx, runEndWithCtx, nil)
		return ctx.Err()
	}
}

func TestWithJobErrorApp(t *testing.T) {
	t.Parallel()

	app := ayaka2.NewApp[*Container](&ayaka2.Options[*Container]{
		Name:        "my-app",
		Description: "my-app description testing",
		Version:     "1.0.0",
	}).WithJob(ayaka2.JobEntry[*Container]{
		Key: "my-test-job",
		Job: &correctJob{
			initDuration: time.Second * 1,
			runDuration:  time.Second * 1,
		},
	})

	assert.Error(t, app.Err())
	assert.Error(t, app.Start())
}

func TestSingleJob(t *testing.T) {
	t.Run("Should correct init and run job", func(t *testing.T) {
		t.Parallel()
		logger := newTestLogger()

		cfg := &ayaka2.Config{
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}

		app := ayaka2.NewApp[*Container](&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithConfig(cfg).WithJob(ayaka2.JobEntry[*Container]{
			Key: "my-test-job",
			Job: &correctJob{
				initDuration: time.Second * 1,
				runDuration:  time.Second * 1,
			},
		})

		err := app.Start()
		assert.NoError(t, err)
		assert.NoError(t, app.Err())

		data := logger.getData()
		data.messages = data.messages[1:]
		data.levels = data.levels[1:]
		data.infos = data.infos[1:]

		assert.Equal(t,
			[]string{"init all job started", "init end", "run all job started", "run end", "run all job finished"},
			data.messages)
		assert.Equal(t,
			[]string{"info", "debug", "info", "debug", "info"},
			data.levels)
		assert.Equal(t, []map[string]any{
			{
				"init_timeout": time.Second * 5,
			},
			nil,
			nil,
			nil,
			nil,
		}, data.infos)
	})

	t.Run("Should correct error handle init job", func(t *testing.T) {
		t.Parallel()
		logger := newTestLogger()
		myErr := errors.New("my error")

		app := ayaka2.NewApp[*Container](&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithConfig(&ayaka2.Config{
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}).WithJob(ayaka2.JobEntry[*Container]{
			Key: "my-test-job-1",
			Job: &correctJob{
				errInit: myErr,
			},
		})

		err := app.Start()
		assert.Error(t, err)
		assert.Equal(t,
			fmt.Errorf(ayaka2.FormatErrJobInitFailed, "my-test-job-1", myErr).Error(),
			errors.Cause(err).Error(),
		)

		data := logger.getData()
		data.messages = data.messages[2:]
		data.levels = data.levels[2:]
		data.infos = data.infos[2:]

		assert.Equal(t, []string{ayaka2.LogMessageInitError}, data.messages)
		assert.Equal(t, []string{"error"}, data.levels)
		assert.Equal(t, []map[string]any{
			{
				ayaka2.LogKeyInfoError: myErr.Error(),
				ayaka2.LogKeyInfoKey:   "my-test-job-1",
			},
		}, data.infos)
	})

	t.Run("Should correct panic handle init job", func(t *testing.T) {
		t.Parallel()
		logger := newTestLogger()
		panicMessage := "panic init haha!!!"

		app := ayaka2.NewApp(&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithConfig(&ayaka2.Config{
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}).WithJob(ayaka2.JobEntry[*Container]{
			Key: "my-test-job-1",
			Job: &correctJob{
				panicInit: panicMessage,
			},
		})

		err := app.Start()
		assert.Error(t, err)
		assert.Equal(t,
			fmt.Errorf(ayaka2.FormatErrJobInitPanic, "my-test-job-1", panicMessage).Error(),
			errors.Cause(err).Error(),
		)

		data := logger.getData()
		data.messages = data.messages[2:]
		data.levels = data.levels[2:]
		data.infos = data.infos[2:]

		assert.Equal(t, []string{ayaka2.LogMessageInitPanic}, data.messages)
		assert.Equal(t, []string{"error"}, data.levels)
		assert.Equal(t, []map[string]any{
			{
				ayaka2.LogKeyInfoPanic: panicMessage,
				ayaka2.LogKeyInfoKey:   "my-test-job-1",
			},
		}, data.infos)
	})

	t.Run("Should correct error handle run job", func(t *testing.T) {
		t.Parallel()
		logger := newTestLogger()
		myErr := errors.New("my error")

		app := ayaka2.NewApp(&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithConfig(&ayaka2.Config{
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}).WithJob(ayaka2.JobEntry[*Container]{
			Key: "my-test-job-1",
			Job: &correctJob{
				errRun: myErr,
			},
		})

		err := app.Start()
		assert.Error(t, err)
		assert.Equal(t,
			fmt.Errorf(ayaka2.FormatErrJobRunFailed, "my-test-job-1", myErr).Error(),
			errors.Cause(err).Error(),
		)

		data := logger.getData()
		data.messages = data.messages[4:]
		data.levels = data.levels[4:]
		data.infos = data.infos[4:]

		assert.Equal(t, []string{ayaka2.LogMessageRunError}, data.messages)
		assert.Equal(t, []string{"error"}, data.levels)
		assert.Equal(t, []map[string]any{
			{
				ayaka2.LogKeyInfoError: myErr.Error(),
				ayaka2.LogKeyInfoKey:   "my-test-job-1",
			},
		}, data.infos)
	})

	t.Run("Should correct panic handle run job", func(t *testing.T) {
		t.Parallel()
		logger := newTestLogger()
		panicMessage := "panic run haha!!!"

		app := ayaka2.NewApp(&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithConfig(&ayaka2.Config{
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}).WithJob(ayaka2.JobEntry[*Container]{
			Key: "my-test-job-1",
			Job: &correctJob{
				panicRun: panicMessage,
			},
		})

		err := app.Start()
		assert.Error(t, err)
		assert.Equal(t,
			fmt.Errorf(ayaka2.FormatErrJobRunPanic, "my-test-job-1", panicMessage).Error(),
			errors.Cause(err).Error(),
		)

		data := logger.getData()
		data.messages = data.messages[4:]
		data.levels = data.levels[4:]
		data.infos = data.infos[4:]

		assert.Equal(t, []string{ayaka2.LogMessageRunPanic}, data.messages)
		assert.Equal(t, []string{"error"}, data.levels)
		assert.Equal(t, []map[string]any{
			{
				ayaka2.LogKeyInfoPanic: panicMessage,
				ayaka2.LogKeyInfoKey:   "my-test-job-1",
			},
		}, data.infos)
	})
}

func TestMultipleJobs(t *testing.T) {
	t.Run("Should correct init and run jobs", func(t *testing.T) {
		t.Parallel()
		logger := newTestLogger()

		cfg := &ayaka2.Config{
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}
		jobCount := 4
		j := 1
		multiJ := 300
		jobEntries := make([]ayaka2.JobEntry[*Container], 0, jobCount)
		expectedMessage := []string{
			"init all job started", "run all job finished", "run all job started",
		}
		expectedLevel := []string{
			"info", "info", "info",
		}
		for i := 0; i < jobCount; i++ {
			expectedMessage = append(expectedMessage, "init end", "run end")
			expectedLevel = append(expectedLevel, "debug", "debug")
			jobEntries = append(jobEntries, ayaka2.JobEntry[*Container]{
				Key: fmt.Sprintf("my-test-job-%d", i+1),
				Job: &correctJob{
					initDuration: time.Millisecond * time.Duration(j*multiJ),
				},
			})
			j++
		}

		app := ayaka2.NewApp(&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithConfig(cfg).WithJob(jobEntries...)

		ti := time.Now()

		err := app.Start()

		duration := time.Since(ti)
		assert.True(t, duration > time.Millisecond*time.Duration((j-1)*multiJ))
		assert.NoError(t, err)
		assert.NoError(t, app.Err())

		data := logger.getData()
		data.messages = data.messages[1:]
		data.levels = data.levels[1:]
		data.infos = data.infos[1:]

		sort.Strings(expectedMessage)
		sort.Strings(expectedLevel)
		sort.Strings(data.messages)
		sort.Strings(data.levels)

		assert.Equal(t,
			expectedMessage,
			data.messages)
		assert.Equal(t,
			expectedLevel,
			data.levels)
	})

	t.Run("Should correct error handle init jobs", func(t *testing.T) {
		t.Parallel()
		logger := newTestLogger()
		myError := errors.New("my error")

		cfg := &ayaka2.Config{
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}
		jobCount := 2
		j := 1
		multiJ := 300
		jobEntries := make([]ayaka2.JobEntry[*Container], 0, jobCount)
		expectedMessage := []string{
			"init all job started",
		}
		expectedLevel := []string{
			"info",
		}
		for i := 0; i < jobCount; i++ {
			expectedMessage = append(expectedMessage, ayaka2.LogMessageInitError, initEndWithCtx)
			expectedLevel = append(expectedLevel, "debug", "error")
			jobEntries = append(jobEntries, ayaka2.JobEntry[*Container]{
				Key: fmt.Sprintf("my-test-job-%d", i+1),
				Job: &correctJob{
					initDuration: time.Millisecond * time.Duration(j*multiJ),
				},
			})
			j++
		}

		// error
		jobEntries = append(jobEntries, ayaka2.JobEntry[*Container]{
			Key: fmt.Sprintf("my-test-job-%d", j),
			Job: &correctJob{
				errInit: myError,
			},
		})
		expectedMessage = append(expectedMessage, ayaka2.LogMessageInitError)
		expectedLevel = append(expectedLevel, "error")

		app := ayaka2.NewApp(&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithConfig(cfg).WithJob(jobEntries...)

		ti := time.Now()

		err := app.Start()

		duration := time.Since(ti)
		assert.True(t, duration < time.Millisecond*time.Duration(j*multiJ))
		assert.Error(t, err)
		assert.NoError(t, app.Err())

		data := logger.getData()
		data.messages = data.messages[1:]
		data.levels = data.levels[1:]
		data.infos = data.infos[1:]

		sort.Strings(expectedMessage)
		sort.Strings(expectedLevel)
		sort.Strings(data.messages)
		sort.Strings(data.levels)

		assert.Equal(t,
			expectedMessage,
			data.messages)
		assert.Equal(t,
			expectedLevel,
			data.levels)
	})

	t.Run("Should correct error panic init jobs", func(t *testing.T) {
		t.Parallel()
		logger := newTestLogger()

		cfg := &ayaka2.Config{
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}
		jobCount := 2
		j := 1
		multiJ := 300
		jobEntries := make([]ayaka2.JobEntry[*Container], 0, jobCount)
		expectedMessage := []string{
			"init all job started",
		}
		expectedLevel := []string{
			"info",
		}
		for i := 0; i < jobCount; i++ {
			expectedMessage = append(expectedMessage, ayaka2.LogMessageInitError, initEndWithCtx)
			expectedLevel = append(expectedLevel, "debug", "error")
			jobEntries = append(jobEntries, ayaka2.JobEntry[*Container]{
				Key: fmt.Sprintf("my-test-job-%d", i+1),
				Job: &correctJob{
					initDuration: time.Millisecond * time.Duration(j*multiJ),
				},
			})
			j++
		}

		// error
		jobEntries = append(jobEntries, ayaka2.JobEntry[*Container]{
			Key: fmt.Sprintf("my-test-job-%d", j),
			Job: &correctJob{
				panicInit: "panic xd",
			},
		})
		expectedMessage = append(expectedMessage, ayaka2.LogMessageInitPanic)
		expectedLevel = append(expectedLevel, "error")

		app := ayaka2.NewApp(&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithConfig(cfg).WithJob(jobEntries...)

		ti := time.Now()

		err := app.Start()

		duration := time.Since(ti)
		assert.True(t, duration < time.Millisecond*time.Duration(j*multiJ))
		assert.Error(t, err)
		assert.NoError(t, app.Err())

		data := logger.getData()
		data.messages = data.messages[1:]
		data.levels = data.levels[1:]
		data.infos = data.infos[1:]

		sort.Strings(expectedMessage)
		sort.Strings(expectedLevel)
		sort.Strings(data.messages)
		sort.Strings(data.levels)

		assert.Equal(t,
			expectedMessage,
			data.messages)
		assert.Equal(t,
			expectedLevel,
			data.levels)
	})

	t.Run("Should correct error handle run jobs", func(t *testing.T) {
		t.Parallel()
		logger := newTestLogger()
		myError := errors.New("my error")

		cfg := &ayaka2.Config{
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}
		jobCount := 2
		j := 1
		multiJ := 300
		jobEntries := make([]ayaka2.JobEntry[*Container], 0, jobCount)
		expectedMessage := []string{
			"init all job started",
		}
		expectedLevel := []string{
			"info",
		}
		for i := 0; i < jobCount; i++ {
			expectedMessage = append(expectedMessage, "init end", ayaka2.LogMessageRunError, runEndWithCtx)
			expectedLevel = append(expectedLevel, "debug", "debug", "error")
			jobEntries = append(jobEntries, ayaka2.JobEntry[*Container]{
				Key: fmt.Sprintf("my-test-job-%d", i+1),
				Job: &correctJob{
					initDuration: time.Millisecond * time.Duration((jobCount-1)*multiJ),
					runDuration:  time.Second * 5,
				},
			})
			j++
		}

		// error
		jobEntries = append(jobEntries, ayaka2.JobEntry[*Container]{
			Key: fmt.Sprintf("my-test-job-%d", j),
			Job: &correctJob{
				initDuration: time.Millisecond * time.Duration((j-1)*multiJ),
				errRun:       myError,
			},
		})
		expectedMessage = append(expectedMessage, ayaka2.LogMessageRunError, "init end", "run all job started")
		expectedLevel = append(expectedLevel, "error", "debug", "info")

		app := ayaka2.NewApp(&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithConfig(cfg).WithJob(jobEntries...)

		ti := time.Now()

		err := app.Start()

		duration := time.Since(ti)
		assert.True(t, duration < time.Millisecond*time.Duration(j*multiJ))
		assert.Error(t, err)
		assert.NoError(t, app.Err())

		data := logger.getData()
		data.messages = data.messages[1:]
		data.levels = data.levels[1:]
		data.infos = data.infos[1:]

		sort.Strings(expectedMessage)
		sort.Strings(expectedLevel)
		sort.Strings(data.messages)
		sort.Strings(data.levels)

		assert.Equal(t,
			expectedMessage,
			data.messages)
		assert.Equal(t,
			expectedLevel,
			data.levels)
	})

	t.Run("Should correct panic handler run jobs", func(t *testing.T) {
		t.Parallel()
		logger := newTestLogger()

		cfg := &ayaka2.Config{
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 5,
		}
		jobCount := 2
		j := 1
		multiJ := 300
		jobEntries := make([]ayaka2.JobEntry[*Container], 0, jobCount)
		expectedMessage := []string{
			"init all job started",
		}
		expectedLevel := []string{
			"info",
		}
		for i := 0; i < jobCount; i++ {
			expectedMessage = append(expectedMessage, "init end", ayaka2.LogMessageRunError, runEndWithCtx)
			expectedLevel = append(expectedLevel, "debug", "debug", "error")
			jobEntries = append(jobEntries, ayaka2.JobEntry[*Container]{
				Key: fmt.Sprintf("my-test-job-%d", i+1),
				Job: &correctJob{
					initDuration: time.Millisecond * time.Duration((jobCount-1)*multiJ),
					runDuration:  time.Second * 5,
				},
			})
			j++
		}

		// error
		jobEntries = append(jobEntries, ayaka2.JobEntry[*Container]{
			Key: fmt.Sprintf("my-test-job-%d", j),
			Job: &correctJob{
				initDuration: time.Millisecond * time.Duration((j-1)*multiJ),
				panicRun:     "panic xd",
			},
		})
		expectedMessage = append(expectedMessage, ayaka2.LogMessageRunPanic, "init end", "run all job started")
		expectedLevel = append(expectedLevel, "error", "debug", "info")

		app := ayaka2.NewApp(&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithConfig(cfg).WithJob(jobEntries...)

		ti := time.Now()

		err := app.Start()

		duration := time.Since(ti)
		assert.True(t, duration < time.Millisecond*time.Duration(j*multiJ))
		assert.Error(t, err)
		assert.NoError(t, app.Err())

		data := logger.getData()
		data.messages = data.messages[1:]
		data.levels = data.levels[1:]
		data.infos = data.infos[1:]

		sort.Strings(expectedMessage)
		sort.Strings(expectedLevel)
		sort.Strings(data.messages)
		sort.Strings(data.levels)

		assert.Equal(t,
			expectedMessage,
			data.messages)
		assert.Equal(t,
			expectedLevel,
			data.levels)
	})
}

func TestJobsTimout(t *testing.T) {
	t.Run("Should correct stop init with start timout 1", func(t *testing.T) {
		t.Parallel()
		logger := newTestLogger()

		cfg := &ayaka2.Config{
			StartTimeout:    time.Second * 1,
			GracefulTimeout: time.Second * 2,
		}

		app := ayaka2.NewApp(&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithConfig(cfg).WithJob(ayaka2.JobEntry[*Container]{
			Key: "my-test-job",
			Job: &correctJob{
				initDuration: time.Second * 2,
				runDuration:  time.Second * 1,
			},
		})

		err := app.Start()
		assert.Error(t, err)
		assert.NoError(t, app.Err())

		data := logger.getData()
		data.messages = data.messages[1:]
		data.levels = data.levels[1:]
		data.infos = data.infos[1:]

		assert.Equal(t,
			[]string{"init all job started", "init end with ctx done", ayaka2.LogMessageInitError},
			data.messages)
		assert.Equal(t,
			[]string{"info", "debug", "error"},
			data.levels)
		assert.Equal(t,
			[]map[string]any{
				{
					"init_timeout": time.Second * 1,
				}, nil, {
					ayaka2.LogKeyInfoKey:   "my-test-job",
					ayaka2.LogKeyInfoError: context.DeadlineExceeded.Error(),
				},
			},
			data.infos)
	})

	t.Run("Should correct stop init with start timout 2", func(t *testing.T) {
		t.Parallel()
		logger := newTestLogger()

		cfg := &ayaka2.Config{
			StartTimeout:    time.Second * 1,
			GracefulTimeout: time.Second * 2,
		}

		app := ayaka2.NewApp(&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithConfig(cfg).WithJob(ayaka2.JobEntry[*Container]{
			Key: "my-test-job",
			Job: &correctJob{
				initDuration: time.Second * 2,
				runDuration:  time.Second * 1,
			},
		}, ayaka2.JobEntry[*Container]{
			Key: "my-test-job-2",
			Job: &correctJob{
				initDuration: time.Second * 0,
				runDuration:  time.Second * 5,
			},
		})

		err := app.Start()
		assert.Error(t, err)
		assert.NoError(t, app.Err())

		data := logger.getData()
		data.messages = data.messages[1:]
		data.levels = data.levels[1:]
		data.infos = data.infos[1:]

		assert.Equal(t,
			[]string{"init all job started", "init end", "init end with ctx done", ayaka2.LogMessageInitError},
			data.messages)
		assert.Equal(t,
			[]string{"info", "debug", "debug", "error"},
			data.levels)
		assert.Equal(t,
			[]map[string]any{
				{
					"init_timeout": time.Second * 1,
				}, nil, nil, {
					ayaka2.LogKeyInfoKey:   "my-test-job",
					ayaka2.LogKeyInfoError: context.DeadlineExceeded.Error(),
				},
			},
			data.infos)
	})

	t.Run("Should correct graceful timeout init job", func(t *testing.T) {
		t.Parallel()
		logger := newTestLogger()

		cfg := &ayaka2.Config{
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 1,
		}

		myErr := errors.New("my error")

		app := ayaka2.NewApp(&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithConfig(cfg).WithJob(ayaka2.JobEntry[*Container]{
			Key: "my-test-job",
			Job: &correctJob{
				initDuration:        time.Second * 1,
				runDuration:         time.Second * 1,
				ctxDoneInitDuration: time.Second * 2,
			},
		}, ayaka2.JobEntry[*Container]{
			Key: "my-test-job-2",
			Job: &correctJob{
				errInit: myErr,
			},
		})

		err := app.Start()
		assert.Error(t, err)
		assert.NoError(t, app.Err())

		data := logger.getData()
		data.messages = data.messages[1:]
		data.levels = data.levels[1:]
		data.infos = data.infos[1:]

		assert.Equal(t,
			[]string{"init all job started", ayaka2.LogMessageInitError, ayaka2.LogMessageGracefulShotdownFailed},
			data.messages)
		assert.Equal(t,
			[]string{"info", "error", "warn"},
			data.levels)
		assert.Equal(t,
			[]map[string]any{
				{
					"init_timeout": time.Second * 5,
				}, {
					ayaka2.LogKeyInfoKey:   "my-test-job-2",
					ayaka2.LogKeyInfoError: myErr.Error(),
				}, nil,
			},
			data.infos)
	})

	t.Run("Should correct graceful timeout run job", func(t *testing.T) {
		t.Parallel()
		logger := newTestLogger()

		cfg := &ayaka2.Config{
			StartTimeout:    time.Second * 5,
			GracefulTimeout: time.Second * 1,
		}

		myErr := errors.New("my error")

		app := ayaka2.NewApp(&ayaka2.Options[*Container]{
			Name:        "my-app",
			Description: "my-app description testing",
			Version:     "1.0.0",
			Container:   &Container{},
			Logger:      logger,
		}).WithConfig(cfg).WithJob(ayaka2.JobEntry[*Container]{
			Key: "my-test-job",
			Job: &correctJob{
				initDuration:       time.Second * 1,
				runDuration:        time.Second * 2,
				ctxDoneRunDuration: time.Second * 2,
			},
		}, ayaka2.JobEntry[*Container]{
			Key: "my-test-job-2",
			Job: &correctJob{
				errRun: myErr,
			},
		})

		err := app.Start()
		assert.Error(t, err)
		assert.NoError(t, app.Err())

		data := logger.getData()
		data.messages = data.messages[1:]
		data.levels = data.levels[1:]
		data.infos = data.infos[1:]

		assert.Equal(t,
			[]string{"init all job started", "init end", "init end", "run all job started", ayaka2.LogMessageRunError, ayaka2.LogMessageGracefulShotdownFailed},
			data.messages)
		assert.Equal(t,
			[]string{"info", "debug", "debug", "info", "error", "warn"},
			data.levels)
		assert.Equal(t,
			[]map[string]any{
				{
					"init_timeout": time.Second * 5,
				}, nil, nil, nil, {
					ayaka2.LogKeyInfoKey:   "my-test-job-2",
					ayaka2.LogKeyInfoError: myErr.Error(),
				}, nil,
			},
			data.infos)
	})
}
