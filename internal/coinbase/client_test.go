package coinbase

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscribeToMatchesForProduct(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		c := Client{
			Dialer: &DialerMock{
				DialContextFunc: func(_ context.Context, _ string, _ http.Header) (Conn, *http.Response, error) {
					return &ConnMock{
						WriteJSONFunc: func(v interface{}) error { return nil },
						CloseFunc:     func() error { return nil },
						ReadJSONFunc: func(v interface{}) error {
							return fmt.Errorf("TestABC")
						},
					}, nil, nil
				},
			},
		}

		ms, err := c.SubscribeToMatchesForProduct(context.Background(), ProductIDBtcUsd)
		if !assert.NoError(t, err, "Subscribe err") {
			return
		}

		assert.NoError(t, ms.Close(context.Background()), "Close")
	})

	for _, tc := range []struct {
		name          string
		with          *Client
		giveCtx       context.Context
		giveProductID ProductID
		expectedErr   string
	}{
		{
			name: "dialer_err",
			with: &Client{Dialer: &DialerMock{
				DialContextFunc: func(_ context.Context, _ string, _ http.Header) (Conn, *http.Response, error) {
					return nil, nil, fmt.Errorf("TestABC")
				},
			}},
			giveCtx:       context.Background(),
			giveProductID: ProductIDEthUsd,
			expectedErr:   "dialing coinbase: TestABC",
		},
		{
			name: "missing_productid",
			with: &Client{Dialer: &DialerMock{
				DialContextFunc: func(_ context.Context, _ string, _ http.Header) (Conn, *http.Response, error) {
					return &ConnMock{}, nil, nil
				},
			}},
			giveCtx:       context.Background(),
			giveProductID: ProductIDUnknown,
			expectedErr:   "productID is required",
		},
		{
			name: "conn_write_json_err",
			with: &Client{Dialer: &DialerMock{
				DialContextFunc: func(_ context.Context, _ string, _ http.Header) (Conn, *http.Response, error) {
					return &ConnMock{
						WriteJSONFunc: func(v interface{}) error { return fmt.Errorf("TestABC") },
					}, nil, nil
				},
			}},
			giveCtx:       context.Background(),
			giveProductID: ProductIDBtcUsd,
			expectedErr:   "subscribing to Matches channel for product BTC-USD: TestABC",
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, actualErr := tc.with.SubscribeToMatchesForProduct(tc.giveCtx, tc.giveProductID)

			assert.Nil(t, actual, "Actual")
			assert.EqualError(t, actualErr, tc.expectedErr, "Actual err")
		})
	}
}
