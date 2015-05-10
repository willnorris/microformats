// microformats project microformats.go
package microformats

import (
	"io"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	RootClassNames     = regexp.MustCompile("h-\\S*")
	PropertyClassNames = regexp.MustCompile("(p|u|dt|e)-(\\S*)")
)

type MicroFormat struct {
	Value      string                 `json:"value,omitempty"`
	Type       []string               `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Shape      string                 `json:"shape,omitempty"`
	Coords     string                 `json:"coords,omitempty"`
}

type Parser struct {
	curData *Data
	curItem *MicroFormat
}

type Data struct {
	Items []*MicroFormat
	Rels  map[string][]string
}

func New() *Parser {
	return &Parser{}

}

func (p *Parser) Parse(r io.Reader) *Data {
	doc, _ := html.Parse(r)
	p.curData = &Data{
		Items: make([]*MicroFormat, 0),
		Rels:  make(map[string][]string),
	}
	p.walk(doc)
	return p.curData
}

func (p *Parser) walk(node *html.Node) {
	var curItem *MicroFormat
	var priorItem *MicroFormat
	rootclasses := RootClassNames.FindAllString(GetAttr(node, "class"), -1)
	if len(rootclasses) > 0 {
		curItem = &MicroFormat{}
		curItem.Type = rootclasses
		curItem.Properties = make(map[string]interface{})
		p.curData.Items = append(p.curData.Items, curItem)
		priorItem = p.curItem
		p.curItem = curItem
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		p.walk(c)
	}

	propertyclasses := PropertyClassNames.FindAllStringSubmatch(GetAttr(node, "class"), -1)
	if len(propertyclasses) > 0 {
		for _, prop := range propertyclasses {

			var value string
			switch prop[1] {
			case "p":
				// TODO: value-class-pattern
				if value == "" && isAtom(node, atom.Abbr) {
					value = GetAttr(node, "title")
				}
				if value == "" && isAtom(node, atom.Data, atom.Input) {
					value = GetAttr(node, "value")
				}
				if value == "" && isAtom(node, atom.Img, atom.Area) {
					value = GetAttr(node, "alt")
				}
				if value == "" {
					value = GetTextContent(node)
				}
			case "u":
				if value == "" && isAtom(node, atom.A, atom.Area) {
					value = GetAttr(node, "href")
				}
				if value == "" && isAtom(node, atom.Img, atom.Audio, atom.Video, atom.Source) {
					value = GetAttr(node, "src")
				}
				if value == "" && isAtom(node, atom.Object) {
					value = GetAttr(node, "data")
				}
				// TODO: normalize
				// TODO: value-class-pattern
				if value == "" && isAtom(node, atom.Abbr) {
					value = GetAttr(node, "title")
				}
				if value == "" && isAtom(node, atom.Data, atom.Input) {
					value = GetAttr(node, "value")
				}
				if value == "" {
					value = GetTextContent(node)
				}
			}
			if curItem != nil && priorItem != nil {
				priorItem.Properties[prop[2]] = &MicroFormat{
					Type:       curItem.Type,
					Properties: curItem.Properties,
					Coords:     curItem.Coords,
					Shape:      curItem.Shape,
					Value:      value,
				}
			}
			if value != "" {
				p.curItem.Properties[prop[2]] = value
			}
		}
	}
	if curItem != nil {
		if _, ok := curItem.Properties["name"]; !ok {
			var name string
			if node.DataAtom == atom.Img || node.DataAtom == atom.Area {
				name = GetAttr(node, "alt")
			}
			if name != "" {
				curItem.Properties["name"] = name
			}
		}
		if _, ok := curItem.Properties["url"]; !ok {
			var url string
			if node.DataAtom == atom.A || node.DataAtom == atom.Area {
				url = GetAttr(node, "href")
			}
			if url != "" {
				curItem.Properties["url"] = url
			}
		}

		p.curItem = priorItem
	}
}

func GetAttr(node *html.Node, name string) string {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, name) {
			return attr.Val
		}
	}
	return ""
}

func isAtom(node *html.Node, atoms ...atom.Atom) bool {
	for _, atom := range atoms {
		if atom == node.DataAtom {
			return true
		}
	}
	return false
}

func ParseValueClass(node *html.Node) string {

	return ""
}

func GetTextContent(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}
	buf := make([]string, 0)
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		buf = append(buf, GetTextContent(c))
	}
	return strings.Join(buf, "")
}
