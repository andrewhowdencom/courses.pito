package main

import (
	"context"
	"log"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"tracing.local/examples/setup"
)

func main() {
	tp, err := setup.NewTracerProvider()
	if err != nil {
		log.Fatal(err)
	}

	tracer := tp.Tracer("context_a.go")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx, srvSpan := tracer.Start(context.Background(), "serve_a", trace.WithSpanKind(trace.SpanKindServer))
		defer srvSpan.End()

		resp, err := otelhttp.Get(ctx, "http://localhost:8083")

		if err != nil {
			srvSpan.SetStatus(codes.Error, err.Error())
			return
		}

		resp.Body.Close()

		if resp.StatusCode > 499 {
			srvSpan.SetStatus(codes.Error, "status code too high")
			w.WriteHeader(resp.StatusCode)
		}
	})

	log.Fatal(http.ListenAndServe("localhost:8084", nil))
}
