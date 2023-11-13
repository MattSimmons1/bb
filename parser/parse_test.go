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
	{"simple no spaces", "a={b:c}\n1a2", `[{"b":"c","quantity":1,"value":2}]`},
	{"comment in definition", "a={//foo\nb:c}\na2", `[{"b":"c","value":2}]`},
	{"quote modifier", "a={\":c}\na2\" a\"foo\"\" a\"", `[{"c":true,"value":2},{"c":true,"value":"foo"},{"c":true}]`},
	{"udt-like strings", "x 1234 z 12x 213 34x", `["x",1234,"z","12x",213,"34x"]`},
	{"negative numbers", "-1 -0.1 -.0 .2 -12x", `[-1,-0.1,-0,0.2,"-12x"]`},
	{"unquoted values", `a = {} a:helloa:`, `[{"value":"helloa:"}]`},
	{"unquoted values disabled", `: = { t: u } a = {} a:hello: a:a`, `[{},{"t":"u"},"hello:",{},{"t":"u"},{}]`},
	//{"decimals_disabled", `. = {} 1.2.`, ``},
	// TODO: how to allow for other types but not this one? what should happen?
	//{"dot modifier", `a = { .: b } 1.1a."a" a.1.1`, `[{"b":"a","quantity":1.1}]`},
	// TODO: not sure if this is what we want to happen - tell people not to do this
	{"dash modifier", `a = { -: b } -1a-2-3`, `[{"b":3,"quantity":-1,"value":-2}]`},

	//{"modifiers", "a = { t: a, *: b, !: c }\na* a*2 2a** a*! a!!`yes`!`no`", ``},  // TODO: not working properly
	//{"dates", "2016-01-01", `[{"type": "date", value: "2016-01-01"}]`},  // TODO: future feature
}

func Test_Parse_simple(t *testing.T) {

	verbose = true

	for i := range testCases {

		data := Parse(testCases[i].raw)
		result, err := json.Marshal(data)

		if err != nil {
			t.Fatalf(`Couldn't parse output as json: %s`, err)
		}

		if string(result) != testCases[i].parsed {
			t.Fatalf(`Failed test case '%s': 
%s 
--> Not the expected output: %s vs %s`, testCases[i].name, testCases[i].raw, result, testCases[i].parsed)
		}

	}
}
