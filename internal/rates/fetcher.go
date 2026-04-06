package rates

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

type OrderBook struct {
	Timestamp int64
	Asks      []SpotLevel
	Bids      []SpotLevel
}

type SpotLevel struct {
	Price  float64
	Amount float64
}

type HTTPClient interface {
	Get(ctx context.Context, url string) ([]byte, int, error)
}

type Fetcher struct {
	httpClient HTTPClient
	apiURL     string
}

type depthResponse struct {
	Timestamp int64       `json:"timestamp"`
	Asks      []bookLevel `json:"asks"`
	Bids      []bookLevel `json:"bids"`
}

type bookLevel struct {
	Price  string `json:"price"`
	Volume string `json:"volume"`
	Amount string `json:"amount"`
}

func NewFetcher(httpClient HTTPClient, apiURL string) *Fetcher {
	return &Fetcher{
		httpClient: httpClient,
		apiURL:     apiURL,
	}
}

func (c *Fetcher) FetchOrderBook(ctx context.Context) (*OrderBook, error) {
	var payload depthResponse

	body, statusCode, err := c.httpClient.Get(ctx, c.apiURL)

	if err != nil {
		return nil, fmt.Errorf("request grinex depth: %w", err)
	}

	if statusCode >= 400 {
		return nil, fmt.Errorf("grinex returned status %d", statusCode)
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode grinex response: %w", err)
	}

	asks, err := parseLevels(payload.Asks)

	if err != nil {
		return nil, fmt.Errorf("parse asks: %w", err)
	}

	bids, err := parseLevels(payload.Bids)

	if err != nil {
		return nil, fmt.Errorf("parse bids: %w", err)
	}

	return &OrderBook{
		Timestamp: payload.Timestamp,
		Asks:      asks,
		Bids:      bids,
	}, nil
}

type Calculator interface {
	Calculate(levels []SpotLevel) (float64, error)
}

func parseLevels(raw []bookLevel) ([]SpotLevel, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("no spot data")
	}

	levels := make([]SpotLevel, 0, len(raw))

	for _, item := range raw {
		price, err := strconv.ParseFloat(item.Price, 64)

		if err != nil {
			return nil, fmt.Errorf("parse price %q: %w", item.Price, err)
		}

		amount, err := strconv.ParseFloat(item.Amount, 64)

		if err != nil {
			return nil, fmt.Errorf("parse amount %q: %w", item.Amount, err)
		}

		levels = append(levels, SpotLevel{
			Price:  price,
			Amount: amount,
		})
	}

	return levels, nil
}
