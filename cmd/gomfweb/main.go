package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"willnorris.com/go/microformats"
)

var addr = flag.String("addr", ":4001", "Address and port to listen on")

func main() {
	flag.Parse()

	http.Handle("/", http.HandlerFunc(index))
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

		w.Write(j)
		return
	}

	html := r.FormValue("html")
	var j []byte
	if html != "" {
		mf := microformats.Parse(strings.NewReader(html), parsedURL)
		j, err = json.MarshalIndent(mf, "", "    ")
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
</head>
<body>
  <h2>Parse a URL</h2>
  <form method="GET">
    <input name="url" type="url" />
    <input type="submit" value="Parse" />
  </form>

  <h2>Parse HTML</h2>
  <form method="POST">
    <textarea name="html" style="width: 100%;" rows="15">{{ .HTML }}</textarea>
    <br>
    <input name="url" type="text" style="width: 100%;" value="{{ .URL }}"></input>
    <br>
    <input type="submit" value="Parse"/>
  </form><br>

  <pre><code>
{{ .JSON }}
</code></pre>
</body>
</html>`))
