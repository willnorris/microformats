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

// Package microformats provides a microformats V2 parser.
//
// See also: http://microformats.org/wiki/microformats2
package microformats // import "willnorris.com/go/microformats"

import (
	"bytes"
	"io"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	rootClassNames     = regexp.MustCompile(`^h-[a-z\-]+$`)
	propertyClassNames = regexp.MustCompile(`^(p|u|dt|e)-([a-z\-]+)$`)
)

// Microformat specifies a single microformat object and its properties.  It
// may represent a person, an address, a blog post, etc.
type Microformat struct {
	Value      string                   `json:"value,omitempty"`
	HTML       string                   `json:"html,omitempty"`
	Type       []string                 `json:"type"`
	Properties map[string][]interface{} `json:"properties"`
	Shape      string                   `json:"shape,omitempty"`
	Coords     string                   `json:"coords,omitempty"`
	Children   []*Microformat           `json:"children,omitempty"`
}

// Data specifies all of the microformats and data parsed from a single HTML
// page.
type Data struct {
	Items   []*Microformat      `json:"items"`
	Rels    map[string][]string `json:"rels"`
	RelURLs map[string]*RelURL  `json:"rel-urls"`
}

// RelURL represents the attributes of a URL.  The URL value itself is the map
// key in the RelURLs field of the Data type.
type RelURL struct {
	Rels     []string `json:"rels,omitempty"`
	Text     string   `json:"text,omitempty"`
	Media    string   `json:"media,omitempty"`
	HrefLang string   `json:"hreflang,omitempty"`
	Type     string   `json:"type,omitempty"`
}

type parser struct {
	curData   *Data
	curItem   *Microformat
	base      *url.URL
	baseFound bool
}

// Parse the microformats found in the HTML document read from r.  baseURL is
// the URL this document was retrieved from which is used to resolve any
// relative URLs.
func Parse(r io.Reader, baseURL *url.URL) *Data {
	doc, _ := html.Parse(r)
	return ParseNode(doc, baseURL)
}

// ParseNode parses the microformats found in doc.  baseURL is the URL this
// document was retrieved from which is used to resolve any relative URLs.
func ParseNode(doc *html.Node, baseURL *url.URL) *Data {
	p := new(parser)
	p.curData = &Data{
		Items:   make([]*Microformat, 0),
		Rels:    make(map[string][]string),
		RelURLs: make(map[string]*RelURL),
	}
	p.base = baseURL
	p.baseFound = false
	p.walk(doc)
	return p.curData
}

func (p *parser) replaceHref(node *html.Node) {
	if isAtom(node, atom.A) {
		href := getAttrPtr(node, "href")
		if href != nil {
			if urlParsed, err := url.Parse(*href); err == nil {
				urlParsed = p.base.ResolveReference(urlParsed)
				*href = urlParsed.String()
			}
		}
		return
	}

	if isAtom(node, atom.Img) {
		href := getAttrPtr(node, "src")
		if href != nil {
			if urlParsed, err := url.Parse(*href); err == nil {
				urlParsed = p.base.ResolveReference(urlParsed)
				*href = urlParsed.String()
			}
		}
		return
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		p.replaceHref(c)
	}
}

