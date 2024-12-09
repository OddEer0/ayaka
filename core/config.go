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
		"StartTimeout":    c.StartTimeout / time.Second,
		"GracefulTimeout": c.GracefulTimeout / time.Second,
		"Info":            c.Info,
	})
}

func (a *App) WithConfig(conf *Config) *App {
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
