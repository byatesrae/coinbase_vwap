package vwap

import (
	"github.com/byatesrae/coinbase_vwap/internal/slidingslice"
)

type position struct {
	units     float64
	unitPrice float64
}

func (p position) TotalPrice() float64 {
	return p.units * p.unitPrice
}

type SlidingWindowVWAP struct {
	values     *slidingslice.SlidingSlice[*position]
	totalUnits float64
	totalPrice float64
}

func NewSlidingWindowVWAP(windowCapacity int) *SlidingWindowVWAP {
	return &SlidingWindowVWAP{
		values: slidingslice.New[*position](windowCapacity),
	}
}

func (s *SlidingWindowVWAP) Add(units, unitPrice float64) float64 {
	var poppedValue *position

	// If len == cap, pushing will pop the first. Keep track of it.
	if s.values.Len() == s.values.Cap() {
		poppedValue = s.values.At(0)
	}

	pushedValue := &position{units: units, unitPrice: unitPrice}
	s.values.Push(pushedValue)

	if poppedValue != nil {
		s.totalUnits -= poppedValue.units
		s.totalPrice -= poppedValue.TotalPrice()
	}

	s.totalUnits += pushedValue.units
	s.totalPrice += pushedValue.TotalPrice()

	return s.totalPrice / s.totalUnits
}
