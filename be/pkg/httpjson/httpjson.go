// Package httpjson calls JSON HTTP APIs with retries on 429 and 5xx (used by the LLM and image adapters).
package httpjson

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// UserAgent: Pixabay sits behind Cloudflare, which blocks generic client user agents.
const UserAgent = "ielts-practice/1.0 (vocabulary app)"

// DefaultClient is used by Get.
var DefaultClient = &http.Client{Timeout: 15 * time.Second}

// StatusError is returned for non-2xx responses so callers can branch on the code.
type StatusError struct {
	Code int
	Body string
}

func (e *StatusError) Error() string { return fmt.Sprintf("http %d: %s", e.Code, e.Body) }

// Get performs a GET with retries on 429 and 5xx (1s, 2s, 4s) and decodes JSON into out.
func Get(ctx context.Context, url string, headers map[string]string, out any) error {
	return Send(ctx, http.MethodGet, url, headers, nil, out, DefaultClient)
}

// Send sends an optional JSON body with retries on 429 and 5xx and decodes the JSON answer into out.
func Send(ctx context.Context, method, url string, headers map[string]string, body []byte, out any, client *http.Client) error {
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		if attempt > 0 {
			wait := time.Duration(1<<(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}
		var reader io.Reader
		if body != nil {
			reader = bytes.NewReader(body)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, reader)
		if err != nil {
			return err
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("User-Agent", UserAgent)
		req.Header.Set("Accept", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		res, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		respBody, _ := io.ReadAll(io.LimitReader(res.Body, 4<<20))
		_ = res.Body.Close()
		if res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500 {
			lastErr = &StatusError{Code: res.StatusCode, Body: Truncate(string(respBody), 200)}
			if ra, _ := strconv.Atoi(res.Header.Get("Retry-After")); ra > 0 && ra <= 10 {
				time.Sleep(time.Duration(ra) * time.Second)
			}
			continue
		}
		if res.StatusCode < 200 || res.StatusCode > 299 {
			return &StatusError{Code: res.StatusCode, Body: Truncate(string(respBody), 200)}
		}
		if out == nil {
			return nil
		}
		return json.Unmarshal(respBody, out)
	}
	return lastErr
}

// Truncate cuts s to at most n bytes.
func Truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
