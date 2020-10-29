// Copyright (c) 2015 Andy Leap, Google
// SPDX-License-Identifier: MIT

// The gomfweb command runs a simple web server that demonstrates the use of
// the go microformats library.  It can parse the microformats found at a URL
// or in a provided snippet of HTML.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"willnorris.com/go/microformats"
)

var addr = flag.String("addr", ":4001", "Address and port to listen on")

func main() {
	flag.Parse()

	http.Handle("/", http.HandlerFunc(index))

	if port := os.Getenv("PORT"); port != "" {
		*addr = ":" + port
	}

	fmt.Printf("gomfweb listening on %s\n", *addr)
	http.ListenAndServe(*addr, nil)
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
		defer resp.Body.Close()

		mf := microformats.Parse(resp.Body, parsedURL)
		if err := enc.Encode(mf); err != nil {
			http.Error(w, fmt.Sprintf("error marshaling json: %v", err), http.StatusInternalServerError)
		}

		if callback := r.FormValue("callback"); callback != "" {
			fmt.Fprintf(w, "%s(%s)", callback, buf.String())
		} else {
			io.Copy(w, buf)
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

	tpl.Execute(w, data)
}

var tpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

  <title>Go Microformats Parser</title>
  <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-alpha.6/css/bootstrap.min.css" integrity="sha384-rwoIResjU2yc3z8GV/NPeZWAv56rSmLldC3R/AZzGRnGxQQKnKkoFVhFQhNUwEyJ" crossorigin="anonymous">
  <style>
    form label { font-weight: bold; }
    form textarea, form input[type=url] { font-family: "SF Mono", Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace; }
    form .form-control:disabled { cursor: default; background: #efefef; color: black; }
  </style>
</head>

<body>
  <main class="container">
    <h1 class="mt-5 mb-3">Microformats Parser (Go)</h1>

    <form method="get">
      <div class="form-group">
        <label for="url">Enter a URL</label>
        <input name="url" type="url" placeholder="https://indieweb.org" class="form-control form-control-lg" />
      </div>

      <button type="submit" class="btn btn-lg btn-success">Parse</button>
    </form>

    <h2 class="h4 my-5">OR parse just a snippet of HTML</h2>

    <form method="post" class="mb-5">
      <div class="form-group">
        <label for="html">HTML</label>
        <textarea id="html" name="html" rows="6" class="form-control form-control-lg">{{ .HTML }}</textarea>
      </div>

      <div class="form-group">
        <label for="base-url">Base URL</label>
        <input id="base-url" name="base-url" type="url" value="{{ .URL }}" placeholder="https://indieweb.org" class="form-control form-control-lg" />
      </div>

      <button type="submit" class="btn btn-lg btn-success">Parse</button>
    </form>

    {{ with .JSON }}
    <div class="form-group mb-5">
      <label for="json">JSON</label>
      <textarea id="json" name="json" rows="10" class="form-control form-control-lg" disabled="disabled">{{ . }}</textarea>
    </div>
    {{ end }}

    <footer class="mb-5">
      <ul>
        <li><a href="https://microformats.io">About Microformats</a></li>
        <li><a href="https://github.com/willnorris/microformats/tree/master/cmd/gomfweb">Source code for this site</a></li>
        <li><a href="https://github.com/willnorris/microformats">Source code for the Microformats Go Parser</a></li>
        <li>
          Other Microformats Parser websites:
          <a href="http://node.microformats.io">Node</a>,
          <a href="https://php.microformats.io">PHP</a>,
          <a href="http://python.microformats.io">Python</a>, and
          <a href="https://ruby.microformats.io">Ruby</a>.
        </li>
	<li><a href="http://microformats.org/wiki/microformats2#Parsers">More Microformats parsers</a></li>
      </ul>
    </footer>
  </main>
</body>
</html>`))
