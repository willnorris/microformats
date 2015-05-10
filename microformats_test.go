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

	output, _ := json.Marshal(data)
	fmt.Println(string(output))
}
