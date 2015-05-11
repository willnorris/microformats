package microformats_test

import (
	"io/ioutil"
	"os"
	"testing"
	"path"
	"bytes"
	"fmt"
	"encoding/json"
	"reflect"

	"github.com/andyleap/microformats"
)

var parser = microformats.New()

func TestSuite(t *testing.T) {
	passes := 0
	count := 0
	testsdir, _ := os.Open("tests")
	suites, _ := testsdir.Readdir(0)
	for _, suite := range suites {
		if suite.IsDir() {
			suitedir, _ := os.Open(path.Join("tests", suite.Name()))
			suitedirs, _ := suitedir.Readdir(0)
			for _, test := range suitedirs {
				if test.IsDir() {
					count = count + 1
					if runTest(path.Join("tests", suite.Name(), test.Name())) {
						fmt.Printf("PASS: %s/%s\n", suite.Name(), test.Name())
						passes = passes + 1
					} else {
						fmt.Printf("FAIL: %s/%s\n", suite.Name(), test.Name())
						t.Fail()
					}
					
				}
			}
		}
	}
	fmt.Printf("PASSING %d OF %d\n", passes, count)
}

func runTest(test string) bool {
	input, _ := ioutil.ReadFile(path.Join(test, "input.html"))
	data := parser.Parse(bytes.NewReader(input))
	
	expectedJson, _ := ioutil.ReadFile(path.Join(test, "output.json"))
	expected := make(map[string]interface{})
	json.Unmarshal(expectedJson, &expected)
	
	outputJson, _ := json.Marshal(data)
	output := make(map[string]interface{})
	json.Unmarshal(outputJson, &output)
	
	
	return reflect.DeepEqual(output, expected)
}
