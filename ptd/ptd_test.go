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
	if got := PostType(nil); got != want {
		t.Errorf("PostType(nil) returned %q, want %q", got, want)
	}
	if got := ResponseType(nil); got != want {
		t.Errorf("ResponseType(nil) returned %q, want %q", got, want)
	}
}

func Test_PostType_Type(t *testing.T) {
	item := &microformats.Microformat{Type: []string{"h-event"}}
	if got, want := PostType(item), "event"; got != want {
		t.Errorf("ResponseType(%v) returned %q, want %q", item, got, want)
	}
}

func Test_PostType_Properties(t *testing.T) {
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
		{pm{"video": {"foo"}}, "video"},
		{pm{"photo": {"foo"}}, "photo"},

		// content and name variations
		{pm{"content": {"foo"}}, "note"},
		{pm{"content": {"foo"}, "name": {"foo"}}, "note"},
		{pm{"summary": {"foo"}, "name": {"foo"}}, "note"},
		{pm{"content": {"foobar"}, "name": {"foo"}}, "note"},
		{pm{"summary": {"foo"}, "name": {"Foo"}}, "article"},
		{pm{"content": {"foo"}, "name": {"bar"}}, "article"},
		{pm{"content": {"foo"}, "name": {"bar"}}, "article"},
		{pm{"content": {"foo"}, "summary": {"bar"}, "name": {"bar"}}, "article"},
		{pm{"content": {"foo \t\n bar"}, "name": {" foo bar "}}, "note"},
	}

	for _, tt := range tests {
		item := &microformats.Microformat{Properties: tt.properties}
		if got, want := PostType(item), tt.want; got != want {
			t.Errorf("PostType(%v) returned %q, want %q", tt.properties, got, want)
		}
	}
}

func Test_ResponseType(t *testing.T) {
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
		if got, want := ResponseType(item), tt.want; got != want {
			t.Errorf("ResponseType(%v) returned %q, want %q", tt.properties, got, want)
		}
	}
}

func Test_ValidURL(t *testing.T) {
	tests := []struct {
		values []interface{}
		want   bool
	}{
		{[]interface{}{}, false},
		{[]interface{}{""}, false},
		{[]interface{}{"%"}, false},
		{[]interface{}{struct{}{}}, false},

		{[]interface{}{"a"}, true},
		{[]interface{}{"a", "b"}, true},
		{[]interface{}{"", "a"}, true},
		{[]interface{}{"%", "a"}, true},
	}

	for _, tt := range tests {
		if got, want := validURL(tt.values), tt.want; got != want {
			t.Errorf("validURL(%q) returned %v, want %v", tt.values, got, want)
		}
	}
}
