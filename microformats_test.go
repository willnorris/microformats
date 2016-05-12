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
	"regexp"
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var body = &html.Node{Type: html.ElementNode, Data: "body", DataAtom: atom.Body}

// parse the first HTML node found in s.
func parseNode(s string) (n *html.Node, err error) {
	nodes, err := html.ParseFragment(strings.NewReader(s), body)
	if len(nodes) > 0 {
		n = nodes[0]
	}
	return n, err
}

func Test_GetClasses(t *testing.T) {
	tests := []struct {
		html    string
		classes []string
	}{
		{``, nil},
		{`<img>`, nil},
		{`<img class_="a">`, nil},

		{`<img class="a">`, []string{"a"}},
		{`<img class="a b">`, []string{"a", "b"}},
		{`<img class="http://example.com/ b">`, []string{"http://example.com/", "b"}},

		{`<img CLASS="a">`, []string{"a"}},
	}

	for _, tt := range tests {
		n, err := parseNode(tt.html)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		if got, want := getClasses(n), tt.classes; !reflect.DeepEqual(got, want) {
			t.Errorf("getClasses(%q) returned %v, want %v", tt.html, got, want)
		}
	}
}

func Test_HasMatchingClass(t *testing.T) {
	root, prop := rootClassNames.String(), propertyClassNames.String()

	tests := []struct {
		html  string
		regex string
		match bool
	}{
		{``, ``, false},
		{`<img>`, ``, false},
		{`<img>`, `.*`, false},

		{`<img class="">`, `.*`, true},
		{`<img class="a">`, `.+`, true},

		// root class names
		{`<img class="h-card">`, root, true},
		{`<img class="h-card vcard">`, root, true},
		{`<img class="vcard h-card">`, root, true},
		{`<img class="hcard">`, root, false},
		{`<img class="p-name">`, root, false},

		// property class names
		{`<img class="p-name">`, prop, true},
		{`<img class="u-url">`, prop, true},
		{`<img class="dt-updated">`, prop, true},
		{`<img class="e-content">`, prop, true},
		{`<img class="p-name name">`, prop, true},
		{`<img class="h-card">`, prop, false},
		{`<img class="pname">`, prop, false},
	}

	for _, tt := range tests {
		n, err := parseNode(tt.html)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		r, err := regexp.Compile(tt.regex)
		if err != nil {
			t.Fatalf("Error compiling regex: %v", err)
		}

		if got, want := hasMatchingClass(n, r), tt.match; got != want {
			t.Errorf("hasMatchingClass(%q, %q) returned %t, want %t", tt.html, tt.regex, got, want)
		}
	}
}

// test both getAttr and getAttrPtr
func Test_GetAttr(t *testing.T) {
	ptr := func(s string) *string { return &s }

	tests := []struct {
		html, attr string
		value      string  // string valuer returned by getAttr
		ptr        *string // pointer value returned by getAttrPtr
	}{
		{``, "", "", nil},
		{`<img>`, "", "", nil},
		{`<img>`, "src", "", nil},
		{`<img src>`, "src", "", ptr("")},
		{`<img src="a">`, "src", "a", ptr("a")},
		{`<img src="a">`, "SRC", "a", ptr("a")},
		{`<img SRC="a">`, "src", "a", ptr("a")},
		{`<img src="a" src="b">`, "src", "a", ptr("a")},
	}

	for _, tt := range tests {
		n, err := parseNode(tt.html)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		if got, want := getAttr(n, tt.attr), tt.value; got != want {
			t.Errorf("getAttr(%q, %q) returned %v, want %v", tt.html, tt.attr, got, want)
		}

		if got, want := getAttrPtr(n, tt.attr), tt.ptr; !reflect.DeepEqual(got, want) {
			t.Errorf("getAttrPtr(%q, %q) returned %v, want %v", tt.html, tt.attr, got, want)
		}
	}
}

func Test_IsAtom(t *testing.T) {
	tests := []struct {
		html  string
		atoms []atom.Atom
		match bool
	}{
		{"", nil, false},
		{"<img>", []atom.Atom{}, false},
		{"<img>", []atom.Atom{atom.A}, false},

		{"<img>", []atom.Atom{atom.Img}, true},
		{"<img>", []atom.Atom{atom.A, atom.Img}, true},
		{"<img>", []atom.Atom{atom.Img, atom.A}, true},
	}

	for _, tt := range tests {
		n, err := parseNode(tt.html)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		if got, want := isAtom(n, tt.atoms...), tt.match; got != want {
			t.Errorf("isAtom(%q, %v) returned %t, want %t", tt.html, tt.atoms, got, want)
		}
	}
}

