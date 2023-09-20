package carriers

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

var (
	ErrNoOffersFound         = errors.New("no offers found")
	ErrFailedToApplyOption   = errors.New("failed to apply option")
	ErrFailedToCreateMetrics = errors.New("failed to create metric from provider")
)

type Option func(car *Carriers) error

var Defaults = []Option{
	WithMeter(otel.Meter("github.com/andrewhowdencom/courses.pito/delivery-service/carriers")),
}

// Carriers is a wrapper around all individual carriers to aggregate the results from those carriers
// into a single set of delivery options.
//
// Later it will be extended to include statistics for each carrier.
type Carriers struct {
	// opts are things that modifiy the structs bootstrap, but are later unused.
	opts struct {
		m metric.Meter
	}

	// Metrics are used
	metrics struct {
		queries metric.Int64Counter
	}

	carriers []Carrier
}

// New generates a new set of carriers. There are a series of default options that should be extended
// when this function is used. For example,
//
//	New(append(Defaults, WithCarrier(...))
func New(opts ...Option) (*Carriers, error) {
	c := &Carriers{
		carriers: make([]Carrier, 0),
	}

	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrFailedToApplyOption, err)
		}
	}

	// If there is no meter, add one so we're safe.
	if c.opts.m == nil {
		c.opts.m = noop.NewMeterProvider().Meter("noop")
	}

	// Setup metrics
	var err error
	if c.metrics.queries, err = c.opts.m.Int64Counter("delivery-option.queries"); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFailedToCreateMetrics, err)
	}

	return c, nil
}

// WithCarrier adds a carrier to the carriers primitive
func WithCarrier(nc Carrier) Option {
	return func(c *Carriers) error {
		c.carriers = append(c.carriers, nc)

		return nil
	}
}

// WithMeter applies a specific meter provider to the carriers. Used mostly in testing.
func WithMeter(mp metric.Meter) Option {
	return func(car *Carriers) error {
		car.opts.m = mp

		return nil
	}
}

// Query takes a single package and returns the aggregated results from all delivery providers.
func (c *Carriers) Query(in *Package) ([]*DeliveryOption, error) {

	c.metrics.queries.Add(context.Background(), 1)

	results := []*DeliveryOption{}

	for _, ic := range c.carriers {
		// Here, we do not want to _fail_ the request if a single provider fails. Instead, we just want to return
		// whatever providers are available. Otherwise, we'd be only as available as a the worst downstream provider!
		// However, that creates a dilemma: How do we know when we need to intervene with a provider?
		opts, _ := ic.Query(in)

		results = append(results, opts...)
	}

	if len(results) == 0 {
		return nil, ErrNoOffersFound
	}

	return results, nil
}
