// Copyright (c) 2015 Andy Leap, Google
// SPDX-License-Identifier: MIT

package microformats

import (
	"bytes"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func renderNode(n *html.Node) string {
	b := new(bytes.Buffer)
	html.Render(b, n)
	return b.String()
}

func ptr(s string) *string { return &s }

func Test_ExpandAttrURLs(t *testing.T) {
	base, _ := url.Parse("/a/")
	p := &parser{base: base}

	tests := []struct {
		html string
		want string
	}{
		{`<form action="b"></form>`, `<form action="/a/b"></form>`},
		{`<blockquote cite="b"></blockquote>`, `<blockquote cite="/a/b"></blockquote>`},
		{`<del cite="b"></del>`, `<del cite="/a/b"></del>`},
		{`<ins cite="b"></ins>`, `<ins cite="/a/b"></ins>`},
		{`<q cite="b"></q>`, `<q cite="/a/b"></q>`},
		{`<object data="b"></object>`, `<object data="/a/b"></object>`},
		{`<button formaction="b"></button>`, `<button formaction="/a/b"></button>`},
		{`<input formaction="b"/>`, `<input formaction="/a/b"/>`},
		{`<a href="b"></a>`, `<a href="/a/b"></a>`},
		{`<area href="b"/>`, `<area href="/a/b"/>`},
		{`<base href="b"/>`, `<base href="/a/b"/>`},
		{`<link href="b"/>`, `<link href="/a/b"/>`},
		{`<a ping="b"></a>`, `<a ping="/a/b"></a>`},
		{`<area ping="b"/>`, `<area ping="/a/b"/>`},
		{`<audio src="b"></audio>`, `<audio src="/a/b"></audio>`},
		{`<embed src="b"/>`, `<embed src="/a/b"/>`},
		{`<iframe src="b"></iframe>`, `<iframe src="/a/b"></iframe>`},
		{`<img src="b"/>`, `<img src="/a/b"/>`},
		{`<input src="b"/>`, `<input src="/a/b"/>`},
		{`<script src="b"></script>`, `<script src="/a/b"></script>`},
		{`<source src="b"/>`, `<source src="/a/b"/>`},
		{`<track src="b"/>`, `<track src="/a/b"/>`},
		{`<video src="b"></video>`, `<video src="/a/b"></video>`},
		{`<video poster="b"></video>`, `<video poster="/a/b"></video>`},

		// multiple attributes
		{`<input formaction="b" src="c"/>`, `<input formaction="/a/b" src="/a/c"/>`},
	}

	for _, tt := range tests {
		n, _ := parseNode(tt.html)
		p.expandAttrURLs(n)
		got := renderNode(n)
		if got != tt.want {
			t.Errorf("expandAttrURL(%q) returned %q, want %q", tt.html, got, tt.want)
		}
	}
}

func Test_ExpandURL(t *testing.T) {
	example, _ := url.Parse("http://example.com/base/")
	tests := []struct {
		relative string
		base     *url.URL
		want     string
	}{
		{"", nil, ""},
		{"", example, "http://example.com/base/"},
		{"/", nil, "/"},
		{"/", example, "http://example.com/"},
		{"foo", example, "http://example.com/base/foo"},
		{"/foo", example, "http://example.com/foo"},
	}

	for _, tt := range tests {
		got := expandURL(tt.relative, tt.base)
		if want := tt.want; got != want {
			t.Errorf("expandURL(%q, %q) returned %q, want %q", tt.relative, tt.base, got, want)
		}
	}
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
		{`<img class=" a
		b	c ">`, []string{"a", "b", "c"}},
		{`<img class="http://example.com/ b">`, []string{"http://example.com/", "b"}},

		{`<img CLASS="a">`, []string{"a"}},
	}

	for _, tt := range tests {
		n, err := parseNode(tt.html)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		if got, want := getClasses(n), tt.classes; !cmp.Equal(got, want) {
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
		{`<img class="">`, `.*`, false},

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

		if got, want := getAttrPtr(n, tt.attr), tt.ptr; !cmp.Equal(got, want) {
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
	base, _ := url.Parse("http://example.com/")
	p := &parser{base: base}
	tests := []struct {
		html    string
		imgFn   func(*html.Node) string
		content string
	}{
		{"", nil, ""},
		{"foo", nil, "foo"},
		{"<a>", nil, ""},
		{"<a>foo</a>", nil, "foo"},
		{"<a><b>foo</b>bar</a>", nil, "foobar"},
		{"<a><b>foo</b><i>bar</i></a>", nil, "foobar"},
		{"<a> <b>foo</b> <i>bar</i> </a>", nil, " foo bar "},
		{"<a><b><i>foo</i></b>bar</a>", nil, "foobar"},

		// test image functions
		{"<a><img alt='foo'></a>", nil, ""},
		{"<a><img alt='foo'></a>", imageAltValue, "foo"},
		{"<a><img src='foo'></a>", imageAltValue, ""},
		{"<a><img src='foo'></a>", p.imageAltSrcValue, " http://example.com/foo "},
		{"<a><img alt='foo' src='bar'></a>", p.imageAltSrcValue, "foo"},
		{"<a><img></a>", p.imageAltSrcValue, ""},
	}

	for _, tt := range tests {
		n, err := parseNode(tt.html)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		if got, want := getTextContent(n, tt.imgFn), tt.content; got != want {
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

func Test_GetImpliedName(t *testing.T) {
	tests := []struct {
		html, name string
	}{
		{"", ""},

		{`<img alt="name">`, "name"},
		{`<area alt="name">`, "name"},
		{`<abbr title="name">`, "name"},

		{`<span><img alt="name"></span>`, "name"},
		{`<span><img alt="name" class="h-card"></span>`, "name"},
		{`<span><area alt="name"></span>`, "name"},
		{`<span><area alt="name" class="h-card"></span>`, ""},
		{`<span><abbr title="name"></span>`, "name"},

		{`<p><span><img alt="name"></span></p>`, "name"},
		{`<p><span><img alt="name" class="h-card"></span></p>`, "name"},
		{`<p><span><area alt="name"></span></p>`, "name"},
		{`<p><span><area alt="name" class="h-card"></span></p>`, ""},
		{`<p><span><abbr title="name"></span></p>`, "name"},

		{`<p><span>name</span></p>`, "name"},
	}

	for _, tt := range tests {
		n, err := parseNode(tt.html)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		if got, want := getImpliedName(n), tt.name; got != want {
			t.Errorf("getImpliedName(%q) returned %v, want %v", tt.html, got, want)
		}
	}
}

func Test_GetImpliedPhoto(t *testing.T) {
	base, _ := url.Parse("http://example.com/")

	tests := []struct {
		html string
		base *url.URL
		url  string
		alt  string
	}{
		{"", nil, "", ""},

		{`<img src="p">`, nil, "p", ""},
		{`<img src="p" alt="a">`, nil, "p", "a"},
		{`<img src="p">`, base, "http://example.com/p", ""},
		{`<img src="p" alt="a">`, base, "http://example.com/p", "a"},

		{`<object data="p">`, nil, "p", ""},
		{`<object data="p">`, base, "http://example.com/p", ""},

		{`<p><img src="p"></p>`, nil, "p", ""},
		{`<p><img src="p" alt="a"></p>`, nil, "p", "a"},
		{`<p><img src="p"></p>`, base, "http://example.com/p", ""},
		{`<p><img src="p" alt="a"></p>`, base, "http://example.com/p", "a"},
		{`<p><img src="p"><img src="q"></p>`, nil, "", ""},
		{`<p><img src="p" class="h-entry"></p>`, nil, "", ""},

		{`<p><object data="p"></p>`, nil, "p", ""},
		{`<p><object data="p"></p>`, base, "http://example.com/p", ""},
		{`<p><object data="p"></object><object data="p"></object></p>`, nil, "", ""},
		{`<p><object data="p" class="h-entry"></p>`, nil, "", ""},

		{`<p><span><img src="p"></span></p>`, nil, "p", ""},
		{`<p><span><img src="p" alt="a"></span></p>`, nil, "p", "a"},
		{`<p><span><object data="p"></span></p>`, nil, "p", ""},
		{`<p><span><object data="p"></span></p>`, base, "http://example.com/p", ""},
		{`<p><span><object data="p" class="h-entry"></span></p>`, nil, "", ""},
	}

	for _, tt := range tests {
		n, err := parseNode(tt.html)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		src, alt := getImpliedPhoto(n, tt.base)
		if got, want := src, tt.url; got != want {
			t.Errorf("getImpliedPhoto(%q, %s) returned src %v, want %v", tt.html, tt.base, got, want)
		}
		if got, want := alt, tt.alt; got != want {
			t.Errorf("getImpliedPhoto(%q, %s) returned alt %v, want %v", tt.html, tt.base, got, want)
		}
	}
}

func Test_GetImpliedURL(t *testing.T) {
	base, _ := url.Parse("http://example.com/")

	tests := []struct {
		html string
		base *url.URL
		url  string
	}{
		{"", nil, ""},

		{`<a href="p">`, nil, "p"},
		{`<a href="p">`, base, "http://example.com/p"},

		{`<area href="p">`, nil, "p"},
		{`<area href="p">`, base, "http://example.com/p"},

		{`<p><a href="p"></p>`, nil, "p"},
		{`<p><a href="p"></p>`, base, "http://example.com/p"},
		{`<p><a href="p" class="h-entry"></p>`, nil, ""},
		{`<p><b><a href="p"></b></p>`, nil, "p"},

		{`<p><area href="p"></p>`, nil, "p"},
		{`<p><area href="p"></p>`, base, "http://example.com/p"},
		{`<p><area href="p" class="h-entry"></p>`, nil, ""},
	}

	for _, tt := range tests {
		n, err := parseNode(tt.html)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		if got, want := getImpliedURL(n, tt.base), tt.url; got != want {
			t.Errorf("getImpliedURL(%q, %s) returned %v, want %v", tt.html, tt.base, got, want)
		}
	}
}

func Test_GetValueClassPattern(t *testing.T) {
	tests := []struct {
		html  string
		value *string
	}{
		{"", nil},

		{`<p><img alt="v"></p>`, nil},
		{`<p><img alt="v" class="value"></p>`, ptr("v")},
		{`<p><area alt="v" class="value"></p>`, ptr("v")},

		{`<p><data value="v"></data></p>`, nil},
		{`<p><data value="v" class="value"></data></p>`, ptr("v")},
		{`<p><data class="value">v</data></p>`, ptr("v")},

		{`<p><abbr title="v"></abbr></p>`, nil},
		{`<p><abbr title="v" class="value"></abbr></p>`, ptr("v")},
		{`<p><abbr class="value">v</abbr></p>`, ptr("v")},

		{`<p><span>v</span></p>`, nil},
		{`<p><span class="value">v</span></p>`, ptr("v")},

		// concatenation
		{`<p><b class="value">a</b><b class="value">b</b></p>`, ptr("ab")},
		{`<p><img class="value" alt="a"><b>b</b><b class="value">c</b></p>`, ptr("ac")},

		// value-title
		{`<p><img alt="v" class="value-title" title="t"></p>`, ptr("t")},
		{`<p><img alt="v" class="value" title="t"><img alt="v" class="value-title" title="t"></p>`, ptr("vt")},
	}

	for _, tt := range tests {
		n, err := parseNode(tt.html)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		if got, want := getValueClassPattern(n), tt.value; !cmp.Equal(got, want) {
			t.Errorf("getValueClassPattern(%q) returned %v, want %v", tt.html, got, want)
		}
	}
}

func Test_GetFirstPropValue(t *testing.T) {
	tests := []struct {
		properties map[string][]interface{}
		prop       string
		value      *string
	}{
		{nil, "", nil},
		{nil, "name", nil},
		{map[string][]interface{}{"name": {"n"}}, "", nil},

		{map[string][]interface{}{"name": {"n"}}, "name", ptr("n")},
		{map[string][]interface{}{"name": {"a", "b"}}, "name", ptr("a")},
		{map[string][]interface{}{"name": {"a", "b"}}, "url", nil},
		{map[string][]interface{}{"name": {1, 2}}, "name", nil},
		{map[string][]interface{}{"name": {"n"}, "url": {"u"}}, "url", ptr("u")},
	}

	for _, tt := range tests {
		mf := &Microformat{Properties: tt.properties}
		if got, want := getFirstPropValue(mf, tt.prop), tt.value; !cmp.Equal(got, want) {
			t.Errorf("getFirstPropValue(%v, %q) returned %v, want %v", tt.properties, tt.prop, got, want)
		}
	}
}
