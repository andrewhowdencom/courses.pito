package setup

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrFailedTracerSetup = errors.New("failed to setup tracer")
)

// NewTracerProvider generates a new, basic tracer provider configured to export telemetry in a way that is
// optimized for learning, rather than for production.
//
// See
// 1. https://medium.com/jaegertracing/introducing-native-support-for-opentelemetry-in-jaeger-eb661be8183c
// 2. https://github.com/open-telemetry/opentelemetry-go/blob/main/example/otel-collector/main.go
func NewTracerProvider() (trace.TracerProvider, error) {
	ctx := context.Background()

	// Setup the "application resource". This is the way in which the application is identified in OpenTelemetry,
	// as well as any "core attributes" it has.
	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName("examples"),
	))

	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFailedTracerSetup, err)
	}

	// Create a new "timeout" context, allowing downstream processes to determine whether they should abandon a
	// process.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// Establish a gRPC connection to the place we're sending traces — in this case, jaeger. We expect jaeger to be
	// running on the same comptuer we're running this application, so we'll just point to it directly at
	// localhost:4317
	conn, err := grpc.DialContext(ctx, "localhost:4317",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)

	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFailedTracerSetup, err)
	}

	// Create the exporter, or the thing that will allow this application to write its diagnostic data where it needs to
	// go.
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFailedTracerSetup, err)
	}

	// Create the "processor", or the thing that usually samples or batches spans. This span processor is deliberately
	// simplified, so that it is more useful for our learning and for local development.
	processor := sdktrace.NewSimpleSpanProcessor(exporter)

	// Create the tracer provier, or factory for tracers.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(processor),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		))

	return tp, nil
}
