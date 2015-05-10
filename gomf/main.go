package main

import (
	"fmt"
	"os"
	"net/http"
	"encoding/json"

	"github.com/andyleap/microformats"
)

func main() {
	parser := microformats.New()
	resp, err := http.Get(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	data := parser.Parse(resp.Body)
	
	json, _ := json.MarshalIndent(data, "", "  ")

	fmt.Println(string(json))
}