func (p *parser) walk(node *html.Node) {
	var curItem *Microformat
	var priorItem *Microformat
	var rootclasses []string
	classes := getClasses(node)
	for _, class := range classes {
		if rootClassNames.MatchString(class) {
			rootclasses = append(rootclasses, class)
		}
	}
	if len(rootclasses) > 0 {
		curItem = &Microformat{}
		curItem.Type = rootclasses
		curItem.Properties = make(map[string][]interface{})
		if p.curItem == nil {
			p.curData.Items = append(p.curData.Items, curItem)
		}
		priorItem = p.curItem
		p.curItem = curItem
	}
	if !p.baseFound && isAtom(node, atom.Base) {
		if href := getAttr(node, "href"); href != "" {
			if newbase, err := url.Parse(href); err == nil {
				newbase = p.base.ResolveReference(newbase)
				p.base = newbase
				p.baseFound = true
			}
		}
	}

	if isAtom(node, atom.A, atom.Link) {
		if rel := getAttr(node, "rel"); rel != "" {
			urlVal := getAttr(node, "href")

			if p.base != nil {
				if urlParsed, err := url.Parse(urlVal); err == nil {
					urlParsed = p.base.ResolveReference(urlParsed)
					urlVal = urlParsed.String()
				}
			}

			rels := strings.Split(rel, " ")
			for _, relval := range rels {
				p.curData.Rels[relval] = append(p.curData.Rels[relval], urlVal)
			}
			p.curData.RelURLs[urlVal] = &RelURL{
				Text:     getTextContent(node),
				Rels:     rels,
				Media:    getAttr(node, "media"),
				HrefLang: getAttr(node, "hreflang"),
				Type:     getAttr(node, "type"),
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		p.walk(c)
	}

	if curItem != nil {
		if _, ok := curItem.Properties["name"]; !ok {
			name := getImpliedName(node)
			if name != "" {
				curItem.Properties["name"] = append(curItem.Properties["name"], name)
			}
		}
		if _, ok := curItem.Properties["photo"]; !ok {
			photo := getImpliedPhoto(node, p.base)
			if photo != "" {
				curItem.Properties["photo"] = append(curItem.Properties["photo"], photo)
			}
		}
		if _, ok := curItem.Properties["url"]; !ok {
			url := getImpliedURL(node, p.base)
			if url != "" {
				curItem.Properties["url"] = append(curItem.Properties["url"], url)
			}
		}
		p.curItem = priorItem
	}
	var propertyclasses [][]string
	for _, class := range classes {
		match := propertyClassNames.FindStringSubmatch(class)
		if match != nil {
			propertyclasses = append(propertyclasses, match)
		}
	}
	if len(propertyclasses) > 0 {
		for _, prop := range propertyclasses {
			prefix, name := prop[1], prop[2]

			var value, embedValue *string
			var htmlbody string
			switch prefix {
			case "p":
				value = getValueClassPattern(node)
				if value == nil && isAtom(node, atom.Abbr, atom.Link) {
					value = getAttrPtr(node, "title")
				}
				if value == nil && isAtom(node, atom.Data, atom.Input) {
					value = getAttrPtr(node, "value")
				}
				if value == nil && isAtom(node, atom.Img, atom.Area) {
					value = getAttrPtr(node, "alt")
				}
				if value == nil {
					value = new(string)
					*value = strings.TrimSpace(getTextContent(node))
				}
				if curItem != nil && p.curItem != nil {
					embedValue = getFirstPropValue(curItem, "name")
				}
			case "u":
				if value == nil && isAtom(node, atom.A, atom.Area, atom.Link) {
					value = getAttrPtr(node, "href")
				}
				if value == nil && isAtom(node, atom.Img, atom.Audio, atom.Video, atom.Source) {
					value = getAttrPtr(node, "src")
				}
				if value == nil && isAtom(node, atom.Object) {
					value = getAttrPtr(node, "data")
				}
				if value == nil && isAtom(node, atom.Video) {
					value = getAttrPtr(node, "poster")
				}
				if p.base != nil && value != nil {
					if urlParsed, err := url.Parse(*value); err == nil {
						urlParsed = p.base.ResolveReference(urlParsed)
						*value = urlParsed.String()
					}
				}
				if value == nil {
					value = getValueClassPattern(node)
				}
				if value == nil && isAtom(node, atom.Abbr) {
					value = getAttrPtr(node, "title")
				}
				if value == nil && isAtom(node, atom.Data, atom.Input) {
					value = getAttrPtr(node, "value")
				}
				if value == nil {
					value = new(string)
					*value = strings.TrimSpace(getTextContent(node))
				}
				if curItem != nil && p.curItem != nil {
					embedValue = getFirstPropValue(curItem, "url")
				}
			case "e":
				value = new(string)
				*value = strings.TrimSpace(getTextContent(node))
				var buf bytes.Buffer

				for c := node.FirstChild; c != nil; c = c.NextSibling {
					p.replaceHref(c)
					html.Render(&buf, c)
				}

				htmlbody = strings.Replace(buf.String(), "/>", " />", -1)
			case "dt":
				if value == nil {
					value = getDateTimeValue(node)
				}
				if value == nil && isAtom(node, atom.Time, atom.Ins, atom.Del) {
					value = getAttrPtr(node, "datetime")
				}
				if value == nil && isAtom(node, atom.Abbr) {
					value = getAttrPtr(node, "title")
				}
				if value == nil && isAtom(node, atom.Data, atom.Input) {
					value = getAttrPtr(node, "value")
				}
				if value == nil {
					value = new(string)
					*value = strings.TrimSpace(getTextContent(node))
				}
			}
			if curItem != nil && p.curItem != nil {
				if embedValue == nil {
					embedValue = value
				}
				p.curItem.Properties[name] = append(p.curItem.Properties[name], &Microformat{
					Type:       curItem.Type,
					Properties: curItem.Properties,
					Coords:     curItem.Coords,
					Shape:      curItem.Shape,
					Value:      *embedValue,
					HTML:       htmlbody,
				})
			} else if value != nil && *value != "" && p.curItem != nil {
				if htmlbody != "" {
					p.curItem.Properties[name] = append(p.curItem.Properties[name], map[string]interface{}{"value": *value, "html": htmlbody})
				} else {
					p.curItem.Properties[name] = append(p.curItem.Properties[name], *value)
				}
			}
		}
	} else {
		if curItem != nil && p.curItem != nil {
			p.curItem.Children = append(p.curItem.Children, curItem)
		}
	}
}

func getClasses(node *html.Node) []string {
	if c := getAttrPtr(node, "class"); c != nil {
		return strings.Split(*c, " ")
	}
	return nil
}

func hasMatchingClass(node *html.Node, regex *regexp.Regexp) bool {
	classes := getClasses(node)
	for _, class := range classes {
		if regex.MatchString(class) {
			return true
		}
	}
	return false
}

func getAttr(node *html.Node, name string) string {
	if v := getAttrPtr(node, name); v != nil {
		return *v
	}
	return ""
}

func getAttrPtr(node *html.Node, name string) *string {
	if node == nil {
		return nil
	}
	for i, attr := range node.Attr {
		if strings.EqualFold(attr.Key, name) {
			return &node.Attr[i].Val
		}
	}
	return nil
}

func isAtom(node *html.Node, atoms ...atom.Atom) bool {
	if node == nil {
		return false
	}
	for _, atom := range atoms {
		if atom == node.DataAtom {
			return true
		}
	}
	return false
}

func getTextContent(node *html.Node) string {
	if node == nil {
		return ""
	}
	if isAtom(node, atom.Script, atom.Style) {
		return ""
	}
	if node.Type == html.TextNode {
		return node.Data
	}
	var buf bytes.Buffer
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		buf.WriteString(getTextContent(c))
	}
	return buf.String()
}

// getOnlyChild returns the sole child of node.  Returns nil if node has zero
// or more than one child.
func getOnlyChild(node *html.Node) *html.Node {
	if node == nil {
		return nil
	}
	var n *html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			if n == nil {
				n = c
			} else {
				return nil
			}
		}
	}
	return n
}

