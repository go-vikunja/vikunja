package client

import "context"

// Info fetches GET /info. No auth required.
func (c *Client) Info(ctx context.Context) (*Info, error) {
	var out Info
	if err := c.Do(ctx, "GET", "/info", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
