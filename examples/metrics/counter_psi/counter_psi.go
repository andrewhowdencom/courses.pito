package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"pito.local/examples/metrics/setup"
)

const Example = "counter_psi"

const (
	PSICpu    = "cpu"
	PSIMemory = "memory"
	PSIIO     = "io"
)

var fLimit = flag.String("limit", "1m", "How long to run the program for")

var (
	ErrFailedToReadPSI = errors.New("failed to read PSI")
)

func main() {
	// Calculate how long to run this test for
	flag.Parse()
	dur, err := time.ParseDuration(*fLimit)
	if err != nil {
		log.Fatal(err)
	}

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

	// Create a series of metrics for each kind of "pressure".
	counters := map[string]metric.Int64ObservableCounter{}
	for _, k := range []string{PSICpu, PSIIO, PSIMemory} {
		var err error
		counters[k], err = meter.Int64ObservableCounter(
			fmt.Sprintf("system.psi.%s.time", k),
			metric.WithDescription(fmt.Sprintf("The total amount of time spent awaiting the %s resource", k)),
			metric.WithUnit("us"),
		)

		if err != nil {
			log.Fatal(err)
		}
	}

	// The creation of instruments _can fail_. Normally, it is better to initialize the counters first with a "noop"
	// object to avoid the null pointer errors and provide a consistent interface, but here we'll just fatal if the
	// application cannot understand it.
	if err != nil {
		log.Fatal(err)
	}

	// Here, we register the (concurrency-safe) callback to collect the metrics.
	_, err = meter.RegisterCallback(
		PSI(counters[PSICpu], counters[PSIIO], counters[PSIMemory]),
		counters[PSICpu],
		counters[PSIIO],
		counters[PSIMemory],
	)

	// As before, while we should handle this, the complexity is skipper for the example.
	if err != nil {
		log.Fatal(err)
	}

	// Wait for the program to run for a little while, collecting metrics.
	<-time.After(dur)

	// Here, we're asking the metrics reader to "shutdown". While normally it periodically exports the metrics,
	// these examples do not run long enough for an evaluation window to go past. So, we have to export them.
	//
	// It is good practice to shut these down if you have an application lifecycle process regardless.
	if err := r.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}
}

// PSI runs as a metric observer and queries the PSI based on information exposed via the proc filesystem.
// The filesystem exposes this via:
//
//	/proc/pressure/{cpu,memory,io}
//
// It exposes several lines in each file in the format:
//
//	 	some avg10=0.00 avg60=0.00 avg300=0.00 total=473806492
//			full avg10=0.00 avg60=0.00 avg300=0.00 total=0
//
// In which
//
//		some: indicates time in which at least some tasks were awaiting the resource (cpu, memory or IO)
//	    full: indicates the time in which all tasks were awaiting the resource (cpu, memory or IO)
//
// There are three aggreagetes and one total exported (in µs), where the aggregates are 10, 60 and 300 seconds
// expressed as avg10, avg60 and avg300 respectively)
//
// This is implemented naively, as it is a demonstration and I didn't think about it too much.
//
// See also,
// 1. https://facebookmicrosites.github.io/psi/docs/overview
// 2. https://docs.kernel.org/accounting/psi.html
// 3. https://github.com/google/cadvisor/issues/3052
func PSI(cpu, io, memory metric.Int64ObservableCounter) func(ctx context.Context, o metric.Observer) error {
	lookup := map[string]metric.Int64ObservableCounter{
		"cpu":    cpu,
		"memory": memory,
		"io":     io,
	}

	return func(ctx context.Context, o metric.Observer) error {
		// Iterate through all of the resource types (cpu, memory, io)
		for _, r := range []string{"cpu", "memory", "io"} {
			// Open t
			fh, err := os.Open("/proc/pressure/" + r)
			if err != nil {
				log.Println(err)
				return fmt.Errorf("%w: %s", ErrFailedToReadPSI, err)
			}

			reader := bufio.NewScanner(fh)
			reader.Split(bufio.ScanLines)
			for reader.Scan() {
				// Iterate through each field.
				fields := strings.Split(reader.Text(), " ")
				t := "unknown"
				var v int64
				for _, f := range fields {
					// The first field gives the type of
					if f == "some" || f == "full" {
						t = f
					}

					if strTime, found := strings.CutPrefix(f, "total="); found {
						v, err = strconv.ParseInt(strTime, 10, 64)
						if err != nil {

							log.Println(err)
							return fmt.Errorf("%w: %s", ErrFailedToReadPSI, err)
						}
						o.ObserveInt64(lookup[r], v, metric.WithAttributes(
							attribute.Key("type").String(t),
						))
					}
				}
			}

		}

		return nil
	}

}
