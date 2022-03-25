package client

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var defaultClient = http.Client{
	Timeout: 60 * time.Second,
}

const maxDownloadBytes = 10e+6

// Download opens the given URL, reads it body and returns it.
func Download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("http.Download: NewRequestWithContext: %w", err)
	}
	resp, err := defaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http.Download: Get(%s): %w", url, err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(io.LimitReader(resp.Body, maxDownloadBytes))
	if err != nil {
		return nil, fmt.Errorf("http.Download: Get(%s): Read Body: %w", url, err)
	}
	return b, nil
}

// Open opens the given URL, returning a handle to the response body. The
// caller is responsible for closing the body.
func Open(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("http.Download: NewRequestWithContext: %w", err)
	}
	resp, err := defaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http.Download: Get(%s): %w", url, err)
	}
	// TODO how to make io.LimitReader a io.ReadCloser?
	return resp.Body, nil
}
