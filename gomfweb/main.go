package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/andyleap/microformats"
)

var indextemplate = template.Must(template.New("index").Parse(index))

var parser = microformats.New()

func main() {
	http.Handle("/", http.HandlerFunc(Index))
	http.ListenAndServe(":4001", nil)

}

func Index(rw http.ResponseWriter, req *http.Request) {
	mf := req.FormValue("MF")
	parsed := parser.Parse(strings.NewReader(mf))
	parsedjson, _ := json.MarshalIndent(parsed, "", "    ")

	data := struct {
		MF     string
		Parsed string
	}{
		mf,
		string(parsedjson),
	}

	indextemplate.Execute(rw, data)
}

var index = `<html>
<head>
</head>
<body>
<form method="POST">
<textarea name="MF" style="width: 100%;" rows="15">{{.MF}}</textarea>
<br/>
<input type="submit" value="Parse"/>
</form><br/>
<code><pre>
{{.Parsed}}
</pre></code>
</body>
</html>`
