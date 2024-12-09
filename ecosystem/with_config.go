package ecosystem

import (
	"context"
	"os"
	"strconv"
	"time"

	ayaka "github.com/OddEer0/ayaka/core"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/pkg/errors"
)

const (
	EnvAyakaStartTimeout    = "AYAKA_START_TIMEOUT"
	EnvAyakaGracefulTimeout = "AYAKA_GRACEFUL_TIMEOUT"
)

type LibCfg struct {
	StartTimeout    time.Duration  `yaml:"start_timeout"`
	GracefulTimeout time.Duration  `yaml:"graceful_timeout"`
	Info            map[string]any `yaml:"info,omitempty"`
}

func AdapterParseLibConfigFromYaml(path string) ayaka.ConfigInterceptor {
	return func(ctx context.Context, conf *ayaka.Config) (*ayaka.Config, error) {
		_, err := os.Stat(path)
		if err != nil {
			return nil, errors.Wrap(err, "cannot stat config file")
		}

		lCfg := &LibCfg{}
		err = cleanenv.ReadConfig(path, lCfg)
		if err != nil {
			return nil, errors.Wrap(err, "cannot read config file")
		}

		conf.Info = lCfg.Info
		conf.StartTimeout = lCfg.StartTimeout
		conf.GracefulTimeout = lCfg.GracefulTimeout

		return conf, nil
	}
}

func AdapterParseConfigFromEnv(ctx context.Context, cfg *ayaka.Config) (*ayaka.Config, error) {
	startTimeoutEnv := os.Getenv(EnvAyakaStartTimeout)
	if startTimeoutEnv == "" {
		return cfg, errors.Errorf("%s env variable is not set", EnvAyakaStartTimeout)
	}
	startTimeout, err := strconv.Atoi(startTimeoutEnv)
	if err != nil {
		return cfg, errors.Wrapf(err, "bad value %s env variable", EnvAyakaStartTimeout)
	}

	gracefulTimeoutEnv := os.Getenv(EnvAyakaGracefulTimeout)
	if gracefulTimeoutEnv == "" {
		return cfg, errors.Errorf("%s env variable is not set", EnvAyakaGracefulTimeout)
	}
	gracefulTimeout, err := strconv.Atoi(gracefulTimeoutEnv)
	if err != nil {
		return cfg, errors.Wrapf(err, "bad value %s env variable", EnvAyakaGracefulTimeout)
	}

	cfg.StartTimeout = time.Second * time.Duration(startTimeout)
	cfg.GracefulTimeout = time.Second * time.Duration(gracefulTimeout)

	return cfg, nil
}
