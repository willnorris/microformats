// Copyright (c) 2015 Andy Leap, Google
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package ptd

import (
	"testing"

	"willnorris.com/go/microformats"
)

// shortcut for "property map"
type pm map[string][]interface{}

func Test_NilItems(t *testing.T) {
	want := ""
	if got := DiscoverPostType(nil); got != want {
		t.Errorf("DiscoverPostType(nil) returned %q, want %q", got, want)
	}
	if got := DiscoverResponseType(nil); got != want {
		t.Errorf("DiscoverResponseType(nil) returned %q, want %q", got, want)
	}
}

func Test_DiscoverPostType_Type(t *testing.T) {
	item := &microformats.Microformat{Type: []string{"h-event"}}
	if got, want := DiscoverResponseType(item), "event"; got != want {
		//t.Errorf("DiscoverResponseType(%v) returned %q, want %q", item, got, want)
	}
}

func Test_DiscoverPostType_Properties(t *testing.T) {
	tests := []struct {
		properties pm
		want       string
	}{
		{nil, "note"},
		{pm{}, "note"},

		// RSVP values
		{pm{"rsvp": {}}, "note"},
		{pm{"rsvp": {""}}, "note"},
		{pm{"rsvp": {"yes"}}, "rsvp"},
		{pm{"rsvp": {"no"}}, "rsvp"},
		{pm{"rsvp": {"maybe"}}, "rsvp"},
		{pm{"rsvp": {"interested"}}, "rsvp"},
		{pm{"rsvp": {"meh"}}, "note"},
		{pm{"rsvp": {"yes", "well, not sure"}}, "rsvp"},
		{pm{"rsvp": {"probably", "most likely", "yes"}}, "rsvp"},

		// URL properties
		{pm{"repost-of": {}}, "note"},
		{pm{"repost-of": {""}}, "note"},
		{pm{"repost-of": {"foo"}}, "repost"},
		{pm{"like-of": {}}, "note"},
		{pm{"like-of": {""}}, "note"},
		{pm{"like-of": {"foo"}}, "like"},
		{pm{"in-reply-to": {}}, "note"},
		{pm{"in-reply-to": {""}}, "note"},
		{pm{"in-reply-to": {"foo"}}, "reply"},

		// content and name variations
		{pm{"content": {"foo"}, "name": {"foo"}}, "note"},
		{pm{"summary": {"foo"}, "name": {"foo"}}, "note"},
		{pm{"summary": {"foo"}, "name": {"Foo"}}, "article"},
		{pm{"content": {"foo"}, "name": {"bar"}}, "article"},
		{pm{"content": {"foo"}, "name": {"bar"}}, "article"},
		{pm{"content": {"foo"}, "summary": {"bar"}, "name": {"bar"}}, "article"},
		{pm{"content": {"foo \t\n bar"}, "name": {" foo bar "}}, "note"},
	}

	for _, tt := range tests {
		item := &microformats.Microformat{Properties: tt.properties}
		if got, want := DiscoverPostType(item), tt.want; got != want {
			t.Errorf("DiscoverPostType(%v) returned %q, want %q", tt.properties, got, want)
		}
	}
}

func Test_DiscoverResponseType(t *testing.T) {
	tests := []struct {
		properties pm
		want       string
	}{
		{nil, "mention"},
		{pm{}, "mention"},

		// RSVP values
		{pm{"rsvp": {}}, "mention"},
		{pm{"rsvp": {""}}, "mention"},
		{pm{"rsvp": {"yes"}}, "rsvp"},
		{pm{"rsvp": {"no"}}, "rsvp"},
		{pm{"rsvp": {"maybe"}}, "rsvp"},
		{pm{"rsvp": {"interested"}}, "rsvp"},
		{pm{"rsvp": {"meh"}}, "mention"},
		{pm{"rsvp": {"yes", "well, not sure"}}, "rsvp"},
		{pm{"rsvp": {"probably", "most likely", "yes"}}, "rsvp"},

		// URL properties
		{pm{"repost-of": {}}, "mention"},
		{pm{"repost-of": {""}}, "mention"},
		{pm{"repost-of": {"foo"}}, "repost"},
		{pm{"like-of": {}}, "mention"},
		{pm{"like-of": {""}}, "mention"},
		{pm{"like-of": {"foo"}}, "like"},
		{pm{"in-reply-to": {}}, "mention"},
		{pm{"in-reply-to": {""}}, "mention"},
		{pm{"in-reply-to": {"foo"}}, "reply"},
	}

	for _, tt := range tests {
		item := &microformats.Microformat{Properties: tt.properties}
		if got, want := DiscoverResponseType(item), tt.want; got != want {
			t.Errorf("DiscoverResponseType(%v) returned %q, want %q", tt.properties, got, want)
		}
	}
}
