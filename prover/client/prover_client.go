package client

import (
	"net/http"
)

// INodeClient is low-level abstraction that
// represent connection to prover
type INodeClient interface {
	// Close all connections to prover
	Close()

	DoRequest(*http.Request) (*http.Response, error)
}

// roundTripper is round-tripper
// transport implementation
type roundTripper struct {
	r http.RoundTripper
}

func newRoundTripper(r http.RoundTripper) *roundTripper {
	return &roundTripper{r}
}

// RoundTrip set default headers to
// all requests no prover
func (mrt roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("accept", "application/json")
	r.Header.Add("content-type", "application/json")

	return mrt.r.RoundTrip(r)
}
