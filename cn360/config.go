package cn360

import "time"

// Site constants.
const (
	SearchHost    = "so.com"
	SearchBaseURL = "https://www.so.com"
	NewsBaseURL   = "https://news.so.com"
	BaikeBaseURL  = "https://baike.so.com"
	SugBaseURL    = "https://sug.so.com"
	HotAPIURL     = SearchBaseURL + "/s/index/top_hot_words"

	DefaultDelay   = 2 * time.Second
	DefaultRetries = 3
	DefaultTimeout = 15 * time.Second
)

// BrowserUA is a real browser User-Agent that reduces anti-bot friction on so.com SERPs.
const BrowserUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

// JSONUA is used for JSON API endpoints (hot, suggest).
const JSONUA = "cn360/dev (+https://github.com/tamnd/360cn-cli)"

// Config holds runtime configuration for a Client.
type Config struct {
	SearchBaseURL string        // so.com search base (default SearchBaseURL)
	NewsBaseURL   string        // news.so.com base (default NewsBaseURL)
	BaikeBaseURL  string        // baike.so.com base (default BaikeBaseURL)
	SugBaseURL    string        // sug.so.com base (default SugBaseURL)
	HotAPIURL     string        // hot JSON API (default HotAPIURL)
	UserAgent     string        // HTTP User-Agent header
	Rate          time.Duration // minimum gap between requests
	Retries       int           // retry attempts on 429/5xx
	Timeout       time.Duration // per-request HTTP timeout
}

// DefaultConfig returns sensible, polite defaults for scraping 360 Search.
func DefaultConfig() Config {
	return Config{
		SearchBaseURL: SearchBaseURL,
		NewsBaseURL:   NewsBaseURL,
		BaikeBaseURL:  BaikeBaseURL,
		SugBaseURL:    SugBaseURL,
		HotAPIURL:     HotAPIURL,
		UserAgent:     BrowserUA,
		Rate:          DefaultDelay,
		Retries:       DefaultRetries,
		Timeout:       DefaultTimeout,
	}
}
