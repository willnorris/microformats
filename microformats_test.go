package microformats

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestHCard(t *testing.T) {
	test := `<div class="h-card">
  <a class="p-name u-url" href="http://blog.lizardwrangler.com/">Mitchell Baker</a> 
  (<a class="p-org h-card" href="http://mozilla.org/">Mozilla Foundation</a>)
</div>`

	p := New()

	data := p.Parse(strings.NewReader(test))

	output, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(output))
}

func TestHCardEProp(t *testing.T) {
	test := `<div class="h-entry"><div class="e-content h-card"><p>Hello</p></div></div>`

	p := New()

	data := p.Parse(strings.NewReader(test))

	output, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(output))
}
