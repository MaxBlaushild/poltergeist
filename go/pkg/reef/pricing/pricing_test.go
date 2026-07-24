package pricing

import "testing"

func defaultRates() Rates {
	return Rates{
		SetupFeeCents:             300,
		MaterialRateCentsPerGram:  8.0,
		MachineRateCentsPerMinute: 4.0,
		FulfillmentFeeCents:       250,
		MarginMultiplier:          1.8,
	}
}

func TestPrice_MatchesFormulaExactly(t *testing.T) {
	// weight=80g, printTime=90min -> raw = (300 + 80*8 + 90*4 + 250) * 1.8
	//   = (300 + 640 + 360 + 250) * 1.8 = 1550 * 1.8 = 2790.0
	got := Price(80, 90*60, defaultRates())
	if got != 2790 {
		t.Fatalf("Price = %d, want 2790", got)
	}
}

func TestPrice_RoundsFractionalCentsUp(t *testing.T) {
	// Pick a weight that produces a non-integer raw price, and confirm the
	// result is ceil'd rather than truncated or rounded-to-nearest.
	rates := Rates{
		SetupFeeCents:             0,
		MaterialRateCentsPerGram:  1.0 / 3.0, // deliberately produces a repeating fraction
		MachineRateCentsPerMinute: 0,
		FulfillmentFeeCents:       0,
		MarginMultiplier:          1.0,
	}
	got := Price(1, 0, rates) // raw = 1 * (1/3) = 0.333...
	if got != 1 {
		t.Fatalf("Price = %d, want 1 (ceil of 0.333...)", got)
	}
}

func TestPrice_ZeroInputsGiveZero(t *testing.T) {
	rates := Rates{}
	if got := Price(0, 0, rates); got != 0 {
		t.Fatalf("Price = %d, want 0", got)
	}
}

func TestPrice_UsesIntegerMinutesConversion(t *testing.T) {
	// 30 seconds = 0.5 minutes; make sure PrintTimeS is converted to
	// minutes as a float, not truncated to 0 minutes.
	rates := Rates{MachineRateCentsPerMinute: 10, MarginMultiplier: 1}
	got := Price(0, 30, rates) // 0.5 min * 10 cents/min = 5 cents
	if got != 5 {
		t.Fatalf("Price = %d, want 5", got)
	}
}

func TestShipping_FreeAtOrAboveThreshold(t *testing.T) {
	rates := ShippingRates{FreeShippingThresholdCents: 4500, FlatShippingCents: 795}

	shipping, remaining := Shipping(4500, rates)
	if shipping != 0 || remaining != 0 {
		t.Fatalf("at threshold: shipping=%d remaining=%d, want 0,0", shipping, remaining)
	}

	shipping, remaining = Shipping(5000, rates)
	if shipping != 0 || remaining != 0 {
		t.Fatalf("above threshold: shipping=%d remaining=%d, want 0,0", shipping, remaining)
	}
}

func TestShipping_ChargedAndRemainingBelowThreshold(t *testing.T) {
	rates := ShippingRates{FreeShippingThresholdCents: 4500, FlatShippingCents: 795}

	shipping, remaining := Shipping(3000, rates)
	if shipping != 795 {
		t.Fatalf("shipping = %d, want 795", shipping)
	}
	if remaining != 1500 {
		t.Fatalf("remaining = %d, want 1500", remaining)
	}
}

func TestShipping_ZeroCartSubtotal(t *testing.T) {
	rates := ShippingRates{FreeShippingThresholdCents: 4500, FlatShippingCents: 795}
	shipping, remaining := Shipping(0, rates)
	if shipping != 795 || remaining != 4500 {
		t.Fatalf("shipping=%d remaining=%d, want 795,4500", shipping, remaining)
	}
}