// getOnlyChild returns the sole child of node with the specified atom.
// Returns nil if node has zero or more than one child with that atom.
func getOnlyChildAtom(node *html.Node, atom atom.Atom) *html.Node {
	if node == nil {
		return nil
	}
	var n *html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.DataAtom == atom {
			if n == nil {
				n = c
			} else {
				return nil
			}
		}
	}
	return n
}

// getOnlyChild returns the sole child of node with the specified atom and
// attribute.  Returns nil if node has zero or more than one child with that
// atom and attribute.
func getOnlyChildAtomWithAttr(node *html.Node, atom atom.Atom, attr string) *html.Node {
	if node == nil {
		return nil
	}
	var n *html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.DataAtom == atom && getAttrPtr(c, attr) != nil {
			if n == nil {
				n = c
			} else {
				return nil
			}
		}
	}
	return n
}

func getImpliedName(node *html.Node) string {
	var name *string
	if isAtom(node, atom.Img, atom.Area) {
		name = getAttrPtr(node, "alt")
	}
	if name == nil && isAtom(node, atom.Abbr) {
		name = getAttrPtr(node, "title")
	}

	if name == nil {
		subnode := getOnlyChild(node)
		if subnode != nil && subnode.DataAtom == atom.Img && !hasMatchingClass(subnode, rootClassNames) {
			name = getAttrPtr(subnode, "alt")
		}
	}
	if name == nil {
		subnode := getOnlyChild(node)
		if subnode != nil && subnode.DataAtom == atom.Area && !hasMatchingClass(subnode, rootClassNames) {
			name = getAttrPtr(subnode, "alt")
		}
	}
	if name == nil {
		subnode := getOnlyChild(node)
		if subnode != nil && subnode.DataAtom == atom.Abbr && !hasMatchingClass(subnode, rootClassNames) {
			name = getAttrPtr(subnode, "title")
		}
	}

	if name == nil {
		subnode := getOnlyChild(node)
		if subnode != nil && !hasMatchingClass(subnode, rootClassNames) {
			subsubnode := getOnlyChild(subnode)
			if subsubnode != nil && subsubnode.DataAtom == atom.Img && !hasMatchingClass(subsubnode, rootClassNames) {
				name = getAttrPtr(subsubnode, "alt")
			}
		}
	}
	if name == nil {
		subnode := getOnlyChild(node)
		if subnode != nil && !hasMatchingClass(subnode, rootClassNames) {
			subsubnode := getOnlyChild(subnode)
			if subsubnode != nil && subsubnode.DataAtom == atom.Area && !hasMatchingClass(subsubnode, rootClassNames) {
				name = getAttrPtr(subsubnode, "alt")
			}
		}
	}
	if name == nil {
		subnode := getOnlyChild(node)
		if subnode != nil && !hasMatchingClass(subnode, rootClassNames) {
			subsubnode := getOnlyChild(subnode)
			if subsubnode != nil && subsubnode.DataAtom == atom.Abbr {
				name = getAttrPtr(subsubnode, "title")
			}
		}
	}

	if name == nil {
		name = new(string)
		*name = getTextContent(node)
	}
	return strings.TrimSpace(*name)
}

