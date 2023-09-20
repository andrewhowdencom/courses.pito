package telemetry

import (
	"errors"
	"fmt"
	"net/http"

	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"

	"go.opentelemetry.io/otel/sdk/metric"
)

var Log *slog.Logger

var (
	ErrFailedSetup    = errors.New("failed to setup telemetry")
	ErrFailedExporter = errors.New("failed to setup exporter")
)

// WithReader allows supplying different metric "readers" to the OpenTelemetry infrastructure. Readers allow
// exporting the data to different "sinks".
type WithReader func() (metric.Reader, error)

// WithPrometheusHTTP sets up the promethus exporter, as well as starts a HTTP instance on the specified address / port
// with the "/metrics" endpoint exposed.
func WithPrometheusHTTP(listenOn string) func() (metric.Reader, error) {
	promRegistry := prometheus.NewRegistry()
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.InstrumentMetricHandler(
		promRegistry, promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{}),
	))

	// Here, unfortunately there's not a lot we can do to determine whether the server was successful. Listening to a
	// socket requires a blocking goroutine. Given this, we simply log if there's a failure to instantiate metrics.
	go func() {
		srv := http.Server{
			Addr:    listenOn,
			Handler: mux,
		}

		if err := srv.ListenAndServe(); err != nil {
			Log.Error("failed to start metrics handler: %s", err)
		}
	}()

	return func() (metric.Reader, error) {
		exporter, err := otelprom.New(
			otelprom.WithRegisterer(promRegistry),
		)

		// If we fail to setup the registry or exporter (somehow), there's no point in bootstrapping the telemetry. We
		// thus return an error to indicate the failure.
		if err != nil {
			return nil, err
		}

		return exporter, nil
	}
}

// SetupOTelMetrics sets up the OpenTelemetry metrics provider with the Prometheus exporter. This allows using the in-process
// OpenTelemetry APIs, but exports them in a way that is easy to hook up to graph visualization software.
//
// The objects it bootstraps are global, largely for convenience.
//
// See:
// 1. https://github.com/open-telemetry/opentelemetry-go/blob/main/example/prometheus/main.go
func SetupOTelMetrics(exporters ...func() (metric.Reader, error)) error {
	opts := []metric.Option{}
	for _, f := range exporters {
		reader, err := f()

		if err != nil {
			return fmt.Errorf("%w: %s", ErrFailedExporter, err)
		}

		opts = append(opts, metric.WithReader(reader))
	}

	// Create the "meter provider" — the thing that will be used to create metrics.
	provider := metric.NewMeterProvider(opts...)

	// Set the global "meter provider". This can be queried on a per package basis as the default "meter provider",
	// and overridden in specific cases where code needs to be tested.
	otel.SetMeterProvider(provider)

	return nil
}
