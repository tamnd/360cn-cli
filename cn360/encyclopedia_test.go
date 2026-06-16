package cn360

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const encyclopediaFixture = `<html><body>
<ul class="baike-results">
  <li>
    <h3><a href="/doc/98765432.html">围棋</a></h3>
    <p class="summary">围棋是一种策略性棋盘游戏，起源于中国，有超过2500年历史...</p>
  </li>
  <li>
    <h3><a href="/doc/12345678.html">围棋规则</a></h3>
    <p class="summary">围棋的基本规则包括落子、提子、禁着点...</p>
  </li>
</ul>
</body></html>`

func TestEncyclopediaParse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(encyclopediaFixture))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.Rate = 0
	cfg.BaikeBaseURL = srv.URL
	c := NewClientWithConfig(cfg)

	results, err := c.EncyclopediaURL(context.Background(), srv.URL+"/search?q=围棋", "围棋")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Title != "围棋" {
		t.Errorf("results[0].Title = %q, want 围棋", results[0].Title)
	}
	if results[0].ArticleID != "98765432" {
		t.Errorf("results[0].ArticleID = %q, want 98765432", results[0].ArticleID)
	}
	if results[0].Summary == "" {
		t.Error("results[0].Summary is empty")
	}
}