func getImpliedPhoto(node *html.Node, baseURL *url.URL) string {
	var photo *string
	if photo == nil && isAtom(node, atom.Img) {
		photo = getAttrPtr(node, "src")
	}
	if photo == nil && isAtom(node, atom.Object) {
		photo = getAttrPtr(node, "data")
	}

	if photo == nil {
		subnode := getOnlyChildAtomWithAttr(node, atom.Img, "src")
		if subnode != nil && !hasMatchingClass(subnode, rootClassNames) {
			photo = getAttrPtr(subnode, "src")
		}
	}
	if photo == nil {
		subnode := getOnlyChildAtomWithAttr(node, atom.Object, "data")
		if subnode != nil && !hasMatchingClass(subnode, rootClassNames) {
			photo = getAttrPtr(subnode, "data")
		}
	}

	if photo == nil {
		subnode := getOnlyChild(node)
		if subnode != nil && !hasMatchingClass(subnode, rootClassNames) {
			subsubnode := getOnlyChildAtomWithAttr(subnode, atom.Img, "src")
			if subsubnode != nil && !hasMatchingClass(subsubnode, rootClassNames) {
				photo = getAttrPtr(subsubnode, "src")
			}
		}
	}
	if photo == nil {
		subnode := getOnlyChild(node)
		if subnode != nil && !hasMatchingClass(subnode, rootClassNames) {
			subsubnode := getOnlyChildAtomWithAttr(subnode, atom.Object, "data")
			if subsubnode != nil && !hasMatchingClass(subsubnode, rootClassNames) {
				photo = getAttrPtr(subsubnode, "data")
			}
		}
	}

	if photo == nil {
		return ""
	}
	if baseURL != nil {
		if urlParsed, err := url.Parse(*photo); err == nil {
			urlParsed = baseURL.ResolveReference(urlParsed)
			*photo = urlParsed.String()
		}
	}
	return *photo
}

