package parser

import (
	"encoding/json"
	"testing"
)

type testCase struct {
	name   string
	raw    string
	parsed string
}

var testCases = []testCase{
	{"simple udt", "a = { b: c }\n1a2 a3b a4b5", `[{"b":"c","quantity":1,"value":2},{"c":true,"value":3},{"c":5,"value":4}]`},
	{"udt-like strings", "x 1234 z 12x 213 34x", `["x",1234,"z","12x",213,"34x"]`},
	{"negative numbers", "-1 -0.1 -.0 .2 -12x", `[-1,-0.1,-0,0.2,"-12x"]`},
	//{"modifiers", "a = { t: a, *: b, !: c }\na* a*2 2a** a*! a!!`yes`!`no`", ``},  // TODO: not working properly
	//{"dates", "2016-01-01", `[{"type": "date", value: "2016-01-01"}]`},  // TODO: future feature
}

func Test_Parse_simple(t *testing.T) {

	for i := range testCases {

		data := Parse(testCases[i].raw)
		result, err := json.Marshal(data)

		if err != nil {
			t.Fatalf(`Couldn't parse output as json: %s`, err)
		}

		if string(result) != testCases[i].parsed {
			t.Fatalf(`Failed test case '%s': Not the expected output: %s vs %s`, testCases[i].name, result, testCases[i].parsed)
		}

	}
}
