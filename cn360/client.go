// Package cn360 is the library behind the cn360 command line: the HTTP client,
// HTML parsers, and typed data models for 360 Search (so.com) and related
// 360 services (news.so.com, baike.so.com, sug.so.com).
//
// The client paces requests to stay polite (2 req/s default), retries transient
// 429 and 5xx responses, and sets browser-like headers to reduce CAPTCHA friction.
// The __guid cookie is picked up automatically on the first request by the shared
// cookie jar.
package cn360

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Client talks to 360 Search surfaces over HTTP with pacing and retries.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured with defaults.
func NewClient() *Client {
	return NewClientWithConfig(DefaultConfig())
}

// NewClientWithConfig returns a Client configured with cfg.
func NewClientWithConfig(cfg Config) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout, Jar: jar},
	}
}

// WarmUp GETs the so.com homepage to pick up the __guid session cookie.
// Call once before issuing SERP requests for best results.
func (c *Client) WarmUp(ctx context.Context) error {
	_, err := c.Get(ctx, c.cfg.SearchBaseURL+"/")
	return err
}

// Get fetches rawURL and returns the response body. It paces, retries 429/5xx,
// detects CAPTCHA pages, and maps outcomes to typed errors.
func (c *Client) Get(ctx context.Context, rawURL string) ([]byte, error) {
	return c.fetch(ctx, rawURL, false)
}

// GetJSON fetches rawURL with a JSON User-Agent and Accept header.
func (c *Client) GetJSON(ctx context.Context, rawURL string) ([]byte, error) {
	return c.fetch(ctx, rawURL, true)
}

// GetHTML fetches rawURL and parses it as an HTML document.
func (c *Client) GetHTML(ctx context.Context, rawURL string) (*goquery.Document, error) {
	body, err := c.Get(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse HTML %s: %w", rawURL, err)
	}
	return doc, nil
}

func (c *Client) fetch(ctx context.Context, rawURL string, jsonMode bool) ([]byte, error) {
	retries := c.cfg.Retries
	if retries < 1 {
		retries = 1
	}
	var lastErr error
	for attempt := 1; attempt <= retries; attempt++ {
		if attempt > 1 {
			wait := retryWait(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}
		body, code, err := c.do(ctx, rawURL, jsonMode)
		if err != nil {
			lastErr = err
			continue
		}
		switch {
		case code == 200:
			if isBlocked(body) {
				return nil, codeErr(ExitBlocked, "360 Search returned a CAPTCHA for %s", rawURL)
			}
			return body, nil
		case code == 404:
			return nil, codeErr(ExitNotFound, "not found: %s", rawURL)
		case code == 429 || code == 503:
			lastErr = ErrRateLimited
			if attempt == retries {
				return nil, codeErr(ExitBlocked, "rate limited (HTTP %d) after %d attempts", code, retries)
			}
			wait := time.Duration(attempt*attempt) * 3 * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		case code >= 500:
			lastErr = fmt.Errorf("server error HTTP %d", code)
			if attempt == retries {
				return nil, codeErr(ExitGeneric, "server error HTTP %d after %d attempts", code, retries)
			}
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		default:
			return nil, codeErr(ExitGeneric, "unexpected HTTP %d for %s", code, rawURL)
		}
	}
	return nil, codeErr(ExitGeneric, "all %d attempts failed: %v", retries, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string, jsonMode bool) ([]byte, int, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, err
	}
	if jsonMode {
		req.Header.Set("User-Agent", JSONUA)
		req.Header.Set("Accept", "application/json, */*;q=0.8")
	} else {
		req.Header.Set("User-Agent", c.cfg.UserAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		req.Header.Set("Referer", c.cfg.SearchBaseURL+"/")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

// pace blocks until at least Rate has elapsed since the previous request.
func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

// isBlocked returns true when 360 Search served a CAPTCHA or anti-bot page.
func isBlocked(body []byte) bool {
	s := string(body)
	return strings.Contains(s, `id="captcha-container"`) ||
		strings.Contains(s, "/captcha") ||
		strings.Contains(s, "请输入验证码")
}

// retryWait returns the backoff duration for a given attempt number.
func retryWait(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		return 5 * time.Second
	}
	return d
}
