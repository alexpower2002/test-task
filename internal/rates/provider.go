package rates

import (
	"context"
	"fmt"
	"time"
)

type Rate struct {
	Ask       float64
	Bid       float64
	Timestamp time.Time
}

type ProviderFetcher interface {
	FetchOrderBook(ctx context.Context) (*OrderBook, error)
}

type Provider struct {
	fetcher       ProviderFetcher
	askCalculator Calculator
	bidCalculator Calculator
}

func NewProvider(fetcher ProviderFetcher, askCalculator, bidCalculator Calculator) *Provider {
	return &Provider{
		fetcher:       fetcher,
		askCalculator: askCalculator,
		bidCalculator: bidCalculator,
	}
}

func (c *Provider) FetchRates(ctx context.Context) (*Rate, error) {
	orderBook, err := c.fetcher.FetchOrderBook(ctx)

	if err != nil {
		return nil, fmt.Errorf("fetch order book: %w", err)
	}

	ask, err := c.askCalculator.Calculate(orderBook.Asks)

	if err != nil {
		return nil, fmt.Errorf("calculate ask: %w", err)
	}

	bid, err := c.bidCalculator.Calculate(orderBook.Bids)

	if err != nil {
		return nil, fmt.Errorf("calculate bid: %w", err)
	}

	return &Rate{
		Ask:       ask,
		Bid:       bid,
		Timestamp: time.Now().UTC(),
	}, nil
}
