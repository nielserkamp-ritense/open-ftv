package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GetFile downloads a file from the GitHub repository using the given download url.
func (r *repo) GetFile(url string, timeout time.Duration) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	ctx, cancel := context.WithTimeout(r.ctx, timeout)
	defer cancel()

	var resp *http.Response
	if resp, err = http.DefaultClient.Do(req.WithContext(ctx)); err != nil {
		return nil, fmt.Errorf("failed to make http request: %w", err)
	}

	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
