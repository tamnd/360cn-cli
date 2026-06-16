package cn360

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const newsFixture = `<html><body>
<ul class="news-list">
  <li>
    <h3 class="news-title"><a href="https://news.example.com/1">Go 1.23 正式发布</a></h3>
    <p class="news-content">Go 团队正式发布 Go 1.23 版本，带来多项改进...</p>
    <a class="news-from">InfoQ</a>
    <time>2024-08-15</time>
  </li>
  <li>
    <h3 class="news-title"><a href="https://news.example.com/2">Go 语言性能测试</a></h3>
    <p class="news-content">最新基准测试显示 Go 在并发场景下...</p>
    <a class="news-from">掘金</a>
    <span class="news-time">2小时前</span>
  </li>
</ul>
</body></html>`

func TestNewsParse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(newsFixture))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.Rate = 0
	cfg.NewsBaseURL = srv.URL
	c := NewClientWithConfig(cfg)

	results, err := c.NewsURL(context.Background(), srv.URL+"/ns?q=go", "go")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Title != "Go 1.23 正式发布" {
		t.Errorf("result[0].Title = %q, want %q", results[0].Title, "Go 1.23 正式发布")
	}
	if results[0].Source != "InfoQ" {
		t.Errorf("result[0].Source = %q, want InfoQ", results[0].Source)
	}
	if results[1].Source != "掘金" {
		t.Errorf("result[1].Source = %q, want 掘金", results[1].Source)
	}
	// result[1] uses relative time "2小时前"; published_at should be non-empty
	if results[1].PublishedAt == "" {
		t.Error("result[1].PublishedAt is empty, want non-empty")
	}
}
