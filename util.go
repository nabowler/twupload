package twupload

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	ErrBadHttpStatus = sentinelError("Non 200-level http status")
)

func handleHttpResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("Got http response status %d %w", resp.StatusCode, ErrBadHttpStatus)
	}

	if target == nil {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Unable to read response body %w", err)
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal body %w", err)
	}

	return nil
}

func post(ctx context.Context, httpClient *http.Client, loc string, data url.Values) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loc, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return httpClient.Do(req)
}

func get(ctx context.Context, httpClient *http.Client, url string) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return httpClient.Do(req)
}
