package cn360

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuggestWrappedForm(t *testing.T) {
	fixture := `{"result":[{"key":"golang"},{"key":"go 语言"},{"key":"go 教程"}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fixture))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.Rate = 0
	cfg.SugBaseURL = srv.URL
	c := NewClientWithConfig(cfg)

	sugs, err := c.SuggestURL(context.Background(), srv.URL, "go", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(sugs) != 3 {
		t.Fatalf("got %d suggestions, want 3", len(sugs))
	}
	if sugs[0].Text != "golang" {
		t.Errorf("sugs[0].Text = %q, want golang", sugs[0].Text)
	}
}

func TestSuggestArrayForm(t *testing.T) {
	fixture := `[["golang",1000],["go 教程",950],["go 面试题",800]]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fixture))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.Rate = 0
	cfg.SugBaseURL = srv.URL
	c := NewClientWithConfig(cfg)

	sugs, err := c.SuggestURL(context.Background(), srv.URL, "go", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(sugs) != 3 {
		t.Fatalf("got %d suggestions, want 3", len(sugs))
	}
	if sugs[0].Text != "golang" {
		t.Errorf("sugs[0].Text = %q, want golang", sugs[0].Text)
	}
	if sugs[0].Score != 1000 {
		t.Errorf("sugs[0].Score = %d, want 1000", sugs[0].Score)
	}
}

func TestSuggestLimit(t *testing.T) {
	fixture := `{"result":[{"key":"a"},{"key":"b"},{"key":"c"},{"key":"d"}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fixture))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.Rate = 0
	cfg.SugBaseURL = srv.URL
	c := NewClientWithConfig(cfg)

	sugs, err := c.SuggestURL(context.Background(), srv.URL, "x", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(sugs) != 2 {
		t.Errorf("got %d suggestions with limit=2, want 2", len(sugs))
	}
}
