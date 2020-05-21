// Copyright (c) 2015 Andy Leap, Google
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

// Run the shared test suite from https://github.com/microformats/tests

package microformats_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kylelemons/godebug/pretty"
	"willnorris.com/go/microformats"
)

// skip the tests which we don't pass yet
var skipTests = []string{
	// most of these breakages are related to https://github.com/microformats/tests/pull/109
	filepath.Join("microformats-v2", "mixed", "id"),
	filepath.Join("microformats-v2", "mixed", "ignoretemplate"),
	filepath.Join("microformats-v2", "mixed", "vendorprefix"),
	filepath.Join("microformats-v2", "mixed", "vendorprefixproperty"),
}

func TestSuite(t *testing.T) {
	for _, version := range []string{"microformats-mixed", "microformats-v1", "microformats-v2"} {
		t.Run(version, func(t *testing.T) {
			base := filepath.Join("testdata", "tests", version)
			tests, err := listTests(base)
			if err != nil {
				t.Fatalf("error reading test cases: %v", err)
			}

			for _, test := range tests {
				t.Run(test, func(t *testing.T) {
					for _, skip := range skipTests {
						if filepath.Join(version, test) == skip {
							t.Skip()
						}
					}

					runTest(t, filepath.Join(base, test))
				})
			}
		})
	}
}

// listTests recursively lists microformat tests in the specified root
// directory.  A test is identified as a pair of matching .html and .json files
// in the same directory.  Returns a slice of named tests, where the test name
// is the path to the html and json files relative to root, excluding any file
// extension.
func listTests(root string) ([]string, error) {
	tests := []string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if ext := filepath.Ext(path); ext == ".json" {
			test := strings.TrimSuffix(path, ext)
			// ensure .html file exists with the same name
			if _, err := os.Stat(test + ".html"); os.IsNotExist(err) {
				return nil
			}
			test, err = filepath.Rel(root, test)
			if err != nil {
				return err
			}
			tests = append(tests, test)
		}
		return nil
	})
	return tests, err
}

func runTest(t *testing.T, test string) {
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
	err = json.Unmarshal(expectedJSON, &want)
	if err != nil {
		t.Fatalf("error unmarshaling json in file %q: %v", test+".json", err)
	}

	outputJSON, _ := json.Marshal(data)
	got := make(map[string]interface{})
	err = json.Unmarshal(outputJSON, &got)
	if err != nil {
		t.Fatalf("error unmarshaling json: %v", err)
	}

	if !cmp.Equal(got, want) {
		t.Fatalf("Parse value differs:\n%v", pretty.Compare(want, got))
	}
}
