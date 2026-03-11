package ayaka

import (
	"encoding/json"
	"time"
)

type Config struct {
	StartTimeout    time.Duration
	GracefulTimeout time.Duration
	Info            map[string]interface{}
}

func (c Config) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"start_timeout":    c.StartTimeout / time.Second,
		"graceful_timeout": c.GracefulTimeout / time.Second,
		"info":             c.Info,
	})
}

func (a *App[T]) WithConfig(conf *Config) *App[T] {
	if a.err != nil {
		return a
	}

	if a.configInterceptor != nil {
		var err error
		conf, err = a.configInterceptor(a.ctx, conf)
		if err != nil {
			a.err = err

			a.logger.Error(a.ctx, "config interceptor failed", nil)
			return a
		}
	}

	a.logger.Info(a.ctx, "add new config", map[string]any{
		"config": conf,
	})
	a.config = conf
	return a
}
