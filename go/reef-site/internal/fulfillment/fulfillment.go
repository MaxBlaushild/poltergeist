// Package fulfillment implements R-7's FulfillmentAdapter interface. Only
// ManualAdapter (R-7.2) exists in v1 — SlantAdapter (R-7.3) is explicitly
// deferred to v1.1 in the requirements doc ("do not integrate until sample
// parts have been ordered and inspected"), so it is not implemented here,
// only planned for behind the same interface.
package fulfillment

import "context"

type Status string

const (
	StatusSubmitted Status = "submitted"
	StatusUnknown   Status = "unknown"
)

type OrderItem struct {
	ProductSlug string
	VariantKey  string // empty for configurable products
	Quantity    int
	// STLKey is set for configurable products (from the winning
	// reef_slice_results row) and empty for fixed SKUs, which the operator
	// prints from their own on-hand STL files per product/variant rather
	// than a generated one.
	STLKey string
}

type Order struct {
	OrderToken      string
	CustomerEmail   string
	ShippingName    string
	ShippingLine1   string
	ShippingLine2   string
	ShippingCity    string
	ShippingState   string
	ShippingZip     string
	ShippingCountry string
	Items           []OrderItem
}

// Adapter is R-7.1's FulfillmentAdapter, named to avoid stuttering
// (fulfillment.Adapter rather than fulfillment.FulfillmentAdapter) — the
// requirements doc's Go snippet names it FulfillmentAdapter because it's
// shown outside any package; within package fulfillment that prefix is
// redundant.
type Adapter interface {
	SubmitOrder(ctx context.Context, o Order) (externalID string, err error)
	GetStatus(ctx context.Context, externalID string) (Status, error)
}
