package cn360

import (
	"context"
	"strings"
	"time"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes cn360 as a kit Domain so a multi-domain host enables it
// with a single blank import:
//
//	import _ "github.com/tamnd/360cn-cli/cn360"
//
// The same Domain builds the standalone cn360 binary (see cli/root.go).
func init() { kit.Register(Domain{}) }

// Domain is the 360 Search driver. It carries no state; the per-run client is
// built by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme, hosts, and binary identity.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "cn360",
		Hosts:  []string{SearchHost, "www.so.com", "news.so.com", "baike.so.com"},
		Identity: kit.Identity{
			Binary: "cn360",
			Short:  "360 Search (so.com) and Encyclopedia from the terminal",
			Long: `cn360 turns 360 Search's public surfaces into a fast, scriptable command line.

Fetch web search results, news articles, hot search terms, autocomplete
suggestions, and 360 Encyclopedia entries -- all over plain HTTPS, no API key.

Quick start:
  cn360 search "golang tutorial"    web search results
  cn360 suggest "go lang"           autocomplete suggestions
  cn360 hot                         today's hot searches
  cn360 news "Go 语言"              news search
  cn360 encyclopedia "围棋"         encyclopedia article search`,
			Site: SearchHost,
			Repo: "https://github.com/tamnd/360cn-cli",
		},
	}
}

// Register installs the client factory and all 360 Search operations onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{
		Name:    "search",
		Group:   "search",
		Summary: "Web search results from so.com",
		Args: []kit.Arg{
			{Name: "query", Help: "search query"},
		},
	}, searchOp)

	kit.Handle(app, kit.OpMeta{
		Name:    "news",
		Group:   "search",
		Summary: "News search results from news.so.com",
		Args: []kit.Arg{
			{Name: "query", Help: "search query"},
		},
	}, newsOp)

	kit.Handle(app, kit.OpMeta{
		Name:    "hot",
		Group:   "search",
		Summary: "Hot search terms from 360 Search",
	}, hotOp)

	kit.Handle(app, kit.OpMeta{
		Name:    "suggest",
		Group:   "search",
		Summary: "Autocomplete suggestions from sug.so.com",
		Args: []kit.Arg{
			{Name: "query", Help: "query prefix"},
		},
	}, suggestOp)

	kit.Handle(app, kit.OpMeta{
		Name:    "encyclopedia",
		Group:   "search",
		Summary: "Encyclopedia article search on baike.so.com",
		Args: []kit.Arg{
			{Name: "query", Help: "search query"},
		},
	}, encyclopediaOp)
}

// newClient builds a Client from the kit Config resolved at startup.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClientWithConfig(c), nil
}

// --- input structs ---

type searchInput struct {
	Query  string  `kit:"arg"          help:"search query"`
	Page   int     `kit:"flag"         help:"result page (1-based)" default:"1"`
	Pages  int     `kit:"flag"         help:"number of pages to fetch" default:"1"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type newsInput struct {
	Query  string  `kit:"arg"          help:"search query"`
	Page   int     `kit:"flag"         help:"result page (1-based)" default:"1"`
	Pages  int     `kit:"flag"         help:"number of pages to fetch" default:"1"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type hotInput struct {
	Limit  int     `kit:"flag,inherit" help:"max items"`
	Client *Client `kit:"inject"`
}

type suggestInput struct {
	Query  string  `kit:"arg"          help:"query prefix"`
	Limit  int     `kit:"flag,inherit" help:"max suggestions"`
	Client *Client `kit:"inject"`
}

type encyclopediaInput struct {
	Query  string  `kit:"arg"          help:"search query"`
	Page   int     `kit:"flag"         help:"result page (1-based)" default:"1"`
	Pages  int     `kit:"flag"         help:"number of pages to fetch" default:"1"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func searchOp(ctx context.Context, in searchInput, emit func(SearchResult) error) error {
	n := 0
	pages := in.Pages
	if pages < 1 {
		pages = 1
	}
	for p := in.Page; p < in.Page+pages; p++ {
		results, err := in.Client.Search(ctx, in.Query, p)
		if err != nil {
			return mapErr(err)
		}
		if len(results) == 0 {
			break
		}
		for _, r := range results {
			if err := emit(r); err != nil {
				return err
			}
			n++
			if in.Limit > 0 && n >= in.Limit {
				return nil
			}
		}
		if p < in.Page+pages-1 {
			time.Sleep(in.Client.cfg.Rate)
		}
	}
	return nil
}

func newsOp(ctx context.Context, in newsInput, emit func(NewsResult) error) error {
	n := 0
	pages := in.Pages
	if pages < 1 {
		pages = 1
	}
	for p := in.Page; p < in.Page+pages; p++ {
		results, err := in.Client.News(ctx, in.Query, p)
		if err != nil {
			return mapErr(err)
		}
		if len(results) == 0 {
			break
		}
		for _, r := range results {
			if err := emit(r); err != nil {
				return err
			}
			n++
			if in.Limit > 0 && n >= in.Limit {
				return nil
			}
		}
		if p < in.Page+pages-1 {
			time.Sleep(in.Client.cfg.Rate)
		}
	}
	return nil
}

func hotOp(ctx context.Context, in hotInput, emit func(HotItem) error) error {
	items, err := in.Client.Hot(ctx, in.Limit)
	if err != nil {
		return mapErr(err)
	}
	for _, h := range items {
		if err := emit(h); err != nil {
			return err
		}
	}
	return nil
}

func suggestOp(ctx context.Context, in suggestInput, emit func(Suggestion) error) error {
	sugs, err := in.Client.Suggest(ctx, in.Query, in.Limit)
	if err != nil {
		return mapErr(err)
	}
	for _, s := range sugs {
		if err := emit(s); err != nil {
			return err
		}
	}
	return nil
}

func encyclopediaOp(ctx context.Context, in encyclopediaInput, emit func(EncyclopediaResult) error) error {
	n := 0
	pages := in.Pages
	if pages < 1 {
		pages = 1
	}
	for p := in.Page; p < in.Page+pages; p++ {
		results, err := in.Client.Encyclopedia(ctx, in.Query, p)
		if err != nil {
			return mapErr(err)
		}
		if len(results) == 0 {
			break
		}
		for _, r := range results {
			if err := emit(r); err != nil {
				return err
			}
			n++
			if in.Limit > 0 && n >= in.Limit {
				return nil
			}
		}
		if p < in.Page+pages-1 {
			time.Sleep(in.Client.cfg.Rate)
		}
	}
	return nil
}

// --- Resolver (URI driver) ---

// Classify turns a 360 URL into (uriType, id) for the kit URI driver.
func (Domain) Classify(input string) (uriType, id string, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", errs.Usage("cn360: empty input")
	}
	if !strings.HasPrefix(input, "http") {
		return "search", input, nil
	}
	return "search", input, nil
}

// Locate returns the canonical URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "search":
		return SearchBaseURL + "/s?q=" + encodeQuery(id), nil
	default:
		return "", errs.Usage("cn360: no resource type %q", uriType)
	}
}

// mapErr translates library errors to kit error kinds with the right exit codes.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	var ce *CodeError
	if isCodeErr(err, &ce) {
		switch ce.Code {
		case ExitNotFound:
			return errs.NotFound("%s", ce.Msg)
		case ExitBlocked:
			return errs.RateLimited("%s", ce.Msg)
		}
	}
	return err
}

func isCodeErr(err error, target **CodeError) bool {
	if ce, ok := err.(*CodeError); ok {
		*target = ce
		return true
	}
	return false
}
