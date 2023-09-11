package carriers

import (
	"time"

	"github.com/andrewhowdencom/courses.pito/delivery-service/money"
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
