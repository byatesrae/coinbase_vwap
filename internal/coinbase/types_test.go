package coinbase

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchResponseToUnitsAndUnitPrice(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name              string
		with              *MatchResponse
		expectedUnits     float64
		expectedUnitPrice float64
		expectedErr       string
	}{
		{
			name:              "match",
			with:              &MatchResponse{Match: Match{Size: "5.5", Price: "10.1"}},
			expectedUnits:     5.5,
			expectedUnitPrice: 10.1,
		},
		{
			name:        "err",
			with:        &MatchResponse{Err: fmt.Errorf("TestABC")},
			expectedErr: "match response: TestABC",
		},
		{
			name:        "invalid_size",
			with:        &MatchResponse{Match: Match{Size: "abc", Price: "10.1"}},
			expectedErr: "parse units from size: strconv.ParseFloat: parsing \"abc\": invalid syntax",
		},
		{
			name:        "invalid_price",
			with:        &MatchResponse{Match: Match{Size: "5.5", Price: "abc"}},
			expectedErr: "parse unitPrice from price: strconv.ParseFloat: parsing \"abc\": invalid syntax",
		},
		{
			name:        "zero_value",
			with:        &MatchResponse{},
			expectedErr: "parse units from size: strconv.ParseFloat: parsing \"\": invalid syntax",
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actualUnits, actualUnitPrice, err := tc.with.ToUnitsAndUnitPrice()

			assert.Equal(t, tc.expectedUnits, actualUnits, "Units")
			assert.Equal(t, tc.expectedUnitPrice, actualUnitPrice, "Unit price")

			if tc.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
