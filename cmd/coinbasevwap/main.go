package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/byatesrae/coinbase_vwap/internal/coinbase"
	"github.com/byatesrae/coinbase_vwap/internal/vwap"
)

func main() {
	err := runApp(log.Writer())
	if err != nil {
		log.Fatal(err)
	}
}

func runApp(output io.Writer) error {
	ctx := context.Background()

	coinbaseClient := coinbase.Client{}

	log.Print("[INF] Creating subscriptions...\n")
	subscriptions, err := subscribeToAll(
		ctx,
		&coinbaseClient,
		[]coinbase.ProductID{
			coinbase.ProductIDBtcUsd,
			coinbase.ProductIDEthUsd,
			coinbase.ProductIDEthBtc,
		},
	)
	if err != nil {
		return fmt.Errorf("subscribe to all: %w", err)
	}

	log.Print("[INF] Starting printing of VWAPS...\n")
	wg := sync.WaitGroup{}
	startPrintingVWAPs(subscriptions, &wg, output)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	<-interrupt
	log.Print("[INF] Interrupted!\n")

	closeCtx, cancelCloseCtx := context.WithTimeout(ctx, time.Second*2)
	defer cancelCloseCtx()

	for _, subscription := range subscriptions {
		subscription := subscription

		go func() {
			if err := subscription.Close(closeCtx); err != nil {
				log.Printf("[ERR] Failed to close subscription for %q: %v", subscription.ProductID(), err)
			}
		}()
	}

	wg.Wait() // or timeout

	return nil
}

func subscribeToAll(ctx context.Context, coinbaseClient *coinbase.Client, productIDs []coinbase.ProductID) ([]*coinbase.MatchesSubscription, error) {
	var subscriptions []*coinbase.MatchesSubscription

	for _, productID := range productIDs {
		subscription, err := coinbaseClient.SubscribeToMatchesForProduct(ctx, productID)
		if err != nil {
			return nil, fmt.Errorf("subscribe to matches (%q): %w", productID, err)
		}

		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

func startPrintingVWAPs(subscriptions []*coinbase.MatchesSubscription, wg *sync.WaitGroup, w io.Writer) {
	for _, subscription := range subscriptions {
		subscription := subscription

		wg.Add(1)
		go func() {
			printVWAP(subscription, w)

			wg.Done()
		}()
	}
}

func printVWAP(subscription *coinbase.MatchesSubscription, w io.Writer) {
	vwapCalculator := vwap.NewSlidingWindowVWAP(200)

	for {
		matchResponse, ok := <-subscription.Read()
		if !ok {
			break
		}

		units, unitPrice, err := matchResponse.ToUnitsAndUnitPrice()
		if err != nil {
			fmt.Fprintf(w, "%q ERROR: %v\n", subscription.ProductID(), err)
			continue
		}

		v := vwapCalculator.Add(units, unitPrice)

		fmt.Fprintf(w, "%q: %v\n", subscription.ProductID(), v)
	}
}
