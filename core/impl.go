package ayaka

import (
	"context"
	"github.com/pkg/errors"
)

func (a *App) Start() error {
	if a.err != nil {
		return a.err
	}
	if len(a.jobs) == 0 {
		return a.err
	}

	ctxWithStartTimeout, cancel := context.WithTimeout(a.ctx, a.Config().StartTimeout)
	defer cancel()

	a.logger.Info(a.ctx, "init all job started", map[string]any{
		"init_timeout": a.Config().StartTimeout,
	})
	err := a.initJob(ctxWithStartTimeout)
	if err != nil {
		return errors.Wrap(err, "[App] initJob")
	}

	ctx, cncl := context.WithCancel(a.ctx)
	defer cncl()

	a.logger.Info(a.ctx, "run all job started", nil)
	err = a.runJob(ctx)
	if err != nil {
		return errors.Wrap(err, "[App] runJob")
	}

	a.logger.Info(a.ctx, "run all job finished", nil)
	return nil
}

func (a *App) WithJob(jobEntries ...JobEntry) *App {
	if a.err != nil {
		return a
	}

	for _, entry := range jobEntries {
		a.jobs[entry.Key] = entry.Job
	}

	return a
}
