package ecosystem

import (
	"context"
	ayaka "github.com/OddEer0/ayaka/core"
	"net/http"
	"net/http/pprof"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	DefaultMonitoringMaxHeaderBytes = 1 << 20
	DefaultMonitoringRequestTimeout = time.Minute
)

type (
	MonitoringJobBuilder[T any] struct {
		address string
		mux     *http.ServeMux
	}

	MonitoringJob[T any] struct {
		address string
		mux     *http.ServeMux
	}
)

func (m *MonitoringJobBuilder[T]) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(&m.address, validation.Required),
	)
}

func (m *MonitoringJobBuilder[T]) Build() (*MonitoringJob[T], error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	if m.mux == nil {
		m.mux = http.NewServeMux()
	}
	return &MonitoringJob[T]{
		address: m.address,
		mux:     m.mux,
	}, nil
}

func (m *MonitoringJob[T]) Init(ctx context.Context, _ T) error {
	app, err := ayaka.AppFromContext[T](ctx)
	if err != nil {
		return errors.Wrap(err, "[MonitoringJob.Init] ayaka.AppFromContext")
	}

	// Register Prometheus metrics handler.
	m.mux.Handle("/metrics", promhttp.Handler())

	// Register pprof handlers
	m.mux.HandleFunc("/debug/pprof/", pprof.Index)
	m.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	m.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	m.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	m.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	err = prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace:   "App",
			Name:        "info",
			Help:        "Application info",
			ConstLabels: prometheus.Labels{"version": app.Info().Version},
		},
		func() float64 { return float64(1) },
	))
	if err != nil {
		return errors.Wrap(err, "[MonitoringJob] prometheus.Register")
	}

	return nil
}

func (m *MonitoringJob[T]) Run(ctx context.Context, container T) error {
	errCh := make(chan error, 1)
	app, err := ayaka.AppFromContext[T](ctx)
	if err != nil {
		return errors.Wrap(err, "[GrpcJob] ayaka.AppFromContext")
	}

	srv := http.Server{
		Addr:           m.address,
		Handler:        m.mux,
		WriteTimeout:   DefaultMonitoringRequestTimeout,
		ReadTimeout:    DefaultMonitoringRequestTimeout,
		MaxHeaderBytes: DefaultMonitoringMaxHeaderBytes,
	}

	go func() {
		app.Logger().Info(ctx, "http monitoring server started...", map[string]any{"address": m.address})
		err = srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		err := srv.Shutdown(ctx)
		if err != nil {
			app.Logger().Warn(ctx, "http monitoring server failed graceful stopped", map[string]any{"address": m.address})
			return errors.Wrap(err, "[MonitoringJob] failed to shutdown http monitoring server")
		}
		app.Logger().Warn(ctx, "http server stopped", map[string]any{"address": m.address})
		return nil
	}
}

func (m *MonitoringJob[T]) Address() string {
	return m.address
}

func (m *MonitoringJobBuilder[T]) Address(addr string) *MonitoringJobBuilder[T] {
	m.address = addr
	return m
}

func (m *MonitoringJobBuilder[T]) Mux(mux *http.ServeMux) *MonitoringJobBuilder[T] {
	m.mux = mux
	return m
}

func NewMonitoringJobBuilder[T any]() *MonitoringJobBuilder[T] {
	return &MonitoringJobBuilder[T]{}
}
