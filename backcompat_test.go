// Copyright (c) The microformats project authors.
// SPDX-License-Identifier: MIT

package microformats

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/net/html"
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
		got := backcompatRootClasses(tt.classes, nil)
		if want := tt.want; !cmp.Equal(got, want) {
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
		if want := tt.want; !cmp.Equal(got, want) {
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
		{"%", "%"}, // invalid URL
	}

	for _, tt := range tests {
		got := backcompatURLCategory(tt.url)
		if want := tt.want; got != want {
			t.Errorf("backcompatURLCategory(%q) returned %q, want %q)", tt.url, got, want)
		}
	}
}

func Test_BackcompatIncludeRefs(t *testing.T) {
	tests := []struct {
		html        string
		wantRefs    []string
		wantReplace bool
	}{
		{`<a></a>`, nil, false},
		{`<a href="#foo"></a>`, nil, false},
		{`<a class="include" href="foo"></a>`, nil, false},
		{`<a class="include" href="#"></a>`, nil, false},
		{
			`<object class="include" data="#foo">`,
			[]string{"foo"},
			true,
		},
		{
			`<a class="include" href="#foo"></a>`,
			[]string{"foo"},
			true,
		},
		{
			`<a itemref="foo"></a>`,
			[]string{"foo"},
			false,
		},
		{
			`<a itemref="foo bar"></a>`,
			[]string{"foo", "bar"},
			false,
		},
	}

	for _, tt := range tests {
		p := &parser{}
		node, _ := parseNode(tt.html)
		refs, replace := p.backcompatIncludeRefs(node)
		if !cmp.Equal(refs, tt.wantRefs) {
			t.Errorf("backcompatIncludeRefs(%v) returned refs %v, want %v", tt.html, refs, tt.wantRefs)
		}
		if replace != tt.wantReplace {
			t.Errorf("backcompatIncludeRefs(%v) returned replace %t, want %t", tt.html, replace, tt.wantReplace)
		}
	}
}

func Test_BackcompatIncludeNode(t *testing.T) {
	n, _ := parseNode("<p></p>")

	tests := []struct {
		node    *html.Node
		refs    []string
		replace bool
		want    *html.Node
	}{
		{n, []string{}, false, n},
	}

	for _, tt := range tests {
		p := &parser{}
		got := p.backcompatIncludeNode(tt.node, tt.refs, tt.replace)
		if want := tt.want; !cmp.Equal(got, want) {
			t.Errorf("backcompatIncludeNode(%v, %v, %v) returned %v, want %v", tt.node, tt.refs, tt.replace, got, want)
		}
	}
}

func Test_IsAncestorNode(t *testing.T) {
	a1, _ := parseNode("<p><b></b></p>")
	a2 := a1.FirstChild
	b1, _ := parseNode("<div></div>")

	tests := []struct {
		parent, child *html.Node
		want          bool // expected return from from isAncestorNode
	}{
		{nil, nil, false},
		{a1, nil, false},
		{nil, a1, false},

		{a1, a1, true},
		{a1, a2, true},
		{a2, a1, false},
		{a1, b1, false},
	}

	for _, tt := range tests {
		got := isAncestorNode(tt.child, tt.parent)
		if got != tt.want {
			t.Errorf("isAncestorNode(%v, %v) returned %v, want %v", tt.child, tt.parent, got, tt.want)
		}
	}
}
