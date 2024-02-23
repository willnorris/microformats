// Copyright (c) The microformats project authors.
// SPDX-License-Identifier: MIT

// Run the shared test suite from https://github.com/microformats/tests

package microformats_test

import (
	"bytes"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"willnorris.com/go/microformats"
)

// skip the tests which we don't pass yet
var skipTests = []string{
	// all tests pass!
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
	input, err := os.ReadFile(test + ".html")
	if err != nil {
		t.Fatalf("error reading file %q: %v", test+".html", err)
	}

	URL, _ := url.Parse("http://example.com/")
	data := microformats.Parse(bytes.NewReader(input), URL)

	expectedJSON, err := os.ReadFile(test + ".json")
	if err != nil {
		t.Fatalf("error reading file %q: %v", test+".json", err)
	}

	want := make(map[string]any)
	err = json.Unmarshal(expectedJSON, &want)
	if err != nil {
		t.Fatalf("error unmarshaling json in file %q: %v", test+".json", err)
	}

	outputJSON, _ := json.Marshal(data)
	got := make(map[string]any)
	err = json.Unmarshal(outputJSON, &got)
	if err != nil {
		t.Fatalf("error unmarshaling json: %v", err)
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("Parse value differs:\n%v", diff)
	}
}
