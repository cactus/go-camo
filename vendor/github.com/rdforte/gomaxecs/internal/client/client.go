// Copyright 2004 Ryan Forte
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package client provides an HTTP client.
package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/rdforte/gomaxecs/internal/config"
)

// New returns a new Client.
func New(cfg config.Config) *Client {
	return &Client{
		log: cfg.DebugLogf,
		client: &http.Client{
			Timeout: cfg.Client.HTTPTimeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: cfg.Client.DialTimeout,
				}).DialContext,
				MaxIdleConns:          cfg.Client.MaxIdleConns,
				MaxIdleConnsPerHost:   cfg.Client.MaxIdleConnsPerHost,
				DisableKeepAlives:     cfg.Client.DisableKeepAlives,
				IdleConnTimeout:       cfg.Client.IdleConnTimeout,
				TLSHandshakeTimeout:   cfg.Client.TLSHandshakeTimeout,
				ResponseHeaderTimeout: cfg.Client.ResponseHeaderTimeout,
			},
		},
	}
}

// Client is an HTTP client.
type Client struct {
	log    config.Logger
	client *http.Client
}

// Get performs an HTTP GET request.
func (c *Client) Get(ctx context.Context, url string) (*Response, error) {
	c.log("Performing HTTP GET request to %s", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.log("Error creating HTTP request to %s: %v", url, err)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		c.log("Error performing HTTP GET request to %s: %v", url, err)
		return nil, fmt.Errorf("failed to perform HTTP GET request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.log("Error reading response body from %s: %v", url, err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	c.log("client: Received response with status code from %s: %d", url, res.StatusCode)
	c.log("client: Response body from url %s: %s", url, string(body))

	return &Response{res.StatusCode, body}, nil
}

type Response struct {
	StatusCode int
	Body       []byte
}
