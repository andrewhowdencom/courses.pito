package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"tracing.local/examples/setup"
)

func main() {
	tp, err := setup.NewTracerProvider()
	if err != nil {
		log.Fatal(err)
	}

	handle := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		fmt.Println(r.Header)

		span := trace.SpanFromContext(r.Context())
		b := baggage.FromContext(r.Context())
		m := b.Member("http.traffic.source")

		span.SetAttributes(attribute.String("http.traffic.source", m.Value()))

		if i := rand.Intn(2); i == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Uh oh! I failed."))
			return
		}

		w.Write([]byte("Hello, #PITO!"))
	})

	instrumented := otelhttp.NewHandler(handle, "serve_b", otelhttp.WithTracerProvider(tp))

	http.Handle("/", instrumented)

	log.Fatal(http.ListenAndServe("localhost:8083", nil))
}
