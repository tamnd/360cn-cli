package cn360

import (
	"testing"
)

// These tests exercise the URI driver's pure string functions (no network).
// Parser tests are in the individual *_test.go files.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "cn360" {
		t.Errorf("Scheme = %q, want cn360", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != SearchHost {
		t.Errorf("Hosts = %v, want first=%s", info.Hosts, SearchHost)
	}
	if info.Identity.Binary != "cn360" {
		t.Errorf("Identity.Binary = %q, want cn360", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct{ in, typ, id string }{
		{"golang tutorial", "search", "golang tutorial"},
		{"https://www.so.com/s?q=go", "search", "https://www.so.com/s?q=go"},
	}
	for _, tc := range cases {
		typ, id, err := Domain{}.Classify(tc.in)
		if err != nil || typ != tc.typ || id != tc.id {
			t.Errorf("Classify(%q) = (%q, %q, %v), want (%q, %q, nil)",
				tc.in, typ, id, err, tc.typ, tc.id)
		}
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("search", "golang")
	want := SearchBaseURL + "/s?q=golang"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}
}

func TestLocateUnknownType(t *testing.T) {
	_, err := Domain{}.Locate("unknown", "foo")
	if err == nil {
		t.Error("expected error for unknown type, got nil")
	}
}