func Test_GetTextContent(t *testing.T) {
	tests := []struct {
		html, content string
	}{
		{"", ""},
		{"foo", "foo"},
		{"<a>", ""},
		{"<a>foo</a>", "foo"},
		{"<a><b>foo</b>bar</a>", "foobar"},
		{"<a><b>foo</b><i>bar</i></a>", "foobar"},
		{"<a> <b>foo</b> <i>bar</i> </a>", " foo bar "},
		{"<a><b><i>foo</i></b>bar</a>", "foobar"},
	}

	for _, tt := range tests {
		n, err := parseNode(tt.html)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		if got, want := getTextContent(n), tt.content; got != want {
			t.Errorf("getTextContent(%q) returned %q, want %q", tt.html, got, want)
		}
	}
}

func Test_GetOnlyChild(t *testing.T) {
	tests := []struct {
		html, child string
	}{
		{"", ""},
		{"<img>", ""},
		{"<a>foo</a>", ""},

		{"<a><img></a>", "<img>"},
		{"<a>foo<img>bar</a>", "<img>"},
		{"<a><b><img></b></a>", "<b><img></b>"},

		// too many children
		{"<a><img><img></a>", ""},
	}

	for _, tt := range tests {
		n, err := parseNode(tt.html)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		want, err := parseNode(tt.child)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		got := getOnlyChild(n)
		if got != nil {
			// for the purposes of comparison, adjacent nodes don't matter
			got.Parent = nil
			got.PrevSibling = nil
			got.NextSibling = nil
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("getOnlyChild(%q) returned %#v, want %#v", tt.html, got, want)
		}
	}
}

func Test_GetOnlyChildAtom(t *testing.T) {
	tests := []struct {
		html, atom, child string
	}{
		{"", "", ""},
		{"<img>", "", ""},
		{"<a>foo</a>", "", ""},

		{"<a><img></a>", "img", "<img>"},
		{"<a>foo<img>bar</a>", "img", "<img>"},
		{"<a><b><img></b></a>", "b", "<b><img></b>"},

		// wrong atom
		{"<a><img></a>", "b", ""},
		// too many children
		{"<a><img><img></a>", "img", ""},
		// child too deep
		{"<a><b><img></b></a>", "img", ""},
	}

	for _, tt := range tests {
		n, err := parseNode(tt.html)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		want, err := parseNode(tt.child)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		got := getOnlyChildAtom(n, atom.Lookup([]byte(tt.atom)))
		if got != nil {
			// for the purposes of comparison, adjacent nodes don't matter
			got.Parent = nil
			got.PrevSibling = nil
			got.NextSibling = nil
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("getOnlyChildAtom(%q, %q) returned %#v, want %#v", tt.html, tt.atom, got, want)
		}
	}
}

func Test_GetOnlyChildAtomWithAttr(t *testing.T) {
	tests := []struct {
		html, atom, attr, child string
	}{
		{"", "", "", ""},
		{`<a><img></a>`, "img", "", ""},

		{`<a><img src></a>`, "img", "src", "<img src>"},
		{`<a><img src=""></a>`, "img", "src", `<img src="">`},
		{`<a><img src=""><img></a>`, "img", "src", `<img src="">`},
		{`<a>foo<img src>bar</a>`, "img", "src", "<img src>"},
		{`<a><b class><img></b></a>`, "b", "class", "<b class><img></b>"},

		// wrong atom
		{`<a><img src></a>`, "b", "", ""},
		// wrong attr
		{`<a><img src></a>`, "img", "class", ""},
		// too many children
		{`<a><img src><img src></a>`, "img", "src", ""},
		// child too deep
		{`<a><b><img src></b></a>`, "img", "src", ""},
	}

	for _, tt := range tests {
		n, err := parseNode(tt.html)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		want, err := parseNode(tt.child)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		got := getOnlyChildAtomWithAttr(n, atom.Lookup([]byte(tt.atom)), tt.attr)
		if got != nil {
			// for the purposes of comparison, adjacent nodes don't matter
			got.Parent = nil
			got.PrevSibling = nil
			got.NextSibling = nil
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("getOnlyChildAtomWithAttr(%q, %q, %q) returned %#v, want %#v", tt.html, tt.atom, tt.attr, got, want)
		}
	}
}
