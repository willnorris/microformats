// The gomf tool is a command line tool which parses microformats from the
// specified URL.
//
// Usage: gomf <URL> [optional selector]
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/PuerkitoBio/goquery"
	"willnorris.com/go/microformats"
)

func main() {
	u, _ := url.Parse(os.Args[1])
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
		data = microformats.ParseNode(doc.Find(os.Args[2]).Get(0), u)
	} else {
		data = microformats.Parse(resp.Body, u)
	}

	json, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(json))
}
