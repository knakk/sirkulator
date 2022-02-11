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
	Timeout: 5 * time.Second,
}

func Download(ctx context.Context, url string) ([]byte, error) {
	const maxDownloadBytes = 10e+6
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

func Open(ctx context.Context, url string) (io.ReadCloser, error) {
	const maxDownloadBytes = 10e+6
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