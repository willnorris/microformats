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
	rootClassNames     = regexp.MustCompile(`^h-\S*$`)
	propertyClassNames = regexp.MustCompile(`^(p|u|dt|e)-(\S*)$`)
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
		if getAttr(node, "href") != "" {
			newbase, _ := url.Parse(getAttr(node, "href"))
			newbase = p.base.ResolveReference(newbase)
			p.base = newbase
			p.baseFound = true
		}
	}

	if isAtom(node, atom.A, atom.Link) {
		if rel := getAttr(node, "rel"); rel != "" {
			urlVal := getAttr(node, "href")

			if p.base != nil {
				urlParsed, _ := url.Parse(urlVal)
				urlParsed = p.base.ResolveReference(urlParsed)
				urlVal = urlParsed.String()
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
			name := p.getImpliedName(node)
			if name != "" {
				curItem.Properties["name"] = append(curItem.Properties["name"], name)
			}
		}
		if _, ok := curItem.Properties["photo"]; !ok {
			photo := p.getImpliedPhoto(node)
			if photo != "" {
				curItem.Properties["photo"] = append(curItem.Properties["photo"], photo)
			}
		}
		if _, ok := curItem.Properties["url"]; !ok {
			url := p.getImpliedURL(node)
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

			var value *string
			var htmlbody string
			switch prop[1] {
			case "p":
				value = p.getValueClassPattern(node)
				if value == nil && isAtom(node, atom.Abbr) {
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
					*value = getTextContent(node)
				}
			case "u":
				if value == nil && isAtom(node, atom.A, atom.Area) {
					value = getAttrPtr(node, "href")
				}
				if value == nil && isAtom(node, atom.Img, atom.Audio, atom.Video, atom.Source) {
					value = getAttrPtr(node, "src")
				}
				if value == nil && isAtom(node, atom.Object) {
					value = getAttrPtr(node, "data")
				}
				if p.base != nil && value != nil {
					urlParsed, _ := url.Parse(*value)
					urlParsed = p.base.ResolveReference(urlParsed)
					*value = urlParsed.String()
				}
				// TODO: value-class-pattern
				if value == nil && isAtom(node, atom.Abbr) {
					value = getAttrPtr(node, "title")
				}
				if value == nil && isAtom(node, atom.Data, atom.Input) {
					value = getAttrPtr(node, "value")
				}
				if value == nil {
					value = new(string)
					*value = getTextContent(node)
				}
			case "e":
				value = new(string)
				*value = getTextContent(node)
				buf := &bytes.Buffer{}
				for c := node.FirstChild; c != nil; c = c.NextSibling {
					html.Render(buf, c)
				}
				htmlbody = buf.String()
			case "dt":
				if value == nil {
					value = p.getValueClassPattern(node)
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
			}
			if curItem != nil && p.curItem != nil {
				p.curItem.Properties[prop[2]] = append(p.curItem.Properties[prop[2]], &Microformat{
					Type:       curItem.Type,
					Properties: curItem.Properties,
					Coords:     curItem.Coords,
					Shape:      curItem.Shape,
					Value:      *value,
					HTML:       htmlbody,
				})
			} else if value != nil && *value != "" && p.curItem != nil {
				if htmlbody != "" {
					p.curItem.Properties[prop[2]] = append(p.curItem.Properties[prop[2]], map[string]interface{}{"value": *value, "html": htmlbody})
				} else {
					p.curItem.Properties[prop[2]] = append(p.curItem.Properties[prop[2]], *value)
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
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, "class") {
			return strings.Split(attr.Val, " ")
		}
	}
	return []string{}
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
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, name) {
			return attr.Val
		}
	}
	return ""
}

func getAttrPtr(node *html.Node, name string) *string {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, name) {
			return &attr.Val
		}
	}
	return nil
}

func isAtom(node *html.Node, atoms ...atom.Atom) bool {
	for _, atom := range atoms {
		if atom == node.DataAtom {
			return true
		}
	}
	return false
}

func parseValueClass(node *html.Node) string {
	return ""
}
