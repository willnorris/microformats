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

// Package ptd implements Post Type Discovery as defined by
// https://www.w3.org/TR/post-type-discovery/
package ptd

import (
	"net/url"
	"strings"

	"willnorris.com/go/microformats"
)

// PostType determines the type of a post identified by the provided
// microformat using the Post Type Algorithm.
//
// See also https://www.w3.org/TR/post-type-discovery/#algorithm
func PostType(item *microformats.Microformat) string {
	if item == nil {
		return ""
	}

	for _, t := range item.Type {
		if t == "h-event" {
			return "event"
		}
	}

	// duplicate rsvp, repost, like, reply detection from response algorithm
	if t := ResponseType(item); t != "" && t != "mention" {
		return t
	}

	if validURL(item.Properties["video"]) {
		return "video"
	}
	if validURL(item.Properties["photo"]) {
		return "photo"
	}

	// compare content and name to determine if post is a note or an article
	var content, name string
	for _, value := range item.Properties["content"] {
		if v, ok := value.(string); ok && v != "" {
			content = v
			break
		}
	}
	if content == "" {
		for _, value := range item.Properties["summary"] {
			if v, ok := value.(string); ok && v != "" {
				content = v
				break
			}
		}
	}
	if content == "" {
		return "note"
	}

	for _, value := range item.Properties["name"] {
		if v, ok := value.(string); ok && v != "" {
			name = v
			break
		}
	}
	if name == "" {
		return "note"
	}

	name = strings.Join(strings.Fields(strings.TrimSpace(name)), " ")
	content = strings.Join(strings.Fields(strings.TrimSpace(content)), " ")

	if !strings.HasPrefix(content, name) {
		return "article"
	}

	return "note"
}

// ResponseType determines the type of a post identified by the
// provided microformat using the Response Type Algorithm.
//
// See also https://www.w3.org/TR/post-type-discovery/#response-algorithm
func ResponseType(item *microformats.Microformat) string {
	if item == nil {
		return ""
	}

	for _, value := range item.Properties["rsvp"] {
		if v, ok := value.(string); ok {
			if v == "yes" || v == "no" || v == "maybe" || v == "interested" {
				return "rsvp"
			}
		}
	}

	if validURL(item.Properties["repost-of"]) {
		return "repost"
	}
	if validURL(item.Properties["like-of"]) {
		return "like"
	}
	if validURL(item.Properties["in-reply-to"]) {
		return "reply"
	}

	return "mention"
}

// Returns true if one of values is a string that is a valid URL.
func validURL(values []interface{}) bool {
	for _, value := range values {
		if s, ok := value.(string); ok && s != "" {
			// url.Parse will happily parse an empty string, but
			// that's probably not really what we want here
			_, err := url.Parse(s)
			return err == nil
		}
	}
	return false
}
