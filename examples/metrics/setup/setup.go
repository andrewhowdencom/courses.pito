package setup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdk "go.opentelemetry.io/otel/sdk/metric"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var (
	ErrFailedMeterSetup = errors.New("failed to setup meter provider")
)

// NewMeterProvider proview a new, basic meter provider configured to export telemetry in a way that is optimized for
// learning, rather than for production. For more information, see:
//
// 1.https://github.com/open-telemetry/opentelemetry-go/commit/e44ea5cc7fa82cbbc839e69829e7660b6e0c7536
// 2. xamples/tracing/setup/setup.go
// 3.https://github.com/MrAlias/otel-otlp-metric-example/blob/main/main.go
//
// It returns both the meter provider and the reader, as the examples have no long lived persistence (or application
// lifecycle management), and thus need to flush the metrics manually.
func NewMetricProvider(applicationName string) (metric.MeterProvider, sdk.Reader, error) {
	ctx := context.Background()
	hostname, _ := os.Hostname()

	// Setup the "application resource". This is the way in which the application is identified in OpenTelemetry,
	// as well as any "core attributes" it has.
	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName(applicationName),
		semconv.HostName(hostname),
	))

	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s", ErrFailedMeterSetup, err)
	}

	// Create a new "timeout" context, allowing downstream processes to determine whether they should abandon a
	// process.
	ctx, cxl := context.WithTimeout(ctx, time.Second)
	defer cxl()

	// Create the new "exporter", which sends protobuf encoded messages from this application to Prometheus.
	exp, err := otlpmetrichttp.New(
		ctx,
		otlpmetrichttp.WithInsecure(),

		// WithEndpoint & WithURLPath supplies the endpoint to the local instances of Prometheus, which will receive
		// the appropriate metrics
		otlpmetrichttp.WithEndpoint("localhost:9090"),
		otlpmetrichttp.WithURLPath("/api/v1/otlp/v1/metrics"),
	)

	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s", ErrFailedMeterSetup, err)
	}

	// WithReader(sdk.NewPeriodicReader) provides a reader that will periodically collect the state
	// of the metrics and export it via the wire transport to Prometheus. We have scheduled it much more frequently
	// than normal (15s instead of 60s) so we do not need to run the examples for a long time, but typically
	// 60s is fine.
	r := sdk.NewPeriodicReader(exp, sdk.WithInterval(time.Second*15))

	// Create the "meter provider", or the thing that we will use to create our metrics.
	mp := sdk.NewMeterProvider(
		sdk.WithReader(r),

		// Withresource supplies the provided resource to the meter provider.
		sdk.WithResource(res),
	)

	return mp, r, nil
}
