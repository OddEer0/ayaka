package ayaka

import (
	"github.com/pkg/errors"
)

func (a *App[T]) Start() error {
	if a.err != nil {
		return a.err
	}
	if len(a.jobs) == 0 {
		return a.err
	}

	a.logger.Info(a.ctx, "init all job started", map[string]any{
		"init_timeout": a.Config().StartTimeout,
	})
	err := a.initJob()
	if err != nil {
		return errors.Wrap(err, "[App] initJob")
	}

	a.logger.Info(a.ctx, "run all job started", nil)
	err = a.runJob()
	if err != nil {
		return errors.Wrap(err, "[App] runJob")
	}

	a.logger.Info(a.ctx, "run all job finished", nil)
	return nil
}

func (a *App[T]) WithJob(jobEntries ...JobEntry[T]) *App[T] {
	if a.err != nil {
		return a
	}

	for _, entry := range jobEntries {
		a.jobs[entry.Key] = entry.Job
	}

	return a
}
