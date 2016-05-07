package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"willnorris.com/go/microformats"
)

func main() {
	resp, err := http.Get(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	urlparsed, _ := url.Parse(os.Args[1])
	data := microformats.Parse(resp.Body, urlparsed)

	json, _ := json.MarshalIndent(data, "", "  ")

	fmt.Println(string(json))
}
