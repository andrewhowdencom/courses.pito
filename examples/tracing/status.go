package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"tracing.local/examples/setup"
)

func main() {
	tp, err := setup.NewTracerProvider()
	if err != nil {
		log.Fatal(err)
	}

	tracer := tp.Tracer("status.go")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, span := tracer.Start(context.Background(), "serve_root", trace.WithSpanKind(trace.SpanKindServer))

		if i := rand.Intn(2); i == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Uh oh! I failed."))
			span.End()
			return
		}

		w.Write([]byte("Hello, #PITO!"))

		span.End()
	})

	log.Fatal(http.ListenAndServe("localhost:8083", nil))
}
