package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"willnorris.com/go/microformats"
)

var indextemplate = template.Must(template.New("index").Parse(index))

func main() {
	http.Handle("/parse", http.HandlerFunc(Parse))
	http.Handle("/", http.HandlerFunc(Index))
	http.ListenAndServe(":4001", nil)
}

func Index(rw http.ResponseWriter, req *http.Request) {
	mf := req.FormValue("html")
	URL := req.FormValue("url")
	urlparsed, _ := url.Parse(URL)
	parsed := microformats.Parse(strings.NewReader(mf), urlparsed)
	parsedjson, _ := json.MarshalIndent(parsed, "", "    ")

	data := struct {
		MF     string
		URL    string
		Parsed string
	}{
		mf,
		URL,
		string(parsedjson),
	}

	indextemplate.Execute(rw, data)
}

func Parse(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		data := struct {
			MF     string
			URL    string
			Parsed string
		}{
			"",
			"",
			"",
		}
		indextemplate.Execute(rw, data)
		return
	}
	mf := req.FormValue("html")
	URL := req.FormValue("url")
	urlparsed, _ := url.Parse(URL)
	parsed := microformats.Parse(strings.NewReader(mf), urlparsed)
	parsedjson, _ := json.MarshalIndent(parsed, "", "    ")

	rw.Write(parsedjson)
}

var index = `<html>
<head>
</head>
<body>
<form method="POST">
<textarea name="html" style="width: 100%;" rows="15">{{.MF}}</textarea>
<br>
<input name="url" type="text" style="width: 100%;" value="{{.URL}}"></input>
<br>
<input type="submit" value="Parse"/>
</form><br>
<code><pre>
{{.Parsed}}
</pre></code>
</body>
</html>`
