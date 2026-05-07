package client

import "context"

// Login posts to /login and returns the JWT bundle. The returned token is a
// JWT good for the user's normal API calls; we use it transiently during init.
func (c *Client) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	var out LoginResponse
	if err := c.Do(ctx, "POST", "/login", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CurrentUser fetches /user — handy for resolving the bot's own user_id from
// its API token without poking the human's data.
func (c *Client) CurrentUser(ctx context.Context) (*User, error) {
	var out User
	if err := c.Do(ctx, "GET", "/user", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
