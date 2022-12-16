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

	"github.com/byatesrae/coinbase_vwap/internal/coinbase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type matchesIn struct {
	matchType coinbase.MessageType
	size      string
	price     string
	message   string
}

func newDialerFake(ctx context.Context, t *testing.T, wg *sync.WaitGroup, matchesIn []*matchesIn) *DialerMock {
	return &DialerMock{
		DialContextFunc: func(_ context.Context, _ string, _ http.Header) (coinbase.Conn, *http.Response, error) {
			subscribedProductID := coinbase.ProductIDUnknown
			subscribedProductIDMu := sync.Mutex{}

			readCountRemaining := int32(len(matchesIn))

			closed := make(chan struct{})

			return &ConnMock{
				WriteJSONFunc: func(v interface{}) error {
					select {
					case <-ctx.Done():
						return fmt.Errorf("test ctx is done: %w", ctx.Err())
					case <-closed:
						return fmt.Errorf("test, WriteJSON called after close")
					default:
					}

					subscribeRequest, ok := v.(coinbase.SubscribeRequest)
					if !ok {
						return nil
					}

					if len(subscribeRequest.ProductIDs) != 0 ||
						len(subscribeRequest.Channels) != 1 ||
						len(subscribeRequest.Channels[0].ProductIDs) != 1 {
						return nil
					}

					subscribedProductIDMu.Lock()
					defer subscribedProductIDMu.Unlock()

					subscribedProductID = subscribeRequest.Channels[0].ProductIDs[0]

					return nil
				},
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

					defer func() {
						if atomic.AddInt32(&readCountRemaining, -1) == 0 {
							wg.Done()
						}
					}()

					require.IsType(t, (*coinbase.Match)(nil), v, "")
					matchIn, ok := v.(*coinbase.Match)
					if !ok {
						return nil
					}

					subscribedProductIDMu.Lock()
					productID := subscribedProductID
					subscribedProductIDMu.Unlock()

					*matchIn = coinbase.Match{
						Type:      matchesIn[count-1].matchType,
						ProductID: productID,
						Size:      matchesIn[count-1].size,
						Price:     matchesIn[count-1].price,
						Message:   matchesIn[count-1].message,
					}

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

type stringBuilderMutex struct {
	sb strings.Builder
	mu sync.Mutex
}

func (s *stringBuilderMutex) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.sb.Write(p)
}

func TestSubscribes(t *testing.T) {
	t.Parallel()

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

	go func() {
		err := runApp(coinbaseClient, &sbWithMutex, interrupt)

		assert.NoError(t, err, "runApp error.")

		close(exited)
	}()

	wgForAllSubscriptionReads.Wait()

	interrupt <- os.Interrupt

	<-exited

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
