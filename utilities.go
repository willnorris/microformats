package microformats

import (
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var whiteSpaceRegex = regexp.MustCompile("(\t|\n|\r|[ ]|&nbsp;)+")

var blockLevelTags = []atom.Atom{atom.H1, atom.H2, atom.H3, atom.H4, atom.H5,
	atom.H6, atom.P, atom.Hr, atom.Pre, atom.Table, atom.Address, atom.Article,
	atom.Aside, atom.Blockquote, atom.Caption, atom.Col, atom.Colgroup, atom.Dd,
	atom.Div, atom.Dt, atom.Dir, atom.Fieldset, atom.Figcaption, atom.Figure,
	atom.Footer, atom.Form, atom.Header, atom.Hgroup, atom.Li, atom.Map,
	atom.Menu, atom.Nav, atom.Optgroup, atom.Option, atom.Section, atom.Tbody,
	atom.Textarea, atom.Tfoot, atom.Th, atom.Thead, atom.Tr, atom.Td,
	atom.Ul, atom.Ol, atom.Dl, atom.Details}

func getTextContent(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}
	buf := make([]string, 0)
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		buf = append(buf, getTextContent(c))
		if isAtom(c, blockLevelTags...) {
			buf = append(buf, " ")
		}
	}
	return strings.TrimSpace(whiteSpaceRegex.ReplaceAllString(strings.Join(buf, ""), " "))
}

func getOnlyChild(node *html.Node) *html.Node {
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

func getOnlyChildAtom(node *html.Node, atom atom.Atom) *html.Node {
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

func getOnlyChildAtomWithAttr(node *html.Node, atom atom.Atom, attr string) *html.Node {
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

func (p *Parser) getImpliedName(node *html.Node) string {
	var name *string
	if isAtom(node, atom.Img, atom.Area) {
		name = getAttrPtr(node, "alt")
	}
	if name == nil && isAtom(node, atom.Abbr) {
		name = getAttrPtr(node, "title")
	}
	if name == nil {
		subnode := getOnlyChildAtomWithAttr(node, atom.Img, "alt")
		if subnode != nil && !RootClassNames.MatchString(GetAttr(subnode, "class")) {
			name = getAttrPtr(subnode, "alt")
		}
	}
	if name == nil {
		subnode := getOnlyChildAtomWithAttr(node, atom.Area, "alt")
		if subnode != nil && !RootClassNames.MatchString(GetAttr(subnode, "class")) {
			name = getAttrPtr(subnode, "alt")
		}
	}
	if name == nil {
		subnode := getOnlyChildAtomWithAttr(node, atom.Abbr, "title")
		if subnode != nil {
			name = getAttrPtr(subnode, "title")
		}
	}
	if name == nil {
		subnode := getOnlyChild(node)
		if subnode != nil {
			subsubnode := getOnlyChildAtomWithAttr(subnode, atom.Img, "alt")
			if subsubnode != nil && !RootClassNames.MatchString(GetAttr(subsubnode, "class")) {
				name = getAttrPtr(subsubnode, "alt")
			}
		}
	}
	if name == nil {
		subnode := getOnlyChild(node)
		if subnode != nil {
			subsubnode := getOnlyChildAtomWithAttr(subnode, atom.Area, "alt")
			if subsubnode != nil && !RootClassNames.MatchString(GetAttr(subsubnode, "class")) {
				name = getAttrPtr(subsubnode, "alt")
			}
		}
	}
	if name == nil {
		subnode := getOnlyChild(node)
		if subnode != nil {
			subsubnode := getOnlyChildAtomWithAttr(subnode, atom.Abbr, "title")
			if subsubnode != nil {
				name = getAttrPtr(subsubnode, "title")
			}
		}
	}
	if name == nil {
		name = new(string)
		*name = getTextContent(node)
	}
	return strings.TrimSpace(whiteSpaceRegex.ReplaceAllString(*name, " "))
}

func (p *Parser) getImpliedPhoto(node *html.Node) string {
	var photo *string
	if photo == nil && isAtom(node, atom.Img) {
		photo = getAttrPtr(node, "src")
	}
	if photo == nil && isAtom(node, atom.Object) {
		photo = getAttrPtr(node, "data")
	}
	if photo == nil {
		subnode := getOnlyChildAtomWithAttr(node, atom.Img, "src")
		if subnode != nil && !RootClassNames.MatchString(GetAttr(subnode, "class")) {
			photo = getAttrPtr(subnode, "src")
		}
	}
	if photo == nil {
		subnode := getOnlyChildAtomWithAttr(node, atom.Object, "data")
		if subnode != nil && !RootClassNames.MatchString(GetAttr(subnode, "class")) {
			photo = getAttrPtr(subnode, "data")
		}
	}
	if photo == nil {
		subnode := getOnlyChild(node)
		if subnode != nil {
			subsubnode := getOnlyChildAtomWithAttr(subnode, atom.Img, "src")
			if subsubnode != nil && !RootClassNames.MatchString(GetAttr(subsubnode, "class")) {
				photo = getAttrPtr(subsubnode, "src")
			}
		}
	}
	if photo == nil {
		subnode := getOnlyChild(node)
		if subnode != nil {
			subsubnode := getOnlyChildAtomWithAttr(subnode, atom.Object, "data")
			if subsubnode != nil && !RootClassNames.MatchString(GetAttr(subsubnode, "class")) {
				photo = getAttrPtr(subsubnode, "data")
			}
		}
	}
	if photo == nil {
		return ""
	}
	return *photo
}

func (p *Parser) getImpliedURL(node *html.Node) string {
	var url *string
	if url == nil && isAtom(node, atom.A, atom.Area) {
		url = getAttrPtr(node, "href")
	}
	if url == nil {
		subnode := getOnlyChildAtomWithAttr(node, atom.A, "href")
		if subnode != nil && !RootClassNames.MatchString(GetAttr(subnode, "class")) {
			url = getAttrPtr(node, "href")
		}
	}
	if url == nil {
		subnode := getOnlyChildAtomWithAttr(node, atom.Area, "href")
		if subnode != nil && !RootClassNames.MatchString(GetAttr(subnode, "class")) {
			url = getAttrPtr(node, "href")
		}
	}
	if url == nil {
		return ""
	}
	//TODO: normalize
	return *url
}

func (p *Parser) getValueClassPattern(node *html.Node) string {
	return ""
}
