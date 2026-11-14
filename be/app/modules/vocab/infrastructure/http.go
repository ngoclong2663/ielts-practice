package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// userAgent: Pixabay sits behind Cloudflare, which blocks generic client user agents.
const userAgent = "ielts-practice/1.0 (vocabulary app)"

var httpClient = &http.Client{Timeout: 15 * time.Second}

// statusError is returned for non-2xx responses so callers can branch on the code.
type statusError struct {
	Code int
	Body string
}

func (e *statusError) Error() string { return fmt.Sprintf("http %d: %s", e.Code, e.Body) }

// getJSON performs a GET with retries on 429 and 5xx (1s, 2s, 4s) and decodes JSON into out.
func getJSON(ctx context.Context, url string, headers map[string]string, out any) error {
	return sendJSON(ctx, http.MethodGet, url, headers, nil, out, httpClient)
}

// sendJSON sends an optional JSON body with retries on 429 and 5xx and decodes the JSON answer into out.
func sendJSON(ctx context.Context, method, url string, headers map[string]string, body []byte, out any, client *http.Client) error {
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
		req.Header.Set("User-Agent", userAgent)
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
			lastErr = &statusError{Code: res.StatusCode, Body: truncate(string(respBody), 200)}
			if ra, _ := strconv.Atoi(res.Header.Get("Retry-After")); ra > 0 && ra <= 10 {
				time.Sleep(time.Duration(ra) * time.Second)
			}
			continue
		}
		if res.StatusCode < 200 || res.StatusCode > 299 {
			return &statusError{Code: res.StatusCode, Body: truncate(string(respBody), 200)}
		}
		if out == nil {
			return nil
		}
		return json.Unmarshal(respBody, out)
	}
	return lastErr
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// ttlCache keeps provider responses for a while (Pixabay requires caching search results for 24 h).
type ttlCache[T any] struct {
	mu      sync.Mutex
	ttl     time.Duration
	max     int
	entries map[string]ttlEntry[T]
}

type ttlEntry[T any] struct {
	value   T
	expires time.Time
}

func newTTLCache[T any](ttl time.Duration, max int) *ttlCache[T] {
	return &ttlCache[T]{ttl: ttl, max: max, entries: map[string]ttlEntry[T]{}}
}

func (c *ttlCache[T]) get(key string) (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expires) {
		var zero T
		return zero, false
	}
	return e.value, true
}

func (c *ttlCache[T]) set(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.max {
		now := time.Now()
		for k, e := range c.entries {
			if now.After(e.expires) {
				delete(c.entries, k)
			}
		}
		if len(c.entries) >= c.max {
			c.entries = map[string]ttlEntry[T]{}
		}
	}
	c.entries[key] = ttlEntry[T]{value: value, expires: time.Now().Add(c.ttl)}
}
