// package money is a utility package providing functions to handle money.
package money

// Money is an amount of cash, represented in the base (non decimal) unit of that currency.
type Money struct {
	Total int64 `json:"total"`

	// The ISO code of the relevant currency. See the standard:
	//
	// * https://www.iso.org/iso-4217-currency-codes.html).
	Currency string `json:"currency"`
}
