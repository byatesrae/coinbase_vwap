package main

//go:generate moq -pkg main -out moq_test.go ./../../internal/coinbase/. Conn Dialer

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/byatesrae/coinbase_vwap/internal/coinbase"
)

func TestSubscribesReadsAndExits(t *testing.T) {
	t.Parallel()

	// Setup

	ctx, ctxCancel := context.WithCancel(context.Background())
	t.Cleanup(ctxCancel)

	wgForAllSubscriptionReads := sync.WaitGroup{}
	wgForAllSubscriptionReads.Add(3) // Number of expected subscriptions

	coinbaseClient := &coinbase.Client{
		Dialer: newDialerFake(ctx, t, &wgForAllSubscriptionReads, []*matchesIn{
			{
				matchType: coinbase.MessageTypeMatch,
				size:      "1",
				price:     "2",
			},
			{
				matchType: coinbase.MessageTypeMatch,
				size:      "2",
				price:     "4",
			},
			{
				matchType: coinbase.MessageTypeMatch,
				size:      "3",
				price:     "6",
			},
		}),
	}

	sbWithMutex := stringBuilderMutex{}

	interrupt := make(chan os.Signal, 1)

	exited := make(chan struct{})

	// Do

	go func() { // Run app
		err := runApp(coinbaseClient, &sbWithMutex, interrupt)

		assert.NoError(t, err, "runApp error.")

		close(exited)
	}()

	wgForAllSubscriptionReads.Wait() // Wait for all expected match reads

	interrupt <- os.Interrupt // Interrupt the app

	<-exited // Wait for app to exit

	// Assert

	output := sbWithMutex.sb.String()
	outputLines := strings.Split(output, "\n")

	require.Len(t, outputLines, 13, "Number of lines outputted.")
	require.Empty(t, outputLines[12], "Last line outputted.")

	outputLines = outputLines[:12]
	assert.Contains(t, outputLines, "\"ETH-USD\": 6")
	assert.Contains(t, outputLines, "\"ETH-USD\": 5.2")
	assert.Contains(t, outputLines, "\"ETH-USD\": 4.666666666666667")
	assert.Contains(t, outputLines, "\"ETH-BTC\": 6")
	assert.Contains(t, outputLines, "\"ETH-BTC\": 5.2")
	assert.Contains(t, outputLines, "\"ETH-BTC\": 4.666666666666667")
	assert.Contains(t, outputLines, "\"BTC-USD\": 6")
	assert.Contains(t, outputLines, "\"BTC-USD\": 5.2")
	assert.Contains(t, outputLines, "\"BTC-USD\": 4.666666666666667")
	assert.Contains(t, outputLines, "\"ETH-USD\" ERROR: match response: read match: test, close called during ReadJSON")
	assert.Contains(t, outputLines, "\"ETH-BTC\" ERROR: match response: read match: test, close called during ReadJSON")
	assert.Contains(t, outputLines, "\"BTC-USD\" ERROR: match response: read match: test, close called during ReadJSON")
}

// stringBuilderMutex wraps a stringbuilder and implements io.writer with a mutex.
type stringBuilderMutex struct {
	sb strings.Builder
	mu sync.Mutex
}

func (s *stringBuilderMutex) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.sb.Write(p)
}

// matchesIn represents matches read from the Dialer fake (see newDialerFake).
type matchesIn struct {
	matchType coinbase.MessageType
	size      string
	price     string
	message   string
}

func (m *matchesIn) toMatch(productID coinbase.ProductID) *coinbase.Match {
	return &coinbase.Match{
		Type:      m.matchType,
		ProductID: productID,
		Size:      m.size,
		Price:     m.price,
		Message:   m.message,
	}
}

// writeJSONRequestToProductID will return the productID of writeJSONRequest, assuming
// it is of type coinbase.SubscribeRequest and is subscribing to the matches channel
// for one productID.
func writeJSONRequestToProductID(writeJSONRequest interface{}) coinbase.ProductID {
	subscribeRequest, ok := writeJSONRequest.(coinbase.SubscribeRequest)

	if !ok ||
		len(subscribeRequest.ProductIDs) != 0 ||
		len(subscribeRequest.Channels) != 1 ||
		subscribeRequest.Channels[0].Name != coinbase.ChannelNameMatches ||
		len(subscribeRequest.Channels[0].ProductIDs) != 1 {
		return coinbase.ProductIDUnknown
	}

	return subscribeRequest.Channels[0].ProductIDs[0]
}

// newDialerFake creates a Dialer fake. It behaves under the assumption that connections
// are created only for Coinbase Subscriptions to the Matches channel for one product.
// This assumption is not validated. It will then fake matches being read from the
// connection, one for each matchesIn, before waiting to be closed. After all reads
// the wg is "Done"'d.
func newDialerFake(ctx context.Context, t *testing.T, wg *sync.WaitGroup, matchesIn []*matchesIn) *DialerMock {
	return &DialerMock{
		DialContextFunc: func(_ context.Context, _ string, _ http.Header) (coinbase.Conn, *http.Response, error) {
			subscribedProductID := coinbase.ProductIDUnknown
			subscribedProductIDMu := sync.Mutex{}

			readCountRemaining := int32(len(matchesIn))

			closed := make(chan struct{})

			return &ConnMock{
				// Tries to set subscribedProductID.
				WriteJSONFunc: func(v interface{}) error {
					select {
					case <-ctx.Done():
						return fmt.Errorf("test ctx is done: %w", ctx.Err())
					case <-closed:
						return fmt.Errorf("test, WriteJSON called after close")
					default:
					}

					productID := writeJSONRequestToProductID(v)
					if productID == coinbase.ProductIDUnknown {
						return nil
					}

					subscribedProductIDMu.Lock()
					subscribedProductID = productID
					subscribedProductIDMu.Unlock()

					return nil
				},

				// Tries to read a match from matchesIn.
				ReadJSONFunc: func(v interface{}) error {
					select {
					case <-ctx.Done():
						return fmt.Errorf("test ctx is done: %w", ctx.Err())
					case <-closed:
						return fmt.Errorf("test, ReadJSON called after close")
					default:
					}

					count := atomic.LoadInt32(&readCountRemaining)
					if count == 0 {
						<-closed
						return fmt.Errorf("test, close called during ReadJSON")
					}

					matchIn, ok := v.(*coinbase.Match)
					if !ok {
						return nil
					}

					subscribedProductIDMu.Lock()
					productID := subscribedProductID
					subscribedProductIDMu.Unlock()

					*matchIn = *matchesIn[count-1].toMatch(productID)

					defer func() {
						if atomic.AddInt32(&readCountRemaining, -1) == 0 {
							wg.Done()
						}
					}()

					return nil
				},

				CloseFunc: func() error {
					select {
					case <-ctx.Done():
						return fmt.Errorf("test ctx is done: %w", ctx.Err())
					default:
					}

					close(closed)

					return nil
				},
			}, nil, nil
		},
	}
}
