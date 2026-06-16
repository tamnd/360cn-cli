package cn360

// SearchResult is one result row from a 360 Search (so.com) SERP page.
type SearchResult struct {
	Query     string `json:"query"`
	Position  int    `json:"position"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	Snippet   string `json:"snippet"`
	SiteName  string `json:"site_name"`
	IsAd      bool   `json:"is_ad"`
	FetchedAt string `json:"fetched_at"`
}

// NewsResult is one news article from a 360 News (news.so.com) search page.
type NewsResult struct {
	Query       string `json:"query"`
	Position    int    `json:"position"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Snippet     string `json:"snippet"`
	Source      string `json:"source"`
	ThumbURL    string `json:"thumb_url"`
	PublishedAt string `json:"published_at"` // RFC3339 UTC or empty
	FetchedAt   string `json:"fetched_at"`
}

// HotItem is one trending search term from 360 Search's hot-search API.
type HotItem struct {
	Rank   int    `json:"rank"`
	Word   string `json:"word"`
	Tag    string `json:"tag"`    // "热", "新", "爆", or empty
	Weight int    `json:"weight"` // relative popularity score
}

// Suggestion is one autocomplete suggestion returned by sug.so.com.
type Suggestion struct {
	Query string `json:"query"`
	Text  string `json:"text"`
	Score int    `json:"score"`
}

// EncyclopediaResult is one search result stub from baike.so.com.
type EncyclopediaResult struct {
	Query     string `json:"query"`
	Position  int    `json:"position"`
	ArticleID string `json:"article_id"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	FetchedAt string `json:"fetched_at"`
}
