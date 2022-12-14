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
	err := runApp(&coinbase.Client{}, log.Writer(), make(chan os.Signal, 1))
	if err != nil {
		log.Fatal(err)
	}
}

// runApp runs the application, connecting to Coinbase with coinbaseClient and outputting
// the VWAPs (or errors) on output. Signal interrupt to exit.
func runApp(coinbaseClient *coinbase.Client, output io.Writer, interrupt chan os.Signal) error {
	ctx := context.Background()

	log.Print("[INF] Creating subscriptions...\n")
	subscriptions, err := subscribeToAll(
		ctx,
		coinbaseClient,
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

	wg.Wait()

	return nil
}

// subscribeToAll establishes subscriptions for each productID in productIDs using
// coinbaseClient.
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

// startPrintingVWAPs will start outputting VWAPs to w for each Subscription in subscriptions.
// wg is used to signal when each VWAP output loop start/stops.
func startPrintingVWAPs(subscriptions []*coinbase.MatchesSubscription, wg *sync.WaitGroup, w io.Writer) {
	for _, subscription := range subscriptions {
		read := subscription.Read()
		productID := subscription.ProductID()

		wg.Add(1)
		go func() {
			printVWAP(read, productID, w)

			wg.Done()
		}()
	}
}

// printVWAP reads a MatchResponse from read and outputs it to w for productID productID.
func printVWAP(read <-chan *coinbase.MatchResponse, productID coinbase.ProductID, w io.Writer) {
	vwapCalculator := vwap.NewSlidingWindowVWAP(200)

	for {
		matchResponse, ok := <-read
		if !ok {
			break
		}

		units, unitPrice, err := matchResponse.ToUnitsAndUnitPrice()
		if err != nil {
			fmt.Fprintf(w, "%q ERROR: %v\n", productID, err)
			continue
		}

		v := vwapCalculator.Add(units, unitPrice)

		fmt.Fprintf(w, "%q: %v\n", productID, v)
	}
}