func getImpliedURL(node *html.Node, baseURL *url.URL) string {
	var value *string
	if value == nil && isAtom(node, atom.A, atom.Area) {
		value = getAttrPtr(node, "href")
	}

	if value == nil {
		subnode := getOnlyChildAtomWithAttr(node, atom.A, "href")
		if subnode != nil && !hasMatchingClass(subnode, rootClassNames) {
			value = getAttrPtr(subnode, "href")
		}
	}
	if value == nil {
		subnode := getOnlyChildAtomWithAttr(node, atom.Area, "href")
		if subnode != nil && !hasMatchingClass(subnode, rootClassNames) {
			value = getAttrPtr(subnode, "href")
		}
	}

	if value == nil {
		subnode := getOnlyChild(node)
		if subnode != nil && !hasMatchingClass(subnode, rootClassNames) {
			subsubnode := getOnlyChildAtomWithAttr(subnode, atom.A, "href")
			if subsubnode != nil && !hasMatchingClass(subsubnode, rootClassNames) {
				value = getAttrPtr(subsubnode, "href")
			}
		}
	}
	if value == nil {
		subnode := getOnlyChild(node)
		if subnode != nil && !hasMatchingClass(subnode, rootClassNames) {
			subsubnode := getOnlyChildAtomWithAttr(subnode, atom.Area, "href")
			if subsubnode != nil && !hasMatchingClass(subsubnode, rootClassNames) {
				value = getAttrPtr(subsubnode, "href")
			}
		}
	}

	if value == nil {
		return ""
	}
	if baseURL != nil {
		if urlParsed, err := url.Parse(*value); err == nil {
			urlParsed = baseURL.ResolveReference(urlParsed)
			*value = urlParsed.String()
		}
	}
	return *value
}

func getValueClassPattern(node *html.Node) *string {
	values := parseValueClassPattern(node, false)
	if len(values) > 0 {
		val := strings.Join(values, "")
		return &val
	}
	return nil
}

func parseValueClassPattern(node *html.Node, dt bool) []string {
	if node == nil {
		return nil
	}
	var values []string
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		classes := getClasses(c)
		var valueClass, valueTitleClass bool
		for _, class := range classes {
			if class == "value" {
				valueClass = true
			}
			if class == "value-title" {
				valueTitleClass = true
			}
		}
		if valueTitleClass {
			values = append(values, *getAttrPtr(c, "title"))
		} else if valueClass {
			if isAtom(c, atom.Img, atom.Area) && getAttrPtr(c, "alt") != nil {
				values = append(values, *getAttrPtr(c, "alt"))
			} else if isAtom(c, atom.Data) && getAttrPtr(c, "value") != nil {
				values = append(values, *getAttrPtr(c, "value"))
			} else if isAtom(c, atom.Abbr) && getAttrPtr(c, "title") != nil {
				values = append(values, *getAttrPtr(c, "title"))
			} else if dt && isAtom(c, atom.Del, atom.Ins, atom.Time) && getAttrPtr(c, "datetime") != nil {
				values = append(values, *getAttrPtr(c, "datetime"))
			} else {
				values = append(values, getTextContent(c))
			}
		}
	}

	return values
}

func getFirstPropValue(item *Microformat, prop string) *string {
	values := item.Properties[prop]
	if len(values) > 0 {
		if v, ok := values[0].(string); ok {
			return &v
		}
	}
	return nil
}
