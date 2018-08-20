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

package microformats

import (
	"reflect"
	"sort"
	"testing"
)

func Test_BackcompatRootClasses(t *testing.T) {
	tests := []struct {
		classes, want []string
	}{
		{nil, nil},
		{[]string{""}, nil},
		{[]string{"foo"}, nil},
		{[]string{"adr"}, []string{"h-adr"}},
		{[]string{"adr", "foo"}, []string{"h-adr"}},
		{[]string{"adr", "vcard"}, []string{"h-adr", "h-card"}},
	}

	for _, tt := range tests {
		got := backcompatRootClasses(tt.classes)
		if want := tt.want; !reflect.DeepEqual(got, want) {
			t.Errorf("backcompatRootClasses(%q) returned %q, want %q)", tt.classes, got, want)
		}
	}
}

func Test_BackcompatPropertyClasses(t *testing.T) {
	tests := []struct {
		classes []string
		rels    []string
		context []string // microformat type that property appears in
		want    []string
	}{
		{nil, nil, nil, nil},
		{[]string{""}, nil, nil, nil},
		{[]string{"foo"}, nil, nil, nil},
		{[]string{"fn"}, nil, []string{"h-card"}, []string{"p-name"}},
		{[]string{"fn", "foo"}, nil, []string{"h-card"}, []string{"p-name"}},
		{[]string{"fn", "email"}, nil, []string{"h-card"}, []string{"p-name", "u-email"}},

		// itemtype-specific property mappings
		{[]string{"summary"}, nil, []string{"h-entry"}, []string{"p-summary"}},
		{[]string{"summary"}, nil, []string{"h-event"}, []string{"p-name"}},

		// duplicate properties
		{[]string{"summary"}, nil, []string{"h-entry", "h-resume"}, []string{"p-summary"}},
		{[]string{"summary"}, nil, []string{"h-entry", "h-event"}, []string{"p-summary", "p-name"}},

		// rels
		{nil, []string{"bookmark"}, nil, nil},
		{nil, []string{"bookmark"}, []string{"h-entry"}, []string{"u-url"}},
		{[]string{"category"}, []string{"tag"}, []string{"h-card"}, []string{"u-category"}},
	}

	for _, tt := range tests {
		got := backcompatPropertyClasses(tt.classes, tt.rels, tt.context)
		sort.Strings(got)
		sort.Strings(tt.want)
		if want := tt.want; !reflect.DeepEqual(got, want) {
			t.Errorf("backcompatPropertyClasses(%q) returned %q, want %q)", tt.classes, got, want)
		}
	}
}

func Test_BackcompatURLCategory(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"", ""},
		{"a", "a"},
		{"a/b", "b"},
		{"/a/b", "b"},
		{"/a/b/", "b"},
		{"http://example.com/a/b", "b"},
	}

	for _, tt := range tests {
		got := backcompatURLCategory(tt.url)
		if want := tt.want; got != want {
			t.Errorf("backcompatURLCategory(%q) returned %q, want %q)", tt.url, got, want)
		}
	}
}
