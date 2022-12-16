package coinbase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchesSubscriptionRead(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		readMatch  *Match
		readErr    error
		expected   *MatchResponse
		expectedOk bool
	}{
		{
			name:       "read_match",
			readMatch:  &Match{Type: MessageTypeMatch},
			expected:   &MatchResponse{Match: Match{Type: MessageTypeMatch}},
			expectedOk: true,
		},
		{
			name:       "read_last_match",
			readMatch:  &Match{Type: MessageTypeLastMatch},
			expected:   &MatchResponse{Match: Match{Type: MessageTypeLastMatch}},
			expectedOk: true,
		},
		{
			name:      "read_subscriptions",
			readMatch: &Match{Type: MessageTypeSubscriptions},
		},
		{
			name:       "read_error",
			readMatch:  &Match{Type: MessageTypeError, Message: "TestABC"},
			expected:   &MatchResponse{Err: fmt.Errorf("error message received: \"TestABC\"")},
			expectedOk: true,
		},
		{
			name:       "read_unknown",
			readMatch:  &Match{Type: MessageTypeUnknown},
			expected:   &MatchResponse{Err: fmt.Errorf("received unexpected message with type \"\"")},
			expectedOk: true,
		},
		{
			name:       "err_reading",
			readErr:    fmt.Errorf("TestABC"),
			expected:   &MatchResponse{Err: fmt.Errorf("read match: %w", fmt.Errorf("TestABC"))},
			expectedOk: true,
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup

			ctx, ctxCancel := context.WithCancel(context.Background())
			t.Cleanup(ctxCancel)

			ms := newMatchesSubscriptionWithNext(ctx, t, nil, nil)
			t.Cleanup(func() {
				ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second)
				defer ctxCancel()

				ms.fillAndDrainRead(ctx, t) // Make sure the read isn't blocked

				if err := ms.matchesSubscription.Close(ctx); err != nil {
					t.Errorf("Failed to close test MatchesSubscription: %v", err)
				}
			})

			ms.readIn <- &matchesNext{match: tc.readMatch, err: tc.readErr}

			// Do

			var actual *MatchResponse
			var actualOk bool

			select {
			case actual, actualOk = <-ms.matchesSubscription.Read():
				break
			case <-time.NewTimer(time.Millisecond * 300).C:
				// Need to give the above read a chance first, but this is prone
				// to flakiness.
				break
			}

			// Assert

			assert.Equal(t, tc.expected, actual, "Actual")
			assert.Equal(t, tc.expectedOk, actualOk, "Actual Ok")
		})
	}
}

func TestMatchesSubscriptionClose(t *testing.T) {
	t.Parallel()

	t.Run("unblocked_reading_closes", func(t *testing.T) {
		t.Parallel()

		// Setup

		ctx, ctxCancel := context.WithCancel(context.Background())
		t.Cleanup(ctxCancel)

		ms := newMatchesSubscriptionWithNext(ctx, t, nil, nil)

		ms.fillAndDrainRead(ctx, t)

		// Do

		closeCtx, closeCtxCancel := context.WithTimeout(context.Background(), time.Millisecond*300)
		t.Cleanup(closeCtxCancel)

		err := ms.matchesSubscription.Close(closeCtx)

		// Assert

		assert.NoError(t, err, "Close err")
	})

	t.Run("unblocked_reading_conn_close_err", func(t *testing.T) {
		t.Parallel()

		// Setup

		ctx, ctxCancel := context.WithCancel(context.Background())
		t.Cleanup(ctxCancel)

		ms := newMatchesSubscriptionWithNext(ctx, t, nil, func() error { return fmt.Errorf("TestABC") })

		ms.fillAndDrainRead(ctx, t)

		// Do

		closeCtx, closeCtxCancel := context.WithTimeout(context.Background(), time.Millisecond*300)
		t.Cleanup(closeCtxCancel)

		err := ms.matchesSubscription.Close(closeCtx)

		// Assert

		assert.EqualError(t, err, "close connection: TestABC", "Close err")
	})

	t.Run("blocked_reading_closes", func(t *testing.T) {
		t.Parallel()

		// Setup

		ctx, ctxCancel := context.WithCancel(context.Background())
		t.Cleanup(ctxCancel)

		ms := newMatchesSubscriptionWithNext(ctx, t, nil, nil)

		ms.readIn <- &matchesNext{match: &Match{Type: MessageTypeMatch}} // Block until we know the read loop is running.

		// Do

		closeCtx, closeCtxCancel := context.WithTimeout(context.Background(), time.Millisecond*300)
		t.Cleanup(closeCtxCancel)

		err := ms.matchesSubscription.Close(closeCtx)

		// Assert

		assert.NoError(t, err, "Close err")
	})

	t.Run("blocked_reading_conn_close_err", func(t *testing.T) {
		t.Parallel()

		// Setup

		ctx, ctxCancel := context.WithCancel(context.Background())
		t.Cleanup(ctxCancel)

		ms := newMatchesSubscriptionWithNext(ctx, t, nil, func() error { return fmt.Errorf("TestABC") })

		ms.readIn <- &matchesNext{match: &Match{Type: MessageTypeMatch}} // Block until we know the read loop is running.

		// Do

		closeCtx, closeCtxCancel := context.WithTimeout(context.Background(), time.Millisecond*300)
		t.Cleanup(closeCtxCancel)

		err := ms.matchesSubscription.Close(closeCtx)

		// Assert

		assert.EqualError(t, err, "close connection: TestABC", "Close err")
	})
}

