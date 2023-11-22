package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/metric"
	"pito.local/examples/metrics/setup"
)

const Example = "counter_gc"

var fLimit = flag.String("limit", "1m", "How long to run the program for")

func main() {
	// Calculate how long to run this test for
	flag.Parse()
	dur, err := time.ParseDuration(*fLimit)
	if err != nil {
		log.Fatal(err)
	}

	// Define a variable that will contain memory statsitcs (when read).
	var ms runtime.MemStats

	// Here, we fetch a new "application level metric provider". This is the thing that we register all metrics against
	// for the entire application.
	//
	// There is convenience function to set a "global" meter provider, but it is omitted here for understanding. See
	//
	// 1. https://opentelemetry.io/docs/specs/otel/metrics/api/#meterprovider
	mp, r, err := setup.NewMetricProvider(Example)

	if err != nil {
		log.Fatal(err)
	}

	// Here, we provide a new "meter" — a library specific provider of metrics instruments. These should be created per
	// library, and are used to give a reference for where these metrics "Came from".
	meter := mp.Meter(fmt.Sprintf("pito.local/examples/metrics/%s", Example))

	// Create the counter for the total number of garbage collections completed Follows the conventions defined in the
	// OpenTelemetry documentation:
	//
	// - https://opentelemetry.io/docs/specs/semconv/general/metrics/#general-guidelines
	// - https://pkg.go.dev/runtime#ReadMemStats
	// - https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/instrumentation/runtime/runtime.go#L187
	//
	// Notably, this is an *asynchronous counter*. This means that rather than incrementing inline within our code,
	// we instead schedule a callback to periodically collect the up to date data.
	counter, err := meter.Int64ObservableCounter(
		"process.runtime.go.gc.count",
		metric.WithDescription("The number of completed garbage collection cycles"),
	)

	// The creation of instruments _can fail_. Normally, it is better to initialize the counters first with a "noop"
	// object to avoid the null pointer errors and provide a consistent interface, but here we'll just fatal if the
	// application cannot understand it.
	if err != nil {
		log.Fatal(err)
	}

	// Here, we register the (concurrency-safe) callback to collect the metrics.
	_, err = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		// Here, just for the purposes of this demonstrate, we're logging the collection of metrics so we can see it.
		log.Println("Metrics were collected!")

		// Read the current state of the memory into a variable. Note: This causes a "stop the world" event, in which
		// all goroutines are paused to capture state.
		runtime.ReadMemStats(&ms)

		// Record the latest state with the number of garbage collections.
		o.ObserveInt64(counter, int64(ms.NumGC))

		return nil
	}, counter)

	// As before, while we should handle this, the complexity is skipper for the example.
	if err != nil {
		log.Fatal(err)
	}

	// Tick for 60s, doing "something" that leaves orphaned objects laying around for the GC to clean up.
	ti := time.NewTicker(time.Second * 1)
	i := 0

	for {
		next := <-ti.C

		// Cancel after 130 seconds
		if i >= int(dur.Seconds()) {
			break
		}

		i++

		// Here, we want to invoke the garbage collector manually. By default, Go doesn't invoke the collector
		// so frequently, so if we want some data, we need to create it.
		//
		// See
		// - https://tip.golang.org/doc/gc-guide
		runtime.GC()

		fmt.Println(strconv.Itoa(i) + ": Time is: " + next.String())
	}

	// Here, we're asking the metrics reader to "shutdown". While normally it periodically exports the metrics,
	// these examples do not run long enough for an evaluation window to go past. So, we have to export them.
	//
	// It is good practice to shut these down if you have an application lifecycle process regardless.
	if err := r.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}
}
