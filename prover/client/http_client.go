package client

import (
	"net/http"
	"time"
)

// HTTPClient is basic http client implementation
// of INodeClient
type HTTPClient struct {
	client *http.Client
	addr   string
}

// NewHttpClient create new HTTPClient with given address
func NewHttpClient(addr string) (INodeClient, error) {
	return &HTTPClient{
		client: &http.Client{
			Transport: newRoundTripper(http.DefaultTransport),
			// TODO set timeout via config
			Timeout: time.Second * 600,
		},
		addr: addr,
	}, nil
}

// Close all connections to prover
func (c *HTTPClient) Close() {
	c.client.CloseIdleConnections()
}

func (c *HTTPClient) DoRequest(*http.Request) (*http.Response, error) {
	return nil, nil
}
