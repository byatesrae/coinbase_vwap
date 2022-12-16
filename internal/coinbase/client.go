package coinbase

import (
	"context"
	"fmt"
)

// Client can be used to integrate with the Coinbase API.
type Client struct {
	// The Dialer used to open a connection. If nil, a default is used.
	Dialer Dialer
}

func (c *Client) dialerOrDefault() Dialer {
	if c.Dialer != nil {
		return c.Dialer
	}

	// By default, use gorilla
	return newGorillaWebsocketDialler(nil)
}

// SubscribeToMatchesForProduct will dial a new websocket connection and [Subscribe]
// to the [Matches Channel] for product by ProductID.
//
// [Subscribe]: https://docs.cloud.coinbase.com/exchange/docs/websocket-overview#subscribe
// [Matches Channel]: https://docs.cloud.coinbase.com/exchange/docs/websocket-channels#matches-channel
func (c *Client) SubscribeToMatchesForProduct(ctx context.Context, productID ProductID) (*MatchesSubscription, error) {
	conn, _, err := c.dialerOrDefault().DialContext(ctx, "wss://ws-feed.exchange.coinbase.com", nil)
	if err != nil {
		// If errors.Is(websocket.ErrBadHandshake, err) we could inspect the response
		// for further information that would aid in debugging.
		return nil, fmt.Errorf("dialing coinbase: %w", err)
	}

	return newMatchesSubscription(ctx, conn, productID)
}
