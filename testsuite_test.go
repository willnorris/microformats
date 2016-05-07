// Run the shared test suite from https://github.com/microformats/tests

package microformats_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"reflect"
	"testing"

	"willnorris.com/go/microformats"
)

// list of tests built using:
//     find testdata/tests/microformats-v2 -name "*.html" | grep -v change-log | sed -E 's/^testdata\/tests\/(.+)\.html$/\1/'
var tests = []string{
	// "microformats-v2/h-adr/geo",
	"microformats-v2/h-adr/geourl",
	"microformats-v2/h-adr/justaname",
	// "microformats-v2/h-adr/simpleproperties",
	// "microformats-v2/h-as-note/note",
	"microformats-v2/h-card/baseurl",
	"microformats-v2/h-card/childimplied",
	"microformats-v2/h-card/extendeddescription",
	"microformats-v2/h-card/hcard",
	"microformats-v2/h-card/horghcard",
	"microformats-v2/h-card/hyperlinkedphoto",
	// "microformats-v2/h-card/impliedname",
	// "microformats-v2/h-card/impliedphoto",
	// "microformats-v2/h-card/impliedurl",
	"microformats-v2/h-card/justahyperlink",
	"microformats-v2/h-card/justaname",
	// "microformats-v2/h-card/nested",
	// "microformats-v2/h-card/p-property",
	"microformats-v2/h-card/relativeurls",
	// "microformats-v2/h-entry/impliedvalue-nested",
	"microformats-v2/h-entry/justahyperlink",
	"microformats-v2/h-entry/justaname",
	// "microformats-v2/h-entry/summarycontent",
	// "microformats-v2/h-entry/u-property",
	// "microformats-v2/h-entry/urlincontent",
	// "microformats-v2/h-event/ampm",
	"microformats-v2/h-event/attendees",
	// "microformats-v2/h-event/combining",
	// "microformats-v2/h-event/concatenate",
	// "microformats-v2/h-event/dates",
	// "microformats-v2/h-event/dt-property",
	"microformats-v2/h-event/justahyperlink",
	"microformats-v2/h-event/justaname",
	// "microformats-v2/h-event/time",
	// "microformats-v2/h-feed/implied-title",
	// "microformats-v2/h-feed/simple",
	// "microformats-v2/h-geo/abbrpattern",
	"microformats-v2/h-geo/altitude",
	// "microformats-v2/h-geo/hidden",
	"microformats-v2/h-geo/justaname",
	"microformats-v2/h-geo/simpleproperties",
	// "microformats-v2/h-geo/valuetitleclass",
	// "microformats-v2/h-news/all",
	// "microformats-v2/h-news/minimum",
	"microformats-v2/h-org/hyperlink",
	"microformats-v2/h-org/simple",
	// "microformats-v2/h-org/simpleproperties",
	// "microformats-v2/h-product/aggregate",
	"microformats-v2/h-product/justahyperlink",
	"microformats-v2/h-product/justaname",
	"microformats-v2/h-product/simpleproperties",
	// "microformats-v2/h-recipe/all",
	"microformats-v2/h-recipe/minimum",
	// "microformats-v2/h-resume/affiliation",
	// "microformats-v2/h-resume/contact",
	// "microformats-v2/h-resume/education",
	"microformats-v2/h-resume/justaname",
	"microformats-v2/h-resume/skill",
	// "microformats-v2/h-resume/work",
	"microformats-v2/h-review/hyperlink",
	// "microformats-v2/h-review/implieditem",
	// "microformats-v2/h-review/item",
	"microformats-v2/h-review/justaname",
	"microformats-v2/h-review/photo",
	// "microformats-v2/h-review/vcard",
	// "microformats-v2/h-review-aggregate/hevent",
	// "microformats-v2/h-review-aggregate/justahyperlink",
	// "microformats-v2/h-review-aggregate/simpleproperties",
	// "microformats-v2/rel/duplicate-rels",
	"microformats-v2/rel/license",
	"microformats-v2/rel/nofollow",
	"microformats-v2/rel/rel-urls",
	// "microformats-v2/rel/varying-text-duplicate-rels",
	"microformats-v2/rel/xfn-all",
	"microformats-v2/rel/xfn-elsewhere",
}

func TestSuite(t *testing.T) {
	passes := 0
	count := 0
	for _, test := range tests {
		count++
		if runTest(t, filepath.Join("testdata", "tests", test)) {
			passes++
		}
	}
	fmt.Printf("PASSING %d OF %d\n", passes, count)
}

func runTest(t *testing.T, test string) bool {
	input, err := ioutil.ReadFile(test + ".html")
	if err != nil {
		t.Fatalf("error reading file %q: %v", test+".html", err)
	}

	URL, _ := url.Parse("http://example.com/")
	data := microformats.Parse(bytes.NewReader(input), URL)

	expectedJSON, err := ioutil.ReadFile(test + ".json")
	if err != nil {
		t.Fatalf("error reading file %q: %v", test+".json", err)
	}
	want := make(map[string]interface{})
	json.Unmarshal(expectedJSON, &want)

	outputJSON, _ := json.Marshal(data)
	got := make(map[string]interface{})
	json.Unmarshal(outputJSON, &got)

	if reflect.DeepEqual(got, want) {
		fmt.Printf("PASS: %s\n", test)
		return true
	}

	fmt.Printf("FAIL: %s\ngot: %v\n\nwant: %v\n\n", test, got, want)
	t.Fail()
	return false
}
