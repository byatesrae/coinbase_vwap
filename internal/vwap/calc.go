// package vwap provides a way to calculate a volume-weighted average price in a
// sliding window of trades.
package vwap

import "github.com/byatesrae/coinbase_vwap/internal/platform/slidingslice"

// position represents a single trade (buy or sell).
type position struct {
	// The number of units traded.
	units float64

	// The price per unit.
	unitPrice float64
}

// TotalPrice is the total price of the trade.
func (p position) TotalPrice() float64 {
	return p.units * p.unitPrice
}

// SlidingWindowVWAP uses a sliding window of positions to calculate a VWAP.
//
// The zero-value of this type has no capacity, and therefore no utility.
type SlidingWindowVWAP struct {
	// all positions
	positions *slidingslice.SlidingSlice[*position]

	// a cumulative total for all units traded in the window.
	totalUnits float64

	// a cumulative price total traded in the window.
	totalPrice float64
}

// NewSlidingWindowVWAP creates a new SlidingWindowVWAP with the specified capacity.
func NewSlidingWindowVWAP(windowCapacity int) *SlidingWindowVWAP {
	return &SlidingWindowVWAP{
		positions: slidingslice.New[*position](windowCapacity),
	}
}

// Add records a new trade (the number of units traded and the price paid per unit)
// in the window. The return value is the new VWAP.
func (s *SlidingWindowVWAP) Add(units, unitPrice float64) float64 {
	if s.positions == nil {
		return 0
	}

	var poppedValue *position

	// If len == cap, pushing will pop the first element. Keep track of it.
	if s.positions.Len() == s.positions.Cap() {
		poppedValue = s.positions.At(0)
	}

	pushedValue := &position{units: units, unitPrice: unitPrice}
	s.positions.Push(pushedValue)

	if poppedValue != nil {
		s.totalUnits -= poppedValue.units
		s.totalPrice -= poppedValue.TotalPrice()
	}

	s.totalUnits += pushedValue.units
	s.totalPrice += pushedValue.TotalPrice()

	return s.totalPrice / s.totalUnits
}
