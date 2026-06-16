package cn360

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const searchFixture = `<html><body>
<ul id="results">
  <li class="res-list" id="result-1">
    <h3 class="res-title"><a href="https://golang.org/" data-mdurl="https://golang.org/">Go 语言官网</a></h3>
    <p class="res-desc">Go 是一种开源编程语言，旨在提高程序员的工作效率...</p>
    <div class="res-linkinfo"><cite>golang.org</cite></div>
  </li>
  <li class="res-list res-list-ad" id="result-2">
    <h3 class="res-title"><a href="https://ad.example.com/" data-mdurl="https://ad.example.com/">Go 培训班</a></h3>
    <span class="ad-label">广告</span>
    <p class="res-desc">专业 Go 培训...</p>
  </li>
</ul>
</body></html>`

func TestSearchParse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(searchFixture))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.Rate = 0
	cfg.SearchBaseURL = srv.URL
	c := NewClientWithConfig(cfg)

	results, err := c.SearchURL(context.Background(), srv.URL+"/s?q=go", "go")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Title != "Go 语言官网" {
		t.Errorf("result[0].Title = %q, want %q", results[0].Title, "Go 语言官网")
	}
	if results[0].IsAd {
		t.Error("result[0].IsAd = true, want false")
	}
	if !results[1].IsAd {
		t.Error("result[1].IsAd = false, want true")
	}
	if results[0].URL != "https://golang.org/" {
		t.Errorf("result[0].URL = %q, want https://golang.org/", results[0].URL)
	}
}
