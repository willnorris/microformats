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

// This file includes backwards compatibility support for microformats v1.

package microformats

import (
	"net/url"
	"path"
	"strings"
)

var (
	backcompatRootMap = map[string]string{
		"adr":               "h-adr",
		"geo":               "h-geo",
		"hentry":            "h-entry",
		"hfeed":             "h-feed",
		"hnews":             "h-news",
		"hproduct":          "h-product",
		"hrecipe":           "h-recipe",
		"hresume":           "h-resume",
		"hreview":           "h-review",
		"hreview-aggregate": "h-review-aggregate",
		"vcard":             "h-card",
		"vevent":            "h-event",
	}

	backcompatPropertyMap = map[string]string{
		"author":            "p-author",
		"description":       "p-description",
		"job-title":         "p-job-title",
		"organization-name": "p-organization-name",
		"organization-unit": "p-organization-unit",
		"published":         "dt-published",
		"summary":           "p-summary",
		"title":             "p-title",
		"worst":             "p-worst",
	}

	backcompatPropertyOverrideMap = map[string]map[string]string{
		"h-adr": map[string]string{
			"country-name":     "p-country-name",
			"extended-address": "p-extended-address",
			"locality":         "p-locality",
			"post-office-box":  "p-post-office-box",
			"postal-code":      "p-postal-code",
			"region":           "p-region",
			"street-address":   "p-street-address",
		},
		"h-card": map[string]string{
			"additional-name":  "p-additional-name",
			"adr":              "p-adr",
			"agent":            "p-agent",
			"bday":             "dt-bday",
			"category":         "p-category",
			"class":            "p-class",
			"email":            "u-email",
			"family-name":      "p-family-name",
			"fn":               "p-name",
			"geo":              "p-geo",
			"given-name":       "p-given-name",
			"honorific-prefix": "p-honorific-prefix",
			"honorific-suffix": "p-honorific-suffix",
			"key":              "u-key",
			"label":            "p-label",
			"logo":             "u-logo",
			"mailer":           "p-mailer",
			"nickname":         "p-nickname",
			"note":             "p-note",
			"org":              "p-org",
			"photo":            "u-photo",
			"rev":              "dt-rev",
			"role":             "p-role",
			"sort-string":      "p-sort-string",
			"sound":            "u-sound",
			"tel":              "p-tel",
			"title":            "p-job-title",
			"tz":               "dt-tz",
			"uid":              "u-uid",
			"url":              "u-url",
		},
		"h-entry": map[string]string{
			"author":        "p-author",
			"entry-content": "e-content",
			"entry-summary": "p-summary",
			"entry-title":   "p-name",
			"published":     "dt-published",
			"summary":       "p-summary",
			"updated":       "dt-updated",
		},
		"h-event": map[string]string{
			"attendee":    "p-attendee",
			"category":    "p-category",
			"description": "p-description",
			"dtend":       "dt-end",
			"dtstart":     "dt-start",
			"duration":    "dt-duration",
			"location":    "p-location",
			"summary":     "p-name",
			"url":         "u-url",
		},
		"h-feed": map[string]string{
			"author": "p-author",
			"entry":  "p-entry",
			"photo":  "u-photo",
			"url":    "u-url",
		},
		"h-geo": map[string]string{
			"latitude":  "p-latitude",
			"longitude": "p-longitude",
		},
		"h-news": map[string]string{
			"dateline":   "p-dateline",
			"entry":      "p-entry",
			"geo":        "p-geo",
			"source-org": "p-source-org",
		},
		"h-product": map[string]string{
			"brand":       "p-brand",
			"category":    "p-category",
			"description": "p-description",
			"fn":          "p-name",
			"listing":     "p-listing",
			"photo":       "u-photo",
			"price":       "p-price",
			"review":      "p-review",
			"url":         "u-url",
		},
		"h-resume": map[string]string{
			"affiliation":  "p-affiliation",
			"contact":      "p-contact",
			"education":    "p-education",
			"experience":   "p-experience",
			"publications": "p-publications",
			"skill":        "p-skill",
			"summary":      "p-summary",
		},
		"h-review": map[string]string{
			"description": "e-content",
			"dtreviewed":  "dt-reviewed",
			"item":        "p-item",
			"rating":      "p-rating",
			"reviewer":    "p-author",
			"summary":     "p-name",
		},
		"h-review-aggregate": map[string]string{
			"average": "p-average",
			"best":    "p-best",
			"count":   "p-count",
			"item":    "p-item",
			"rating":  "p-rating",
			"summary": "p-name",
			"votes":   "p-votes",
		},
	}

	backcompatRelMap = map[string]map[string]string{
		"h-entry": map[string]string{
			"bookmark": "u-url",
		},
		"h-feed": map[string]string{
			"tag": "u-category",
		},
		"h-news": map[string]string{
			"principles": "u-principles",
		},
		"h-review": map[string]string{
			"bookmark": "u-url",
			"tag":      "u-category",
		},
	}
)

func backcompatRootClasses(classes []string) []string {
	var rootclasses []string
	for _, class := range classes {
		if c, ok := backcompatRootMap[class]; ok {
			rootclasses = append(rootclasses, c)
		}
	}
	return rootclasses
}

func backcompatPropertyClasses(classes []string, rels []string, context []string) []string {
	var classmap = make(map[string]string)
	for _, class := range classes {
		for _, ctx := range context {
			if c, ok := backcompatPropertyOverrideMap[ctx][class]; ok {
				parts := strings.SplitN(c, "-", 2)
				classmap[parts[1]] = c
			}
		}
	}
	for _, rel := range rels {
		for _, ctx := range context {
			if c, ok := backcompatRelMap[ctx][rel]; ok {
				parts := strings.SplitN(c, "-", 2)
				classmap[parts[1]] = c
			}
		}
	}

	var propertyclasses []string
	for _, c := range classmap {
		propertyclasses = append(propertyclasses, c)
	}
	return propertyclasses
}

// strip provided URL to its last path segment to serve as a category value.
func backcompatURLCategory(s string) string {
	if s == "" {
		return s
	}
	if p, err := url.Parse(s); err == nil {
		return path.Base(p.Path)
	}
	return s
}
