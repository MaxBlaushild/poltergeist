// Package pricing implements R-6's formula. Price is computed here, in Go,
// from a persisted slice_result (R-6.2: server authority) — the client only
// ever displays what this package returns, never computes it itself.
// Integer cents are the only thing that crosses this package's boundary in
// either direction; the per-unit rates are inherently fractional (e.g.
// "8.2 cents per gram"), but nothing downstream of Price ever sees a
// fractional cent.
package pricing

import "math"

// Rates are R-6.1's config-driven inputs — [DECIDE] in the requirements
// doc: real values must come from fulfillment quotes, not these defaults
// (see internal/config's REEF_PRICE_* defaults, which are explicitly
// placeholders).
type Rates struct {
	SetupFeeCents             int64
	MaterialRateCentsPerGram  float64
	MachineRateCentsPerMinute float64
	FulfillmentFeeCents       int64
	MarginMultiplier          float64
}

// Price implements R-6.1 exactly:
//
//	price = ceil((setup_fee + weight_g * material_rate + print_minutes *
//	  machine_rate + fulfillment_fee) * margin_multiplier)
//
// weightG and printTimeS must come from a persisted reef_slice_results row
// — never from a client-supplied value or an estimate.
func Price(weightG float64, printTimeS int64, rates Rates) int64 {
	printMinutes := float64(printTimeS) / 60.0
	raw := (float64(rates.SetupFeeCents) +
		weightG*rates.MaterialRateCentsPerGram +
		printMinutes*rates.MachineRateCentsPerMinute +
		float64(rates.FulfillmentFeeCents)) * rates.MarginMultiplier
	return int64(math.Ceil(raw))
}

// ShippingRates are R-6.3's free-shipping-threshold config.
type ShippingRates struct {
	FreeShippingThresholdCents int64
	FlatShippingCents          int64
}

// Shipping returns the shipping cost for a cart subtotal and, when still
// below the free-shipping threshold, how many more cents would reach it
// (R-6.3: "Display remaining-to-threshold in the cart"). At or above the
// threshold, shipping is free and remaining is 0.
func Shipping(subtotalCents int64, rates ShippingRates) (shippingCents int64, remainingToFreeCents int64) {
	if subtotalCents >= rates.FreeShippingThresholdCents {
		return 0, 0
	}
	return rates.FlatShippingCents, rates.FreeShippingThresholdCents - subtotalCents
}