// matchesNext is used by matchesSubscriptionWithNext to return a Match or err from
// the MatchesSubscription.Conn's ReadJSON method.
type matchesNext struct {
	match *Match
	err   error
}

// matchesSubscriptionWithNext wraps a MatchesSubscription and provides some helper
// functionality for testing.
type matchesSubscriptionWithNext struct {
	matchesSubscription *MatchesSubscription

	// Pushing onto this will return a Match or err from the MatchesSubscription.Conn's
	// ReadJSON method.
	readIn chan<- *matchesNext
}

// newMatchesSubscriptionWithNext creates a matchesSubscriptionWithNext.
//
//   - ctx - can be used to signal that the push loop should exit. See field matchesSubscriptionWithNext.readIn.
//   - writeJSONFunc- override for MatchesSubscription.Conn's WriteJSON method. If nil, a default is used.
//   - closeFunc - override for MatchesSubscription.Conn's Close method. If nil, a default is used.
func newMatchesSubscriptionWithNext(ctx context.Context, t *testing.T, writeJSONFunc func(v interface{}) error, closeFunc func() error) *matchesSubscriptionWithNext {
	t.Helper()

	if writeJSONFunc == nil {
		writeJSONFunc = func(v interface{}) error { return nil }
	}

	if closeFunc == nil {
		closeFunc = func() error { return nil }
	}

	read := make(chan *matchesNext)

	matchesSubscription, err := newMatchesSubscription(
		context.Background(), // This shouldn't matter given the tests relying on this method.
		&ConnMock{
			WriteJSONFunc: writeJSONFunc,
			ReadJSONFunc: func(v interface{}) error {
				require.IsType(t, (*Match)(nil), v, "")
				matchIn := v.(*Match)

				select {
				case <-ctx.Done():
					return fmt.Errorf("matchesSubscriptionWithNext ctx expired: %w", ctx.Err())
				case matchOut, ok := <-read:
					if !ok {
						break
					}

					if matchOut.err != nil {
						return matchOut.err
					}

					*matchIn = *matchOut.match
				}

				return nil
			},
			CloseFunc: closeFunc,
		},
		ProductIDBtcUsd, // This shouldn't matter given the tests relying on this method.
	)
	require.NoError(t, err, "create newMatchesSubscriptionWithNext")

	return &matchesSubscriptionWithNext{matchesSubscription: matchesSubscription, readIn: read}
}

// fillAndDrainRead will fill and drain the MatchesSubscription's Read channel. This
// is useful in simulating a read loop not blocked on MatchesSubscription.Conn's
// ReadJSON method.
func (m *matchesSubscriptionWithNext) fillAndDrainRead(ctx context.Context, t *testing.T) {
	t.Helper()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				m.readIn <- &matchesNext{match: &Match{Type: MessageTypeMatch}}
			}
		}
	}()

	go func() {
		for {
			if _, ok := <-m.matchesSubscription.Read(); ok != true {
				return
			}
		}
	}()
}
