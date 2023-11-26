// Copyright (c) The microformats project authors.
// SPDX-License-Identifier: MIT

// Package rhc implements Representative h-card parsing as defined by
// http://microformats.org/wiki/representative-h-card-parsing
package rhc

import (
	"net/url"

	"willnorris.com/go/microformats"
)

// RepresentativeHcard returns the representative h-card for the given parsed data from srcURL.
func RepresentativeHcard(data *microformats.Data, srcURL string) *microformats.Microformat {
	if len(data.Items) == 0 || srcURL == "" {
		return nil
	}

	var (
		// candidate h-card based on rel=me matching
		relMeCard *microformats.Microformat

		// candidate h-card based on srcURL matching
		urlMatchCard *microformats.Microformat
	)

	hcards := findByType(data.Items, "h-card")
	for _, h := range hcards {
		// If the page contains an h-card with uid and url properties both matching the page URL,
		// the first such h-card is the representative h-card
		if hasURLValue(h.Properties["url"], srcURL) {
			if hasURLValue(h.Properties["uid"], srcURL) {
				return h
			}
			if urlMatchCard == nil {
				urlMatchCard = h
			}
		}

		// If no representative h-card was found, if the page contains an h-card with a
		// url property value which also has a rel=me relation (i.e. matches a URL in
		// parse_results.rels.me), the first such h-card is the representative h-card
		if relMeCard == nil {
			for _, u := range h.Properties["url"] {
				if s, ok := u.(string); ok {
					for _, r := range data.Rels["me"] {
						if urlMatch(s, r) {
							relMeCard = h
						}
					}
				}
			}
		}
	}

	if relMeCard != nil {
		return relMeCard
	}

	// If no representative h-card was found, if the page contains one single h-card,
	// and the h-card has a url property matching the page URL,
	// that h-card is the representative h-card
	if len(hcards) == 1 && urlMatchCard != nil {
		return urlMatchCard
	}

	return nil
}

func findByType(in []*microformats.Microformat, typ string) (out []*microformats.Microformat) {
	for _, mf := range in {
		// if mf is the type we're looking for, add it
		for _, t := range mf.Type {
			if t == typ {
				out = append(out, mf)
				continue
			}
		}

		// check each property value that is a microformat
		for _, props := range mf.Properties {
			for _, p := range props {
				if pm, ok := p.(*microformats.Microformat); ok {
					out = append(out, findByType([]*microformats.Microformat{pm}, typ)...)
				}
			}
		}

		// check all children
		out = append(out, findByType(mf.Children, typ)...)
	}
	return out
}

func hasURLValue(values []any, s string) bool {
	for _, v := range values {
		if vs, ok := v.(string); ok {
			if urlMatch(vs, s) {
				return true
			}
		}
	}
	return false
}

func urlMatch(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}

	au, err := url.Parse(a)
	if err != nil {
		return false
	}
	if au.Path == "" {
		au.Path = "/"
	}

	bu, err := url.Parse(b)
	if err != nil {
		return false
	}
	if bu.Path == "" {
		bu.Path = "/"
	}

	return au.String() == bu.String()
}
