package rates

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type successOrderBookFetcherStub struct {
	orderBook *OrderBook
}

func (s successOrderBookFetcherStub) FetchOrderBook(context.Context) (*OrderBook, error) {
	return s.orderBook, nil
}

type failureOrderBookFetcherStub struct {
	err error
}

func (s failureOrderBookFetcherStub) FetchOrderBook(context.Context) (*OrderBook, error) {
	return nil, s.err
}

type successCalculatorStub struct {
	value float64
}

func (s successCalculatorStub) Calculate([]SpotLevel) (float64, error) {
	return s.value, nil
}

type failureCalculatorStub struct {
	err error
}

func (s failureCalculatorStub) Calculate([]SpotLevel) (float64, error) {
	return 0, s.err
}

func TestProviderFetchRates(t *testing.T) {
	tests := []struct {
		name          string
		fetcher       ProviderFetcher
		askCalculator Calculator
		bidCalculator Calculator
		wantAsk       float64
		wantBid       float64
		wantErr       bool
	}{
		{
			name: "fetches and calculates rate",
			fetcher: successOrderBookFetcherStub{
				orderBook: &OrderBook{
					Timestamp: 1775406674,
					Asks:      []SpotLevel{{Price: 100, Amount: 100}},
					Bids:      []SpotLevel{{Price: 90, Amount: 90}},
				},
			},
			askCalculator: successCalculatorStub{value: 101},
			bidCalculator: successCalculatorStub{value: 89},
			wantAsk:       101,
			wantBid:       89,
		},
		{
			name:          "returns fetcher error",
			fetcher:       failureOrderBookFetcherStub{err: errors.New("fetch failed")},
			askCalculator: successCalculatorStub{value: 101},
			bidCalculator: successCalculatorStub{value: 89},
			wantErr:       true,
		},
		{
			name: "returns ask calculator error",
			fetcher: successOrderBookFetcherStub{
				orderBook: &OrderBook{
					Asks: []SpotLevel{{Price: 100, Amount: 100}},
					Bids: []SpotLevel{{Price: 90, Amount: 90}},
				},
			},
			askCalculator: failureCalculatorStub{err: errors.New("ask failed")},
			bidCalculator: successCalculatorStub{value: 89},
			wantErr:       true,
		},
		{
			name: "returns bid calculator error",
			fetcher: successOrderBookFetcherStub{
				orderBook: &OrderBook{
					Asks: []SpotLevel{{Price: 100, Amount: 100}},
					Bids: []SpotLevel{{Price: 90, Amount: 90}},
				},
			},
			askCalculator: successCalculatorStub{value: 101},
			bidCalculator: failureCalculatorStub{err: errors.New("bid failed")},
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewProvider(tt.fetcher, tt.askCalculator, tt.bidCalculator)

			rate, err := provider.FetchRates(context.Background())
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantAsk, rate.Ask)
			assert.Equal(t, tt.wantBid, rate.Bid)
		})
	}
}
