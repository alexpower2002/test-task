package utils

import (
	"context"
	"time"

	"github.com/go-resty/resty/v2"
)

type RestyClient struct {
	client *resty.Client
}

func NewRestyClient(timeout time.Duration) *RestyClient {
	client := resty.New().
		SetTimeout(timeout).
		SetRetryCount(0)

	return &RestyClient{client: client}
}

func (c *RestyClient) Get(ctx context.Context, url string) ([]byte, int, error) {
	response, err := c.client.R().
		SetContext(ctx).
		Get(url)
	if err != nil {
		return nil, 0, err
	}

	return response.Body(), response.StatusCode(), nil
}
