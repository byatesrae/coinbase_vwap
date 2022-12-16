package coinbase

import (
	"context"
	"fmt"
	"log"
)

// MatchesSubscription is created by a Client to manage a subscription to the [Matches Channel].
//
// [Matches Channel]: https://docs.cloud.coinbase.com/exchange/docs/websocket-channels#matches-channel
type MatchesSubscription struct {
	productID ProductID

	conn Conn

	// A read channel, pushed to by the connection read loop
	read chan *MatchResponse

	// A channel to signal the read loop to stop.
	stopReading chan struct{}

	// True if the reading loop has been stopped.
	isReadingStopped bool
}

// newMatchesSubscription creates a new MatchesSubscription. It will first subscribe
// to the Matches Channel for productID over conn (using ctx). If this is successful,
// the read loop is started.
func newMatchesSubscription(ctx context.Context, conn Conn, productID ProductID) (*MatchesSubscription, error) {
	if productID == ProductIDUnknown {
		return nil, fmt.Errorf("productID is required")
	}

	request := SubscribeRequest{
		Type: "subscribe",
		Channels: []SubscribeChannelRequest{
			{
				Name: ChannelNameMatches,
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

// ProductID returns the Product ID for this subscription.
func (m *MatchesSubscription) ProductID() ProductID {
	return m.productID
}

// Read can be used to read from this subscription. Only messages of type "match",
// "last_match" or "error" are read. On connection error, no further messages are
// read.
func (m *MatchesSubscription) Read() <-chan *MatchResponse {
	return m.read
}

// Close can be used to close the subscription.
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
		// Ideally this package wouldn't be logging on a whim like this.
		log.Printf("read loop took to long to exit: %v", ctx.Err())
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
				// if m.read's buffer is full and not being drained, this would
				// block indefinitely.
				m.read <- &MatchResponse{Err: fmt.Errorf("read match: %w", err)}

				// There's no recovery here, if we keep invoking conn.ReadJSON it
				// will eventually panic. Wait for the signal to exit.
				<-m.stopReading

				return
			}

			switch message.Type {
			case MessageTypeError:
				m.read <- &MatchResponse{Err: fmt.Errorf("error message received: %q", message.Message)}
			case MessageTypeLastMatch, MessageTypeMatch:
				m.read <- &MatchResponse{Match: message}
			case MessageTypeSubscriptions:
			default:
				m.read <- &MatchResponse{Err: fmt.Errorf("received unexpected message with type %q", message.Type)}
			}
		}
	}()
}
