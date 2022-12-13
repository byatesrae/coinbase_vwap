package coinbase

import (
	"context"
	"fmt"
)

// MatchesSubscription is created by a Client to manage a subscription to the [Matches Channel].
//
// [Matches Channel]: https://docs.cloud.coinbase.com/exchange/docs/websocket-channels#matches-channel
type MatchesSubscription struct {
	productID ProductID

	conn Conn
	read chan *MatchResponse // Read channel, pushed to by the connection read loop

	stopReading      chan struct{}
	isReadingStopped bool
}

func newMatchesSubscription(ctx context.Context, conn Conn, productID ProductID) (*MatchesSubscription, error) {
	if productID == ProductIDUnknown {
		return nil, fmt.Errorf("productID is required")
	}

	request := subscribeRequest{
		Type: "subscribe",
		Channels: []subscribeChannelRequest{
			{
				Name: channelNameMatches,
				ProductIDs: []ProductID{
					productID,
				},
			},
		},
	}
	if err := conn.WriteJSON(request); err != nil {
		return nil, fmt.Errorf("subscribing to Matches channel for product %s: %w", productID, err)
	}

	m := &MatchesSubscription{
		productID:   productID,
		conn:        conn,
		read:        make(chan *MatchResponse, 10),
		stopReading: make(chan struct{}),
	}

	m.startReading()

	return m, nil
}

func (m *MatchesSubscription) ProductID() ProductID {
	return m.productID
}

// Read can be used to read from this subscription. Only messages of type "match"
// or "last_match" are read. On error, no further messages are read.
func (m *MatchesSubscription) Read() <-chan *MatchResponse {
	return m.read
}

func (m *MatchesSubscription) Close(ctx context.Context) error {
	stopReadingDone := make(chan struct{})
	go func() {
		m.signalStopReading()

		close(stopReadingDone)
	}()

	ctxDoneBeforeReadStopped := false

	// Wait for ctx or read loop to exit gracefully.
	select {
	case <-ctx.Done():
		ctxDoneBeforeReadStopped = true
	case <-stopReadingDone:
	}

	// If the read loop is blocked, this will unblock it.
	if err := m.conn.Close(); err != nil {
		return fmt.Errorf("close connection: %w", err)
	}

	if ctxDoneBeforeReadStopped {
		return fmt.Errorf("read loop took to long to exit: %w", ctx.Err())
	}

	return nil
}

func (m *MatchesSubscription) signalStopReading() {
	if !m.isReadingStopped {
		m.stopReading <- struct{}{} // blocks until the read loop has exited.

		m.isReadingStopped = true

		close(m.stopReading)
	}
}

func (m *MatchesSubscription) startReading() {
	go func() {
		defer func() {
			close(m.read)
		}()

		for {
			select {
			case <-m.stopReading:
				return
			default:
			}

			message := Match{}
			if err := m.conn.ReadJSON(&message); err != nil {
				m.read <- &MatchResponse{Err: fmt.Errorf("read match: %w", err)}

				// There's no recovery here, if we keep invoking conn.ReadJSON it
				// will eventually panic. Wait for the signal to exit.
				<-m.stopReading

				return
			}

			switch message.Type {
			case "error":
				m.read <- &MatchResponse{Err: fmt.Errorf("error message received: %q", message.Message)}
			case "last_match", "match":
				m.read <- &MatchResponse{Match: message}
			case "subscriptions":
			default:
				m.read <- &MatchResponse{Err: fmt.Errorf("received unexpected message with type %q", message.Type)}
			}
		}
	}()
}
