package main

import (
	"log"
	"math/rand"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"tracing.local/examples/setup"
)

func main() {
	tp, err := setup.NewTracerProvider()
	if err != nil {
		log.Fatal(err)
	}

	handle := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
