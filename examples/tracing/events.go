package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"tracing.local/examples/setup"
)

func main() {
	tp, err := setup.NewTracerProvider()
	if err != nil {
		log.Fatal(err)
	}

	handle := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())

		span.AddEvent("start sleep", trace.WithAttributes(attribute.String("sleep.duration", "300ms")))
		time.Sleep(time.Millisecond * 300)
		span.AddEvent("end sleep")

		if i := rand.Intn(2); i == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Uh oh! I failed."))
			return
		}

		w.Write([]byte("Hello, #PITO!"))
	})

	instrumented := otelhttp.NewHandler(handle, "root_page", otelhttp.WithTracerProvider(tp))

	http.Handle("/", instrumented)

	log.Fatal(http.ListenAndServe("localhost:8083", nil))
}
