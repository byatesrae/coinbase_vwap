package coinbase

import (
	"context"
	"fmt"

	"github.com/gorilla/websocket"
)

type Client struct {
	Dialer *websocket.Dialer
}

func (c *Client) dialerOrDefault() *websocket.Dialer {
	if c.Dialer != nil {
		return c.Dialer
	}

	return &websocket.Dialer{}
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
