package vwap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlidingWindowVWAPAdd(t *testing.T) {
	t.Parallel()

	newWithAdds := func(windowCapacity int, positions []position) *SlidingWindowVWAP {
		s := NewSlidingWindowVWAP(windowCapacity)

		for _, position := range positions {
			s.Add(position.units, position.unitPrice)
		}

		return s
	}

	for _, tc := range []struct {
		name          string
		with          *SlidingWindowVWAP
		giveUnits     float64
		giveUnitPrice float64
		expected      float64
	}{
		{
			name:          "add_to_empty_5capacity",
			with:          NewSlidingWindowVWAP(5),
			giveUnits:     2.5,
			giveUnitPrice: 1.2,
			expected:      1.2,
		},
		{
			name:          "add_to_empty_1capacity",
			with:          NewSlidingWindowVWAP(1),
			giveUnits:     2.5,
			giveUnitPrice: 1.2,
			expected:      1.2,
		},
		{
			name:          "add_to_zero_value",
			with:          &SlidingWindowVWAP{},
			giveUnits:     2.5,
			giveUnitPrice: 1.2,
			expected:      0,
		},
		{
			name:          "add_to_full_1capacity",
			with:          newWithAdds(1, []position{{units: 10, unitPrice: 5}}),
			giveUnits:     2.5,
			giveUnitPrice: 1.2,
			expected:      1.2,
		},
		{
			name:          "add_to_full_2capacity",
			with:          newWithAdds(2, []position{{units: 1, unitPrice: 1}, {units: 1, unitPrice: 1}}),
			giveUnits:     1,
			giveUnitPrice: 2,
			expected:      1.5, // 3 total price / 2 total units
		},
		{
			name:          "add_reaches_capacity",
			with:          newWithAdds(3, []position{{units: 1, unitPrice: 1}, {units: 2, unitPrice: 1}}),
			giveUnits:     1,
			giveUnitPrice: 2,
			expected:      1.25, // 5 total price / 4 total units
		},
		{
			name:          "add_to_1len_3capacity",
			with:          newWithAdds(3, []position{{units: 1, unitPrice: 1}}),
			giveUnits:     1,
			giveUnitPrice: 2,
			expected:      1.5, // 3 total price / 2 total units
		},
		{
			name: "slide_lots",
			with: newWithAdds(
				3,
				[]position{
					{units: 1, unitPrice: 1},
					{units: 2, unitPrice: 2},
					{units: 3, unitPrice: 3},
					{units: 4, unitPrice: 4},
					{units: 5, unitPrice: 5},
					{units: 6, unitPrice: 6},
					{units: 7, unitPrice: 7},
					{units: 8, unitPrice: 8},
					{units: 9, unitPrice: 9},
					{units: 10, unitPrice: 10},
				}),
			giveUnits:     11,
			giveUnitPrice: 11,
			expected:      10.066666666666666, // 302 total price / 30 total units
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := tc.with.Add(tc.giveUnits, tc.giveUnitPrice)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
