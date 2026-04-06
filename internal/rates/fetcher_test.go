package rates

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type successHTTPClientStub struct {
	body       []byte
	statusCode int
}

func (s successHTTPClientStub) Get(context.Context, string) ([]byte, int, error) {
	return s.body, s.statusCode, nil
}

type failureHTTPClientStub struct {
	err error
}

func (s failureHTTPClientStub) Get(context.Context, string) ([]byte, int, error) {
	return nil, 0, s.err
}

func TestFetcherFetchOrderBook(t *testing.T) {
	tests := []struct {
		name       string
		httpClient HTTPClient
		wantBook   *OrderBook
		wantErr    bool
	}{
		{
			name: "returns parsed order book",
			httpClient: successHTTPClientStub{
				body: []byte(`{
					"timestamp": 1775406674,
					"asks": [
						{"price": "100.0", "volume": "1", "amount": "100.0"},
						{"price": "110.0", "volume": "1", "amount": "110.0"}
					],
					"bids": [
						{"price": "90.0", "volume": "1", "amount": "90.0"},
						{"price": "80.0", "volume": "1", "amount": "80.0"}
					]
				}`),
				statusCode: 200,
			},
			wantBook: &OrderBook{
				Timestamp: 1775406674,
				Asks: []SpotLevel{
					{Price: 100, Amount: 100},
					{Price: 110, Amount: 110},
				},
				Bids: []SpotLevel{
					{Price: 90, Amount: 90},
					{Price: 80, Amount: 80},
				},
			},
		},
		{
			name: "returns error on http client failure",
			httpClient: failureHTTPClientStub{
				err: errors.New("network failed"),
			},
			wantErr: true,
		},
		{
			name: "returns error on non-success status",
			httpClient: successHTTPClientStub{
				body:       []byte(`{"error":"unavailable"}`),
				statusCode: 503,
			},
			wantErr: true,
		},
		{
			name: "returns error on invalid json",
			httpClient: successHTTPClientStub{
				body:       []byte(`{invalid json`),
				statusCode: 200,
			},
			wantErr: true,
		},
		{
			name: "returns error on invalid numeric value",
			httpClient: successHTTPClientStub{
				body: []byte(`{
					"timestamp": 1775406674,
					"asks": [{"price": "broken", "volume": "1", "amount": "100.0"}],
					"bids": [{"price": "90.0", "volume": "1", "amount": "90.0"}]
				}`),
				statusCode: 200,
			},
			wantErr: true,
		},
		{
			name: "returns error on empty asks",
			httpClient: successHTTPClientStub{
				body: []byte(`{
					"timestamp": 1775406674,
					"asks": [],
					"bids": [{"price": "90.0", "volume": "1", "amount": "90.0"}]
				}`),
				statusCode: 200,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewFetcher(tt.httpClient, "http://grinex.test/api/v1/spot/depth?symbol=usdta7a5")

			orderBook, err := fetcher.FetchOrderBook(context.Background())

			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantBook, orderBook)
		})
	}
}
