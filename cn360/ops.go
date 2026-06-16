package cn360

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Search fetches one page of so.com web search results for query.
func (c *Client) Search(ctx context.Context, query string, page int) ([]SearchResult, error) {
	if page < 1 {
		page = 1
	}
	u := fmt.Sprintf("%s/s?q=%s&pn=%d", c.cfg.SearchBaseURL, encodeQuery(query), page)
	return c.SearchURL(ctx, u, query)
}

// SearchURL fetches a SERP page at an arbitrary URL (used by tests with httptest).
func (c *Client) SearchURL(ctx context.Context, u, query string) ([]SearchResult, error) {
	doc, err := c.GetHTML(ctx, u)
	if err != nil {
		return nil, err
	}
	return parseSearchResults(doc, query), nil
}

// News fetches one page of news.so.com news search results for query.
func (c *Client) News(ctx context.Context, query string, page int) ([]NewsResult, error) {
	if page < 1 {
		page = 1
	}
	u := fmt.Sprintf("%s/ns?q=%s&src=sug&pn=%d", c.cfg.NewsBaseURL, encodeQuery(query), page)
	return c.NewsURL(ctx, u, query)
}

// NewsURL fetches news results from an arbitrary URL (used by tests).
func (c *Client) NewsURL(ctx context.Context, u, query string) ([]NewsResult, error) {
	doc, err := c.GetHTML(ctx, u)
	if err != nil {
		return nil, err
	}
	return parseNewsResults(doc, query), nil
}

// Hot fetches trending search terms from the 360 hot-search JSON API.
func (c *Client) Hot(ctx context.Context, limit int) ([]HotItem, error) {
	return c.HotURL(ctx, c.cfg.HotAPIURL, limit)
}

// HotURL fetches hot items from an arbitrary URL (used by tests).
func (c *Client) HotURL(ctx context.Context, u string, limit int) ([]HotItem, error) {
	body, err := c.GetJSON(ctx, u)
	if err != nil {
		return nil, err
	}
	items := parseHotItems(body)
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

// Suggest fetches autocomplete suggestions for key from sug.so.com.
func (c *Client) Suggest(ctx context.Context, key string, limit int) ([]Suggestion, error) {
	u := fmt.Sprintf("%s/suggest?key=%s&encode=utf-8&type=json&src=0", c.cfg.SugBaseURL, encodeQuery(key))
	return c.SuggestURL(ctx, u, key, limit)
}

// SuggestURL fetches suggestions from an arbitrary URL (used by tests).
func (c *Client) SuggestURL(ctx context.Context, u, key string, limit int) ([]Suggestion, error) {
	body, err := c.GetJSON(ctx, u)
	if err != nil {
		return nil, err
	}
	return parseSuggestions(body, key, limit), nil
}

// Encyclopedia searches baike.so.com for query and returns result stubs.
func (c *Client) Encyclopedia(ctx context.Context, query string, page int) ([]EncyclopediaResult, error) {
	if page < 1 {
		page = 1
	}
	u := fmt.Sprintf("%s/search?q=%s&pn=%d", c.cfg.BaikeBaseURL, encodeQuery(query), page)
	return c.EncyclopediaURL(ctx, u, query)
}

// EncyclopediaURL fetches encyclopedia results from an arbitrary URL (used by tests).
func (c *Client) EncyclopediaURL(ctx context.Context, u, query string) ([]EncyclopediaResult, error) {
	doc, err := c.GetHTML(ctx, u)
	if err != nil {
		return nil, err
	}
	return parseEncyclopediaResults(doc, query), nil
}

// --- parsers ---

func parseSearchResults(doc *goquery.Document, query string) []SearchResult {
	now := time.Now().UTC().Format(time.RFC3339)
	var results []SearchResult
	pos := 0
	doc.Find("ul#results li.res-list, #main .res-list .res-item, li.res-list").Each(func(_ int, s *goquery.Selection) {
		pos++
		anchor := s.Find("h3.res-title a, .res-title a")
		title := strings.TrimSpace(anchor.Text())
		if title == "" {
			return
		}
		// Prefer data-mdurl over href for the real destination.
		href, _ := anchor.Attr("href")
		if mdURL, ok := anchor.Attr("data-mdurl"); ok && mdURL != "" {
			href = mdURL
		}
		// Reject internal search relative URLs.
		if strings.HasPrefix(href, "/s?") || strings.HasPrefix(href, "/") {
			href, _ = anchor.Attr("data-mdurl")
		}
		snippet := strings.TrimSpace(s.Find("p.res-desc, p.res-desc2").Text())
		site := strings.TrimSpace(s.Find("div.res-linkinfo cite, div.res-linkinfo span").First().Text())
		isAd := s.HasClass("res-list-ad") || s.Find("span.ad-label").Length() > 0
		results = append(results, SearchResult{
			Query:     query,
			Position:  pos,
			URL:       href,
			Title:     title,
			Snippet:   snippet,
			SiteName:  site,
			IsAd:      isAd,
			FetchedAt: now,
		})
	})
	return results
}

func parseNewsResults(doc *goquery.Document, query string) []NewsResult {
	now := time.Now().UTC().Format(time.RFC3339)
	var results []NewsResult
	pos := 0
	doc.Find("ul.news-list li, .cm-news-list li").Each(func(_ int, s *goquery.Selection) {
		pos++
		anchor := s.Find("h3.news-title a, .news-title a")
		title := strings.TrimSpace(anchor.Text())
		if title == "" {
			return
		}
		href, _ := anchor.Attr("href")
		snippet := strings.TrimSpace(s.Find("p.news-content, div.news-abstract, .news-summary").Text())
		source := strings.TrimSpace(s.Find("a.news-from, .news-source").Text())
		timeStr := strings.TrimSpace(s.Find("time, span.news-time, .pub-time").Text())
		thumbURL, _ := s.Find("div.news-thumb img, .news-img img").Attr("src")
		pubAt := parseNewsTime(timeStr)
		results = append(results, NewsResult{
			Query:       query,
			Position:    pos,
			URL:         href,
			Title:       title,
			Snippet:     snippet,
			Source:      source,
			ThumbURL:    thumbURL,
			PublishedAt: pubAt,
			FetchedAt:   now,
		})
	})
	return results
}

type hotResponse struct {
	Code int `json:"code"`
	Data struct {
		HotWords []struct {
			Word   string `json:"word"`
			Weight int    `json:"weight"`
			Tag    string `json:"tag"`
		} `json:"hot_words"`
	} `json:"data"`
}

func parseHotItems(body []byte) []HotItem {
	var resp hotResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil
	}
	var items []HotItem
	for i, hw := range resp.Data.HotWords {
		if hw.Word == "" {
			continue
		}
		items = append(items, HotItem{
			Rank:   i + 1,
			Word:   hw.Word,
			Tag:    hw.Tag,
			Weight: hw.Weight,
		})
	}
	return items
}

