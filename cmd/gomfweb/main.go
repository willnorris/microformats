// Copyright (c) The microformats project authors.
// SPDX-License-Identifier: MIT

// The gomfweb command runs a simple web server that demonstrates the use of
// the go microformats library.  It can parse the microformats found at a URL
// or in a provided snippet of HTML.
package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"willnorris.com/go/microformats"
)

var addr = flag.String("addr", ":4001", "Address and port to listen on")

func main() {
	flag.Parse()

	srv := &http.Server{
		Addr:    *addr,
		Handler: http.HandlerFunc(index),

		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if port := os.Getenv("PORT"); port != "" {
		srv.Addr = ":" + port
	}

	fmt.Printf("gomfweb listening on %s\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func index(w http.ResponseWriter, r *http.Request) {
	var parsedURL *url.URL
	var err error

	u := strings.TrimSpace(r.FormValue("url"))
	if u != "" {
		parsedURL, err = url.Parse(u)
		if err != nil {
			http.Error(w, fmt.Sprintf("error parsing url: %v", err), http.StatusBadRequest)
		}
	}

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if r.Method == "GET" && parsedURL != nil {
		resp, err := http.Get(parsedURL.String())
		if err != nil {
			http.Error(w, fmt.Sprintf("error fetching url content: %v", err), http.StatusInternalServerError)
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		mf := microformats.Parse(resp.Body, parsedURL)
		if err := enc.Encode(mf); err != nil {
			http.Error(w, fmt.Sprintf("error marshaling json: %v", err), http.StatusInternalServerError)
		}

		if callback := r.FormValue("callback"); callback != "" {
			_, _ = fmt.Fprintf(w, "%s(%s)", callback, buf.String())
		} else {
			w.Header().Set("Content-Type", "application/mf2+json")
			if _, err := io.Copy(w, buf); err != nil {
				log.Print(err)
			}
		}
		return
	}

	html := r.FormValue("html")
	if html != "" {
		mf := microformats.Parse(strings.NewReader(html), parsedURL)
		if err := enc.Encode(mf); err != nil {
			http.Error(w, fmt.Sprintf("error marshaling json: %v", err), http.StatusInternalServerError)
		}
	}

	data := struct {
		HTML string
		URL  string
		JSON string
	}{
		html,
		u,
		buf.String(),
	}

	if err := tpl.Execute(w, data); err != nil {
		log.Print(err)
	}
}

//go:embed index.html
var indexHTML string
var tpl = template.Must(template.New("").Parse(indexHTML))
