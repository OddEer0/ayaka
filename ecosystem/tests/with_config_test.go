package grpc_job

import (
	"context"
	ayaka "github.com/OddEer0/ayaka/core"
	"github.com/OddEer0/ayaka/ecosystem"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
	"time"
)

func TestConfigInterceptors(t *testing.T) {
	t.Run("Should correct work parse config from env", func(t *testing.T) {
		t.Run("Should correct work", func(t *testing.T) {
			defer os.Clearenv()

			ctx := context.Background()
			cfg := &ayaka.Config{}
			startTimeout := "5"
			gracefulTimeout := "10"
			err := os.Setenv(ecosystem.EnvAyakaStartTimeout, startTimeout)
			assert.NoError(t, err)
			err = os.Setenv(ecosystem.EnvAyakaGracefulTimeout, gracefulTimeout)
			assert.NoError(t, err)

			conf, err := ecosystem.AdapterParseConfigFromEnv(ctx, cfg)
			assert.NoError(t, err)
			assert.Equal(t, time.Second*5, conf.StartTimeout)
			assert.Equal(t, time.Second*10, conf.GracefulTimeout)
		})

		t.Run("Should correctly handling not set env", func(t *testing.T) {
			defer os.Clearenv()

			ctx := context.Background()
			cfg := &ayaka.Config{}
			startTimeout := "5"
			gracefulTimeout := "10"

			// graceful timeout not set
			err := os.Setenv(ecosystem.EnvAyakaStartTimeout, startTimeout)
			assert.NoError(t, err)

			conf, err := ecosystem.AdapterParseConfigFromEnv(ctx, cfg)
			assert.Error(t, err)
			assert.True(t, strings.Contains(err.Error(), ecosystem.EnvAyakaGracefulTimeout))
			assert.Equal(t, cfg, conf)
			os.Clearenv()

			// start timeout not set
			err = os.Setenv(ecosystem.EnvAyakaGracefulTimeout, gracefulTimeout)
			assert.NoError(t, err)
			conf, err = ecosystem.AdapterParseConfigFromEnv(ctx, cfg)
			assert.Error(t, err)
			assert.True(t, strings.Contains(err.Error(), ecosystem.EnvAyakaStartTimeout))
			assert.Equal(t, cfg, conf)
		})

		t.Run("Should correctly handling set invalid value", func(t *testing.T) {
			defer os.Clearenv()

			ctx := context.Background()
			cfg := &ayaka.Config{}
			startTimeout := "incorrect"
			gracefulTimeout := "incorrect"

			err := os.Setenv(ecosystem.EnvAyakaStartTimeout, startTimeout)
			assert.NoError(t, err)
			err = os.Setenv(ecosystem.EnvAyakaGracefulTimeout, "4")
			assert.NoError(t, err)

			// handle incorrect start timeout
			conf, err := ecosystem.AdapterParseConfigFromEnv(ctx, cfg)
			assert.Error(t, err)
			assert.True(t, strings.Contains(err.Error(), ecosystem.EnvAyakaStartTimeout))
			assert.Equal(t, cfg, conf)

			err = os.Setenv(ecosystem.EnvAyakaStartTimeout, "5")
			assert.NoError(t, err)
			err = os.Setenv(ecosystem.EnvAyakaGracefulTimeout, gracefulTimeout)
			assert.NoError(t, err)

			// handle incorrect graceful timeout
			conf, err = ecosystem.AdapterParseConfigFromEnv(ctx, cfg)
			assert.Error(t, err)
			assert.True(t, strings.Contains(err.Error(), ecosystem.EnvAyakaGracefulTimeout))
			assert.Equal(t, cfg, conf)
		})
	})
}
