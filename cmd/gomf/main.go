// Copyright (c) 2015 Andy Leap, Google
// SPDX-License-Identifier: MIT

// The gomf tool is a command line tool which parses microformats from the
// specified URL.  If selector is provided, the first element that matches the
// selector will be used as the root node for parsing.
//
// Usage: gomf <URL> [optional selector]
//
// For example, to parse all microformats from https://microformats.io inside
// the <main> element, call:
//
//     gomf "https://microformats.io" "main"
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"willnorris.com/go/microformats"
)

func main() {
	u, _ := url.Parse(strings.TrimSpace(os.Args[1]))
	resp, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var data *microformats.Data
	if len(os.Args) > 2 {
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		s := doc.Find(os.Args[2])
		if s.Length() == 0 {
			log.Fatal("selector did not match any elements")
		}
		data = microformats.ParseNode(s.Get(0), u)
	} else {
		data = microformats.Parse(resp.Body, u)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}
