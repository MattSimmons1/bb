package parser

import (
	"encoding/json"
	"testing"
)

func Test_Parse_simpleUDT(t *testing.T) {

	data := Parse("a = { b: c }\n1a2 a3b a4b5")
	expectation := `[{"b":"c","quantity":1,"value":2},{"c":true,"value":3},{"c":5,"value":4}]`

	result, err := json.Marshal(data)

	if err != nil {
		t.Fatalf(`Couldn't parse output as json: %s`, err)
	}

	if string(result) != expectation {
		t.Fatalf(`Not the expected output: %s vs %s`, result, expectation)
	}

}
