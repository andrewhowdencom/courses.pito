package main

import (
	"context"
	"log"

	"pito.local/examples/metrics/setup"
)

func main() {
	// Here, we fetch a new "application level metric provider". This is the thing that we register all metrics against
	// for the entire application.
	//
	// There is convenience function to set a "global" meter provider, but it is omitted here for understanding. See
	//
	// 1. https://opentelemetry.io/docs/specs/otel/metrics/api/#meterprovider
	mp, r, err := setup.NewMetricProvider("counter")

	if err != nil {
		log.Fatal(err)
	}

	// Here, we provide a new "meter" — a library specific provider of metrics instruments. These should be created per
	// library, and are used to give a reference for where these metrics "Came from".
	meter := mp.Meter("pito.local/examples/metrics/counter")

	// Create the counter object. Here, the the only option is "Int64Counter", and allows counting up to
	// 18446744073709551615
	counter, err := meter.Int64Counter("coffees")

	// The creation of instruments _can fail_. Normally, it is better to initialize the counters first with a "noop"
	// object to avoid the null pointer errors and provide a consistent interface, but here we'll just fatal if the
	// application cannot understand it.
	if err != nil {
		log.Fatal(err)
	}

	// Express the internal state change. Here, we're expressing that I drank 3 coffees during the making of this
	// section. I now have a stomach ache.
	counter.Add(context.Background(), 1)
	counter.Add(context.Background(), 1)
	counter.Add(context.Background(), 1)

	// Here, we're asking the metrics reader to "shutdown". While normally it periodically exports the metrics,
	// these examples do not run long enough for an evaluation window to go past. So, we have to export them.
	//
	// It is good practice to shut these down if you have an application lifecycle process regardless.
	if err := r.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}
}
