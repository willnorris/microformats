package rhc

import (
	"net/url"
	"strings"
	"testing"

	"willnorris.com/go/microformats"
)

//nolint:funlen
func TestRepresentativeHcard(t *testing.T) {
	tests := []struct {
		name string
		html string
	}{
		{
			name: "url+uid",
			html: `
			<p class="h-card"><a href="/" class="u-url p-name">only url</a></p>
			<p class="h-card"><a href="/" class="u-uid p-name">only uid</a></p>
			<p class="h-card"><a href="/" class="u-url u-uid p-name">rhc</a></p>`,
		},
		{
			name: "rel-me",
			html: `
			<link rel="me" href="http://me/">
			<p class="h-card"><a href="/" class="u-url p-name">only url</a></p>
			<p class="h-card"><a href="/" class="u-uid p-name">only uid</a></p>
			<p class="h-card"><a href="http://me/" p-name">rhc</a></p>`,
		},
		{
			name: "single-url-match",
			html: `
			<p class="h-card"><a href="/" p-name">rhc</a></p>`,
		},
		{
			name: "nested microformat",
			html: `
			<div class="h-feed">
			  <div class="h-entry">
			    <p class="p-author h-card"><a href="/" p-name">rhc</a></p>
			  </div>
			</div>`,
		},
	}

	srcURL, _ := url.Parse("http://example.com")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := microformats.Parse(strings.NewReader(tt.html), srcURL)
			got := RepresentativeHcard(data, srcURL.String())
			if got == nil {
				t.Errorf("RepresentativeHcard() = nil")
			} else if got.Properties["name"][0] != "rhc" {
				t.Errorf("RepresentativeHcard() = %v", got)
			}
		})
	}

	// additional tests for zero values
	if got := RepresentativeHcard(nil, srcURL.String()); got != nil {
		t.Errorf("RepresentativeHcard() = %v, want nil", got)
	}
	if got := RepresentativeHcard(&microformats.Data{}, srcURL.String()); got != nil {
		t.Errorf("RepresentativeHcard() = %v, want nil", got)
	}
	if got := RepresentativeHcard(&microformats.Data{Items: []*microformats.Microformat{}}, srcURL.String()); got != nil {
		t.Errorf("RepresentativeHcard() = %v, want nil", got)
	}
	if got := RepresentativeHcard(&microformats.Data{Items: []*microformats.Microformat{{}}}, ""); got != nil {
		t.Errorf("RepresentativeHcard() = %v, want nil", got)
	}
	if got := RepresentativeHcard(&microformats.Data{Items: []*microformats.Microformat{{}}}, srcURL.String()); got != nil {
		t.Errorf("RepresentativeHcard() = %v, want nil", got)
	}
}

func TestURLMatch(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"", "", false},   // empty input
		{"a", "", false},  // empty input
		{"", "b", false},  // empty input
		{"%", "b", false}, // url.Parse error
		{"a", "%", false}, // url.Parse error
		{"a", "b", false}, // mismatched inputs

		{"a", "a", true},
		{"http://x", "http://x/", true}, // missing trailing slash
		{"http://x/", "http://x", true}, // missing trailing slash
	}

	for _, tt := range tests {
		if got := urlMatch(tt.a, tt.b); got != tt.want {
			t.Errorf("urlMatch(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