func parseSuggestions(body []byte, key string, limit int) []Suggestion {
	// Try wrapped object form first: {"result":[{"key":"..."},...]}
	var wrapped struct {
		Result []struct {
			Key string `json:"key"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && len(wrapped.Result) > 0 {
		var out []Suggestion
		for _, r := range wrapped.Result {
			if r.Key == "" {
				continue
			}
			out = append(out, Suggestion{Query: key, Text: r.Key})
			if limit > 0 && len(out) >= limit {
				break
			}
		}
		return out
	}
	// Fall back to array-of-arrays form: [[text, score], ...]
	var raw [][]json.RawMessage
	if err := json.Unmarshal(body, &raw); err == nil {
		var out []Suggestion
		for _, pair := range raw {
			if len(pair) == 0 {
				continue
			}
			var text string
			if err2 := json.Unmarshal(pair[0], &text); err2 != nil || text == "" {
				continue
			}
			score := 0
			if len(pair) >= 2 {
				var sv string
				if err2 := json.Unmarshal(pair[1], &sv); err2 == nil {
					score, _ = strconv.Atoi(sv)
				} else {
					_ = json.Unmarshal(pair[1], &score)
				}
			}
			out = append(out, Suggestion{Query: key, Text: text, Score: score})
			if limit > 0 && len(out) >= limit {
				break
			}
		}
		return out
	}
	return nil
}

var docIDRE = regexp.MustCompile(`/doc/(\d+)\.html`)

func parseEncyclopediaResults(doc *goquery.Document, query string) []EncyclopediaResult {
	now := time.Now().UTC().Format(time.RFC3339)
	var results []EncyclopediaResult
	pos := 0
	doc.Find("ul.baike-results li, .baike-list li, .result-list li").Each(func(_ int, s *goquery.Selection) {
		pos++
		anchor := s.Find("h3 a, .result-title a")
		title := strings.TrimSpace(anchor.Text())
		if title == "" {
			return
		}
		href, _ := anchor.Attr("href")
		articleID := ""
		if m := docIDRE.FindStringSubmatch(href); m != nil {
			articleID = m[1]
		}
		fullURL := href
		if !strings.HasPrefix(href, "http") {
			fullURL = BaikeBaseURL + href
		}
		summary := strings.TrimSpace(s.Find("p.summary, .abstract, .desc").Text())
		results = append(results, EncyclopediaResult{
			Query:     query,
			Position:  pos,
			ArticleID: articleID,
			URL:       fullURL,
			Title:     title,
			Summary:   summary,
			FetchedAt: now,
		})
	})
	return results
}

// --- helpers ---

func encodeQuery(q string) string {
	return strings.ReplaceAll(q, " ", "+")
}

// parseNewsTime converts a 360 News relative/absolute date string to RFC3339 UTC.
// Returns empty string if unparseable.
func parseNewsTime(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	switch {
	case strings.HasSuffix(s, "小时前"):
		n, _ := strconv.Atoi(strings.TrimSuffix(s, "小时前"))
		return now.Add(-time.Duration(n) * time.Hour).UTC().Format(time.RFC3339)
	case strings.HasPrefix(s, "今天"):
		// "今天 14:30"
		rest := strings.TrimSpace(strings.TrimPrefix(s, "今天"))
		t, err := time.ParseInLocation("15:04", rest, loc)
		if err == nil {
			full := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, loc)
			return full.UTC().Format(time.RFC3339)
		}
		return now.UTC().Format(time.RFC3339)
	default:
		// Try absolute formats
		for _, layout := range []string{"2006-01-02", "01-02", "1-2"} {
			t, err := time.ParseInLocation(layout, s, loc)
			if err == nil {
				if t.Year() == 0 {
					t = t.AddDate(now.Year(), 0, 0)
				}
				return t.UTC().Format(time.RFC3339)
			}
		}
	}
	return ""
}
