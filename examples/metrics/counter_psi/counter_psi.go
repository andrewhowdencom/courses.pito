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
	// Common Errors
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

	// Create the counter instrument. A single instrument (system.psi.time) is created, with each resource excpected
	// to be reflected as an attribute, and each delay type the same.
	//
	// This allows queries like:
	//
	// * sum(rate(system_psi_time_microseconds_total{delay="some"}[1m]))
	//
	// Which indicates whether the machine is "loaded somehow". The diagnostic utility of such a query might be
	// limited.
	counter, err := meter.Int64ObservableCounter(
		"system.psi.time",
		metric.WithDescription("The delay (some or full) a task experienced awaiting a resource"),
		metric.WithUnit("us"),
	)

	// The creation of instruments _can fail_. Normally, it is better to initialize the counters first with a "noop"
	// object to avoid the null pointer errors and provide a consistent interface, but here we'll just fatal if the
	// application cannot understand it.
	if err != nil {
		log.Fatal(err)
	}

	// Here, we register each of the collectors to query the state of the relevant `/proc/pressure/<resource>` file,
	// and fetch the appropriate data.
	for _, v := range []string{PSICpu, PSIIO, PSIMemory} {
		_, err := meter.RegisterCallback(PSI(v, counter), counter)

		// As before, while we should handle this, the complexity is skipped cfor the example.
		if err != nil {
			log.Fatal(err)
		}
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
// This is implemented naively, as it is a demonstration and I didn't think about it too much. Implemented as a
// function generator, as the way in which the callback is registered, the function that it receives
//
// See also,
// 1. https://facebookmicrosites.github.io/psi/docs/overview
// 2. https://docs.kernel.org/accounting/psi.html
// 3. https://github.com/google/cadvisor/issues/3052
func PSI(resource string, counter metric.Int64ObservableCounter) func(ctx context.Context, o metric.Observer) error {
	// Define a noop function to return in the case this fails.
	noop := func(ctx context.Context, o metric.Observer) error { return nil }

	// Open the file, and keep it open.
	handle, err := os.Open("/proc/pressure/" + resource)

	if err != nil {
		fmt.Println(ErrFailedToReadPSI, err)
		return noop
	}

	return func(ctx context.Context, o metric.Observer) error {
		// Rewind the handle to the start of the file, so we can iterate through it agian.
		if _, err := handle.Seek(0, 0); err != nil {
			return fmt.Errorf("%w: %s", ErrFailedToReadPSI, err)
		}

		// Create a "scanner" to tokenize and iterate through the file content. THe scanner
		// is created each run, as it is cheap to do. The scanner iterate through "words" (sets of runes)
		// delimited by a whitespace character)
		// https://groups.google.com/g/golang-nuts/c/_eqP4nU4Cjw?pli=1
		sc := bufio.NewScanner(handle)
		sc.Split(bufio.ScanWords)
		delay := "unknown"

		for sc.Scan() {
			// If the word that comes up is either "some" or "full", it indicates we're at the start of a new line.
			if sc.Text() == "some" || sc.Text() == "full" {
				delay = sc.Text()
			}

			// If the word is prefixed with "total=<blah>", than we're at the total ms delayed and we need to
			// capture that data point
			if after, isPresent := strings.CutPrefix(sc.Text(), "total="); isPresent {
				total, err := strconv.ParseInt(after, 10, 64)
				if err != nil {
					return fmt.Errorf("%w: %s", ErrFailedToReadPSI, err)
				}

				// Record the metric
				o.ObserveInt64(counter, total, metric.WithAttributes(
					attribute.String("resource", resource),
					attribute.String("delay", delay),
				))
			}
		}

		return nil
	}

}
