package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"tracing.local/examples/setup"
)

func main() {
	tp, err := setup.NewTracerProvider()
	if err != nil {
		log.Fatal(err)
	}

	tracer := tp.Tracer("attributes.go")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, span := tracer.Start(context.Background(), "serve_root", trace.WithSpanKind(trace.SpanKindServer))

		span.SetAttributes(semconv.HTTPMethod(r.Method))
		span.SetAttributes(attribute.String("http.client.user_agent", r.UserAgent()))
		span.SetAttributes(attribute.String("http.host", r.Host))

		if i := rand.Intn(2); i == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Uh oh! I failed."))
			span.SetStatus(codes.Error, "Random number was "+strconv.Itoa(i))
			span.End()
			return
		}

		w.Write([]byte("Hello, #PITO!"))

		span.End()
	})

	log.Fatal(http.ListenAndServe("localhost:8083", nil))
}
