// The gomfweb command runs a simple web server that demonstrates the use of
// the go microformats library.  It can parse the microformats found at a URL
// or in a provided snippet of HTML.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
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

	u := r.FormValue("url")
	if u != "" {
		parsedURL, err = url.Parse(u)
		if err != nil {
			http.Error(w, fmt.Sprintf("error parsing url: %v", err), http.StatusBadRequest)
		}
	}

	if r.Method == "GET" && parsedURL != nil {
		resp, err := http.Get(parsedURL.String())
		if err != nil {
			http.Error(w, fmt.Sprintf("error fetching url content: %v", err), http.StatusInternalServerError)
		}
		defer resp.Body.Close()

		mf := microformats.Parse(resp.Body, parsedURL)
		j, err := json.MarshalIndent(mf, "", "    ")
		if err != nil {
			http.Error(w, fmt.Sprintf("error marshaling json: %v", err), http.StatusInternalServerError)
		}

		if callback := r.FormValue("callback"); callback != "" {
			fmt.Fprintf(w, "%s(%s)", callback, j)
		} else {
			w.Write(j)
		}
		return
	}

	html := r.FormValue("html")
	var j []byte
	if html != "" {
		mf := microformats.Parse(strings.NewReader(html), parsedURL)
		j, err = json.MarshalIndent(mf, "", "  ")
		if err != nil {
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
		string(j),
	}

	tpl.Execute(w, data)
}

var tpl = template.Must(template.New("").Parse(`<!doctype html>
<html>
<head>
<style>
  input, textarea { font-size: 1rem; }
  input[type=url], textarea { width: calc(100% - 1rem); }
  input[type=url], textarea, pre { border: 1px solid #999; border-radius: 2px; padding: 0.5rem; }
  label, input { display: block; }
  input[type=submit] { margin: 0.5em 0; }
  pre { background: #eee; }
</style>
</head>
<body>
  <h1>go microformats parser</h1>
  <h2>Parse a URL</h2>
  <form method="GET">
    <input name="url" type="url" placeholder="https://indieweb.org/" />
    <input type="submit" value="Parse" />
  </form>

  <h2>Parse HTML</h2>
  <form method="POST">
    <label for="html">HTML</label>
    <textarea id="html" name="html" rows="15">{{ .HTML }}</textarea>
    <label for="url">Base URL</label>
    <input id="url" name="url" type="url" value="{{ .URL }}" placeholder="https://indieweb.org/" />
    <input type="submit" value="Parse"/>
  </form><br>

{{ with .JSON }}
  <h2>JSON</h2>
  <pre><code>{{ . }}
</code></pre>
{{ end }}
<ul>
  <li><a href="http://microformats.org/wiki/about">About microformats</a></li>
  <li><a href="https://github.com/willnorris/microformats/tree/master/cmd/gomfweb">Source code for this site</a></li>
  <li><a href="http://microformats.org/wiki/microformats2#Parsers">Other microformats parsers</a></li>
</ul>
</body>
</html>`))
