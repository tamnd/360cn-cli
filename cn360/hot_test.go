package cn360

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const hotFixture = `{"code":0,"data":{"hot_words":[{"word":"世界杯","weight":9999,"tag":"热"},{"word":"高考录取","weight":8000,"tag":""},{"word":"台风天气","weight":7500,"tag":"新"}]}}`

func TestHotParse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(hotFixture))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.Rate = 0
	cfg.HotAPIURL = srv.URL
	c := NewClientWithConfig(cfg)

	items, err := c.HotURL(context.Background(), srv.URL, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}
	if items[0].Word != "世界杯" {
		t.Errorf("items[0].Word = %q, want 世界杯", items[0].Word)
	}
	if items[0].Tag != "热" {
		t.Errorf("items[0].Tag = %q, want 热", items[0].Tag)
	}
	if items[0].Rank != 1 {
		t.Errorf("items[0].Rank = %d, want 1", items[0].Rank)
	}
	if items[0].Weight != 9999 {
		t.Errorf("items[0].Weight = %d, want 9999", items[0].Weight)
	}
}

func TestHotLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(hotFixture))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.Rate = 0
	cfg.HotAPIURL = srv.URL
	c := NewClientWithConfig(cfg)

	items, err := c.HotURL(context.Background(), srv.URL, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Errorf("got %d items with limit=2, want 2", len(items))
	}
}
