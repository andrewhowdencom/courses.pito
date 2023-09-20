package carriers

import (
	"errors"
	"time"

	"github.com/andrewhowdencom/courses.pito/delivery-service/money"
)

var (
	ErrNoOffersFound = errors.New("no offers found")
)

// Carrier is the interface that all carriers must meet. It ensure that we can provide a standard set of
// information, and get an appropriate response.
type Carrier interface {
	// Query allows a provider to return a list of possible delivery options, or an error if there is a failure
	// in some way to query the service.
	Query(*Package) ([]*DeliveryOption, error)
}

// Package is a request for a delivery options.
type Package struct {
	// The distance between two points, measured in milimeters
	Width, Height, Depth int64

	// The weight of an object, measured in grams.
	Weight int64
}

// DeliveryOption is an option that can be booked for a delivery.
type DeliveryOption struct {
	// The provider that expects to fulfil this method
	Provider string `json:"provider"`

	// The cost of the delivery option, should it be booked
	Cost *money.Money `json:"cost"`

	// The estimated arrival (within 6 hours) that the package will be delivered.
	Arrival time.Time `json:"arrival"`
}

// Carriers is a wrapper around all individual carriers to aggregate the results from those carriers
// into a single set of delivery options.
//
// Later it will be extended to include statistics for each carrier.
type Carriers struct {
	Carriers []Carrier
}

// Query takes a single package and returns the aggregated results from all delivery providers.
func (c *Carriers) Query(in *Package) ([]*DeliveryOption, error) {
	results := []*DeliveryOption{}

	for _, ic := range c.Carriers {
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

// TODO: Statistic on how many carriers there are.
